// Command seed fills a development database with a season already in flight.
//
// It runs seeds/demo.sql for the facts — accounts, teams, the event calendar,
// the prize pool — and then generates the activity by calling the same use
// cases the running app calls. That is the point of it existing as a program
// rather than a longer SQL script: transaction scores, pair penalties, the
// ledger, badges and raffle tickets are all produced by the production rules,
// so the demo cannot drift away from them. An earlier version reimplemented
// those rules in SQL and had already fallen out of step.
//
//	go run ./cmd/seed                 # against DATABASE_URL
//	bash scripts/seed-demo.sh         # against the compose database
//
// Never run this against production: it deletes the season's activity first.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/joho/godotenv"

	"github.com/ikurniawann/brag2026/backend/internal/config"
	"github.com/ikurniawann/brag2026/backend/internal/domain"
	"github.com/ikurniawann/brag2026/backend/internal/repository/postgres"
	"github.com/ikurniawann/brag2026/backend/internal/usecase"
)

// Fixed seed: the same demo every time, so a screenshot taken today still
// matches the app tomorrow.
const randomSeed = 2026

// How much activity to generate. Enough to fill every screen and to make the
// leaderboards non-trivial, without making the seed slow.
const (
	tyfcbEntries    = 180
	visitorCount    = 60
	oneToOneCount   = 40
	membersWithData = 90
	// Teams guaranteed to qualify for the Full Roster bonus.
	fullRosterTeams = 3
	// Members given a deep season, so every badge has something to fire on.
	standoutMembers  = 6
	standoutPartners = 12
	standoutVisitors = 4
)

// Transaction values chosen to land in every band, including two above the
// HIGH_ROLLER threshold of Rp 250 juta.
var nilaiLadder = []float64{
	350_000, 1_200_000, 4_500_000, 18_000_000,
	62_000_000, 140_000_000, 320_000_000, 780_000_000,
}

// Real reasons, because "Bukti transaksi tidak terbaca." on every rejected row
// makes the rejection screen look broken rather than moderated.
var rejectionReasons = []string{
	"Bukti transaksi tidak terbaca.",
	"Nilai tidak sesuai dengan bukti yang dilampirkan.",
	"Transaksi ini sudah dicatat pada entri lain.",
	"Tanggal transaksi di luar periode season.",
	"Penerima bukan anggota aktif di season ini.",
}

func main() {
	sqlPath := flag.String("sql", "", "path to demo.sql (default: seeds/demo.sql next to the module)")
	skipSQL := flag.Bool("skip-sql", false, "generate activity only, leaving existing facts alone")
	flag.Parse()

	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("configuration: %v", err)
	}

	ctx := context.Background()
	db, err := postgres.Connect(ctx, cfg.DatabaseURL, cfg.DBMaxConns)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer db.Close()

	if !*skipSQL {
		path := *sqlPath
		if path == "" {
			path = defaultSQLPath()
		}
		if err := runSQL(ctx, db, path); err != nil {
			log.Fatalf("facts: %v", err)
		}
		fmt.Println("→ facts seeded from", path)
	}

	if err := generate(ctx, db); err != nil {
		log.Fatalf("activity: %v", err)
	}
}

func defaultSQLPath() string {
	// Run from the backend module root in every documented invocation.
	if wd, err := os.Getwd(); err == nil {
		return filepath.Join(wd, "seeds", "demo.sql")
	}
	return "seeds/demo.sql"
}

func runSQL(ctx context.Context, db *postgres.DB, path string) error {
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return db.ExecScript(ctx, string(body))
}

// world is everything the generator needs to look up while it works.
type world struct {
	season  *domain.Season
	adminID string
	members []domain.Member
}

