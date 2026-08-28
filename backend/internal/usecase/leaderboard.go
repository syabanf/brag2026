package usecase

import (
	"context"

	"github.com/ikurniawann/brag2026/backend/internal/domain"
)

type Leaderboard struct {
	ledger   domain.LedgerRepository
	members  domain.MemberRepository
	tyfcb    domain.TyfcbRepository
	visitors domain.VisitorRepository
	boosters domain.BoosterRepository
	badges   domain.BadgeRepository
	activity domain.ActivityRepository
	seasons  domain.SeasonRepository
}

func NewLeaderboard(
	ledger domain.LedgerRepository,
	members domain.MemberRepository,
	tyfcb domain.TyfcbRepository,
	visitors domain.VisitorRepository,
	boosters domain.BoosterRepository,
	badges domain.BadgeRepository,
	activity domain.ActivityRepository,
	seasons domain.SeasonRepository,
) *Leaderboard {
	return &Leaderboard{
		ledger: ledger, members: members, tyfcb: tyfcb,
		visitors: visitors, boosters: boosters, badges: badges,
		activity: activity, seasons: seasons,
	}
}

func (u *Leaderboard) season(ctx context.Context) (*domain.Season, error) {
	s, err := u.seasons.FindActive(ctx)
	if err != nil {
		return nil, err
	}
	if s == nil {
		return nil, domain.NotFound("Season aktif tidak ditemukan.")
	}
	return s, nil
}

// Standings is the public leaderboard: every team, ranked.
func (u *Leaderboard) Standings(ctx context.Context) ([]domain.TeamScore, error) {
	season, err := u.season(ctx)
	if err != nil {
		return nil, err
	}
	return u.ledger.TeamScores(ctx, season.ID)
}

func (u *Leaderboard) IndividualStandings(ctx context.Context, limit int) ([]domain.MemberScore, error) {
	season, err := u.season(ctx)
	if err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 50
	}
	return u.ledger.MemberScores(ctx, season.ID, limit)
}

// TeamHistory backs the drill-down dialog on a leaderboard row. kategori is
// optional and narrows to "tyfcb" or "visitor".
func (u *Leaderboard) TeamHistory(ctx context.Context, teamID, kategori string) ([]domain.LedgerEntry, error) {
	season, err := u.season(ctx)
	if err != nil {
		return nil, err
	}
	return u.ledger.TeamHistory(ctx, teamID, season.ID, kategori)
}

// Dashboard is the member home screen: season-wide totals, the caller's own
// team, their recent activity and the boosters running right now.
type Dashboard struct {
	Season         *domain.Season        `json:"season"`
	Member         *domain.Member        `json:"member"`
	MemberScore    *domain.MemberScore   `json:"member_score"`
	Teams          []domain.TeamScore    `json:"teams"`
	MyTeam         *domain.TeamScore     `json:"my_team"`
	ActiveBoosters []domain.BoosterEvent `json:"active_boosters"`
	RecentTyfcb    []domain.TyfcbEntry   `json:"recent_tyfcb"`
	Badges         []domain.Badge        `json:"badges"`
	TotalTyfcbTx   int                   `json:"total_tyfcb_tx"`
	TotalTyfcbIDR  float64               `json:"total_tyfcb_idr"`
	TotalVisitor   int                   `json:"total_visitor"`
}

func (u *Leaderboard) Dashboard(ctx context.Context, userID string) (*Dashboard, error) {
	season, err := u.season(ctx)
	if err != nil {
		return nil, err
	}

	member, err := u.members.FindByUserAndSeason(ctx, userID, season.ID)
	if err != nil {
		return nil, err
	}

	teams, err := u.ledger.TeamScores(ctx, season.ID)
	if err != nil {
		return nil, err
	}

	boosters, err := u.boosters.ListBySeason(ctx, season.ID, true)
	if err != nil {
		return nil, err
	}

	out := &Dashboard{
		Season:         season,
		Member:         member,
		Teams:          teams,
		ActiveBoosters: boosters,
	}

	// Season-wide totals are derived from the standings already fetched, so
	// the dashboard costs no extra aggregate queries.
	for i := range teams {
		out.TotalTyfcbTx += teams[i].CountTyfcb
		out.TotalTyfcbIDR += teams[i].NilaiTyfcb
		out.TotalVisitor += teams[i].CountVisitor
	}

	if member == nil {
		return out, nil
	}

	if member.TeamID != nil {
		for i := range teams {
			if teams[i].TeamID == *member.TeamID {
				out.MyTeam = &teams[i]
				break
			}
		}
	}

	if out.MemberScore, err = u.ledger.MemberScore(ctx, member.ID, season.ID); err != nil {
		return nil, err
	}

	if out.RecentTyfcb, err = u.tyfcb.List(ctx, domain.TyfcbFilter{
		SeasonID:   season.ID,
		ReceiverID: member.ID,
		Limit:      5,
	}); err != nil {
		return nil, err
	}

	if out.Badges, err = u.badges.ListForMember(ctx, member.ID); err != nil {
		return nil, err
	}

	return out, nil
}

func (u *Leaderboard) Badges(ctx context.Context) ([]domain.Badge, error) {
	return u.badges.List(ctx)
}

// Activity backs both the feed page and the notification bell; the bell simply
// asks for fewer rows.
func (u *Leaderboard) Activity(ctx context.Context, limit int) ([]domain.ActivityItem, error) {
	season, err := u.season(ctx)
	if err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	return u.activity.Recent(ctx, season.ID, limit)
}
