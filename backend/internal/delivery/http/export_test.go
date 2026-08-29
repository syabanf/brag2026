package http

import (
	"bytes"
	"encoding/csv"
	"io"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/xuri/excelize/v2"

	"github.com/ikurniawann/brag2026/backend/internal/usecase"
)

func sampleReport() *usecase.Report {
	return &usecase.Report{
		Title:    "Transaksi TYFCB — BRAG 2026",
		Subtitle: "Filter: status Ditolak",
		Basename: "tyfcb",
		Sheets: []usecase.Sheet{{
			Name: "TYFCB",
			Columns: []usecase.ReportColumn{
				{Header: "Tanggal", Kind: usecase.CellDate},
				{Header: "Pembeli", Kind: usecase.CellText},
				{Header: "Nilai", Kind: usecase.CellCurrency},
				{Header: "Poin", Kind: usecase.CellNumber},
			},
			Rows: [][]any{
				{time.Date(2026, 8, 17, 0, 0, 0, 0, time.UTC), "Budi Santoso", 18_000_000.0, 80},
				{time.Date(2026, 8, 18, 0, 0, 0, 0, time.UTC), "Ani Wijaya", 350_000.0, nil},
			},
		}},
	}
}

// ── Excel ─────────────────────────────────────────────────────────────────

// The whole reason for offering Excel rather than only a PDF is that someone
// wants to sort and total the numbers. Writing "Rp 18.000.000" as text would
// defeat that, so the cell must hold a number.
func TestXLSXKeepsNumbersNumeric(t *testing.T) {
	var buf bytes.Buffer
	if err := writeXLSX(&buf, sampleReport()); err != nil {
		t.Fatalf("write: %v", err)
	}

	f, err := excelize.OpenReader(&buf)
	if err != nil {
		t.Fatalf("the file is not readable as a workbook: %v", err)
	}
	defer f.Close()

	// The stored value is what Excel computes on. A number is written without
	// a type attribute, so the check is that the raw cell holds the amount
	// itself rather than a rendered string like "Rp 18.000.000".
	raw, err := f.GetCellValue("TYFCB", "C2", excelize.Options{RawCellValue: true})
	if err != nil {
		t.Fatalf("cell value: %v", err)
	}
	if _, err := strconv.ParseFloat(raw, 64); err != nil {
		t.Fatalf("the currency cell holds %q, which is not a number", raw)
	}
	if raw != "18000000" {
		t.Errorf("stored %q, want the raw amount", raw)
	}

	// The rupiah presentation comes from a number format, so the displayed
	// value differs from the stored one without costing the arithmetic.
	shown, err := f.GetCellValue("TYFCB", "C2")
	if err != nil {
		t.Fatalf("formatted value: %v", err)
	}
	if !strings.Contains(shown, "Rp") {
		t.Errorf("the cell displays %q, want it formatted as rupiah", shown)
	}
}

func TestXLSXWritesHeadersAndEveryRow(t *testing.T) {
	var buf bytes.Buffer
	if err := writeXLSX(&buf, sampleReport()); err != nil {
		t.Fatalf("write: %v", err)
	}

	f, err := excelize.OpenReader(&buf)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer f.Close()

	rows, err := f.GetRows("TYFCB")
	if err != nil {
		t.Fatalf("rows: %v", err)
	}
	if len(rows) != 3 {
		t.Fatalf("got %d rows, want a header and two entries", len(rows))
	}
	if rows[0][0] != "Tanggal" {
		t.Errorf("first header is %q", rows[0][0])
	}
	if rows[1][1] != "Budi Santoso" {
		t.Errorf("first row reads %q", rows[1][1])
	}
}

// A nil cell is "no value", not zero. Writing 0 would let someone sum a column
// of scores and get a total that includes entries which never scored.
func TestXLSXLeavesAbsentValuesEmpty(t *testing.T) {
	var buf bytes.Buffer
	if err := writeXLSX(&buf, sampleReport()); err != nil {
		t.Fatalf("write: %v", err)
	}

	f, err := excelize.OpenReader(&buf)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer f.Close()

	value, err := f.GetCellValue("TYFCB", "D3")
	if err != nil {
		t.Fatalf("cell: %v", err)
	}
	if value != "" {
		t.Errorf("the absent score rendered as %q", value)
	}
}

func TestXLSXNamesOneSheetPerReportSheet(t *testing.T) {
	report := sampleReport()
	report.Sheets = append(report.Sheets, usecase.Sheet{
		Name:    "Individu",
		Columns: []usecase.ReportColumn{{Header: "Nama", Kind: usecase.CellText}},
		Rows:    [][]any{{"Citra"}},
	})

	var buf bytes.Buffer
	if err := writeXLSX(&buf, report); err != nil {
		t.Fatalf("write: %v", err)
	}

	f, err := excelize.OpenReader(&buf)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer f.Close()

	got := f.GetSheetList()
	if len(got) != 2 || got[0] != "TYFCB" || got[1] != "Individu" {
		t.Errorf("sheets are %v, want [TYFCB Individu]", got)
	}
}

