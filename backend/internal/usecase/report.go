package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/ikurniawann/brag2026/backend/internal/domain"
)

// A report is a table meant to leave the app: headers, and rows of typed
// cells. There is no formatting here on purpose. A currency belongs in an
// Excel cell as a number the reader can sum and in a PDF as "Rp 18.000.000",
// and both come from this one description — so the delivery layer decides how
// each kind is rendered and the report only says what each column holds.
type CellKind string

const (
	CellText     CellKind = "text"
	CellNumber   CellKind = "number"
	CellCurrency CellKind = "currency"
	CellDate     CellKind = "date"
)

type ReportColumn struct {
	Header string
	Kind   CellKind
}

// Sheet is one table: a tab in a workbook, a section in a document.
type Sheet struct {
	Name    string
	Columns []ReportColumn
	Rows    [][]any
}

type Report struct {
	Title    string
	Subtitle string
	// Basename is the download name without an extension.
	Basename string
	Sheets   []Sheet
	// Truncated names any sheet that hit ExportRowCap, so the reader is told
	// rather than handed a short file that looks complete.
	Truncated []string
}

// ExportRowCap bounds one sheet. A season cannot realistically exceed this,
// but an export is one request holding one result set in memory, so it needs a
// ceiling — and one that is reported rather than silently applied.
const ExportRowCap = 5000

type Reports struct {
	ledger   domain.LedgerRepository
	members  domain.MemberRepository
	tyfcb    domain.TyfcbRepository
	visitors domain.VisitorRepository
	prizes   domain.PrizeRepository
	seasons  domain.SeasonRepository
}

func NewReports(
	ledger domain.LedgerRepository,
	members domain.MemberRepository,
	tyfcb domain.TyfcbRepository,
	visitors domain.VisitorRepository,
	prizes domain.PrizeRepository,
	seasons domain.SeasonRepository,
) *Reports {
	return &Reports{
		ledger: ledger, members: members, tyfcb: tyfcb,
		visitors: visitors, prizes: prizes, seasons: seasons,
	}
}

func (u *Reports) season(ctx context.Context) (*domain.Season, error) {
	s, err := u.seasons.FindActive(ctx)
	if err != nil {
		return nil, err
	}
	if s == nil {
		return nil, domain.NotFound("Season aktif tidak ditemukan.")
	}
	return s, nil
}

// Build dispatches by report name. Keeping the names in one place means the
// route can validate before doing any work, and an unknown name is a 400
// rather than an empty file.
func (u *Reports) Build(ctx context.Context, name string, f ReportFilter) (*Report, error) {
	switch name {
	case "leaderboard":
		return u.leaderboard(ctx)
	case "tyfcb":
		return u.tyfcbReport(ctx, f)
	case "visitors":
		return u.visitorReport(ctx, f)
	case "members":
		return u.memberReport(ctx, f)
	case "prizes":
		return u.prizeReport(ctx)
	default:
		return nil, domain.Invalid("Jenis laporan tidak dikenal.")
	}
}

// ReportFilter carries the same narrowing the list screens use, so "export"
// means "export what I am looking at" rather than "export everything".
type ReportFilter struct {
	Status      string
	TeamID      string
	Search      string
	Converted   *bool
	ColorStatus string
	DateFrom    *time.Time
	DateTo      *time.Time
}

func (u *Reports) leaderboard(ctx context.Context) (*Report, error) {
	season, err := u.season(ctx)
	if err != nil {
		return nil, err
	}

	teams, err := u.ledger.TeamScores(ctx, season.ID)
	if err != nil {
		return nil, err
	}
	members, err := u.ledger.MemberScores(ctx, season.ID, domain.ScoreOverall, ExportRowCap)
	if err != nil {
		return nil, err
	}

	teamRows := make([][]any, 0, len(teams))
	for i, t := range teams {
		teamRows = append(teamRows, []any{
			i + 1, t.NamaTim, t.ScoreOverall, t.ScoreTyfcb, t.ScoreVisitor,
			t.ScoreBonus, t.NilaiTyfcb, t.CountTyfcb, t.CountVisitor,
		})
	}

	memberRows := make([][]any, 0, len(members))
	for i, m := range members {
		memberRows = append(memberRows, []any{
			i + 1, m.FullName, deref(m.NamaTim), m.ScoreOverall,
			m.ScoreTyfcb, m.ScoreVisitor, m.ScoreBonus,
		})
	}

	return &Report{
		Title:    "Leaderboard " + season.Nama,
		Subtitle: "Peringkat tim dan individu per kategori",
		Basename: "leaderboard",
		Sheets: []Sheet{
			{
				Name: "Tim",
				Columns: []ReportColumn{
					{"Peringkat", CellNumber}, {"Tim", CellText},
					{"Overall", CellNumber}, {"TYFCB", CellNumber},
					{"Visitor", CellNumber}, {"Bonus", CellNumber},
					{"Nilai TYFCB", CellCurrency},
					{"Transaksi", CellNumber}, {"Tamu", CellNumber},
				},
				Rows: teamRows,
			},
			{
				Name: "Individu",
				Columns: []ReportColumn{
					{"Peringkat", CellNumber}, {"Nama", CellText}, {"Tim", CellText},
					{"Overall", CellNumber}, {"TYFCB", CellNumber},
					{"Visitor", CellNumber}, {"Bonus", CellNumber},
				},
				Rows: memberRows,
			},
		},
	}, nil
}

