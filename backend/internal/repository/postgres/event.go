package postgres

import (
	"context"
	"time"

	"github.com/ikurniawann/brag2026/backend/internal/domain"
)

type WeeklyEventRepo struct{ db *DB }

func NewWeeklyEventRepo(db *DB) *WeeklyEventRepo { return &WeeklyEventRepo{db: db} }

const weeklyEventSelect = `
	select id, season_id, minggu_ke, event_code, target_classification_id,
	       tanggal_mulai, tanggal_selesai
	from weekly_events
`

func (r *WeeklyEventRepo) scan(row interface{ Scan(...any) error }) (*domain.WeeklyEvent, error) {
	var e domain.WeeklyEvent
	err := row.Scan(&e.ID, &e.SeasonID, &e.MingguKe, &e.EventCode,
		&e.TargetClassificationID, &e.TanggalMulai, &e.TanggalSelesai)

	if noRows(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	e.Nama, e.Mekanik = domain.DescribeEvent(e.EventCode)
	return &e, nil
}

// ActiveOn finds the event covering a date. The unique constraint is on
// (season, week) rather than on the date range, so an admin who edits dates can
// leave two windows overlapping. Ordering makes the outcome predictable: the
// window that started most recently wins, and the later week breaks a tie.
func (r *WeeklyEventRepo) ActiveOn(ctx context.Context, seasonID string, day time.Time) (*domain.WeeklyEvent, error) {
	return r.scan(r.db.q(ctx).QueryRow(ctx, weeklyEventSelect+`
		where season_id = $1 and $2::date between tanggal_mulai and tanggal_selesai
		order by tanggal_mulai desc, minggu_ke desc
		limit 1
	`, seasonID, day))
}

func (r *WeeklyEventRepo) ListBySeason(ctx context.Context, seasonID string) ([]domain.WeeklyEvent, error) {
	rows, err := r.db.q(ctx).Query(ctx, weeklyEventSelect+` where season_id = $1 order by minggu_ke`, seasonID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.WeeklyEvent{}
	for rows.Next() {
		var e domain.WeeklyEvent
		if err := rows.Scan(&e.ID, &e.SeasonID, &e.MingguKe, &e.EventCode,
			&e.TargetClassificationID, &e.TanggalMulai, &e.TanggalSelesai); err != nil {
			return nil, err
		}
		e.Nama, e.Mekanik = domain.DescribeEvent(e.EventCode)
		out = append(out, e)
	}
	return out, rows.Err()
}

// Upsert keys on (season, week) because the schema permits only one event per
// week — scheduling a different one replaces it rather than erroring.
func (r *WeeklyEventRepo) Upsert(ctx context.Context, e *domain.WeeklyEvent) (string, error) {
	var id string
	err := r.db.q(ctx).QueryRow(ctx, `
		insert into weekly_events
			(season_id, minggu_ke, event_code, target_classification_id, tanggal_mulai, tanggal_selesai)
		values ($1, $2, $3, $4, $5, $6)
		on conflict (season_id, minggu_ke) do update
		set event_code = excluded.event_code,
		    target_classification_id = excluded.target_classification_id,
		    tanggal_mulai = excluded.tanggal_mulai,
		    tanggal_selesai = excluded.tanggal_selesai
		returning id
	`, e.SeasonID, e.MingguKe, string(e.EventCode), e.TargetClassificationID,
		e.TanggalMulai, e.TanggalSelesai).Scan(&id)
	return id, err
}

func (r *WeeklyEventRepo) Delete(ctx context.Context, id string) error {
	tag, err := r.db.q(ctx).Exec(ctx, `delete from weekly_events where id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.NotFound("Event tidak ditemukan.")
	}
	return nil
}

var _ domain.WeeklyEventRepository = (*WeeklyEventRepo)(nil)