// Excel refuses some characters outright and caps a tab name at 31 runes, so
// a report name has to be made safe rather than passed through.
func TestSheetNameIsMadeSafeForExcel(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"TYFCB", "TYFCB"},
		{"Tim/Individu", "Tim-Individu"},
		{"A[1]:B?", "A-1--B-"},
		{strings.Repeat("x", 40), strings.Repeat("x", 31)},
		{"   ", "Sheet1"},
		{"", "Sheet1"},
	}

	for _, c := range cases {
		if got := sheetName(c.in, 0); got != c.want {
			t.Errorf("sheetName(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// ── PDF ───────────────────────────────────────────────────────────────────

func TestPDFIsWellFormed(t *testing.T) {
	var buf bytes.Buffer
	if err := writePDF(&buf, sampleReport()); err != nil {
		t.Fatalf("write: %v", err)
	}

	out := buf.Bytes()
	if !bytes.HasPrefix(out, []byte("%PDF-")) {
		t.Errorf("the output does not start with a PDF header: %q", firstBytes(out))
	}
	if !bytes.Contains(out, []byte("%%EOF")) {
		t.Error("the document was not closed")
	}
}

func TestPDFHandlesAnEmptyReport(t *testing.T) {
	report := sampleReport()
	report.Sheets[0].Rows = nil

	var buf bytes.Buffer
	if err := writePDF(&buf, report); err != nil {
		t.Fatalf("an empty report should still produce a document: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("nothing was written")
	}
}

// ── CSV ───────────────────────────────────────────────────────────────────

func TestCSVStartsWithABOM(t *testing.T) {
	var buf bytes.Buffer
	if err := writeCSV(&buf, sampleReport()); err != nil {
		t.Fatalf("write: %v", err)
	}
	// Without it, Excel on Windows reads the file as Latin-1 and turns every
	// accented name into mojibake.
	if !bytes.HasPrefix(buf.Bytes(), []byte("\ufeff")) {
		t.Error("the file has no byte order mark")
	}
}

func TestCSVFormatsValuesForReading(t *testing.T) {
	var buf bytes.Buffer
	if err := writeCSV(&buf, sampleReport()); err != nil {
		t.Fatalf("write: %v", err)
	}

	records := readCSV(t, &buf)
	if len(records) != 3 {
		t.Fatalf("got %d records, want a header and two rows", len(records))
	}

	if records[1][0] != "17/08/2026" {
		t.Errorf("date rendered as %q", records[1][0])
	}
	if records[1][2] != "Rp 18.000.000" {
		t.Errorf("currency rendered as %q", records[1][2])
	}
	// An absent score stays blank rather than becoming a zero.
	if records[2][3] != "" {
		t.Errorf("the absent score rendered as %q", records[2][3])
	}
}

// One CSV cannot hold tabs, so the sheets are stacked with their names as
// separators — dropping all but the first would lose half a leaderboard.
func TestCSVStacksMultipleSheets(t *testing.T) {
	report := sampleReport()
	report.Sheets = append(report.Sheets, usecase.Sheet{
		Name:    "Individu",
		Columns: []usecase.ReportColumn{{Header: "Nama", Kind: usecase.CellText}},
		Rows:    [][]any{{"Citra"}},
	})

	var buf bytes.Buffer
	if err := writeCSV(&buf, report); err != nil {
		t.Fatalf("write: %v", err)
	}

	body := buf.String()
	for _, want := range []string{"TYFCB", "Individu", "Citra", "Budi Santoso"} {
		if !strings.Contains(body, want) {
			t.Errorf("the file is missing %q", want)
		}
	}
}

// ── shared formatting ─────────────────────────────────────────────────────

func TestThousandsGroupsRupiah(t *testing.T) {
	cases := map[string]string{
		"0":         "0",
		"1":         "1",
		"999":       "999",
		"1000":      "1.000",
		"18000000":  "18.000.000",
		"780000000": "780.000.000",
		"-5000":     "-5.000",
	}

	for in, want := range cases {
		if got := thousands(in); got != want {
			t.Errorf("thousands(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestFormatCell(t *testing.T) {
	cases := []struct {
		name  string
		value any
		kind  usecase.CellKind
		want  string
	}{
		{"absent", nil, usecase.CellText, ""},
		{"text", "Budi", usecase.CellText, "Budi"},
		{"count", 12, usecase.CellNumber, "12"},
		{"rupiah from int", 250_000, usecase.CellCurrency, "Rp 250.000"},
		{"rupiah from float", 18_000_000.0, usecase.CellCurrency, "Rp 18.000.000"},
		// A multiplier is a plain number and must keep its decimal.
		{"multiplier", 1.5, usecase.CellNumber, "1.5"},
		{"date", time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC), usecase.CellDate, "05/01/2026"},
		{"flag", true, usecase.CellText, "Ya"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := formatCell(c.value, c.kind); got != c.want {
				t.Errorf("got %q, want %q", got, c.want)
			}
		})
	}
}

// The stamp on every export is read in Indonesia, so it carries local time and
// an Indonesian month rather than a UTC clock and an English one.
func TestStampIsLocalAndIndonesian(t *testing.T) {
	got := stampNow()

	if !strings.HasSuffix(got, "WIB") {
		t.Errorf("%q does not name its timezone", got)
	}
	for _, english := range []string{"January", "August", "December"} {
		if strings.Contains(got, english) {
			t.Errorf("%q uses an English month", got)
		}
	}

	month := indonesianMonths[time.Now().In(exportLocation).Month()-1]
	if !strings.Contains(got, month) {
		t.Errorf("%q does not contain %q", got, month)
	}
}

func firstBytes(b []byte) []byte {
	if len(b) > 16 {
		return b[:16]
	}
	return b
}

func readCSV(t *testing.T, r io.Reader) [][]string {
	t.Helper()

	body, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	body = bytes.TrimPrefix(body, []byte("\ufeff"))

	records, err := csv.NewReader(bytes.NewReader(body)).ReadAll()
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	return records
}
