package postgres

import (
	"strconv"
	"strings"
	"testing"
)

// The builder's job is to keep request data out of the SQL text. These tests
// pin that: whatever a caller passes must end up in args, never in the string.

func TestClauseNumbersPlaceholdersInOrder(t *testing.T) {
	c := &clause{}
	c.add("t.season_id = ", "s1")
	c.add("t.status = ", "approved")
	c.add("t.nilai > ", 100)

	want := " where t.season_id = $1 and t.status = $2 and t.nilai > $3"
	if got := c.sql(); got != want {
		t.Errorf("got %q\nwant %q", got, want)
	}
	if len(c.args) != 3 {
		t.Fatalf("got %d args, want 3", len(c.args))
	}
}

func TestClauseIsEmptyWhenNothingWasAdded(t *testing.T) {
	c := &clause{}
	// An unfiltered list must not emit a dangling WHERE.
	if got := c.sql(); got != "" {
		t.Errorf("got %q, want an empty string", got)
	}
	c.addIf("t.status = ", "")
	c.addSearch("   ", "t.nama")
	if got := c.sql(); got != "" {
		t.Errorf("blank filters produced %q", got)
	}
	if len(c.args) != 0 {
		t.Errorf("blank filters bound %d args", len(c.args))
	}
}

// An injection attempt is data. It must survive as one bound argument and
// leave no trace in the statement.
func TestClauseBindsHostileValuesRatherThanInterpolatingThem(t *testing.T) {
	hostile := []string{
		"'; drop table tyfcb_entries; --",
		"approved' or '1'='1",
		"\\'; delete from members where '1'='1",
		"$1) or true --",
	}

	for _, value := range hostile {
		t.Run(value, func(t *testing.T) {
			c := &clause{}
			c.addIf("t.status = ", value)
			c.addSearch(value, "t.nama", "u.full_name")

			sql := c.sql()
			// The rendered SQL is built only from code-side fragments.
			if strings.ContainsAny(sql, "';\\") || strings.Contains(sql, "--") {
				t.Fatalf("hostile input reached the statement: %q", sql)
			}
			if strings.Contains(sql, "drop") || strings.Contains(sql, "delete") {
				t.Fatalf("hostile input reached the statement: %q", sql)
			}

			if len(c.args) != 2 {
				t.Fatalf("got %d args, want 2", len(c.args))
			}
			if c.args[0] != value {
				t.Errorf("the value was altered: %v", c.args[0])
			}
			if c.args[1] != "%"+value+"%" {
				t.Errorf("the search term was altered: %v", c.args[1])
			}
		})
	}
}

func TestAddSearchFansOutAcrossColumnsWithOneArgument(t *testing.T) {
	c := &clause{}
	c.addIf("v.season_id = ", "s1")
	c.addSearch("budi", "v.nama", "v.kontak", "u.full_name")

	want := " where v.season_id = $1 and (v.nama ilike $2 or v.kontak ilike $2 or u.full_name ilike $2)"
	if got := c.sql(); got != want {
		t.Errorf("got %q\nwant %q", got, want)
	}
	// One term, one argument — repeating it per column would renumber wrongly.
	if len(c.args) != 2 {
		t.Errorf("got %d args, want 2", len(c.args))
	}
}

// paginate runs last, so its placeholders must continue the filter's numbering
// rather than collide with it.
func TestPaginateContinuesTheNumbering(t *testing.T) {
	c := &clause{}
	c.add("v.season_id = ", "s1")
	c.add("v.status_hadir::text = ", "hadir_penuh")

	tail, args := c.paginate(25, 50)

	if tail != " limit $3 offset $4" {
		t.Errorf("got %q, want %q", tail, " limit $3 offset $4")
	}
	if len(args) != 4 {
		t.Fatalf("got %d args, want 4", len(args))
	}
	if args[2] != 25 || args[3] != 50 {
		t.Errorf("limit/offset = %v/%v, want 25/50", args[2], args[3])
	}
}

// The count query and the page query share one clause value; building the tail
// must not mutate it, or the count would silently gain two arguments.
func TestPaginateDoesNotMutateTheClause(t *testing.T) {
	c := &clause{}
	c.add("v.season_id = ", "s1")

	first, _ := c.paginate(25, 0)
	second, args := c.paginate(25, 25)

	if first != second {
		t.Errorf("a second call rendered %q, want %q", second, first)
	}
	if len(c.args) != 1 {
		t.Errorf("the clause kept %d args, want 1", len(c.args))
	}
	if len(args) != 3 {
		t.Errorf("the returned list had %d args, want 3", len(args))
	}
}

// Every placeholder in the final statement must have a matching argument.
// A gap here is what turns a filter bug into a wrong-results bug.
func TestEveryPlaceholderHasAnArgument(t *testing.T) {
	c := &clause{}
	c.addIf("t.season_id = ", "s1")
	c.addIf("t.status = ", "approved")
	c.addIf("m.team_id = ", "")
	c.addSearch("andi", "u.full_name")
	c.add("t.tanggal >= ", "2026-01-01")

	tail, args := c.paginate(10, 0)
	sql := c.sql() + tail

	for i := 1; i <= len(args); i++ {
		if !strings.Contains(sql, "$"+strconv.Itoa(i)) {
			t.Errorf("$%d is bound but never used: %q", i, sql)
		}
	}
	if strings.Contains(sql, "$"+strconv.Itoa(len(args)+1)) {
		t.Errorf("the statement refers to $%d with only %d args", len(args)+1, len(args))
	}
}
