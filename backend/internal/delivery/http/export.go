package http

import (
	"encoding/csv"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-pdf/fpdf"
	"github.com/xuri/excelize/v2"

	"github.com/ikurniawann/brag2026/backend/internal/domain"
	"github.com/ikurniawann/brag2026/backend/internal/usecase"
)

// Rendering lives here rather than in the use case because a format is a
// presentation decision: the same report is a spreadsheet someone will sort
// and total, a PDF someone will print, or a CSV someone will feed to another
// tool. The use case says what the columns mean; this decides how they look.

const exportDateLayout = "02/01/2006"

// Exports are read in Indonesia, so the header stamp is local time with an
// Indonesian month name. Go's own formatting would print a UTC clock and an
// English month, which reads as somebody else's document.
var exportLocation = mustLoadJakarta()

func mustLoadJakarta() *time.Location {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		// A container without tzdata falls back to a fixed offset rather than
		// to UTC: WIB has no daylight saving, so this is exact.
		return time.FixedZone("WIB", 7*60*60)
	}
	return loc
}

var indonesianMonths = [...]string{
	"Januari", "Februari", "Maret", "April", "Mei", "Juni",
	"Juli", "Agustus", "September", "Oktober", "November", "Desember",
}

func stampNow() string {
	t := time.Now().In(exportLocation)
	return fmt.Sprintf("%d %s %d %02d:%02d WIB",
		t.Day(), indonesianMonths[t.Month()-1], t.Year(), t.Hour(), t.Minute())
}

// handleExport serves whichever report the path names. It sits behind
// requireAdmin, so every report is reachable here.
func (s *Server) handleExport(w http.ResponseWriter, r *http.Request) {
	s.export(w, r, chi.URLParam(r, "report"))
}

// handleExportLeaderboard is the members' own route. The name is fixed rather
// than read from the path, so no amount of URL editing turns it into an export
// of the roster or of somebody's transactions.
func (s *Server) handleExportLeaderboard(w http.ResponseWriter, r *http.Request) {
	s.export(w, r, "leaderboard")
}

