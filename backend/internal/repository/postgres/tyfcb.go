package postgres

import (
	"context"
	"strconv"

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

// tyfcbFilterClause is shared by the list, the paged list and the count, so
// all three can never disagree about what is being filtered.
func tyfcbFilterClause(f domain.TyfcbFilter) *clause {
	c := &clause{}
	c.addIf("te.season_id = ", f.SeasonID)
	c.addIf("te.status::text = ", f.Status)
	c.addIf("te.giver_id = ", f.GiverID)
	c.addIf("te.receiver_id = ", f.ReceiverID)
	c.addIf("gm.team_id = ", f.TeamID)
	if f.DateFrom != nil {
		c.add("te.tanggal >= ", *f.DateFrom)
	}
	if f.DateTo != nil {
		c.add("te.tanggal <= ", *f.DateTo)
	}
	// Either side of the transaction, since an admin rarely knows who filed.
	c.addSearch(f.Search, "gu.full_name", "ru.full_name")
	return c
}

func collectTyfcb(rows pgx.Rows) ([]domain.TyfcbEntry, error) {
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

func (r *TyfcbRepo) List(ctx context.Context, f domain.TyfcbFilter) ([]domain.TyfcbEntry, error) {
	c := tyfcbFilterClause(f)
	sql := tyfcbSelect + c.sql() + " order by te.created_at desc"
	args := c.args

	if f.Limit > 0 {
		args = append(args, f.Limit)
		sql += " limit $" + strconv.Itoa(len(args))
	}

	rows, err := r.db.q(ctx).Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return collectTyfcb(rows)
}

// ListPaged adds the total so the admin screen can page through verification
// instead of loading every submission in the season.
func (r *TyfcbRepo) ListPaged(ctx context.Context, f domain.TyfcbFilter) ([]domain.TyfcbEntry, int, error) {
	c := tyfcbFilterClause(f)

	var total int
	if err := r.db.q(ctx).QueryRow(ctx, `
		select count(*)::int
		from tyfcb_entries te
		left join members gm on gm.id = te.giver_id
		left join app_users gu on gu.id = gm.user_id
		left join members rm on rm.id = te.receiver_id
		left join app_users ru on ru.id = rm.user_id`+c.sql(), c.args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	page := f.Page.Normalise()
	tail, args := c.paginate(page.Limit, page.Offset)

	rows, err := r.db.q(ctx).Query(ctx,
		tyfcbSelect+c.sql()+" order by te.created_at desc"+tail, args...)
	if err != nil {
		return nil, 0, err
	}

	entries, err := collectTyfcb(rows)
	return entries, total, err
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

// UpdateStatusGuarded only writes when the row still holds `From`. A losing
// concurrent request gets false rather than crediting the entry a second time.
func (r *TyfcbRepo) UpdateStatusGuarded(ctx context.Context, id string, c domain.TyfcbStatusChange) (bool, error) {
	tag, err := r.db.q(ctx).Exec(ctx, `
		update tyfcb_entries
		set status = $1::tyfcb_status, verified_by = $2, verified_at = $3,
		    rejection_reason = $4
		where id = $5 and status = $6::tyfcb_status
	`, string(c.To), c.VerifiedBy, c.VerifiedAt, c.Reason, id, string(c.From))
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() == 1, nil
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
