// Package postgres implements the domain repository contracts on PostgreSQL.
// It is the only package that knows SQL.
package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ikurniawann/brag2026/backend/internal/domain"
)

// DB owns the pool and doubles as the TxManager.
type DB struct {
	pool *pgxpool.Pool
}

func Connect(ctx context.Context, dsn string, maxConns int32) (*DB, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}
	if maxConns < 1 {
		maxConns = 1
	}
	cfg.MaxConns = maxConns
	// A few warm connections so the first requests after a deploy do not each
	// pay for a handshake.
	cfg.MinConns = min(4, maxConns)
	cfg.MaxConnLifetime = time.Hour
	cfg.MaxConnIdleTime = 5 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping: %w", err)
	}

	return &DB{pool: pool}, nil
}

func (db *DB) Close() { db.pool.Close() }

func (db *DB) Ping(ctx context.Context) error { return db.pool.Ping(ctx) }

// txKey carries an open transaction on the context so repositories can join it
// without every method growing a tx parameter.
type txKey struct{}

// querier is the subset shared by the pool and a transaction.
type querier interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

// q returns the ambient transaction when there is one, else the pool.
func (db *DB) q(ctx context.Context) querier {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return tx
	}
	return db.pool
}

// WithinTx runs fn inside a transaction, rolling back on any error or panic.
// Nested calls join the outer transaction rather than opening a second one.
func (db *DB) WithinTx(ctx context.Context, fn func(context.Context) error) error {
	if _, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return fn(ctx)
	}

	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(context.WithoutCancel(ctx))
			panic(p)
		}
	}()

	if err := fn(context.WithValue(ctx, txKey{}, tx)); err != nil {
		_ = tx.Rollback(context.WithoutCancel(ctx))
		return err
	}

	return tx.Commit(ctx)
}

// isUniqueViolation reports a duplicate-key error so callers can turn it into
// a friendly conflict rather than a 500.
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

// noRows normalises pgx's sentinel into "nothing found", letting callers
// return a nil entity instead of an error.
func noRows(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}

var _ domain.TxManager = (*DB)(nil)