func (s *Server) export(w http.ResponseWriter, r *http.Request, name string) {
	report, err := s.reports.Build(r.Context(), name, usecase.ReportFilter{
		Status:      r.URL.Query().Get("status"),
		TeamID:      r.URL.Query().Get("team_id"),
		Search:      searchParam(r),
		Converted:   boolParam(r, "converted"),
		ColorStatus: r.URL.Query().Get("color_status"),
		DateFrom:    dateParam(r, "from"),
		DateTo:      dateParam(r, "to"),
	})
	if err != nil {
		fail(w, err)
		return
	}

	format := strings.ToLower(r.URL.Query().Get("format"))
	if format == "" {
		format = "xlsx"
	}

	// The filename carries the date so a folder of exports stays sortable.
	stamp := time.Now().In(exportLocation).Format("2006-01-02")
	filename := fmt.Sprintf("%s-%s.%s", report.Basename, stamp, format)

	switch format {
	case "xlsx":
		writeDownloadHeaders(w, filename,
			"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		err = writeXLSX(w, report)
	case "pdf":
		writeDownloadHeaders(w, filename, "application/pdf")
		err = writePDF(w, report)
	case "csv":
		writeDownloadHeaders(w, filename, "text/csv; charset=utf-8")
		err = writeCSV(w, report)
	default:
		fail(w, domain.Invalid("Format harus xlsx, pdf, atau csv."))
		return
	}

	if err != nil {
		// The status line is already sent, so there is no way to turn this
		// into a clean error response. Log it and let the truncated download
		// fail on the client rather than pretending it succeeded.
		slog.Error("export failed mid-stream", "report", name, "format", format, "err", err)
	}
}

func writeDownloadHeaders(w http.ResponseWriter, filename, contentType string) {
	w.Header().Set("Content-Type", contentType)
	// The quoted form is what every browser agrees on for ASCII names, and
	// every name this builds is ASCII.
	w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	w.Header().Set("X-Content-Type-Options", "nosniff")
}

// ── Excel ─────────────────────────────────────────────────────────────────

func writeXLSX(w io.Writer, report *usecase.Report) error {
	f := excelize.NewFile()
	defer f.Close()

	header, err := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"C8102E"}},
		Alignment: &excelize.Alignment{Vertical: "center", WrapText: true},
	})
	if err != nil {
		return err
	}

	// Numbers stay numbers, with a display format. Writing "Rp 18.000.000" as
	// text would make the column impossible to sum, which is the main reason
	// someone asked for Excel rather than a PDF.
	currency, err := f.NewStyle(&excelize.Style{CustomNumFmt: strptr(`"Rp" #,##0`)})
	if err != nil {
		return err
	}
	date, err := f.NewStyle(&excelize.Style{CustomNumFmt: strptr("dd/mm/yyyy")})
	if err != nil {
		return err
	}

	for i, sheet := range report.Sheets {
		name := sheetName(sheet.Name, i)
		if i == 0 {
			if err := f.SetSheetName("Sheet1", name); err != nil {
				return err
			}
		} else if _, err := f.NewSheet(name); err != nil {
			return err
		}

		for c, col := range sheet.Columns {
			cell, err := excelize.CoordinatesToCellName(c+1, 1)
			if err != nil {
				return err
			}
			if err := f.SetCellStr(name, cell, col.Header); err != nil {
				return err
			}
			if err := f.SetCellStyle(name, cell, cell, header); err != nil {
				return err
			}

			letter, err := excelize.ColumnNumberToName(c + 1)
			if err != nil {
				return err
			}
			if err := f.SetColWidth(name, letter, letter, columnWidth(col)); err != nil {
				return err
			}
		}

		for rowIdx, row := range sheet.Rows {
			for c, value := range row {
				if value == nil {
					continue
				}
				cell, err := excelize.CoordinatesToCellName(c+1, rowIdx+2)
				if err != nil {
					return err
				}
				if err := f.SetCellValue(name, cell, value); err != nil {
					return err
				}

				if c < len(sheet.Columns) {
					switch sheet.Columns[c].Kind {
					case usecase.CellCurrency:
						err = f.SetCellStyle(name, cell, cell, currency)
					case usecase.CellDate:
						err = f.SetCellStyle(name, cell, cell, date)
					}
					if err != nil {
						return err
					}
				}
			}
		}

		// Freeze the header and switch on filters, because the first thing
		// anyone does with an exported table is sort it.
		if err := f.SetPanes(name, &excelize.Panes{
			Freeze: true, Split: false, YSplit: 1,
			TopLeftCell: "A2", ActivePane: "bottomLeft",
		}); err != nil {
			return err
		}
		if len(sheet.Rows) > 0 && len(sheet.Columns) > 0 {
			last, err := excelize.ColumnNumberToName(len(sheet.Columns))
			if err != nil {
				return err
			}
			if err := f.AutoFilter(name, "A1:"+last+"1", nil); err != nil {
				return err
			}
		}
	}

	_, err = f.WriteTo(w)
	return err
}

// Excel rejects some characters in a tab name and caps it at 31 runes.
func sheetName(name string, index int) string {
	cleaned := strings.Map(func(r rune) rune {
		if strings.ContainsRune(`:\/?*[]`, r) {
			return '-'
		}
		return r
	}, name)

	runes := []rune(cleaned)
	if len(runes) > 31 {
		cleaned = string(runes[:31])
	}
	if strings.TrimSpace(cleaned) == "" {
		return "Sheet" + strconv.Itoa(index+1)
	}
	return cleaned
}

func columnWidth(col usecase.ReportColumn) float64 {
	switch col.Kind {
	case usecase.CellCurrency:
		return 18
	case usecase.CellDate:
		return 14
	case usecase.CellNumber:
		return 12
	}
	// Text columns vary a lot; the header length is a decent floor.
	if w := float64(len([]rune(col.Header))) + 6; w > 22 {
		return w
	}
	return 22
}

// ── PDF ───────────────────────────────────────────────────────────────────

