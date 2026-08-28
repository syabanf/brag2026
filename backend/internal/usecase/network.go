package usecase

import (
	"context"
	"strings"
	"time"

	"github.com/ikurniawann/brag2026/backend/internal/domain"
)

// Network covers the relationship data behind two weekly events: contact
// spheres (which trades refer each other) and one-to-one meeting logs.
type Network struct {
	spheres  domain.ContactSphereRepository
	oneToOne domain.OneToOneRepository
	members  domain.MemberRepository
	seasons  domain.SeasonRepository
	tx       domain.TxManager
}

func NewNetwork(
	spheres domain.ContactSphereRepository,
	oneToOne domain.OneToOneRepository,
	members domain.MemberRepository,
	seasons domain.SeasonRepository,
	tx domain.TxManager,
) *Network {
	return &Network{spheres: spheres, oneToOne: oneToOne, members: members, seasons: seasons, tx: tx}
}

func (u *Network) season(ctx context.Context) (*domain.Season, error) {
	s, err := u.seasons.FindActive(ctx)
	if err != nil {
		return nil, err
	}
	if s == nil {
		return nil, domain.NotFound("Season aktif tidak ditemukan.")
	}
	return s, nil
}

// ── Contact spheres ───────────────────────────────────────────────────────

func (u *Network) ListSpheres(ctx context.Context) ([]domain.ContactSphere, error) {
	season, err := u.season(ctx)
	if err != nil {
		return nil, err
	}
	return u.spheres.ListBySeason(ctx, season.ID)
}

func (u *Network) CreateSphere(ctx context.Context, nama string, deskripsi *string, klasifikasiIDs []string) (string, error) {
	nama = strings.TrimSpace(nama)
	if nama == "" {
		return "", domain.Invalid("Nama contact sphere wajib diisi.")
	}

	season, err := u.season(ctx)
	if err != nil {
		return "", err
	}

	var id string
	err = u.tx.WithinTx(ctx, func(ctx context.Context) error {
		id, err = u.spheres.Create(ctx, season.ID, nama, deskripsi)
		if err != nil {
			return err
		}
		return u.spheres.SetMembers(ctx, id, klasifikasiIDs)
	})

	return id, err
}

func (u *Network) SetSphereMembers(ctx context.Context, sphereID string, klasifikasiIDs []string) error {
	return u.spheres.SetMembers(ctx, sphereID, klasifikasiIDs)
}

func (u *Network) DeleteSphere(ctx context.Context, id string) error {
	return u.spheres.Delete(ctx, id)
}

// ── One-to-one logs ───────────────────────────────────────────────────────

func (u *Network) ListOneToOne(ctx context.Context, memberID string, limit int) ([]domain.OneToOne, error) {
	season, err := u.season(ctx)
	if err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	return u.oneToOne.List(ctx, season.ID, memberID, limit)
}

type LogOneToOneInput struct {
	MemberA string
	MemberB string
	Tanggal time.Time
	Catatan *string
}

// LogOneToOne records a meeting. It carries no points on its own — the payoff
// only lands if the pair also closes business during a ONE_TO_ONE week, which
// the weekly pass settles.
func (u *Network) LogOneToOne(ctx context.Context, in LogOneToOneInput, submittedBy *string) (string, error) {
	if in.MemberA == "" || in.MemberB == "" || in.Tanggal.IsZero() {
		return "", domain.Invalid("Kedua member dan tanggal wajib diisi.")
	}
	if in.MemberA == in.MemberB {
		return "", domain.Invalid("Tidak bisa mencatat 1-2-1 dengan diri sendiri.")
	}

	season, err := u.season(ctx)
	if err != nil {
		return "", err
	}

	for _, id := range []string{in.MemberA, in.MemberB} {
		member, err := u.members.FindByID(ctx, id)
		if err != nil {
			return "", err
		}
		if member == nil || member.SeasonID != season.ID {
			return "", domain.Invalid("Member tidak ditemukan di season ini.")
		}
	}

	return u.oneToOne.Create(ctx, &domain.OneToOne{
		SeasonID: season.ID,
		MemberA:  in.MemberA,
		MemberB:  in.MemberB,
		Tanggal:  in.Tanggal,
		Catatan:  in.Catatan,
	}, submittedBy)
}

func (u *Network) DeleteOneToOne(ctx context.Context, id string) error {
	return u.oneToOne.Delete(ctx, id)
}