func generate(ctx context.Context, db *postgres.DB) error {
	// The same wiring as cmd/api, minus the parts only HTTP needs.
	seasons := postgres.NewSeasonRepo(db)
	members := postgres.NewMemberRepo(db)
	users := postgres.NewUserRepo(db)
	tyfcbRepo := postgres.NewTyfcbRepo(db)
	visitorRepo := postgres.NewVisitorRepo(db)
	ledger := postgres.NewLedgerRepo(db)
	badgeRepo := postgres.NewBadgeRepo(db)
	events := postgres.NewWeeklyEventRepo(db)
	passRepo := postgres.NewScoringPassRepo(db)
	prizeRepo := postgres.NewPrizeRepo(db)
	sphereRepo := postgres.NewContactSphereRepo(db)
	oneToOneRepo := postgres.NewOneToOneRepo(db)

	badges := usecase.NewBadges(badgeRepo)
	tyfcb := usecase.NewTyfcb(tyfcbRepo, members, ledger, seasons, events, sphereRepo, badges, db)
	visitors := usecase.NewVisitor(visitorRepo, members, ledger, events, badges, db)
	passes := usecase.NewScoringPass(passRepo, ledger, events, oneToOneRepo, seasons, badges, db)
	prizes := usecase.NewPrize(prizeRepo, members, seasons, badges, db)
	network := usecase.NewNetwork(sphereRepo, oneToOneRepo, members, seasons, db)

	w, err := loadWorld(ctx, seasons, members, users)
	if err != nil {
		return err
	}
	fmt.Printf("→ %d members in %s\n", len(w.members), w.season.Nama)

	rng := rand.New(rand.NewSource(randomSeed))

	if err := seedTyfcb(ctx, tyfcb, w, rng); err != nil {
		return fmt.Errorf("tyfcb: %w", err)
	}
	if err := seedVisitors(ctx, visitors, w, rng); err != nil {
		return fmt.Errorf("visitors: %w", err)
	}
	if err := seedStandouts(ctx, tyfcb, visitors, w, rng); err != nil {
		return fmt.Errorf("standouts: %w", err)
	}
	if err := seedOneToOne(ctx, network, w, rng); err != nil {
		return fmt.Errorf("one-to-one: %w", err)
	}
	// Before the passes, not after: the weekly pass windows on these
	// timestamps, so settling first would award a bonus the data no longer
	// supports once the rows moved.
	if err := alignLedgerDates(ctx, db); err != nil {
		return fmt.Errorf("ledger dates: %w", err)
	}
	if err := runPasses(ctx, passes, w); err != nil {
		return fmt.Errorf("passes: %w", err)
	}
	if err := seedRaffle(ctx, prizes, prizeRepo, w); err != nil {
		return fmt.Errorf("raffle: %w", err)
	}
	return awardBadges(ctx, badges, w)
}

func loadWorld(
	ctx context.Context,
	seasons domain.SeasonRepository,
	members domain.MemberRepository,
	users domain.UserRepository,
) (*world, error) {
	season, err := seasons.FindActive(ctx)
	if err != nil {
		return nil, err
	}
	if season == nil {
		return nil, fmt.Errorf("no active season — apply migrations and run the SQL first")
	}

	roster, err := members.ListBySeason(ctx, season.ID)
	if err != nil {
		return nil, err
	}
	if len(roster) < 2 {
		return nil, fmt.Errorf("need at least two members, found %d", len(roster))
	}

	admin, _, err := users.FindByEmail(ctx, "demo.admin@brag2026.id")
	if err != nil {
		return nil, err
	}
	if admin == nil {
		return nil, fmt.Errorf("demo.admin@brag2026.id is missing — run the SQL first")
	}

	// A stable order, so the same members get the same activity on a re-run.
	sortMembersByID(roster)
	return &world{season: season, adminID: admin.ID, members: roster}, nil
}