func writePDF(w io.Writer, report *usecase.Report) error {
	// Landscape: these tables are wider than they are tall, and portrait would
	// squeeze every column to illegibility.
	pdf := fpdf.New("L", "mm", "A4", "")
	pdf.SetTitle(report.Title, true)
	pdf.SetAutoPageBreak(true, 15)
	d := doc{Fpdf: pdf, tr: pdf.UnicodeTranslatorFromDescriptor("")}

	for _, sheet := range report.Sheets {
		pdf.AddPage()
		writePDFHeading(d, report, sheet.Name)
		writePDFTable(d, sheet)
	}

	if len(report.Truncated) > 0 {
		pdf.SetFont("Helvetica", "I", 8)
		pdf.SetTextColor(150, 40, 40)
		pdf.Ln(4)
		pdf.MultiCell(0, 4, d.tr(fmt.Sprintf(
			"Catatan: %s dipotong pada %d baris pertama.",
			strings.Join(report.Truncated, ", "), usecase.ExportRowCap)), "", "L", false)
	}

	return pdf.Output(w)
}

func writePDFHeading(pdf doc, report *usecase.Report, sheetName string) {
	pdf.SetFont("Helvetica", "B", 15)
	pdf.SetTextColor(200, 16, 46) // BNI red
	pdf.CellFormat(0, 8, pdf.tr(report.Title), "", 1, "L", false, 0, "")

	pdf.SetFont("Helvetica", "", 9)
	pdf.SetTextColor(110, 110, 110)
	subtitle := report.Subtitle
	if sheetName != "" {
		subtitle = sheetName + " · " + subtitle
	}
	pdf.CellFormat(0, 5, pdf.tr(subtitle), "", 1, "L", false, 0, "")
	pdf.CellFormat(0, 5, pdf.tr("Diekspor "+stampNow()), "", 1, "L", false, 0, "")
	pdf.Ln(2)
}

func writePDFTable(pdf doc, sheet usecase.Sheet) {
	if len(sheet.Columns) == 0 {
		return
	}

	// Share the printable width in proportion to how much each kind needs.
	usable, _ := pdf.GetPageSize()
	left, _, right, _ := pdf.GetMargins()
	usable -= left + right

	weights := make([]float64, len(sheet.Columns))
	var total float64
	for i, col := range sheet.Columns {
		weights[i] = pdfColumnWeight(col)
		total += weights[i]
	}
	widths := make([]float64, len(weights))
	for i, weight := range weights {
		widths[i] = usable * weight / total
	}

	drawHeader := func() {
		pdf.SetFont("Helvetica", "B", 7.5)
		pdf.SetFillColor(200, 16, 46)
		pdf.SetTextColor(255, 255, 255)
		for i, col := range sheet.Columns {
			// Headers are clipped like any other cell: an overflowing one
			// spills across the next column and reads as that column's label.
			pdf.CellFormat(widths[i], 7, pdf.tr(clip(pdf, col.Header, widths[i])),
				"1", 0, "L", true, 0, "")
		}
		pdf.Ln(-1)
	}
	drawHeader()

	pdf.SetFont("Helvetica", "", 7.5)
	pdf.SetTextColor(30, 30, 30)

	for rowIdx, row := range sheet.Rows {
		// Repeat the header whenever the table spills onto a new page,
		// otherwise the continuation is a wall of unlabelled numbers.
		if pdf.GetY() > 185 {
			pdf.AddPage()
			drawHeader()
			pdf.SetFont("Helvetica", "", 7.5)
			pdf.SetTextColor(30, 30, 30)
		}

		// Banding, so the eye can follow a row across nine columns.
		fill := rowIdx%2 == 1
		if fill {
			pdf.SetFillColor(248, 244, 245)
		}

		for i := range sheet.Columns {
			var value any
			if i < len(row) {
				value = row[i]
			}
			text := formatCell(value, sheet.Columns[i].Kind)
			pdf.CellFormat(widths[i], 5.5, pdf.tr(clip(pdf, text, widths[i])),
				"1", 0, alignFor(sheet.Columns[i].Kind), fill, 0, "")
		}
		pdf.Ln(-1)
	}

	if len(sheet.Rows) == 0 {
		pdf.SetFont("Helvetica", "I", 8)
		pdf.SetTextColor(130, 130, 130)
		pdf.CellFormat(0, 8, pdf.tr("Tidak ada data untuk filter ini."), "", 1, "L", false, 0, "")
	}
}

