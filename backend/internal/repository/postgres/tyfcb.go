package postgres

import (
	"context"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/ikurniawann/brag2026/backend/internal/domain"
)

type TyfcbRepo struct{ db *DB }

func NewTyfcbRepo(db *DB) *TyfcbRepo { return &TyfcbRepo{db: db} }

const tyfcbSelect = `
	select te.id, te.season_id, te.giver_id, te.receiver_id, te.nilai, te.tanggal,
	       te.status::text, te.computed_score, te.pair_ordinal,
	       te.event_multiplier_applied, te.rejection_reason, te.created_at,
	       gu.full_name, ru.full_name
	from tyfcb_entries te
	left join members gm on gm.id = te.giver_id
	left join app_users gu on gu.id = gm.user_id
	left join members rm on rm.id = te.receiver_id
	left join app_users ru on ru.id = rm.user_id
`

func scanTyfcb(row pgx.Row) (*domain.TyfcbEntry, error) {
	var e domain.TyfcbEntry
	var giver, receiver *string

	err := row.Scan(&e.ID, &e.SeasonID, &e.GiverID, &e.ReceiverID, &e.Nilai, &e.Tanggal,
		&e.Status, &e.ComputedScore, &e.PairOrdinal,
		&e.EventMultiplierApplied, &e.RejectionReason, &e.CreatedAt,
		&giver, &receiver)

	if noRows(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if giver != nil {
		e.GiverName = *giver
	}
	if receiver != nil {
		e.ReceiverName = *receiver
	}
	return &e, nil
}

func (r *TyfcbRepo) FindByID(ctx context.Context, id string) (*domain.TyfcbEntry, error) {
	return scanTyfcb(r.db.q(ctx).QueryRow(ctx, tyfcbSelect+` where te.id = $1 limit 1`, id))
}

// List builds the filter incrementally so callers only pay for the predicates
// they actually set.
func (r *TyfcbRepo) List(ctx context.Context, f domain.TyfcbFilter) ([]domain.TyfcbEntry, error) {
	var (
		where []string
		args  []any
	)
	add := func(clause string, value any) {
		args = append(args, value)
		where = append(where, clause+"$"+itoa(len(args)))
	}

	if f.SeasonID != "" {
		add("te.season_id = ", f.SeasonID)
	}
	if f.Status != "" {
		add("te.status::text = ", f.Status)
	}
	if f.GiverID != "" {
		add("te.giver_id = ", f.GiverID)
	}
	if f.ReceiverID != "" {
		add("te.receiver_id = ", f.ReceiverID)
	}
	if f.TeamID != "" {
		add("gm.team_id = ", f.TeamID)
	}

	sql := tyfcbSelect
	if len(where) > 0 {
		sql += " where " + strings.Join(where, " and ")
	}
	sql += " order by te.created_at desc"
	if f.Limit > 0 {
		args = append(args, f.Limit)
		sql += " limit $" + itoa(len(args))
	}

	rows, err := r.db.q(ctx).Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.TyfcbEntry{}
	for rows.Next() {
		var e domain.TyfcbEntry
		var giver, receiver *string
		if err := rows.Scan(&e.ID, &e.SeasonID, &e.GiverID, &e.ReceiverID, &e.Nilai, &e.Tanggal,
			&e.Status, &e.ComputedScore, &e.PairOrdinal,
			&e.EventMultiplierApplied, &e.RejectionReason, &e.CreatedAt,
			&giver, &receiver); err != nil {
			return nil, err
		}
		if giver != nil {
			e.GiverName = *giver
		}
		if receiver != nil {
			e.ReceiverName = *receiver
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// CountPair drives the pair penalty: how many transactions this exact
// giver→receiver pair already has in the season.
func (r *TyfcbRepo) CountPair(ctx context.Context, giverID, receiverID, seasonID string) (int, error) {
	var n int
	err := r.db.q(ctx).QueryRow(ctx, `
		select count(*)::int from tyfcb_entries
		where giver_id = $1 and receiver_id = $2 and season_id = $3
	`, giverID, receiverID, seasonID).Scan(&n)
	return n, err
}

func (r *TyfcbRepo) Create(ctx context.Context, e *domain.TyfcbEntry, submittedBy *string) (string, error) {
	var id string
	err := r.db.q(ctx).QueryRow(ctx, `
		insert into tyfcb_entries
			(season_id, giver_id, receiver_id, nilai, tanggal, status,
			 computed_score, pair_ordinal, event_multiplier_applied, submitted_by)
		values ($1, $2, $3, $4, $5, $6::tyfcb_status, $7, $8, $9, $10)
		returning id
	`, e.SeasonID, e.GiverID, e.ReceiverID, e.Nilai, e.Tanggal, string(e.Status),
		e.ComputedScore, e.PairOrdinal, e.EventMultiplierApplied, submittedBy).Scan(&id)
	return id, err
}

func (r *TyfcbRepo) UpdateStatus(ctx context.Context, id string, status domain.TyfcbStatus, verifiedBy *string, verifiedAt *time.Time) error {
	tag, err := r.db.q(ctx).Exec(ctx, `
		update tyfcb_entries
		set status = $1::tyfcb_status, verified_by = $2, verified_at = $3
		where id = $4
	`, string(status), verifiedBy, verifiedAt, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.NotFound("Entry tidak ditemukan.")
	}
	return nil
}

func (r *TyfcbRepo) Void(ctx context.Context, id, voidedBy string) error {
	tag, err := r.db.q(ctx).Exec(ctx, `
		update tyfcb_entries
		set status = 'void'::tyfcb_status, voided_by = $1, voided_at = now()
		where id = $2 and status <> 'void'
	`, voidedBy, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.Conflict("Entry sudah di-void.")
	}
	return nil
}

func (r *TyfcbRepo) CountByStatus(ctx context.Context, seasonID string) (map[string]int, error) {
	rows, err := r.db.q(ctx).Query(ctx, `
		select status::text, count(*)::int
		from tyfcb_entries where season_id = $1 group by status
	`, seasonID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := map[string]int{}
	for rows.Next() {
		var status string
		var n int
		if err := rows.Scan(&status, &n); err != nil {
			return nil, err
		}
		out[status] = n
	}
	return out, rows.Err()
}

var _ domain.TyfcbRepository = (*TyfcbRepo)(nil)
