package postgres

import (
	"context"

	"github.com/ikurniawann/brag2026/backend/internal/domain"
)

type LedgerRepo struct{ db *DB }

func NewLedgerRepo(db *DB) *LedgerRepo { return &LedgerRepo{db: db} }

// Append is the only write path. There is deliberately no Update or Delete:
// corrections are new rows with the opposite sign, which keeps the audit trail
// intact and makes every score reproducible from the ledger alone.
func (r *LedgerRepo) Append(ctx context.Context, e *domain.LedgerEntry) error {
	_, err := r.db.q(ctx).Exec(ctx, `
		insert into score_ledger (season_id, member_id, team_id, kategori, points, sumber_ref, keterangan)
		values ($1, $2, $3, $4::ledger_kategori, $5, $6, $7)
	`, e.SeasonID, e.MemberID, e.TeamID, string(e.Kategori), e.Points, e.SumberRef, e.Keterangan)
	return err
}

// TeamScores is the leaderboard aggregate. TYFCB value and counts come from
// the entries themselves; points always come from the ledger.
func (r *LedgerRepo) TeamScores(ctx context.Context, seasonID string) ([]domain.TeamScore, error) {
	rows, err := r.db.q(ctx).Query(ctx, `
		with ledger_by_team as (
			select team_id,
			       coalesce(sum(points), 0)::int                                           as score_overall,
			       coalesce(sum(points) filter (where kategori = 'tyfcb'), 0)::int          as score_tyfcb,
			       coalesce(sum(points) filter (where kategori = 'visitor'), 0)::int        as score_visitor,
			       coalesce(sum(points) filter (where kategori = 'bonus'), 0)::int          as score_bonus
			from score_ledger
			where season_id = $1 and team_id is not null
			group by team_id
		),
		tyfcb_by_team as (
			select m.team_id,
			       coalesce(sum(te.nilai), 0)::float8 as nilai_tyfcb,
			       count(te.id)::int                  as count_tyfcb
			from tyfcb_entries te
			join members m on m.id = te.giver_id and m.season_id = $1
			where te.status = 'verified'
			group by m.team_id
		),
		visitor_by_team as (
			select m.team_id, count(v.id)::int as count_visitor
			from visitors v
			join members m on m.id = v.inviter_id and m.season_id = $1
			where v.is_void = false
			group by m.team_id
		)
		select t.id, t.nama_tim,
		       coalesce(l.score_overall, 0), coalesce(l.score_tyfcb, 0),
		       coalesce(l.score_visitor, 0), coalesce(l.score_bonus, 0),
		       coalesce(tt.nilai_tyfcb, 0), coalesce(tt.count_tyfcb, 0),
		       coalesce(vt.count_visitor, 0)
		from teams t
		left join ledger_by_team l   on l.team_id = t.id
		left join tyfcb_by_team tt   on tt.team_id = t.id
		left join visitor_by_team vt on vt.team_id = t.id
		where t.season_id = $1
		order by coalesce(l.score_overall, 0) desc,
		         nullif(regexp_replace(t.nama_tim, '\D', '', 'g'), '')::int nulls last
	`, seasonID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.TeamScore{}
	for rows.Next() {
		var t domain.TeamScore
		if err := rows.Scan(&t.TeamID, &t.NamaTim,
			&t.ScoreOverall, &t.ScoreTyfcb, &t.ScoreVisitor, &t.ScoreBonus,
			&t.NilaiTyfcb, &t.CountTyfcb, &t.CountVisitor); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

const memberScoreSelect = `
	select m.id, u.full_name, t.nama_tim,
	       coalesce(sum(sl.points), 0)::int                                    as score_overall,
	       coalesce(sum(sl.points) filter (where sl.kategori = 'tyfcb'), 0)::int   as score_tyfcb,
	       coalesce(sum(sl.points) filter (where sl.kategori = 'visitor'), 0)::int as score_visitor,
	       coalesce(sum(sl.points) filter (where sl.kategori = 'bonus'), 0)::int   as score_bonus
	from members m
	join app_users u on u.id = m.user_id
	left join teams t on t.id = m.team_id
	left join score_ledger sl on sl.member_id = m.id and sl.season_id = m.season_id
`

func (r *LedgerRepo) MemberScores(ctx context.Context, seasonID string, kategori domain.ScoreCategory, limit int) ([]domain.MemberScore, error) {
	// Column() returns one of three constants from the domain package, so this
	// is not a place user input reaches the statement.
	rows, err := r.db.q(ctx).Query(ctx, memberScoreSelect+`
		where m.season_id = $1
		group by m.id, u.full_name, t.nama_tim
		order by `+kategori.Column()+` desc, score_overall desc, u.full_name
		limit $2
	`, seasonID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.MemberScore{}
	for rows.Next() {
		var m domain.MemberScore
		if err := rows.Scan(&m.MemberID, &m.FullName, &m.NamaTim,
			&m.ScoreOverall, &m.ScoreTyfcb, &m.ScoreVisitor, &m.ScoreBonus); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r *LedgerRepo) MemberScore(ctx context.Context, memberID, seasonID string) (*domain.MemberScore, error) {
	var m domain.MemberScore
	err := r.db.q(ctx).QueryRow(ctx, memberScoreSelect+`
		where m.id = $1 and m.season_id = $2
		group by m.id, u.full_name, t.nama_tim
	`, memberID, seasonID).Scan(&m.MemberID, &m.FullName, &m.NamaTim,
		&m.ScoreOverall, &m.ScoreTyfcb, &m.ScoreVisitor, &m.ScoreBonus)

	if noRows(err) {
		return &domain.MemberScore{MemberID: memberID}, nil
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// TeamHistory lists the ledger rows behind a team's score. kategori is
// optional; empty means every category.
func (r *LedgerRepo) TeamHistory(ctx context.Context, teamID, seasonID, kategori string) ([]domain.LedgerEntry, error) {
	sql := `
		select id, season_id, member_id, team_id, kategori::text, points,
		       sumber_ref, keterangan, created_at
		from score_ledger
		where team_id = $1 and season_id = $2`
	args := []any{teamID, seasonID}

	if kategori != "" {
		args = append(args, kategori)
		sql += ` and kategori::text = $3`
	}
	sql += ` order by created_at desc limit 200`

	rows, err := r.db.q(ctx).Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.LedgerEntry{}
	for rows.Next() {
		var e domain.LedgerEntry
		if err := rows.Scan(&e.ID, &e.SeasonID, &e.MemberID, &e.TeamID, &e.Kategori,
			&e.Points, &e.SumberRef, &e.Keterangan, &e.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (r *LedgerRepo) SumBySource(ctx context.Context, sumberRef string) (int, error) {
	var total int
	err := r.db.q(ctx).QueryRow(ctx,
		`select coalesce(sum(points), 0)::int from score_ledger where sumber_ref = $1`,
		sumberRef).Scan(&total)
	return total, err
}

var _ domain.LedgerRepository = (*LedgerRepo)(nil)