// seedTyfcb files transactions and then moderates them. Submitting through the
// use case is what gives each entry a real pair ordinal and event multiplier;
// verifying through it is what writes the ledger.
func seedTyfcb(ctx context.Context, uc *usecase.Tyfcb, w *world, rng *rand.Rand) error {
	pool := w.members
	if len(pool) > membersWithData {
		pool = pool[:membersWithData]
	}

	var verified, pending, rejected int

	// Full Roster pays a team only when every one of its active members has
	// scored, which random activity essentially never produces. A few teams
	// are covered deliberately so the bonus has something to fire on — and
	// the rest are left uncovered, because a demo where every team qualifies
	// shows the rule no better than one where none does.
	covered, err := coverTeams(ctx, uc, w, rng, fullRosterTeams)
	if err != nil {
		return err
	}
	verified += covered

	for i := 0; i < tyfcbEntries; i++ {
		seller := pool[i%len(pool)]
		buyer := pool[(i*7+3)%len(pool)]
		if buyer.ID == seller.ID {
			buyer = pool[(i*7+4)%len(pool)]
		}

		entry, err := uc.Submit(ctx, usecase.SubmitTyfcbInput{
			BuyerID:  buyer.ID,
			SellerID: seller.ID,
			Nilai:    nilaiLadder[rng.Intn(len(nilaiLadder))],
			// Spread across the fortnight so the activity feed and the
			// STREAK_MASTER rule both have distinct days to work with.
			Tanggal: time.Now().AddDate(0, 0, -(1 + rng.Intn(13))),
		}, &w.adminID)
		if err != nil {
			return err
		}

		// Roughly 70% verified, 20% pending, 10% rejected — a queue that
		// looks worked-through but still has something to do.
		switch roll := rng.Intn(10); {
		case roll < 7:
			if err := uc.SetStatus(ctx, entry.ID, domain.TyfcbVerified, w.adminID, nil); err != nil {
				return err
			}
			verified++
		case roll < 9:
			pending++
		default:
			reason := rejectionReasons[rng.Intn(len(rejectionReasons))]
			if err := uc.SetStatus(ctx, entry.ID, domain.TyfcbRejected, w.adminID, &reason); err != nil {
				return err
			}
			rejected++
		}
	}

	fmt.Printf("→ %d TYFCB (%d verified, %d pending, %d rejected)\n",
		verified+pending+rejected, verified, pending, rejected)
	return nil
}

// coverTeams gives every active member of the first n teams one verified
// transaction, which is what Full Roster measures.
func coverTeams(ctx context.Context, uc *usecase.Tyfcb, w *world, rng *rand.Rand, n int) (int, error) {
	byTeam := map[string][]domain.Member{}
	var order []string
	for _, m := range w.members {
		if m.TeamID == nil || !m.IsActive {
			continue
		}
		if _, seen := byTeam[*m.TeamID]; !seen {
			order = append(order, *m.TeamID)
		}
		byTeam[*m.TeamID] = append(byTeam[*m.TeamID], m)
	}
	sort.Strings(order)
	if len(order) > n {
		order = order[:n]
	}

	written := 0
	for _, teamID := range order {
		roster := byTeam[teamID]
		for i, seller := range roster {
			// Buy from the next member along, wrapping — so the pair is
			// always two different people and the transactions stay inside
			// the covered set.
			buyer := roster[(i+1)%len(roster)]
			if buyer.ID == seller.ID {
				continue
			}

			entry, err := uc.Submit(ctx, usecase.SubmitTyfcbInput{
				BuyerID:  buyer.ID,
				SellerID: seller.ID,
				Nilai:    nilaiLadder[rng.Intn(len(nilaiLadder))],
				// Inside the running week, since that is the window the
				// Full Roster pass settles.
				Tanggal: daysIntoThisWeek(rng),
			}, &w.adminID)
			if err != nil {
				return written, err
			}
			if err := uc.SetStatus(ctx, entry.ID, domain.TyfcbVerified, w.adminID, nil); err != nil {
				return written, err
			}
			written++
		}
	}
	return written, nil
}

// seedVisitors registers guests and walks some of them up the milestones. The
// increments are cumulative, so going through Update is what makes the ledger
// match what an admin clicking the same buttons would have produced.
func seedVisitors(ctx context.Context, uc *usecase.Visitor, w *world, rng *rand.Rand) error {
	pool := w.members
	if len(pool) > visitorCount {
		pool = pool[:visitorCount]
	}

	var hadir, penuh, converted int

	for i := 0; i < visitorCount; i++ {
		inviter := pool[i%len(pool)]

		visitor, err := uc.Register(ctx, usecase.RegisterVisitorInput{
			Nama:          fmt.Sprintf("Tamu %02d", i+1),
			Kontak:        fmt.Sprintf("08%010d", 81_000_000+i*137),
			TanggalUndang: time.Now().AddDate(0, 0, -(1 + rng.Intn(14))),
			InviterID:     inviter.ID,
		}, &w.adminID)
		if err != nil {
			return err
		}

		// Two in five stay merely registered, which is what the real funnel
		// looks like and keeps the visitor screen honest.
		roll := rng.Intn(5)
		if roll < 2 {
			continue
		}

		status := string(domain.VisitorHadir)
		if roll >= 3 {
			status = string(domain.VisitorHadirPenuh)
		}
		if err := uc.Update(ctx, visitor.ID, usecase.UpdateVisitorInput{StatusHadir: &status}); err != nil {
			return err
		}
		if status == string(domain.VisitorHadirPenuh) {
			penuh++
		} else {
			hadir++
		}

		// One in seven of those who turned up joins.
		if roll == 4 && rng.Intn(7) == 0 {
			yes := true
			if err := uc.Update(ctx, visitor.ID, usecase.UpdateVisitorInput{IsConverted: &yes}); err != nil {
				return err
			}
			converted++
		}
	}

	fmt.Printf("→ %d visitors (%d hadir, %d hadir penuh, %d konversi)\n",
		visitorCount, hadir, penuh, converted)
	return nil
}

