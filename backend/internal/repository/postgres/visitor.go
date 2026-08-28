package postgres

import (
	"context"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/ikurniawann/brag2026/backend/internal/domain"
)

// itoa keeps placeholder building readable in the filter helpers.
func itoa(i int) string { return strconv.Itoa(i) }

type VisitorRepo struct{ db *DB }

func NewVisitorRepo(db *DB) *VisitorRepo { return &VisitorRepo{db: db} }

const visitorSelect = `
	select v.id, v.season_id, v.nama, v.kontak, v.inviter_id, v.tanggal_undang,
	       v.status_hadir::text, v.is_converted, v.is_void, v.tanggal_konversi,
	       v.created_at, u.full_name, t.nama_tim
	from visitors v
	left join members m on m.id = v.inviter_id
	left join app_users u on u.id = m.user_id
	left join teams t on t.id = m.team_id
`

func scanVisitor(row pgx.Row) (*domain.Visitor, error) {
	var v domain.Visitor
	var inviter *string

	err := row.Scan(&v.ID, &v.SeasonID, &v.Nama, &v.Kontak, &v.InviterID, &v.TanggalUndang,
		&v.StatusHadir, &v.IsConverted, &v.IsVoid, &v.TanggalKonversi,
		&v.CreatedAt, &inviter, &v.NamaTim)

	if noRows(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if inviter != nil {
		v.InviterName = *inviter
	}
	return &v, nil
}

func (r *VisitorRepo) FindByID(ctx context.Context, id string) (*domain.Visitor, error) {
	return scanVisitor(r.db.q(ctx).QueryRow(ctx, visitorSelect+` where v.id = $1 limit 1`, id))
}

func (r *VisitorRepo) List(ctx context.Context, f domain.VisitorFilter) ([]domain.Visitor, error) {
	var (
		where []string
		args  []any
	)
	add := func(clause string, value any) {
		args = append(args, value)
		where = append(where, clause+"$"+itoa(len(args)))
	}

	if f.SeasonID != "" {
		add("v.season_id = ", f.SeasonID)
	}
	if f.Status != "" {
		add("v.status_hadir::text = ", f.Status)
	}
	if f.InviterID != "" {
		add("v.inviter_id = ", f.InviterID)
	}
	if f.TeamID != "" {
		add("m.team_id = ", f.TeamID)
	}

	sql := visitorSelect
	if len(where) > 0 {
		sql += " where " + strings.Join(where, " and ")
	}
	sql += " order by v.created_at desc"
	if f.Limit > 0 {
		args = append(args, f.Limit)
		sql += " limit $" + itoa(len(args))
	}

	rows, err := r.db.q(ctx).Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.Visitor{}
	for rows.Next() {
		var v domain.Visitor
		var inviter *string
		if err := rows.Scan(&v.ID, &v.SeasonID, &v.Nama, &v.Kontak, &v.InviterID, &v.TanggalUndang,
			&v.StatusHadir, &v.IsConverted, &v.IsVoid, &v.TanggalKonversi,
			&v.CreatedAt, &inviter, &v.NamaTim); err != nil {
			return nil, err
		}
		if inviter != nil {
			v.InviterName = *inviter
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func (r *VisitorRepo) Create(ctx context.Context, v *domain.Visitor, submittedBy *string) (string, error) {
	var id string
	err := r.db.q(ctx).QueryRow(ctx, `
		insert into visitors (season_id, inviter_id, nama, kontak, tanggal_undang, status_hadir, submitted_by)
		values ($1, $2, $3, $4, $5, $6::visitor_status, $7)
		returning id
	`, v.SeasonID, v.InviterID, v.Nama, v.Kontak, v.TanggalUndang, string(v.StatusHadir), submittedBy).Scan(&id)

	if isUniqueViolation(err) {
		return "", domain.NewError(domain.ErrAlreadyExists, "Kontak ini sudah terdaftar di season ini.")
	}
	return id, err
}

// UpdateStatusGuarded only writes when the row still holds `from`. A losing
// concurrent request gets false rather than awarding the points twice.
func (r *VisitorRepo) UpdateStatusGuarded(ctx context.Context, id string, from, to domain.VisitorStatus) (bool, error) {
	tag, err := r.db.q(ctx).Exec(ctx, `
		update visitors set status_hadir = $1::visitor_status
		where id = $2 and status_hadir = $3::visitor_status
	`, string(to), id, string(from))
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() == 1, nil
}

func (r *VisitorRepo) UpdateConversionGuarded(ctx context.Context, id string, from, to bool) (bool, error) {
	tag, err := r.db.q(ctx).Exec(ctx, `
		update visitors
		set is_converted = $1,
		    tanggal_konversi = case when $1 then current_date else null end
		where id = $2 and is_converted = $3
	`, to, id, from)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() == 1, nil
}

func (r *VisitorRepo) Void(ctx context.Context, id, voidedBy string) error {
	tag, err := r.db.q(ctx).Exec(ctx, `
		update visitors set is_void = true, voided_by = $1, voided_at = now()
		where id = $2 and is_void = false
	`, voidedBy, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.Conflict("Visitor sudah di-void.")
	}
	return nil
}

var _ domain.VisitorRepository = (*VisitorRepo)(nil)
