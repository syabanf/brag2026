// Package apidocs serves the OpenAPI description of this API, and renders it
// as a reference page.
//
// The page is built here rather than handed to a CDN-hosted viewer so it works
// on a laptop with no internet, loads in one request, and looks like the rest
// of the product. The specification is the single source: nothing on the page
// is written twice.
package apidocs

import (
	_ "embed"
	"encoding/json"
	"html/template"
	"net/http"
	"sort"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

//go:embed openapi.yaml
var specYAML []byte

// Spec is the parsed document, loaded once. A parse failure is a programming
// error — the file ships inside the binary — so it panics rather than
// degrading a route that can never work.
type Spec struct {
	Info struct {
		Title       string `yaml:"title"`
		Version     string `yaml:"version"`
		Description string `yaml:"description"`
	} `yaml:"info"`
	Tags []struct {
		Name        string `yaml:"name"`
		Description string `yaml:"description"`
	} `yaml:"tags"`
	Paths      map[string]map[string]Operation `yaml:"paths"`
	Components struct {
		Parameters map[string]Parameter `yaml:"parameters"`
		Responses  map[string]Response  `yaml:"responses"`
	} `yaml:"components"`
}

// refName turns "#/components/parameters/PathID" into "PathID". A $ref that
// points anywhere else returns empty, so the caller falls back to what it has.
func refName(ref, section string) string {
	prefix := "#/components/" + section + "/"
	if !strings.HasPrefix(ref, prefix) {
		return ""
	}
	return strings.TrimPrefix(ref, prefix)
}

type Operation struct {
	Tags        []string    `yaml:"tags"`
	Summary     string      `yaml:"summary"`
	Description string      `yaml:"description"`
	Parameters  []Parameter `yaml:"parameters"`
	RequestBody *struct {
		Required bool `yaml:"required"`
	} `yaml:"requestBody"`
	Responses map[string]Response `yaml:"responses"`
	// Security absent means the operation inherits the document default, which
	// requires a credential. Present but empty is the only thing that means
	// open — writing this check the other way round labelled every admin route
	// as needing no authentication.
	Security *[]map[string][]string `yaml:"security"`
}

// Open reports whether the operation overrides the document security with an
// empty list.
func (o Operation) Open() bool {
	return o.Security != nil && len(*o.Security) == 0
}

type Response struct {
	Description string `yaml:"description"`
	Ref         string `yaml:"$ref"`
}

type Parameter struct {
	Name        string `yaml:"name"`
	In          string `yaml:"in"`
	Required    bool   `yaml:"required"`
	Description string `yaml:"description"`
	Ref         string `yaml:"$ref"`
	Schema      struct {
		Type    string   `yaml:"type"`
		Enum    []string `yaml:"enum"`
		Default any      `yaml:"default"`
	} `yaml:"schema"`
}

var (
	once   sync.Once
	loaded Spec
	asJSON []byte
)

// Load parses the embedded specification. Safe to call repeatedly.
func Load() Spec {
	once.Do(func() {
		if err := yaml.Unmarshal(specYAML, &loaded); err != nil {
			panic("apidocs: openapi.yaml does not parse: " + err.Error())
		}

		// The JSON form is what Postman and code generators import, so it is
		// produced from the same bytes rather than maintained separately.
		var generic any
		if err := yaml.Unmarshal(specYAML, &generic); err != nil {
			panic("apidocs: " + err.Error())
		}
		body, err := json.Marshal(normalise(generic))
		if err != nil {
			panic("apidocs: " + err.Error())
		}
		asJSON = body
	})
	return loaded
}

// normalise turns YAML's map[any]any into something encoding/json accepts.
func normalise(v any) any {
	switch t := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(t))
		for k, val := range t {
			out[k] = normalise(val)
		}
		return out
	case map[any]any:
		out := make(map[string]any, len(t))
		for k, val := range t {
			out[toString(k)] = normalise(val)
		}
		return out
	case []any:
		for i := range t {
			t[i] = normalise(t[i])
		}
		return t
	}
	return v
}

func toString(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	body, _ := json.Marshal(v)
	return strings.Trim(string(body), `"`)
}

// Handler returns the three routes: the YAML, the JSON, and the page.
func Handler() http.Handler {
	Load()

	mux := http.NewServeMux()

	mux.HandleFunc("/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
		_, _ = w.Write(specYAML)
	})

	mux.HandleFunc("/openapi.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_, _ = w.Write(asJSON)
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := page.Execute(w, view()); err != nil {
			// Headers are already out; nothing useful is left to say.
			return
		}
	})

	return mux
}

// ── the view the template renders ─────────────────────────────────────────

type Section struct {
	Name        string
	Description string
	Operations  []RenderedOp
}

type RenderedOp struct {
	Method     string
	Path       string
	Summary    string
	Body       template.HTML
	Anchor     string
	Open       bool
	Parameters []Parameter
	HasBody    bool
	Responses  []RenderedResponse
}

type RenderedResponse struct {
	Code        string
	Description string
	Bad         bool
}

type pageView struct {
	Title    string
	Version  string
	Intro    template.HTML
	Sections []Section
	Total    int
}

var methodOrder = map[string]int{"get": 0, "post": 1, "patch": 2, "put": 3, "delete": 4}

