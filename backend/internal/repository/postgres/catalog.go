package postgres

import (
	"context"

	"github.com/ikurniawann/brag2026/backend/internal/domain"
)

// teamOrder sorts by the numeric suffix in the team name, so "Tim 10" follows
// "Tim 9" instead of "Tim 1".
const teamOrder = `order by nullif(regexp_replace(t.nama_tim, '\D', '', 'g'), '')::int nulls last, t.nama_tim`

type TeamRepo struct{ db *DB }

func NewTeamRepo(db *DB) *TeamRepo { return &TeamRepo{db: db} }

func (r *TeamRepo) ListBySeason(ctx context.Context, seasonID string) ([]domain.Team, error) {
	rows, err := r.db.q(ctx).Query(ctx, `
		select t.id, t.season_id, t.nama_tim,
		       (select count(*)::int from members m where m.team_id = t.id) as member_count
		from teams t
		where t.season_id = $1
	`+teamOrder, seasonID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.Team{}
	for rows.Next() {
		var t domain.Team
		if err := rows.Scan(&t.ID, &t.SeasonID, &t.NamaTim, &t.MemberCount); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r *TeamRepo) FindByID(ctx context.Context, id string) (*domain.Team, error) {
	var t domain.Team
	err := r.db.q(ctx).QueryRow(ctx,
		`select id, season_id, nama_tim from teams where id = $1 limit 1`, id).
		Scan(&t.ID, &t.SeasonID, &t.NamaTim)

	if noRows(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TeamRepo) Create(ctx context.Context, seasonID, namaTim string) (string, error) {
	var id string
	err := r.db.q(ctx).QueryRow(ctx,
		`insert into teams (season_id, nama_tim) values ($1, $2) returning id`,
		seasonID, namaTim).Scan(&id)

	if isUniqueViolation(err) {
		return "", domain.NewError(domain.ErrAlreadyExists, "Nama tim sudah dipakai di season ini.")
	}
	return id, err
}

func (r *TeamRepo) Rename(ctx context.Context, id, namaTim string) error {
	tag, err := r.db.q(ctx).Exec(ctx, `update teams set nama_tim = $1 where id = $2`, namaTim, id)
	if isUniqueViolation(err) {
		return domain.NewError(domain.ErrAlreadyExists, "Nama tim sudah dipakai di season ini.")
	}
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.NotFound("Tim tidak ditemukan.")
	}
	return nil
}

func (r *TeamRepo) Delete(ctx context.Context, id string) error {
	tag, err := r.db.q(ctx).Exec(ctx, `delete from teams where id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.NotFound("Tim tidak ditemukan.")
	}
	return nil
}

type ClassificationRepo struct{ db *DB }

func NewClassificationRepo(db *DB) *ClassificationRepo { return &ClassificationRepo{db: db} }

func (r *ClassificationRepo) List(ctx context.Context) ([]domain.Classification, error) {
	rows, err := r.db.q(ctx).Query(ctx, `select id, nama from classifications order by nama`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.Classification{}
	for rows.Next() {
		var c domain.Classification
		if err := rows.Scan(&c.ID, &c.Nama); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *ClassificationRepo) Create(ctx context.Context, nama string) (string, error) {
	var id string
	err := r.db.q(ctx).QueryRow(ctx,
		`insert into classifications (nama) values ($1) returning id`, nama).Scan(&id)

	if isUniqueViolation(err) {
		return "", domain.NewError(domain.ErrAlreadyExists, "Klasifikasi sudah ada.")
	}
	return id, err
}

func (r *ClassificationRepo) Rename(ctx context.Context, id, nama string) error {
	tag, err := r.db.q(ctx).Exec(ctx, `update classifications set nama = $1 where id = $2`, nama, id)
	if isUniqueViolation(err) {
		return domain.NewError(domain.ErrAlreadyExists, "Klasifikasi sudah ada.")
	}
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.NotFound("Klasifikasi tidak ditemukan.")
	}
	return nil
}

func (r *ClassificationRepo) Delete(ctx context.Context, id string) error {
	tag, err := r.db.q(ctx).Exec(ctx, `delete from classifications where id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.NotFound("Klasifikasi tidak ditemukan.")
	}
	return nil
}

func (r *ClassificationRepo) CountMembers(ctx context.Context, id string) (int, error) {
	var n int
	err := r.db.q(ctx).QueryRow(ctx,
		`select count(*)::int from members where klasifikasi_id = $1`, id).Scan(&n)
	return n, err
}

type BoosterRepo struct{ db *DB }

func NewBoosterRepo(db *DB) *BoosterRepo { return &BoosterRepo{db: db} }

func (r *BoosterRepo) ListBySeason(ctx context.Context, seasonID string, activeOnly bool) ([]domain.BoosterEvent, error) {
	sql := `
		select id, season_id, judul, deskripsi, tanggal_mulai, tanggal_berakhir, poin, status
		from booster_events
		where season_id = $1`
	if activeOnly {
		sql += ` and status = 'aktif' and current_date between tanggal_mulai and tanggal_berakhir`
	}
	sql += ` order by tanggal_mulai`

	rows, err := r.db.q(ctx).Query(ctx, sql, seasonID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.BoosterEvent{}
	for rows.Next() {
		var b domain.BoosterEvent
		if err := rows.Scan(&b.ID, &b.SeasonID, &b.Judul, &b.Deskripsi,
			&b.TanggalMulai, &b.TanggalBerakhir, &b.Poin, &b.Status); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func (r *BoosterRepo) FindByID(ctx context.Context, id string) (*domain.BoosterEvent, error) {
	var b domain.BoosterEvent
	err := r.db.q(ctx).QueryRow(ctx, `
		select id, season_id, judul, deskripsi, tanggal_mulai, tanggal_berakhir, poin, status
		from booster_events where id = $1 limit 1
	`, id).Scan(&b.ID, &b.SeasonID, &b.Judul, &b.Deskripsi,
		&b.TanggalMulai, &b.TanggalBerakhir, &b.Poin, &b.Status)

	if noRows(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *BoosterRepo) Create(ctx context.Context, b *domain.BoosterEvent) (string, error) {
	var id string
	err := r.db.q(ctx).QueryRow(ctx, `
		insert into booster_events (season_id, judul, deskripsi, tanggal_mulai, tanggal_berakhir, poin, status)
		values ($1, $2, $3, $4, $5, $6, $7)
		returning id
	`, b.SeasonID, b.Judul, b.Deskripsi, b.TanggalMulai, b.TanggalBerakhir, b.Poin, b.Status).Scan(&id)
	return id, err
}

func (r *BoosterRepo) Update(ctx context.Context, b *domain.BoosterEvent) error {
	tag, err := r.db.q(ctx).Exec(ctx, `
		update booster_events
		set judul = $1, deskripsi = $2, tanggal_mulai = $3,
		    tanggal_berakhir = $4, poin = $5, status = $6
		where id = $7
	`, b.Judul, b.Deskripsi, b.TanggalMulai, b.TanggalBerakhir, b.Poin, b.Status, b.ID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.NotFound("Booster tidak ditemukan.")
	}
	return nil
}

func (r *BoosterRepo) Delete(ctx context.Context, id string) error {
	tag, err := r.db.q(ctx).Exec(ctx, `delete from booster_events where id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.NotFound("Booster tidak ditemukan.")
	}
	return nil
}

type BadgeRepo struct{ db *DB }

func NewBadgeRepo(db *DB) *BadgeRepo { return &BadgeRepo{db: db} }

func (r *BadgeRepo) List(ctx context.Context) ([]domain.Badge, error) {
	rows, err := r.db.q(ctx).Query(ctx,
		`select badge_code, nama, deskripsi, ikon from badges order by badge_code`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.Badge{}
	for rows.Next() {
		var b domain.Badge
		if err := rows.Scan(&b.BadgeCode, &b.Nama, &b.Deskripsi, &b.Ikon); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func (r *BadgeRepo) ListForMember(ctx context.Context, memberID string) ([]domain.Badge, error) {
	rows, err := r.db.q(ctx).Query(ctx, `
		select b.badge_code, b.nama, b.deskripsi, b.ikon, to_char(mb.earned_at, 'DD Mon YYYY')
		from member_badges mb
		join badges b on b.badge_code = mb.badge_code
		where mb.member_id = $1
		order by mb.earned_at desc
	`, memberID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.Badge{}
	for rows.Next() {
		var b domain.Badge
		if err := rows.Scan(&b.BadgeCode, &b.Nama, &b.Deskripsi, &b.Ikon, &b.EarnedAt); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func (r *BadgeRepo) Award(ctx context.Context, memberID, badgeCode string) error {
	_, err := r.db.q(ctx).Exec(ctx, `
		insert into member_badges (member_id, badge_code) values ($1, $2)
		on conflict (member_id, badge_code) do nothing
	`, memberID, badgeCode)
	return err
}

var (
	_ domain.TeamRepository           = (*TeamRepo)(nil)
	_ domain.ClassificationRepository = (*ClassificationRepo)(nil)
	_ domain.BoosterRepository        = (*BoosterRepo)(nil)
	_ domain.BadgeRepository          = (*BadgeRepo)(nil)
)

// Stats gathers every fact the badge rules need in one round trip. Each
// subquery is independent, so this stays a single scan per source table
// rather than a fan-out of calls from the use case.
func (r *BadgeRepo) Stats(ctx context.Context, memberID, seasonID string) (domain.BadgeStats, error) {
	var s domain.BadgeStats

	err := r.db.q(ctx).QueryRow(ctx, `
		select
		  (select count(*)::int from tyfcb_entries
		     where giver_id = $1 and season_id = $2 and status = 'verified'),
		  (select count(distinct receiver_id)::int from tyfcb_entries
		     where giver_id = $1 and season_id = $2 and status = 'verified'),
		  (select coalesce(max(nilai), 0)::float8 from tyfcb_entries
		     where giver_id = $1 and season_id = $2 and status = 'verified'),
		  (select coalesce(sum(points), 0)::int from score_ledger
		     where member_id = $1 and season_id = $2),
		  (select count(*)::int from visitors
		     where inviter_id = $1 and season_id = $2 and is_void = false
		       and status_hadir in ('hadir', 'hadir_penuh')),
		  (select count(*)::int from visitors
		     where inviter_id = $1 and season_id = $2 and is_void = false
		       and status_hadir = 'hadir_penuh'),
		  (select count(*)::int from visitors
		     where inviter_id = $1 and season_id = $2 and is_void = false and is_converted),
		  (select count(distinct created_at::date)::int from score_ledger
		     where member_id = $1 and season_id = $2 and points > 0),
		  (select exists (select 1 from members
		     where id = $1 and color_status <> 'merah'))
	`, memberID, seasonID).Scan(
		&s.VerifiedTyfcbCount, &s.DistinctReceivers, &s.LargestTyfcb,
		&s.ScoreOverall, &s.VisitorsHadir, &s.VisitorsHadirPenuh,
		&s.Conversions, &s.DistinctScoringDays, &s.ColorStatusRaised,
	)
	if err != nil {
		return domain.BadgeStats{}, err
	}

	// TEAM_PLAYER needs the weekly Full Roster pass and PATRON needs the prize
	// pool; neither exists yet, so both stay false rather than being guessed.
	return s, nil
}
