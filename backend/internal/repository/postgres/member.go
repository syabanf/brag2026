package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/ikurniawann/brag2026/backend/internal/domain"
)

type MemberRepo struct{ db *DB }

func NewMemberRepo(db *DB) *MemberRepo { return &MemberRepo{db: db} }

const memberColumns = `
	m.id, m.user_id, m.season_id, m.team_id, m.klasifikasi_id,
	m.color_status::text, m.is_active,
	u.full_name, u.email, u.role::text,
	t.nama_tim, c.nama
`

const memberFrom = `
	from members m
	join app_users u on u.id = m.user_id
	left join teams t on t.id = m.team_id
	left join classifications c on c.id = m.klasifikasi_id
`

func scanMember(row pgx.Row) (*domain.Member, error) {
	var m domain.Member
	err := row.Scan(
		&m.ID, &m.UserID, &m.SeasonID, &m.TeamID, &m.KlasifikasiID,
		&m.ColorStatus, &m.IsActive,
		&m.FullName, &m.Email, &m.Role,
		&m.NamaTim, &m.KlasifikasiNama,
	)
	if noRows(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func collectMembers(rows pgx.Rows) ([]domain.Member, error) {
	defer rows.Close()

	out := []domain.Member{}
	for rows.Next() {
		var m domain.Member
		if err := rows.Scan(
			&m.ID, &m.UserID, &m.SeasonID, &m.TeamID, &m.KlasifikasiID,
			&m.ColorStatus, &m.IsActive,
			&m.FullName, &m.Email, &m.Role,
			&m.NamaTim, &m.KlasifikasiNama,
		); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r *MemberRepo) FindByUserAndSeason(ctx context.Context, userID, seasonID string) (*domain.Member, error) {
	return scanMember(r.db.q(ctx).QueryRow(ctx,
		`select `+memberColumns+memberFrom+` where m.user_id = $1 and m.season_id = $2 limit 1`,
		userID, seasonID))
}

func (r *MemberRepo) FindByID(ctx context.Context, id string) (*domain.Member, error) {
	return scanMember(r.db.q(ctx).QueryRow(ctx,
		`select `+memberColumns+memberFrom+` where m.id = $1 limit 1`, id))
}

// ListBySeason orders by the numeric suffix of the team name so "Tim 10" sorts
// after "Tim 9" rather than after "Tim 1".
func (r *MemberRepo) ListBySeason(ctx context.Context, seasonID string) ([]domain.Member, error) {
	rows, err := r.db.q(ctx).Query(ctx,
		`select `+memberColumns+memberFrom+`
		 where m.season_id = $1
		 order by nullif(regexp_replace(coalesce(t.nama_tim, ''), '\D', '', 'g'), '')::int nulls last,
		          u.full_name`, seasonID)
	if err != nil {
		return nil, err
	}
	return collectMembers(rows)
}

func (r *MemberRepo) ListByTeam(ctx context.Context, teamID string) ([]domain.Member, error) {
	rows, err := r.db.q(ctx).Query(ctx,
		`select `+memberColumns+memberFrom+` where m.team_id = $1 order by u.full_name`, teamID)
	if err != nil {
		return nil, err
	}
	return collectMembers(rows)
}

func (r *MemberRepo) Search(ctx context.Context, seasonID, term string, limit int) ([]domain.Member, error) {
	rows, err := r.db.q(ctx).Query(ctx,
		`select `+memberColumns+memberFrom+`
		 where m.season_id = $1 and m.is_active
		   and (u.full_name ilike '%' || $2 || '%' or u.email ilike '%' || $2 || '%')
		 order by u.full_name
		 limit $3`, seasonID, term, limit)
	if err != nil {
		return nil, err
	}
	return collectMembers(rows)
}

func (r *MemberRepo) Create(ctx context.Context, m *domain.Member) (string, error) {
	var id string
	err := r.db.q(ctx).QueryRow(ctx, `
		insert into members (user_id, season_id, team_id, klasifikasi_id, color_status, is_active)
		values ($1, $2, $3, $4, $5::color_status, $6)
		returning id
	`, m.UserID, m.SeasonID, m.TeamID, m.KlasifikasiID, string(m.ColorStatus), m.IsActive).Scan(&id)

	if isUniqueViolation(err) {
		return "", domain.NewError(domain.ErrAlreadyExists, "Member sudah terdaftar di season ini.")
	}
	return id, err
}

func (r *MemberRepo) Update(ctx context.Context, m *domain.Member) error {
	_, err := r.db.q(ctx).Exec(ctx, `
		update members
		set team_id = $1, klasifikasi_id = $2, color_status = $3::color_status, is_active = $4
		where id = $5
	`, m.TeamID, m.KlasifikasiID, string(m.ColorStatus), m.IsActive, m.ID)
	return err
}

func (r *MemberRepo) CountBySeason(ctx context.Context, seasonID string) (int, error) {
	var n int
	err := r.db.q(ctx).QueryRow(ctx,
		`select count(*)::int from members where season_id = $1`, seasonID).Scan(&n)
	return n, err
}

var _ domain.MemberRepository = (*MemberRepo)(nil)
