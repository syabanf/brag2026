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

// ReplaceTickets rewrites a member's entitlement rather than appending, so the
// issue pass can run as often as the committee likes without inflating anyone's
// odds.
func (r *PrizeRepo) ReplaceTickets(ctx context.Context, seasonID, memberID string, bySource map[domain.RaffleSource]int) error {
	if _, err := r.db.q(ctx).Exec(ctx,
		`delete from raffle_tickets where season_id = $1 and member_id = $2`,
		seasonID, memberID); err != nil {
		return err
	}

	for source, count := range bySource {
		for i := 0; i < count; i++ {
			if _, err := r.db.q(ctx).Exec(ctx, `
				insert into raffle_tickets (season_id, member_id, sumber)
				values ($1, $2, $3::raffle_sumber)
			`, seasonID, memberID, string(source)); err != nil {
				return err
			}
		}
	}
	return nil
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

// RaffleInputs gathers the three entitlement sources in one round trip.
func (r *PrizeRepo) RaffleInputs(ctx context.Context, seasonID, memberID string) (int, int, int, error) {
	var score, visitors, newPairs int

	err := r.db.q(ctx).QueryRow(ctx, `
		select
		  (select coalesce(sum(points), 0)::int from score_ledger
		     where member_id = $1 and season_id = $2),
		  (select count(*)::int from visitors
		     where inviter_id = $1 and season_id = $2 and is_void = false
		       and status_hadir in ('hadir', 'hadir_penuh')),
		  (select count(*)::int from tyfcb_entries
		     where giver_id = $1 and season_id = $2
		       and status = 'verified' and pair_ordinal = 1)
	`, memberID, seasonID).Scan(&score, &visitors, &newPairs)

	return score, visitors, newPairs, err
}

var _ domain.PrizeRepository = (*PrizeRepo)(nil)