func (u *Reports) tyfcbReport(ctx context.Context, f ReportFilter) (*Report, error) {
	season, err := u.season(ctx)
	if err != nil {
		return nil, err
	}

	entries, err := u.tyfcb.List(ctx, domain.TyfcbFilter{
		SeasonID: season.ID,
		Status:   f.Status,
		TeamID:   f.TeamID,
		Search:   f.Search,
		DateFrom: f.DateFrom,
		DateTo:   f.DateTo,
		Limit:    ExportRowCap,
	})
	if err != nil {
		return nil, err
	}

	rows := make([][]any, 0, len(entries))
	for _, e := range entries {
		rows = append(rows, []any{
			e.Tanggal, e.GiverName, e.ReceiverName,
			e.Nilai, statusLabel(string(e.Status)),
			intOrBlank(e.ComputedScore), floatOrBlank(e.EventMultiplierApplied),
			intOrBlank(e.PairOrdinal), deref(e.RejectionReason),
		})
	}

	return &Report{
		Title:    "Transaksi TYFCB — " + season.Nama,
		Subtitle: describeFilter(f),
		Basename: "tyfcb",
		Sheets: []Sheet{{
			Name: "TYFCB",
			Columns: []ReportColumn{
				{"Tanggal", CellDate}, {"Pembeli", CellText}, {"Penjual", CellText},
				{"Nilai", CellCurrency}, {"Status", CellText},
				{"Poin", CellNumber}, {"Pengali", CellNumber},
				{"Pasangan ke-", CellNumber}, {"Alasan Ditolak", CellText},
			},
			Rows: rows,
		}},
		Truncated: capped("TYFCB", len(entries)),
	}, nil
}

func (u *Reports) visitorReport(ctx context.Context, f ReportFilter) (*Report, error) {
	season, err := u.season(ctx)
	if err != nil {
		return nil, err
	}

	visitors, err := u.visitors.List(ctx, domain.VisitorFilter{
		SeasonID:  season.ID,
		Status:    f.Status,
		TeamID:    f.TeamID,
		Search:    f.Search,
		Converted: f.Converted,
		Limit:     ExportRowCap,
	})
	if err != nil {
		return nil, err
	}

	rows := make([][]any, 0, len(visitors))
	for _, v := range visitors {
		rows = append(rows, []any{
			v.TanggalUndang, v.Nama, v.Kontak, v.InviterName, deref(v.NamaTim),
			domain.VisitorStatusLabel(v.StatusHadir), yesNo(v.IsConverted),
			dateOrBlank(v.TanggalKonversi),
		})
	}

	return &Report{
		Title:    "Visitor — " + season.Nama,
		Subtitle: describeFilter(f),
		Basename: "visitor",
		Sheets: []Sheet{{
			Name: "Visitor",
			Columns: []ReportColumn{
				{"Tanggal Undang", CellDate}, {"Nama Tamu", CellText},
				{"Kontak", CellText}, {"Pengundang", CellText}, {"Tim", CellText},
				{"Status Hadir", CellText}, {"Konversi", CellText},
				{"Tanggal Konversi", CellDate},
			},
			Rows: rows,
		}},
		Truncated: capped("Visitor", len(visitors)),
	}, nil
}

func (u *Reports) memberReport(ctx context.Context, f ReportFilter) (*Report, error) {
	season, err := u.season(ctx)
	if err != nil {
		return nil, err
	}

	members, _, err := u.members.ListFiltered(ctx, domain.MemberFilter{
		SeasonID:    season.ID,
		TeamID:      f.TeamID,
		ColorStatus: f.ColorStatus,
		Search:      f.Search,
		Page:        domain.Page{Limit: ExportRowCap},
	})
	if err != nil {
		return nil, err
	}

	// One lookup for every member's running total beats a query per row.
	scores, err := u.ledger.MemberScores(ctx, season.ID, domain.ScoreOverall, ExportRowCap)
	if err != nil {
		return nil, err
	}
	byMember := make(map[string]domain.MemberScore, len(scores))
	for _, s := range scores {
		byMember[s.MemberID] = s
	}

	rows := make([][]any, 0, len(members))
	for _, m := range members {
		s := byMember[m.ID]
		rows = append(rows, []any{
			m.FullName, m.Email, deref(m.NamaTim), deref(m.KlasifikasiNama),
			string(m.Role), string(m.ColorStatus), yesNo(m.IsActive),
			s.ScoreOverall, s.ScoreTyfcb, s.ScoreVisitor,
		})
	}

	return &Report{
		Title:    "Daftar Member — " + season.Nama,
		Subtitle: describeFilter(f),
		Basename: "member",
		Sheets: []Sheet{{
			Name: "Member",
			Columns: []ReportColumn{
				{"Nama", CellText}, {"Email", CellText}, {"Tim", CellText},
				{"Klasifikasi", CellText}, {"Peran", CellText},
				{"Status Warna", CellText}, {"Aktif", CellText},
				{"Overall", CellNumber}, {"TYFCB", CellNumber}, {"Visitor", CellNumber},
			},
			Rows: rows,
		}},
		Truncated: capped("Member", len(members)),
	}, nil
}

