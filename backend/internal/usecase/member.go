package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/ikurniawann/brag2026/backend/internal/domain"
)

type Member struct {
	members domain.MemberRepository
	users   domain.UserRepository
	teams   domain.TeamRepository
	classes domain.ClassificationRepository
	seasons domain.SeasonRepository
	ledger  domain.LedgerRepository
	badges  *Badges
	tx      domain.TxManager
}

func NewMember(
	members domain.MemberRepository,
	users domain.UserRepository,
	teams domain.TeamRepository,
	classes domain.ClassificationRepository,
	seasons domain.SeasonRepository,
	ledger domain.LedgerRepository,
	badges *Badges,
	tx domain.TxManager,
) *Member {
	return &Member{
		members: members, users: users, teams: teams, classes: classes,
		seasons: seasons, ledger: ledger, badges: badges, tx: tx,
	}
}

// Profile is a user's member record for the active season, or nil when the
// committee has not enrolled them yet.
func (u *Member) Profile(ctx context.Context, userID string) (*domain.Member, error) {
	season, err := u.seasons.FindActive(ctx)
	if err != nil {
		return nil, err
	}
	if season == nil {
		return nil, domain.NotFound("Season aktif tidak ditemukan.")
	}
	return u.members.FindByUserAndSeason(ctx, userID, season.ID)
}

func (u *Member) List(ctx context.Context) ([]domain.Member, error) {
	season, err := u.seasons.FindActive(ctx)
	if err != nil {
		return nil, err
	}
	if season == nil {
		return nil, domain.NotFound("Season aktif tidak ditemukan.")
	}
	return u.members.ListBySeason(ctx, season.ID)
}

func (u *Member) ListByTeam(ctx context.Context, teamID string) ([]domain.Member, error) {
	return u.members.ListByTeam(ctx, teamID)
}

// Search backs the member picker on the submission form. It stays quiet until
// the term is long enough to be selective.
func (u *Member) Search(ctx context.Context, term string) ([]domain.Member, error) {
	term = strings.TrimSpace(term)
	if len(term) < 3 {
		return []domain.Member{}, nil
	}

	season, err := u.seasons.FindActive(ctx)
	if err != nil {
		return nil, err
	}
	if season == nil {
		return []domain.Member{}, nil
	}
	return u.members.Search(ctx, season.ID, term, 20)
}

type CreateMemberInput struct {
	FullName      string
	Email         string
	Password      string
	TeamID        *string
	KlasifikasiID *string
	ColorStatus   string
	Role          string
}

// Create provisions the login and the competition profile together, so a
// half-created member cannot exist.
func (u *Member) Create(ctx context.Context, in CreateMemberInput) (string, error) {
	in.FullName = strings.TrimSpace(in.FullName)
	in.Email = strings.TrimSpace(in.Email)

	if in.FullName == "" || in.Email == "" || in.Password == "" {
		return "", domain.Invalid("Nama, email, dan kata sandi wajib diisi.")
	}
	if len(in.Password) < 6 {
		return "", domain.Invalid("Kata sandi minimal 6 karakter.")
	}
	if in.ColorStatus == "" {
		in.ColorStatus = string(domain.ColorMerah)
	}
	if !domain.ValidColorStatus(in.ColorStatus) {
		return "", domain.Invalid("Status warna tidak valid.")
	}
	role := domain.Role(in.Role)
	if role == "" {
		role = domain.RoleMember
	}

	season, err := u.seasons.FindActive(ctx)
	if err != nil {
		return "", err
	}
	if season == nil {
		return "", domain.NotFound("Season aktif tidak ditemukan.")
	}

	existing, _, err := u.users.FindByEmail(ctx, in.Email)
	if err != nil {
		return "", err
	}
	if existing != nil {
		return "", domain.NewError(domain.ErrAlreadyExists, "Email sudah terdaftar.")
	}

	hash, err := HashPassword(in.Password)
	if err != nil {
		return "", err
	}

	var memberID string
	err = u.tx.WithinTx(ctx, func(ctx context.Context) error {
		userID, err := u.users.Create(ctx, in.Email, hash, in.FullName, role)
		if err != nil {
			return err
		}

		memberID, err = u.members.Create(ctx, &domain.Member{
			UserID:        userID,
			SeasonID:      season.ID,
			TeamID:        in.TeamID,
			KlasifikasiID: in.KlasifikasiID,
			ColorStatus:   domain.ColorStatus(in.ColorStatus),
			IsActive:      true,
		})
		return err
	})

	return memberID, err
}

