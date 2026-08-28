package usecase

import (
	"context"
	"strings"
	"time"

	"github.com/ikurniawann/brag2026/backend/internal/domain"
)

// Catalog covers the committee's master data: teams, business classifications
// and booster events.
type Catalog struct {
	teams    domain.TeamRepository
	classes  domain.ClassificationRepository
	boosters domain.BoosterRepository
	seasons  domain.SeasonRepository
}

func NewCatalog(
	teams domain.TeamRepository,
	classes domain.ClassificationRepository,
	boosters domain.BoosterRepository,
	seasons domain.SeasonRepository,
) *Catalog {
	return &Catalog{teams: teams, classes: classes, boosters: boosters, seasons: seasons}
}

func (u *Catalog) activeSeason(ctx context.Context) (*domain.Season, error) {
	season, err := u.seasons.FindActive(ctx)
	if err != nil {
		return nil, err
	}
	if season == nil {
		return nil, domain.NotFound("Season aktif tidak ditemukan.")
	}
	return season, nil
}

// ── Teams ─────────────────────────────────────────────────────────────────

func (u *Catalog) ListTeams(ctx context.Context) ([]domain.Team, error) {
	season, err := u.activeSeason(ctx)
	if err != nil {
		return nil, err
	}
	return u.teams.ListBySeason(ctx, season.ID)
}

func (u *Catalog) CreateTeam(ctx context.Context, nama string) (string, error) {
	nama = strings.TrimSpace(nama)
	if nama == "" {
		return "", domain.Invalid("Nama tim wajib diisi.")
	}
	season, err := u.activeSeason(ctx)
	if err != nil {
		return "", err
	}
	return u.teams.Create(ctx, season.ID, nama)
}

func (u *Catalog) RenameTeam(ctx context.Context, id, nama string) error {
	nama = strings.TrimSpace(nama)
	if nama == "" {
		return domain.Invalid("Nama tim wajib diisi.")
	}
	return u.teams.Rename(ctx, id, nama)
}

func (u *Catalog) DeleteTeam(ctx context.Context, id string) error {
	return u.teams.Delete(ctx, id)
}

// ── Classifications ───────────────────────────────────────────────────────

func (u *Catalog) ListClassifications(ctx context.Context) ([]domain.Classification, error) {
	return u.classes.List(ctx)
}

func (u *Catalog) CreateClassification(ctx context.Context, nama string) (string, error) {
	nama = strings.TrimSpace(nama)
	if nama == "" {
		return "", domain.Invalid("Nama klasifikasi wajib diisi.")
	}
	return u.classes.Create(ctx, nama)
}

func (u *Catalog) RenameClassification(ctx context.Context, id, nama string) error {
	nama = strings.TrimSpace(nama)
	if nama == "" {
		return domain.Invalid("Nama klasifikasi wajib diisi.")
	}
	return u.classes.Rename(ctx, id, nama)
}

// DeleteClassification refuses while members still reference it, rather than
// silently orphaning their profiles.
func (u *Catalog) DeleteClassification(ctx context.Context, id string) error {
	used, err := u.classes.CountMembers(ctx, id)
	if err != nil {
		return err
	}
	if used > 0 {
		return domain.Conflict("Klasifikasi masih dipakai oleh member.")
	}
	return u.classes.Delete(ctx, id)
}

// ── Boosters ──────────────────────────────────────────────────────────────

func (u *Catalog) ListBoosters(ctx context.Context, activeOnly bool) ([]domain.BoosterEvent, error) {
	season, err := u.activeSeason(ctx)
	if err != nil {
		return nil, err
	}
	return u.boosters.ListBySeason(ctx, season.ID, activeOnly)
}

func (u *Catalog) FindBooster(ctx context.Context, id string) (*domain.BoosterEvent, error) {
	b, err := u.boosters.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if b == nil {
		return nil, domain.NotFound("Booster tidak ditemukan.")
	}
	return b, nil
}

type BoosterInput struct {
	Judul           string
	Deskripsi       *string
	TanggalMulai    time.Time
	TanggalBerakhir time.Time
	Poin            int
	Status          string
}

func (in BoosterInput) validate() error {
	if strings.TrimSpace(in.Judul) == "" {
		return domain.Invalid("Judul booster wajib diisi.")
	}
	if in.TanggalMulai.IsZero() || in.TanggalBerakhir.IsZero() {
		return domain.Invalid("Tanggal mulai dan berakhir wajib diisi.")
	}
	if in.TanggalBerakhir.Before(in.TanggalMulai) {
		return domain.Invalid("Tanggal berakhir tidak boleh sebelum tanggal mulai.")
	}
	if in.Status != "" && in.Status != "aktif" && in.Status != "nonaktif" {
		return domain.Invalid("Status booster tidak valid.")
	}
	return nil
}

func (u *Catalog) CreateBooster(ctx context.Context, in BoosterInput) (string, error) {
	if err := in.validate(); err != nil {
		return "", err
	}
	season, err := u.activeSeason(ctx)
	if err != nil {
		return "", err
	}
	if in.Status == "" {
		in.Status = "aktif"
	}

	return u.boosters.Create(ctx, &domain.BoosterEvent{
		SeasonID:        season.ID,
		Judul:           strings.TrimSpace(in.Judul),
		Deskripsi:       in.Deskripsi,
		TanggalMulai:    in.TanggalMulai,
		TanggalBerakhir: in.TanggalBerakhir,
		Poin:            in.Poin,
		Status:          in.Status,
	})
}

func (u *Catalog) UpdateBooster(ctx context.Context, id string, in BoosterInput) error {
	if err := in.validate(); err != nil {
		return err
	}
	existing, err := u.boosters.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return domain.NotFound("Booster tidak ditemukan.")
	}

	existing.Judul = strings.TrimSpace(in.Judul)
	existing.Deskripsi = in.Deskripsi
	existing.TanggalMulai = in.TanggalMulai
	existing.TanggalBerakhir = in.TanggalBerakhir
	existing.Poin = in.Poin
	if in.Status != "" {
		existing.Status = in.Status
	}

	return u.boosters.Update(ctx, existing)
}

func (u *Catalog) DeleteBooster(ctx context.Context, id string) error {
	return u.boosters.Delete(ctx, id)
}
