package postgres

import (
	"context"
	"time"

	"github.com/ikurniawann/brag2026/backend/internal/domain"
)

type UserRepo struct{ db *DB }

func NewUserRepo(db *DB) *UserRepo { return &UserRepo{db: db} }

func (r *UserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, string, error) {
	var u domain.User
	var hash string

	err := r.db.q(ctx).QueryRow(ctx, `
		select id, email, full_name, role::text, password_hash
		from app_users
		where lower(email) = lower($1)
		limit 1
	`, email).Scan(&u.ID, &u.Email, &u.FullName, &u.Role, &hash)

	if noRows(err) {
		return nil, "", nil
	}
	if err != nil {
		return nil, "", err
	}
	return &u, hash, nil
}

func (r *UserRepo) FindByID(ctx context.Context, id string) (*domain.User, error) {
	var u domain.User
	err := r.db.q(ctx).QueryRow(ctx, `
		select id, email, full_name, role::text from app_users where id = $1 limit 1
	`, id).Scan(&u.ID, &u.Email, &u.FullName, &u.Role)

	if noRows(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) UpdatePassword(ctx context.Context, userID, passwordHash string) error {
	_, err := r.db.q(ctx).Exec(ctx, `
		update app_users set password_hash = $1, updated_at = now() where id = $2
	`, passwordHash, userID)
	return err
}

func (r *UserRepo) UpdateRole(ctx context.Context, userID string, role domain.Role) error {
	_, err := r.db.q(ctx).Exec(ctx, `
		update app_users set role = $1::app_role, updated_at = now() where id = $2
	`, string(role), userID)
	return err
}

func (r *UserRepo) UpdateProfile(ctx context.Context, userID, fullName, email string) error {
	_, err := r.db.q(ctx).Exec(ctx, `
		update app_users set full_name = $1, email = $2, updated_at = now() where id = $3
	`, fullName, email, userID)
	if isUniqueViolation(err) {
		return domain.NewError(domain.ErrAlreadyExists, "Email sudah dipakai akun lain.")
	}
	return err
}

func (r *UserRepo) Create(ctx context.Context, email, passwordHash, fullName string, role domain.Role) (string, error) {
	var id string
	err := r.db.q(ctx).QueryRow(ctx, `
		insert into app_users (email, password_hash, full_name, role)
		values ($1, $2, $3, $4::app_role)
		returning id
	`, email, passwordHash, fullName, string(role)).Scan(&id)

	if isUniqueViolation(err) {
		return "", domain.NewError(domain.ErrAlreadyExists, "Email sudah terdaftar.")
	}
	return id, err
}

type SessionRepo struct{ db *DB }

func NewSessionRepo(db *DB) *SessionRepo { return &SessionRepo{db: db} }

func (r *SessionRepo) Create(ctx context.Context, userID, tokenHash string, expiresAt time.Time) error {
	_, err := r.db.q(ctx).Exec(ctx, `
		insert into user_sessions (user_id, token_hash, expires_at) values ($1, $2, $3)
	`, userID, tokenHash, expiresAt)
	return err
}

func (r *SessionRepo) FindUserByTokenHash(ctx context.Context, tokenHash string) (*domain.User, error) {
	var u domain.User
	err := r.db.q(ctx).QueryRow(ctx, `
		select u.id, u.email, u.full_name, u.role::text
		from user_sessions s
		join app_users u on u.id = s.user_id
		where s.token_hash = $1 and s.expires_at > now()
		limit 1
	`, tokenHash).Scan(&u.ID, &u.Email, &u.FullName, &u.Role)

	if noRows(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *SessionRepo) Delete(ctx context.Context, tokenHash string) error {
	_, err := r.db.q(ctx).Exec(ctx, `delete from user_sessions where token_hash = $1`, tokenHash)
	return err
}

func (r *SessionRepo) DeleteAllForUser(ctx context.Context, userID string) error {
	_, err := r.db.q(ctx).Exec(ctx, `delete from user_sessions where user_id = $1`, userID)
	return err
}

func (r *SessionRepo) DeleteExpired(ctx context.Context) (int64, error) {
	tag, err := r.db.q(ctx).Exec(ctx, `delete from user_sessions where expires_at <= now()`)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

type SeasonRepo struct{ db *DB }

func NewSeasonRepo(db *DB) *SeasonRepo { return &SeasonRepo{db: db} }

func scanSeason(row interface{ Scan(...any) error }) (*domain.Season, error) {
	var s domain.Season
	err := row.Scan(&s.ID, &s.Nama, &s.StartsOn, &s.EndsOn, &s.Status)
	if noRows(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// FindActive prefers the season flagged active and falls back to the most
// recently created one, so a half-configured database still renders.
func (r *SeasonRepo) FindActive(ctx context.Context) (*domain.Season, error) {
	return scanSeason(r.db.q(ctx).QueryRow(ctx, `
		select id, nama, starts_on, ends_on, status::text
		from event_seasons
		order by (status = 'active') desc, created_at desc
		limit 1
	`))
}

func (r *SeasonRepo) FindByName(ctx context.Context, nama string) (*domain.Season, error) {
	return scanSeason(r.db.q(ctx).QueryRow(ctx, `
		select id, nama, starts_on, ends_on, status::text
		from event_seasons where nama = $1 limit 1
	`, nama))
}

var (
	_ domain.UserRepository    = (*UserRepo)(nil)
	_ domain.SessionRepository = (*SessionRepo)(nil)
	_ domain.SeasonRepository  = (*SeasonRepo)(nil)
)