// pdfColumnWeight shares the page between columns. The kind sets a base — a
// currency needs more room than a count — but a column never gets less than
// its own header needs, or the label ends up truncated to explain nothing.
func pdfColumnWeight(col usecase.ReportColumn) float64 {
	base := 2.0
	switch col.Kind {
	case usecase.CellCurrency:
		base = 1.4
	case usecase.CellDate:
		base = 1.0
	case usecase.CellNumber:
		base = 0.8
	}

	if header := float64(len([]rune(col.Header))) / 7; header > base {
		return header
	}
	return base
}

func alignFor(kind usecase.CellKind) string {
	switch kind {
	case usecase.CellCurrency, usecase.CellNumber:
		return "R"
	case usecase.CellDate:
		return "C"
	}
	return "L"
}

// clip trims a value that will not fit its column, so a long name pushes an
// ellipsis rather than the rest of the row.
func clip(pdf doc, text string, width float64) string {
	const padding = 2.0
	if pdf.GetStringWidth(text) <= width-padding {
		return text
	}

	runes := []rune(text)
	for len(runes) > 1 {
		runes = runes[:len(runes)-1]
		if pdf.GetStringWidth(string(runes)+"…") <= width-padding {
			return string(runes) + "…"
		}
	}
	return ""
}

// doc pairs the document with its text translator. fpdf's core fonts are
// cp1252, so every string has to go through tr or an unmapped rune comes out
// blank — Indonesian names are Latin, so this is enough without embedding a
// Unicode font.
type doc struct {
	*fpdf.Fpdf
	tr func(string) string
}

// ── CSV ───────────────────────────────────────────────────────────────────

func writeCSV(w io.Writer, report *usecase.Report) error {
	// A BOM, because Excel on Windows reads a plain UTF-8 CSV as Latin-1 and
	// turns every accented name into mojibake.
	if _, err := io.WriteString(w, "\ufeff"); err != nil {
		return err
	}

	out := csv.NewWriter(w)
	defer out.Flush()

	for i, sheet := range report.Sheets {
		// One file, so multiple sheets are stacked with their names as
		// separators rather than silently losing all but the first.
		if len(report.Sheets) > 1 {
			if i > 0 {
				if err := out.Write([]string{}); err != nil {
					return err
				}
			}
			if err := out.Write([]string{sheet.Name}); err != nil {
				return err
			}
		}

		headers := make([]string, len(sheet.Columns))
		for c, col := range sheet.Columns {
			headers[c] = col.Header
		}
		if err := out.Write(headers); err != nil {
			return err
		}

		for _, row := range sheet.Rows {
			record := make([]string, len(sheet.Columns))
			for c := range sheet.Columns {
				if c < len(row) {
					record[c] = formatCell(row[c], sheet.Columns[c].Kind)
				}
			}
			if err := out.Write(record); err != nil {
				return err
			}
		}
	}

	out.Flush()
	return out.Error()
}

// ── shared formatting ─────────────────────────────────────────────────────

// formatCell renders one value for a human-readable format. Excel gets the raw
// value and a number format instead, so this is only used by PDF and CSV.
func formatCell(value any, kind usecase.CellKind) string {
	if value == nil {
		return ""
	}

	switch v := value.(type) {
	case time.Time:
		return v.Format(exportDateLayout)
	case string:
		return v
	case int:
		if kind == usecase.CellCurrency {
			return "Rp " + thousands(strconv.Itoa(v))
		}
		return strconv.Itoa(v)
	case float64:
		if kind == usecase.CellCurrency {
			return "Rp " + thousands(strconv.FormatFloat(v, 'f', 0, 64))
		}
		return strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		if v {
			return "Ya"
		}
		return "Tidak"
	}
	return fmt.Sprint(value)
}

// thousands groups digits with dots, which is how rupiah is written here.
func thousands(digits string) string {
	sign := ""
	if strings.HasPrefix(digits, "-") {
		sign, digits = "-", digits[1:]
	}

	var b strings.Builder
	for i, r := range digits {
		if i > 0 && (len(digits)-i)%3 == 0 {
			b.WriteByte('.')
		}
		b.WriteRune(r)
	}
	return sign + b.String()
}

func strptr(s string) *string { return &s }
