package postgres

import (
	"strconv"
	"strings"
)

// clause accumulates predicates and their arguments together, so a filter is
// built without hand-numbering placeholders — the source of most off-by-one
// bugs in dynamic SQL, and the reason those bugs turn into wrong results
// rather than errors.
type clause struct {
	where []string
	args  []any
}

// add appends a predicate. The fragment must end where its placeholder goes:
//
//	c.add("te.season_id = ", seasonID)   →  te.season_id = $1
func (c *clause) add(fragment string, value any) {
	c.args = append(c.args, value)
	c.where = append(c.where, fragment+"$"+strconv.Itoa(len(c.args)))
}

// addIf skips empty strings, which is what an unset query parameter looks like.
func (c *clause) addIf(fragment, value string) {
	if value != "" {
		c.add(fragment, value)
	}
}

// addSearch matches one term against several columns at once. The term is
// still a bound parameter — only the column list is interpolated, and that
// comes from the code rather than the request.
func (c *clause) addSearch(term string, columns ...string) {
	term = strings.TrimSpace(term)
	if term == "" || len(columns) == 0 {
		return
	}

	c.args = append(c.args, "%"+term+"%")
	placeholder := "$" + strconv.Itoa(len(c.args))

	parts := make([]string, 0, len(columns))
	for _, col := range columns {
		parts = append(parts, col+" ilike "+placeholder)
	}
	c.where = append(c.where, "("+strings.Join(parts, " or ")+")")
}

// sql renders the WHERE clause, or an empty string when nothing was added.
func (c *clause) sql() string {
	if len(c.where) == 0 {
		return ""
	}
	return " where " + strings.Join(c.where, " and ")
}

// paginate appends LIMIT/OFFSET and returns the full argument list. It is the
// last thing called, so its placeholders always follow the filter's.
func (c *clause) paginate(limit, offset int) (string, []any) {
	args := append(append([]any{}, c.args...), limit, offset)
	return " limit $" + strconv.Itoa(len(args)-1) + " offset $" + strconv.Itoa(len(args)), args
}
