package postgres

import (
	"context"

	"github.com/ikurniawann/brag2026/backend/internal/domain"
)

type PrizeRepo struct{ db *DB }

func NewPrizeRepo(db *DB) *PrizeRepo { return &PrizeRepo{db: db} }

const prizeSelect = `
	select p.id, p.season_id, p.nama_hadiah, p.deskripsi, p.nilai_estimasi,
	       p.donatur_id, du.full_name, p.alokasi::text, p.kategori_target,
	       p.status::text, p.pemenang_id, wu.full_name
	from prize_pool p
	left join members dm on dm.id = p.donatur_id
	left join app_users du on du.id = dm.user_id
	left join members wm on wm.id = p.pemenang_id
	left join app_users wu on wu.id = wm.user_id
`

func scanPrize(row interface{ Scan(...any) error }) (*domain.Prize, error) {
	var p domain.Prize
	err := row.Scan(&p.ID, &p.SeasonID, &p.NamaHadiah, &p.Deskripsi, &p.NilaiEstimasi,
		&p.DonaturID, &p.DonaturNama, &p.Alokasi, &p.KategoriTarget,
		&p.Status, &p.PemenangID, &p.PemenangNama)

	if noRows(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PrizeRepo) List(ctx context.Context, seasonID, status string) ([]domain.Prize, error) {
	sql := prizeSelect + ` where p.season_id = $1`
	args := []any{seasonID}

	if status != "" {
		args = append(args, status)
		sql += ` and p.status::text = $2`
	}
	sql += ` order by p.created_at desc`

	rows, err := r.db.q(ctx).Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.Prize{}
	for rows.Next() {
		var p domain.Prize
		if err := rows.Scan(&p.ID, &p.SeasonID, &p.NamaHadiah, &p.Deskripsi, &p.NilaiEstimasi,
			&p.DonaturID, &p.DonaturNama, &p.Alokasi, &p.KategoriTarget,
			&p.Status, &p.PemenangID, &p.PemenangNama); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *PrizeRepo) FindByID(ctx context.Context, id string) (*domain.Prize, error) {
	return scanPrize(r.db.q(ctx).QueryRow(ctx, prizeSelect+` where p.id = $1 limit 1`, id))
}

func (r *PrizeRepo) Create(ctx context.Context, p *domain.Prize) (string, error) {
	var id string
	err := r.db.q(ctx).QueryRow(ctx, `
		insert into prize_pool
			(season_id, nama_hadiah, deskripsi, nilai_estimasi, donatur_id,
			 alokasi, kategori_target, status)
		values ($1, $2, $3, $4, $5, $6::prize_alokasi, $7, $8::prize_status)
		returning id
	`, p.SeasonID, p.NamaHadiah, p.Deskripsi, p.NilaiEstimasi, p.DonaturID,
		p.Alokasi, p.KategoriTarget, p.Status).Scan(&id)
	return id, err
}

func (r *PrizeRepo) SetStatus(ctx context.Context, id, status string, pemenangID *string) error {
	tag, err := r.db.q(ctx).Exec(ctx, `
		update prize_pool
		set status = $1::prize_status,
		    pemenang_id = coalesce($2, pemenang_id)
		where id = $3
	`, status, pemenangID, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.NotFound("Hadiah tidak ditemukan.")
	}
	return nil
}

func (r *PrizeRepo) Delete(ctx context.Context, id string) error {
	tag, err := r.db.q(ctx).Exec(ctx, `delete from prize_pool where id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.NotFound("Hadiah tidak ditemukan.")
	}
	return nil
}

// CountApprovedDonations backs the PATRON badge.
func (r *PrizeRepo) CountApprovedDonations(ctx context.Context, memberID string) (int, error) {
	var n int
	err := r.db.q(ctx).QueryRow(ctx, `
		select count(*)::int from prize_pool
		where donatur_id = $1 and status in ('approved', 'awarded')
	`, memberID).Scan(&n)
	return n, err
}

// RebuildTickets recomputes every member's entitlement in three statements
// rather than a query per member and an insert per ticket. generate_series
// expands the counts into rows inside the database, so two hundred tickets
// cost one round trip instead of two hundred.
func (r *PrizeRepo) RebuildTickets(ctx context.Context, seasonID string) ([]domain.TicketCount, error) {
	if _, err := r.db.q(ctx).Exec(ctx,
		`delete from raffle_tickets where season_id = $1`, seasonID); err != nil {
		return nil, err
	}

	if _, err := r.db.q(ctx).Exec(ctx, `
		with entitlement as (
			select m.id as member_id,
			       greatest((select coalesce(sum(sl.points), 0)::int / 100
			                 from score_ledger sl
			                 where sl.member_id = m.id and sl.season_id = $1), 0) as from_score,
			       (select count(*)::int from visitors v
			        where v.inviter_id = m.id and v.season_id = $1
			          and v.is_void = false
			          and v.status_hadir in ('hadir', 'hadir_penuh'))             as from_visitor,
			       (select count(*)::int from tyfcb_entries te
			        where te.giver_id = m.id and te.season_id = $1
			          and te.status = 'verified' and te.pair_ordinal = 1)         as from_pair
			from members m
			where m.season_id = $1 and m.is_active
		)
		insert into raffle_tickets (season_id, member_id, sumber)
		select $1, e.member_id, s.sumber
		from entitlement e
		cross join lateral (values
			('score'::raffle_sumber, e.from_score),
			('visitor'::raffle_sumber, e.from_visitor),
			('tyfcb_pair'::raffle_sumber, e.from_pair)
		) as s(sumber, n)
		cross join lateral generate_series(1, s.n)
	`, seasonID); err != nil {
		return nil, err
	}

	rows, err := r.db.q(ctx).Query(ctx, `
		select m.id::text, u.full_name, t.nama_tim, count(rt.id)::int
		from members m
		join app_users u on u.id = m.user_id
		left join teams t on t.id = m.team_id
		join raffle_tickets rt on rt.member_id = m.id and rt.season_id = $1
		where m.season_id = $1
		group by m.id, u.full_name, t.nama_tim
		order by count(rt.id) desc, u.full_name
	`, seasonID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.TicketCount{}
	for rows.Next() {
		var t domain.TicketCount
		if err := rows.Scan(&t.MemberID, &t.FullName, &t.NamaTim, &t.Tickets); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r *PrizeRepo) TicketCounts(ctx context.Context, seasonID string) (map[string]int, error) {
	rows, err := r.db.q(ctx).Query(ctx, `
		select member_id::text, count(*)::int
		from raffle_tickets where season_id = $1 group by member_id
	`, seasonID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := map[string]int{}
	for rows.Next() {
		var id string
		var n int
		if err := rows.Scan(&id, &n); err != nil {
			return nil, err
		}
		out[id] = n
	}
	return out, rows.Err()
}

var _ domain.PrizeRepository = (*PrizeRepo)(nil)

// DrawWinner is deliberately one statement. Reading a random ticket and then
// writing the winner would leave a window where a second draw picks a
// different member and overwrites the first.
func (r *PrizeRepo) DrawWinner(ctx context.Context, seasonID, prizeID string) (*domain.Prize, error) {
	row := r.db.q(ctx).QueryRow(ctx, `
		with pick as (
			select member_id
			from raffle_tickets
			where season_id = $1
			order by random()
			limit 1
		)
		update prize_pool p
		set pemenang_id = (select member_id from pick),
		    status = 'awarded'::prize_status
		where p.id = $2
		  and p.season_id = $1
		  and p.pemenang_id is null
		  and exists (select 1 from pick)
		returning p.id
	`, seasonID, prizeID)

	var id string
	if err := row.Scan(&id); err != nil {
		if noRows(err) {
			// Nothing was written. The caller has already established that the
			// prize exists and is drawable, so this is an empty ticket pool or
			// a concurrent draw that got there first.
			return nil, nil
		}
		return nil, err
	}

	// Re-read so the caller gets the winner's name, not just their id.
	return r.FindByID(ctx, id)
}