func (u *Reports) prizeReport(ctx context.Context) (*Report, error) {
	season, err := u.season(ctx)
	if err != nil {
		return nil, err
	}

	prizes, err := u.prizes.List(ctx, season.ID, "")
	if err != nil {
		return nil, err
	}

	counts, err := u.prizes.TicketCounts(ctx, season.ID)
	if err != nil {
		return nil, err
	}
	roster, err := u.members.ListBySeason(ctx, season.ID)
	if err != nil {
		return nil, err
	}

	prizeRows := make([][]any, 0, len(prizes))
	for _, p := range prizes {
		prizeRows = append(prizeRows, []any{
			p.NamaHadiah, deref(p.Deskripsi), floatOrBlank(p.NilaiEstimasi),
			alokasiLabel(p.Alokasi), p.Status,
			deref(p.DonaturNama), deref(p.PemenangNama),
		})
	}

	ticketRows := make([][]any, 0, len(roster))
	for _, m := range roster {
		if n := counts[m.ID]; n > 0 {
			ticketRows = append(ticketRows, []any{m.FullName, deref(m.NamaTim), n})
		}
	}

	return &Report{
		Title:    "Prize Pool & Undian — " + season.Nama,
		Subtitle: "Daftar hadiah dan jumlah tiket per member",
		Basename: "hadiah",
		Sheets: []Sheet{
			{
				Name: "Hadiah",
				Columns: []ReportColumn{
					{"Hadiah", CellText}, {"Deskripsi", CellText},
					{"Nilai Estimasi", CellCurrency}, {"Alokasi", CellText},
					{"Status", CellText}, {"Donatur", CellText}, {"Pemenang", CellText},
				},
				Rows: prizeRows,
			},
			{
				Name: "Tiket Undian",
				Columns: []ReportColumn{
					{"Nama", CellText}, {"Tim", CellText}, {"Tiket", CellNumber},
				},
				Rows: ticketRows,
			},
		},
	}, nil
}

// ── small helpers ─────────────────────────────────────────────────────────

func capped(sheet string, n int) []string {
	if n >= ExportRowCap {
		return []string{sheet}
	}
	return nil
}

// describeFilter turns the active narrowing into a line under the title, so a
// printed page cannot be mistaken for the whole season.
func describeFilter(f ReportFilter) string {
	var parts []string
	if f.Status != "" {
		parts = append(parts, "status "+statusLabel(f.Status))
	}
	if f.ColorStatus != "" {
		parts = append(parts, "warna "+f.ColorStatus)
	}
	if f.Search != "" {
		parts = append(parts, fmt.Sprintf("pencarian %q", f.Search))
	}
	if f.Converted != nil {
		parts = append(parts, "konversi "+yesNo(*f.Converted))
	}
	if f.DateFrom != nil {
		parts = append(parts, "sejak "+shortDate(*f.DateFrom))
	}
	if f.DateTo != nil {
		parts = append(parts, "sampai "+shortDate(*f.DateTo))
	}

	if len(parts) == 0 {
		return "Semua data"
	}
	out := "Filter: "
	for i, p := range parts {
		if i > 0 {
			out += ", "
		}
		out += p
	}
	return out
}

// shortDate keeps the filter line in Indonesian, matching the rest of the
// document rather than switching language mid-sentence.
func shortDate(t time.Time) string {
	months := [...]string{"Jan", "Feb", "Mar", "Apr", "Mei", "Jun",
		"Jul", "Agu", "Sep", "Okt", "Nov", "Des"}
	return fmt.Sprintf("%d %s %d", t.Day(), months[t.Month()-1], t.Year())
}

func statusLabel(s string) string {
	switch s {
	case "verified":
		return "Verified"
	case "pending":
		return "Pending"
	case "rejected":
		return "Ditolak"
	case "void":
		return "Void"
	}
	return s
}

func alokasiLabel(s string) string {
	if s == "undian" {
		return "Undian"
	}
	return "Kategori"
}

func yesNo(v bool) string {
	if v {
		return "Ya"
	}
	return "Tidak"
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// The blank helpers return nil rather than a zero, so an absent score shows as
// an empty cell instead of a 0 someone might add up.
func intOrBlank(v *int) any {
	if v == nil {
		return nil
	}
	return *v
}

func floatOrBlank(v *float64) any {
	if v == nil {
		return nil
	}
	return *v
}

func dateOrBlank(v *time.Time) any {
	if v == nil {
		return nil
	}
	return *v
}