func view() pageView {
	spec := Load()

	order := map[string]int{}
	byTag := map[string][]RenderedOp{}
	descriptions := map[string]string{}
	for i, t := range spec.Tags {
		order[t.Name] = i
		descriptions[t.Name] = t.Description
	}

	total := 0
	for path, methods := range spec.Paths {
		for method, op := range methods {
			total++

			rendered := RenderedOp{
				Method:     strings.ToUpper(method),
				Path:       path,
				Summary:    op.Summary,
				Body:       markdown(op.Description),
				Anchor:     anchor(method, path),
				Open:       op.Open(),
				Parameters: resolveParams(spec, op.Parameters),
				HasBody:    op.RequestBody != nil,
			}

			codes := make([]string, 0, len(op.Responses))
			for code := range op.Responses {
				codes = append(codes, code)
			}
			sort.Strings(codes)
			for _, code := range codes {
				rendered.Responses = append(rendered.Responses, RenderedResponse{
					Code:        code,
					Description: resolveResponse(spec, op.Responses[code]),
					Bad:         strings.HasPrefix(code, "4") || strings.HasPrefix(code, "5"),
				})
			}

			tag := "Lainnya"
			if len(op.Tags) > 0 {
				tag = op.Tags[0]
			}
			byTag[tag] = append(byTag[tag], rendered)
		}
	}

	sections := make([]Section, 0, len(byTag))
	for name, ops := range byTag {
		sort.Slice(ops, func(a, b int) bool {
			if ops[a].Path != ops[b].Path {
				return ops[a].Path < ops[b].Path
			}
			return methodOrder[strings.ToLower(ops[a].Method)] < methodOrder[strings.ToLower(ops[b].Method)]
		})
		sections = append(sections, Section{
			Name:        name,
			Description: descriptions[name],
			Operations:  ops,
		})
	}
	sort.Slice(sections, func(a, b int) bool {
		ia, oka := order[sections[a].Name]
		ib, okb := order[sections[b].Name]
		if oka != okb {
			return oka
		}
		if !oka {
			return sections[a].Name < sections[b].Name
		}
		return ia < ib
	})

	return pageView{
		Title:    spec.Info.Title,
		Version:  spec.Info.Version,
		Intro:    markdown(spec.Info.Description),
		Sections: sections,
		Total:    total,
	}
}

// resolveParams replaces each $ref with the parameter it names, so the table
// shows "id" rather than the path to where "id" is defined.
func resolveParams(spec Spec, params []Parameter) []Parameter {
	out := make([]Parameter, 0, len(params))
	for _, p := range params {
		if name := refName(p.Ref, "parameters"); name != "" {
			if shared, ok := spec.Components.Parameters[name]; ok {
				out = append(out, shared)
				continue
			}
		}
		out = append(out, p)
	}
	return out
}

func resolveResponse(spec Spec, r Response) string {
	if name := refName(r.Ref, "responses"); name != "" {
		if shared, ok := spec.Components.Responses[name]; ok {
			return shared.Description
		}
	}
	return r.Description
}

func anchor(method, path string) string {
	slug := strings.NewReplacer("/", "-", "{", "", "}", "", ".", "-").Replace(path)
	return strings.ToLower(method) + strings.ToLower(slug)
}

// markdown covers the subset the specification actually uses: headings,
// fenced code, paragraphs, and inline code. A full parser would be a
// dependency earning its keep on four constructs.
func markdown(src string) template.HTML {
	if strings.TrimSpace(src) == "" {
		return ""
	}

	var out strings.Builder
	var para []string

	flush := func() {
		if len(para) == 0 {
			return
		}
		out.WriteString("<p>" + inline(strings.Join(para, " ")) + "</p>")
		para = nil
	}

	lines := strings.Split(src, "\n")
	for i := 0; i < len(lines); i++ {
		line := strings.TrimRight(lines[i], " ")

		switch {
		case strings.HasPrefix(line, "```"):
			flush()
			var code []string
			for i++; i < len(lines) && !strings.HasPrefix(lines[i], "```"); i++ {
				code = append(code, lines[i])
			}
			out.WriteString("<pre><code>" +
				template.HTMLEscapeString(strings.Join(code, "\n")) + "</code></pre>")

		case strings.HasPrefix(line, "## "):
			flush()
			out.WriteString("<h3>" + inline(strings.TrimPrefix(line, "## ")) + "</h3>")

		case strings.TrimSpace(line) == "":
			flush()

		default:
			para = append(para, strings.TrimSpace(line))
		}
	}
	flush()

	return template.HTML(out.String())
}

// inline handles `code`, **bold** and *italic*, escaping everything else.
// Order matters: ** is consumed before * so a bold run is not mistaken for
// two italics.
func inline(s string) string {
	escaped := template.HTMLEscapeString(s)

	for _, r := range []struct{ mark, open, close string }{
		{"**", "<strong>", "</strong>"},
		{"`", "<code>", "</code>"},
		{"*", "<em>", "</em>"},
	} {
		parts := strings.Split(escaped, r.mark)
		if len(parts) < 3 {
			continue
		}
		var b strings.Builder
		for i, part := range parts {
			if i%2 == 1 {
				b.WriteString(r.open + part + r.close)
			} else {
				b.WriteString(part)
			}
		}
		escaped = b.String()
	}
	return escaped
}