type UpdateMemberInput struct {
	FullName      *string
	Email         *string
	NewPassword   *string
	TeamID        *string
	ClearTeam     bool
	KlasifikasiID *string
	ClearKlas     bool
	ColorStatus   *string
	IsActive      *bool
	Role          *string
}

func (u *Member) Update(ctx context.Context, id string, in UpdateMemberInput, auth *Auth) error {
	member, err := u.members.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if member == nil {
		return domain.NotFound("Member tidak ditemukan.")
	}

	if in.ColorStatus != nil && !domain.ValidColorStatus(*in.ColorStatus) {
		return domain.Invalid("Status warna tidak valid.")
	}

	previousColor := member.ColorStatus
	if in.ColorStatus != nil {
		member.ColorStatus = domain.ColorStatus(*in.ColorStatus)
	}
	if in.IsActive != nil {
		member.IsActive = *in.IsActive
	}
	if in.ClearTeam {
		member.TeamID = nil
	} else if in.TeamID != nil {
		member.TeamID = in.TeamID
	}
	if in.ClearKlas {
		member.KlasifikasiID = nil
	} else if in.KlasifikasiID != nil {
		member.KlasifikasiID = in.KlasifikasiID
	}

	err = u.tx.WithinTx(ctx, func(ctx context.Context) error {
		if err := u.members.Update(ctx, member); err != nil {
			return err
		}

		// Raising a member's colour rewards their whole team, per the PRD.
		// Only upward steps pay; a correction downwards is worth nothing
		// rather than negative, since the team never chose it.
		if bonus := domain.LevelUpBonus(previousColor, member.ColorStatus); bonus > 0 {
			ref := fmt.Sprintf("level_up:%s:%s", member.ID, member.ColorStatus)
			keterangan := fmt.Sprintf("Naik level: %s → %s", previousColor, member.ColorStatus)

			if err := u.ledger.Append(ctx, &domain.LedgerEntry{
				SeasonID:   member.SeasonID,
				TeamID:     member.TeamID,
				Kategori:   domain.CategoryBonus,
				Points:     bonus,
				SumberRef:  &ref,
				Keterangan: &keterangan,
			}); err != nil {
				return err
			}
		}

		if in.FullName != nil || in.Email != nil {
			name, email := member.FullName, member.Email
			if in.FullName != nil {
				name = strings.TrimSpace(*in.FullName)
			}
			if in.Email != nil {
				email = strings.TrimSpace(*in.Email)
			}
			if name == "" || email == "" {
				return domain.Invalid("Nama dan email tidak boleh kosong.")
			}
			if err := u.users.UpdateProfile(ctx, member.UserID, name, email); err != nil {
				return err
			}
		}

		if in.Role != nil {
			role := domain.Role(*in.Role)
			if role != domain.RoleMember && role != domain.RoleCaptain && role != domain.RoleAdmin {
				return domain.Invalid("Role tidak valid.")
			}
			if err := u.users.UpdateRole(ctx, member.UserID, role); err != nil {
				return err
			}
		}

		if in.NewPassword != nil && *in.NewPassword != "" {
			if err := auth.SetPassword(ctx, member.UserID, *in.NewPassword); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	u.badges.EvaluateQuietly(ctx, member.ID, member.SeasonID)
	return nil
}

// SetPasswordFor lets a captain reset a password, but only for someone on
// their own team.
func (u *Member) SetPasswordFor(ctx context.Context, actorTeamID *string, memberID, password string, auth *Auth) error {
	member, err := u.members.FindByID(ctx, memberID)
	if err != nil {
		return err
	}
	if member == nil {
		return domain.NotFound("Member tidak ditemukan.")
	}
	if actorTeamID != nil {
		if member.TeamID == nil || *member.TeamID != *actorTeamID {
			return domain.Forbidden("Member ini bukan anggota tim Anda.")
		}
	}
	return auth.SetPassword(ctx, member.UserID, password)
}
