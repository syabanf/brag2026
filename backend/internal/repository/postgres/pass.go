package postgres

import (
	"context"
	"time"

	"github.com/ikurniawann/brag2026/backend/internal/domain"
)

type ScoringPassRepo struct{ db *DB }

func NewScoringPassRepo(db *DB) *ScoringPassRepo { return &ScoringPassRepo{db: db} }

// RosterForWeek counts active members per team against how many of them earned
// points in the window, which is exactly the Full Roster test.
func (r *ScoringPassRepo) RosterForWeek(ctx context.Context, seasonID string, from, to time.Time) ([]domain.RosterStatus, error) {
	rows, err := r.db.q(ctx).Query(ctx, `
		with active as (
			select m.id, m.team_id
			from members m
			where m.season_id = $1 and m.is_active and m.team_id is not null
		),
		scored as (
			select distinct a.id, a.team_id
			from active a
			join score_ledger sl on sl.member_id = a.id
			where sl.season_id = $1
			  and sl.points > 0
			  and sl.created_at >= $2 and sl.created_at < $3::date + 1
		)
		select t.id, t.nama_tim,
		       count(distinct a.id)::int as active_count,
		       count(distinct s.id)::int as scoring_count,
		       coalesce(array_agg(distinct a.id::text) filter (where a.id is not null), '{}')
		from teams t
		left join active a on a.team_id = t.id
		left join scored s on s.id = a.id
		where t.season_id = $1
		group by t.id, t.nama_tim
	`, seasonID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.RosterStatus{}
	for rows.Next() {
		var s domain.RosterStatus
		if err := rows.Scan(&s.TeamID, &s.NamaTim, &s.ActiveCount, &s.ScoringCount, &s.MemberIDs); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *ScoringPassRepo) TopTyfcbOfDay(ctx context.Context, seasonID string, day time.Time) (*domain.TyfcbEntry, error) {
	return scanTyfcb(r.db.q(ctx).QueryRow(ctx, tyfcbSelect+`
		where te.season_id = $1 and te.tanggal = $2::date and te.status = 'verified'
		order by te.nilai desc
		limit 1
	`, seasonID, day))
}

func (r *ScoringPassRepo) MembersWithScoringDays(ctx context.Context, seasonID string, from, to time.Time, days int) ([]string, error) {
	rows, err := r.db.q(ctx).Query(ctx, `
		select member_id::text
		from score_ledger
		where season_id = $1 and member_id is not null and points > 0
		  and created_at >= $2 and created_at < $3::date + 1
		group by member_id
		having count(distinct created_at::date) >= $4
	`, seasonID, from, to, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

// AlreadySettled makes every pass re-runnable: the source key encodes what was
// settled and for which period, so a second run finds the row and skips.
func (r *ScoringPassRepo) AlreadySettled(ctx context.Context, sumberRef string) (bool, error) {
	var exists bool
	err := r.db.q(ctx).QueryRow(ctx,
		`select exists (select 1 from score_ledger where sumber_ref = $1)`, sumberRef).Scan(&exists)
	return exists, err
}

func (r *ScoringPassRepo) TeamOf(ctx context.Context, memberID string) (*string, error) {
	var teamID *string
	err := r.db.q(ctx).QueryRow(ctx,
		`select team_id from members where id = $1`, memberID).Scan(&teamID)
	if noRows(err) {
		return nil, nil
	}
	return teamID, err
}

var _ domain.ScoringPassRepository = (*ScoringPassRepo)(nil)