// seedStandouts gives a handful of members a deep season. Every league has a
// few, and without them four of the twelve badges have nothing to fire on:
// CONNECTOR and SPREADER count distinct trading partners, HAT_TRICK counts
// visitors who stayed the whole meeting.
func seedStandouts(
	ctx context.Context,
	tyfcb *usecase.Tyfcb,
	visitors *usecase.Visitor,
	w *world,
	rng *rand.Rand,
) error {
	stars := w.members
	if len(stars) > standoutMembers {
		stars = stars[:standoutMembers]
	}

	partners, guests := 0, 0
	for i, star := range stars {
		// Buy from distinct sellers, past the SPREADER threshold of ten.
		for n := 0; n < standoutPartners; n++ {
			seller := w.members[(i*standoutPartners+n+11)%len(w.members)]
			if seller.ID == star.ID {
				continue
			}

			entry, err := tyfcb.Submit(ctx, usecase.SubmitTyfcbInput{
				BuyerID:  star.ID,
				SellerID: seller.ID,
				Nilai:    nilaiLadder[rng.Intn(len(nilaiLadder))],
				Tanggal:  time.Now().AddDate(0, 0, -rng.Intn(13)),
			}, &w.adminID)
			if err != nil {
				return err
			}
			if err := tyfcb.SetStatus(ctx, entry.ID, domain.TyfcbVerified, w.adminID, nil); err != nil {
				return err
			}
			partners++
		}

		// Enough guests staying the full meeting to clear HAT_TRICK.
		for n := 0; n < standoutVisitors; n++ {
			visitor, err := visitors.Register(ctx, usecase.RegisterVisitorInput{
				Nama:          fmt.Sprintf("Tamu Unggulan %02d-%d", i+1, n+1),
				Kontak:        fmt.Sprintf("08%010d", 91_000_000+i*50+n),
				TanggalUndang: time.Now().AddDate(0, 0, -(1 + rng.Intn(13))),
				InviterID:     star.ID,
			}, &w.adminID)
			if err != nil {
				return err
			}

			penuh := string(domain.VisitorHadirPenuh)
			if err := visitors.Update(ctx, visitor.ID, usecase.UpdateVisitorInput{StatusHadir: &penuh}); err != nil {
				return err
			}
			guests++

			// The first guest of each standout joins, so CLOSER is not down
			// to a single lucky member across the whole season.
			if n == 0 {
				yes := true
				if err := visitors.Update(ctx, visitor.ID, usecase.UpdateVisitorInput{IsConverted: &yes}); err != nil {
					return err
				}
			}
		}
	}

	fmt.Printf("→ %d standouts (%d partners, %d guests)\n", len(stars), partners, guests)
	return nil
}

// alignLedgerDates backdates each ledger row to the day of the thing it
// records. Without this every row carries the moment the seed ran, which makes
// the activity feed a wall of identical timestamps and leaves STREAK_MASTER —
// which counts distinct scoring days — unable to fire at all.
//
// This shapes data, it does not decide points: the amounts stay exactly as the
// use cases computed them.
func alignLedgerDates(ctx context.Context, db *postgres.DB) error {
	return db.ExecScript(ctx, `
		update score_ledger sl
		set created_at = te.tanggal + (sl.created_at - date_trunc('day', sl.created_at))
		from tyfcb_entries te
		where te.id::text = sl.sumber_ref and sl.season_id = te.season_id;

		update score_ledger sl
		set created_at = v.tanggal_undang + (sl.created_at - date_trunc('day', sl.created_at))
		from visitors v
		where v.id::text = split_part(sl.sumber_ref, ':', 1)
		  and sl.season_id = v.season_id;
	`)
}

