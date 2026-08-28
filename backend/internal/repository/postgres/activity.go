package postgres

import (
	"context"

	"github.com/ikurniawann/brag2026/backend/internal/domain"
)

type ActivityRepo struct{ db *DB }

func NewActivityRepo(db *DB) *ActivityRepo { return &ActivityRepo{db: db} }

// Recent merges the two contribution types into a single stream. The union is
// done in SQL so the database handles the ordering and the limit, rather than
// fetching both lists whole and sorting in Go.
func (r *ActivityRepo) Recent(ctx context.Context, seasonID string, limit int) ([]domain.ActivityItem, error) {
	rows, err := r.db.q(ctx).Query(ctx, `
		(
			select te.id, 'tyfcb' as type,
			       ug.full_name as actor_name,
			       ur.full_name as target_name,
			       te.nilai::float8 as amount,
			       te.status::text as status,
			       te.computed_score as points,
			       te.created_at
			from tyfcb_entries te
			join members mg on mg.id = te.giver_id
			join app_users ug on ug.id = mg.user_id
			join members mr on mr.id = te.receiver_id
			join app_users ur on ur.id = mr.user_id
			where te.season_id = $1
		)
		union all
		(
			select v.id, 'visitor' as type,
			       ui.full_name as actor_name,
			       v.nama as target_name,
			       null::float8 as amount,
			       v.status_hadir::text as status,
			       null::int as points,
			       v.created_at
			from visitors v
			join members mi on mi.id = v.inviter_id
			join app_users ui on ui.id = mi.user_id
			where v.season_id = $1 and v.is_void = false
		)
		order by created_at desc
		limit $2
	`, seasonID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.ActivityItem{}
	for rows.Next() {
		var a domain.ActivityItem
		if err := rows.Scan(&a.ID, &a.Type, &a.ActorName, &a.TargetName,
			&a.Amount, &a.Status, &a.Points, &a.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

var _ domain.ActivityRepository = (*ActivityRepo)(nil)
