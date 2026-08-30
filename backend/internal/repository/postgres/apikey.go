package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/ikurniawann/brag2026/backend/internal/domain"
)

type APIKeyRepo struct{ db *DB }

func NewAPIKeyRepo(db *DB) *APIKeyRepo { return &APIKeyRepo{db: db} }

const apiKeySelect = `
	select k.id, k.nama, k.prefix, k.user_id, k.read_only,
	       k.last_used_at, k.expires_at, k.revoked_at, k.created_at,
	       u.full_name, u.email
	from api_keys k
	join app_users u on u.id = k.user_id
`

func scanAPIKey(row pgx.Row) (*domain.APIKey, error) {
	var k domain.APIKey
	err := row.Scan(&k.ID, &k.Nama, &k.Prefix, &k.UserID, &k.ReadOnly,
		&k.LastUsedAt, &k.ExpiresAt, &k.RevokedAt, &k.CreatedAt,
		&k.UserName, &k.UserEmail)

	if noRows(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &k, nil
}

// FindByHash returns the key and its owner in one round trip, since every
// request authenticated by a key needs both.
func (r *APIKeyRepo) FindByHash(ctx context.Context, hash string) (*domain.APIKey, *domain.User, error) {
	var k domain.APIKey
	var u domain.User

	err := r.db.q(ctx).QueryRow(ctx, `
		select k.id, k.nama, k.prefix, k.user_id, k.read_only,
		       k.last_used_at, k.expires_at, k.revoked_at, k.created_at,
		       u.id, u.full_name, u.email, u.role::text
		from api_keys k
		join app_users u on u.id = k.user_id
		where k.key_hash = $1
		limit 1
	`, hash).Scan(&k.ID, &k.Nama, &k.Prefix, &k.UserID, &k.ReadOnly,
		&k.LastUsedAt, &k.ExpiresAt, &k.RevokedAt, &k.CreatedAt,
		&u.ID, &u.FullName, &u.Email, &u.Role)

	if noRows(err) {
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, err
	}

	k.UserName, k.UserEmail = u.FullName, u.Email
	return &k, &u, nil
}

func collectAPIKeys(rows pgx.Rows) ([]domain.APIKey, error) {
	defer rows.Close()

	out := []domain.APIKey{}
	for rows.Next() {
		var k domain.APIKey
		if err := rows.Scan(&k.ID, &k.Nama, &k.Prefix, &k.UserID, &k.ReadOnly,
			&k.LastUsedAt, &k.ExpiresAt, &k.RevokedAt, &k.CreatedAt,
			&k.UserName, &k.UserEmail); err != nil {
			return nil, err
		}
		out = append(out, k)
	}
	return out, rows.Err()
}

func (r *APIKeyRepo) ListByUser(ctx context.Context, userID string) ([]domain.APIKey, error) {
	rows, err := r.db.q(ctx).Query(ctx,
		apiKeySelect+` where k.user_id = $1 order by k.created_at desc`, userID)
	if err != nil {
		return nil, err
	}
	return collectAPIKeys(rows)
}

func (r *APIKeyRepo) List(ctx context.Context) ([]domain.APIKey, error) {
	rows, err := r.db.q(ctx).Query(ctx, apiKeySelect+` order by k.created_at desc`)
	if err != nil {
		return nil, err
	}
	return collectAPIKeys(rows)
}

func (r *APIKeyRepo) Create(ctx context.Context, k *domain.APIKey, hash, createdBy string) (string, error) {
	var id string
	err := r.db.q(ctx).QueryRow(ctx, `
		insert into api_keys (nama, prefix, key_hash, user_id, read_only, expires_at, created_by)
		values ($1, $2, $3, $4, $5, $6, $7)
		returning id
	`, k.Nama, k.Prefix, hash, k.UserID, k.ReadOnly, k.ExpiresAt, createdBy).Scan(&id)

	if isUniqueViolation(err) {
		// Two keys colliding on 32 random bytes is not something to explain to
		// a user; asking them to try again is the whole remedy.
		return "", domain.Conflict("Gagal membuat kunci. Coba lagi.")
	}
	return id, err
}

// Revoke is guarded on the key still being live, so a second press reports
// that it was already revoked rather than silently rewriting the timestamp.
func (r *APIKeyRepo) Revoke(ctx context.Context, id, revokedBy string) (bool, error) {
	tag, err := r.db.q(ctx).Exec(ctx, `
		update api_keys set revoked_at = now(), revoked_by = $1
		where id = $2 and revoked_at is null
	`, revokedBy, id)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() == 1, nil
}

func (r *APIKeyRepo) TouchLastUsed(ctx context.Context, id string) error {
	// Written at most once a minute per key: the column exists to answer "is
	// this key still in use", which a coarse timestamp answers just as well as
	// a write on every single request.
	_, err := r.db.q(ctx).Exec(ctx, `
		update api_keys set last_used_at = now()
		where id = $1 and (last_used_at is null or last_used_at < now() - interval '1 minute')
	`, id)
	return err
}

var _ domain.APIKeyRepository = (*APIKeyRepo)(nil)