// seedOneToOne gives the ONE_TO_ONE week something to pay out on.
func seedOneToOne(ctx context.Context, uc *usecase.Network, w *world, rng *rand.Rand) error {
	logged := 0
	for i := 0; i < oneToOneCount; i++ {
		a := w.members[i%len(w.members)]
		b := w.members[(i*11+5)%len(w.members)]
		if a.ID == b.ID {
			continue
		}

		catatan := "Diskusi peluang referral."
		_, err := uc.LogOneToOne(ctx, usecase.LogOneToOneInput{
			MemberA: a.ID,
			MemberB: b.ID,
			Tanggal: time.Now().AddDate(0, 0, -rng.Intn(14)),
			Catatan: &catatan,
		}, &w.adminID)
		if err != nil {
			// A duplicate pair on the same day is expected and uninteresting.
			continue
		}
		logged++
	}

	fmt.Printf("→ %d 1-2-1 logs\n", logged)
	return nil
}

// runPasses settles the team bonuses. Running the real pass rather than
// inserting the rows keeps the settlement keys identical, so a later manual
// run correctly reports the week as already settled instead of paying twice.
func runPasses(ctx context.Context, uc *usecase.ScoringPass, w *world) error {
	// The pass windows on the ledger row's created_at, and a fresh seed writes
	// every row today — so the current week is the only one with anything in
	// it to settle.
	week, err := uc.RunWeekly(ctx, time.Now())
	if err != nil {
		return err
	}
	today, err := uc.RunDaily(ctx, time.Now())
	if err != nil {
		return err
	}

	fmt.Printf("→ passes: full roster for %d teams, %d points added\n",
		len(week.FullRoster), week.PointsAdded+today.PointsAdded)
	// STREAK and ONE_TO_ONE only pay during their own weeks, and the demo
	// opens on FOUNDER week, so nothing there is expected to fire.
	return nil
}

// seedRaffle issues tickets from the totals the steps above produced, then
// draws one prize so the demo shows a settled winner alongside open ones.
func seedRaffle(ctx context.Context, uc *usecase.Prize, repo domain.PrizeRepository, w *world) error {
	holders, err := uc.IssueTickets(ctx)
	if err != nil {
		return err
	}

	tickets := 0
	for _, h := range holders {
		tickets += h.Tickets
	}
	fmt.Printf("→ %d raffle tickets across %d members\n", tickets, len(holders))

	pool, err := repo.List(ctx, w.season.ID, "approved")
	if err != nil {
		return err
	}
	for _, prize := range pool {
		if prize.Alokasi != "undian" || prize.PemenangID != nil {
			continue
		}
		won, err := uc.Draw(ctx, prize.ID)
		if err != nil {
			return err
		}
		name := "—"
		if won.PemenangNama != nil {
			name = *won.PemenangNama
		}
		fmt.Printf("→ drew %q → %s\n", won.NamaHadiah, name)
		// One drawn prize is enough to show the state; the rest stay open so
		// the draw button has something to do in the demo.
		break
	}
	return nil
}

// awardBadges runs the real evaluator over the whole roster. The use cases
// above already evaluate the members they touch, so this exists to catch the
// rules that depend on totals rather than on a single action — CENTURION and
// TEAM_PLAYER above all.
func awardBadges(ctx context.Context, uc *usecase.Badges, w *world) error {
	for _, m := range w.members {
		if err := uc.Evaluate(ctx, m.ID, w.season.ID); err != nil {
			return err
		}
	}
	fmt.Printf("→ badges evaluated for %d members\n", len(w.members))
	return nil
}

// daysIntoThisWeek returns a date in the current Monday-to-Sunday week, never
// in the future.
func daysIntoThisWeek(rng *rand.Rand) time.Time {
	now := time.Now()
	elapsed := (int(now.Weekday()) + 6) % 7 // Monday = 0
	return now.AddDate(0, 0, -rng.Intn(elapsed+1))
}

// sortMembersByID fixes the order the generator walks the roster in, so a
// re-run produces the same demo rather than one that depends on how the
// database felt like returning rows.
func sortMembersByID(members []domain.Member) {
	sort.Slice(members, func(a, b int) bool { return members[a].ID < members[b].ID })
}
