package postgres

import (
	"context"
	"time"

	"github.com/ikurniawann/brag2026/backend/internal/domain"
)

type ContactSphereRepo struct{ db *DB }

func NewContactSphereRepo(db *DB) *ContactSphereRepo { return &ContactSphereRepo{db: db} }

func (r *ContactSphereRepo) ListBySeason(ctx context.Context, seasonID string) ([]domain.ContactSphere, error) {
	rows, err := r.db.q(ctx).Query(ctx, `
		select s.id, s.season_id, s.nama, s.deskripsi,
		       coalesce(json_agg(json_build_object('id', c.id, 'nama', c.nama)
		                order by c.nama) filter (where c.id is not null), '[]')
		from contact_spheres s
		left join contact_sphere_members m on m.sphere_id = s.id
		left join classifications c on c.id = m.klasifikasi_id
		where s.season_id = $1
		group by s.id, s.season_id, s.nama, s.deskripsi
		order by s.nama
	`, seasonID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.ContactSphere{}
	for rows.Next() {
		var s domain.ContactSphere
		if err := rows.Scan(&s.ID, &s.SeasonID, &s.Nama, &s.Deskripsi, &s.Klasifikasi); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *ContactSphereRepo) Create(ctx context.Context, seasonID, nama string, deskripsi *string) (string, error) {
	var id string
	err := r.db.q(ctx).QueryRow(ctx, `
		insert into contact_spheres (season_id, nama, deskripsi) values ($1, $2, $3) returning id
	`, seasonID, nama, deskripsi).Scan(&id)

	if isUniqueViolation(err) {
		return "", domain.NewError(domain.ErrAlreadyExists, "Contact sphere dengan nama itu sudah ada.")
	}
	return id, err
}

func (r *ContactSphereRepo) Delete(ctx context.Context, id string) error {
	tag, err := r.db.q(ctx).Exec(ctx, `delete from contact_spheres where id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.NotFound("Contact sphere tidak ditemukan.")
	}
	return nil
}

// SetMembers replaces the whole membership, so the caller sends the desired
// end state rather than a diff.
func (r *ContactSphereRepo) SetMembers(ctx context.Context, sphereID string, klasifikasiIDs []string) error {
	if _, err := r.db.q(ctx).Exec(ctx,
		`delete from contact_sphere_members where sphere_id = $1`, sphereID); err != nil {
		return err
	}

	for _, id := range klasifikasiIDs {
		if _, err := r.db.q(ctx).Exec(ctx, `
			insert into contact_sphere_members (sphere_id, klasifikasi_id) values ($1, $2)
			on conflict do nothing
		`, sphereID, id); err != nil {
			return err
		}
	}
	return nil
}

// SharesSphere is false when either side has no classification: an unclassified
// member cannot be in anyone's power team.
func (r *ContactSphereRepo) SharesSphere(ctx context.Context, seasonID string, a, b *string) (bool, error) {
	if a == nil || b == nil {
		return false, nil
	}

	var shares bool
	err := r.db.q(ctx).QueryRow(ctx, `
		select exists (
			select 1
			from contact_sphere_members ma
			join contact_sphere_members mb on mb.sphere_id = ma.sphere_id
			join contact_spheres s on s.id = ma.sphere_id
			where s.season_id = $1 and ma.klasifikasi_id = $2 and mb.klasifikasi_id = $3
		)
	`, seasonID, *a, *b).Scan(&shares)

	return shares, err
}

type OneToOneRepo struct{ db *DB }

func NewOneToOneRepo(db *DB) *OneToOneRepo { return &OneToOneRepo{db: db} }

const oneToOneSelect = `
	select o.id, o.season_id, o.member_a, o.member_b, ua.full_name, ub.full_name,
	       o.tanggal, o.catatan, o.created_at
	from one_to_one_logs o
	join members ma on ma.id = o.member_a
	join app_users ua on ua.id = ma.user_id
	join members mb on mb.id = o.member_b
	join app_users ub on ub.id = mb.user_id
`

func (r *OneToOneRepo) List(ctx context.Context, seasonID, memberID string, limit int) ([]domain.OneToOne, error) {
	sql := oneToOneSelect + ` where o.season_id = $1`
	args := []any{seasonID}

	if memberID != "" {
		args = append(args, memberID)
		sql += ` and (o.member_a = $2 or o.member_b = $2)`
	}
	args = append(args, limit)
	sql += ` order by o.tanggal desc, o.created_at desc limit $` + itoa(len(args))

	rows, err := r.db.q(ctx).Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.OneToOne{}
	for rows.Next() {
		var o domain.OneToOne
		if err := rows.Scan(&o.ID, &o.SeasonID, &o.MemberA, &o.MemberB,
			&o.MemberAName, &o.MemberBName, &o.Tanggal, &o.Catatan, &o.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	return out, rows.Err()
}

func (r *OneToOneRepo) Create(ctx context.Context, o *domain.OneToOne, submittedBy *string) (string, error) {
	// The table stores the pair in a canonical order, so normalise here rather
	// than trusting the caller.
	a, b := o.MemberA, o.MemberB
	if a > b {
		a, b = b, a
	}

	var id string
	err := r.db.q(ctx).QueryRow(ctx, `
		insert into one_to_one_logs (season_id, member_a, member_b, tanggal, catatan, submitted_by)
		values ($1, $2, $3, $4, $5, $6)
		returning id
	`, o.SeasonID, a, b, o.Tanggal, o.Catatan, submittedBy).Scan(&id)

	if isUniqueViolation(err) {
		return "", domain.NewError(domain.ErrAlreadyExists,
			"1-2-1 dengan member ini pada tanggal tersebut sudah dicatat.")
	}
	return id, err
}

func (r *OneToOneRepo) Delete(ctx context.Context, id string) error {
	tag, err := r.db.q(ctx).Exec(ctx, `delete from one_to_one_logs where id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.NotFound("Catatan 1-2-1 tidak ditemukan.")
	}
	return nil
}

// PairsWithTyfcbInWindow finds the meetings that actually turned into business
// in the same window — the payoff the ONE_TO_ONE event rewards. Direction does
// not matter, so the TYFCB is matched on either orientation of the pair.
func (r *OneToOneRepo) PairsWithTyfcbInWindow(ctx context.Context, seasonID string, from, to time.Time) ([][2]string, error) {
	rows, err := r.db.q(ctx).Query(ctx, `
		select distinct o.member_a::text, o.member_b::text
		from one_to_one_logs o
		where o.season_id = $1
		  and o.tanggal between $2 and $3
		  and exists (
			select 1 from tyfcb_entries te
			where te.season_id = $1
			  and te.status = 'verified'
			  and te.tanggal between $2 and $3
			  and ((te.giver_id = o.member_a and te.receiver_id = o.member_b)
			    or (te.giver_id = o.member_b and te.receiver_id = o.member_a))
		  )
	`, seasonID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := [][2]string{}
	for rows.Next() {
		var pair [2]string
		if err := rows.Scan(&pair[0], &pair[1]); err != nil {
			return nil, err
		}
		out = append(out, pair)
	}
	return out, rows.Err()
}

var (
	_ domain.ContactSphereRepository = (*ContactSphereRepo)(nil)
	_ domain.OneToOneRepository      = (*OneToOneRepo)(nil)
)
