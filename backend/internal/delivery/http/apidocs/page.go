package apidocs

import "html/template"

// page is the whole reference in one document: no external stylesheet, no
// script, nothing to fetch. It renders on a laptop with the network off.
var page = template.Must(template.New("apidocs").Parse(`<!doctype html>
<html lang="id">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<meta name="robots" content="noindex, nofollow">
<title>{{.Title}} — Referensi API</title>
<style>
  :root {
    --brand: #c8102e; --brand-dark: #a60b24; --brand-50: #fff1f2;
    --ink: #171923; --muted: #687082; --line: rgba(23,25,35,.10);
    --bg: #fbf8f8; --code: #f6f2f3;
    --get: #0f766e; --post: #1d4ed8; --patch: #b45309; --delete: #b91c1c;
  }
  * { box-sizing: border-box; }
  body {
    margin: 0; background: var(--bg); color: var(--ink);
    font: 15px/1.6 "DM Sans", -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
  }
  code, pre { font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace; }

  .shell { display: grid; grid-template-columns: 260px minmax(0,1fr); max-width: 1180px; margin: 0 auto; }
  @media (max-width: 900px) { .shell { grid-template-columns: 1fr; } nav.side { position: static; height: auto; border: 0; } }

  nav.side {
    position: sticky; top: 0; height: 100dvh; overflow-y: auto;
    padding: 28px 20px; border-right: 1px solid var(--line);
  }
  nav.side h1 { margin: 0; font-size: 1.05rem; font-weight: 800; letter-spacing: -.01em; }
  nav.side .ver { color: var(--muted); font-size: .75rem; }
  nav.side .grp { margin-top: 20px; font-size: .68rem; font-weight: 700; letter-spacing: .12em;
                  text-transform: uppercase; color: var(--muted); }
  nav.side a { display: block; padding: 3px 0; color: var(--ink); text-decoration: none;
               font-size: .82rem; opacity: .8; }
  nav.side a:hover { color: var(--brand); opacity: 1; }

  main { padding: 32px 28px 80px; min-width: 0; }
  header.intro h2 { margin: 0 0 4px; font-size: 1.9rem; font-weight: 800; letter-spacing: -.02em; }
  header.intro .count { color: var(--muted); font-size: .82rem; }
  header.intro h3 { margin: 26px 0 6px; font-size: 1rem; font-weight: 800; }
  header.intro p { margin: 0 0 10px; color: #3b4252; }
  .links { display: flex; flex-wrap: wrap; gap: 8px; margin: 20px 0 8px; }
  .links a {
    display: inline-flex; align-items: center; gap: 6px; padding: 7px 13px;
    border: 1px solid var(--line); border-radius: 999px; background: #fff;
    color: var(--ink); text-decoration: none; font-size: .8rem; font-weight: 700;
  }
  .links a:hover { border-color: var(--brand); color: var(--brand); }

  pre { background: var(--code); padding: 12px 14px; border-radius: 12px; overflow-x: auto;
        font-size: .8rem; line-height: 1.55; }
  :not(pre) > code { background: var(--code); padding: 1px 5px; border-radius: 5px; font-size: .86em; }

  section.tag { margin-top: 44px; }
  section.tag > h3 { margin: 0; font-size: 1.15rem; font-weight: 800; letter-spacing: -.01em; }
  section.tag > .lead { margin: 2px 0 0; color: var(--muted); font-size: .85rem; }

  article.op {
    margin-top: 14px; padding: 16px 18px; background: #fff;
    border: 1px solid var(--line); border-radius: 16px; scroll-margin-top: 16px;
  }
  .row { display: flex; align-items: center; gap: 10px; flex-wrap: wrap; }
  .verb { font-size: .68rem; font-weight: 800; letter-spacing: .06em; padding: 3px 8px;
          border-radius: 6px; color: #fff; }
  .GET { background: var(--get); } .POST { background: var(--post); }
  .PATCH { background: var(--patch); } .DELETE { background: var(--delete); }
  .path { font-family: ui-monospace, monospace; font-size: .86rem; font-weight: 600; word-break: break-all; }
  .open { margin-left: auto; font-size: .66rem; font-weight: 700; color: var(--muted);
          border: 1px solid var(--line); border-radius: 999px; padding: 2px 8px; }
  .summary { margin: 8px 0 0; font-weight: 600; }
  article.op p { margin: 8px 0 0; color: #3b4252; font-size: .9rem; }

  h4 { margin: 16px 0 6px; font-size: .68rem; font-weight: 800; letter-spacing: .12em;
       text-transform: uppercase; color: var(--muted); }
  table { width: 100%; border-collapse: collapse; font-size: .84rem; }
  th { text-align: left; font-weight: 700; color: var(--muted); font-size: .7rem;
       text-transform: uppercase; letter-spacing: .08em; padding: 0 10px 6px 0; }
  td { padding: 5px 10px 5px 0; border-top: 1px solid var(--line); vertical-align: top; }
  td.name { font-family: ui-monospace, monospace; font-weight: 600; white-space: nowrap; }
  .req { color: var(--brand); font-weight: 700; }
  .status { font-family: ui-monospace, monospace; font-weight: 700; }
  .status.bad { color: var(--brand-dark); }
</style>
</head>
<body>
<div class="shell">
  <nav class="side">
    <h1>{{.Title}}</h1>
    <p class="ver">v{{.Version}}</p>
    {{range .Sections}}
      <p class="grp">{{.Name}}</p>
      {{range .Operations}}<a href="#{{.Anchor}}">{{.Method}} {{.Path}}</a>{{end}}
    {{end}}
  </nav>

  <main>
    <header class="intro">
      <h2>Referensi API</h2>
      <p class="count">{{.Total}} endpoint</p>

      <div class="links">
        <a href="openapi.yaml">Unduh OpenAPI (YAML)</a>
        <a href="openapi.json">Unduh OpenAPI (JSON)</a>
      </div>

      {{.Intro}}
    </header>

    {{range .Sections}}
    <section class="tag">
      <h3>{{.Name}}</h3>
      {{if .Description}}<p class="lead">{{.Description}}</p>{{end}}

      {{range .Operations}}
      <article class="op" id="{{.Anchor}}">
        <div class="row">
          <span class="verb {{.Method}}">{{.Method}}</span>
          <span class="path">{{.Path}}</span>
          {{if .Open}}<span class="open">tanpa otentikasi</span>{{end}}
        </div>
        <p class="summary">{{.Summary}}</p>
        {{.Body}}

        {{if .Parameters}}
        <h4>Parameter</h4>
        <table>
          <tr><th>Nama</th><th>Di</th><th>Keterangan</th></tr>
          {{range .Parameters}}
          <tr>
            <td class="name">{{if .Ref}}{{.Ref}}{{else}}{{.Name}}{{if .Required}} <span class="req">*</span>{{end}}{{end}}</td>
            <td>{{.In}}</td>
            <td>{{.Description}}{{if .Schema.Enum}} <code>{{range $i, $e := .Schema.Enum}}{{if $i}} | {{end}}{{$e}}{{end}}</code>{{end}}</td>
          </tr>
          {{end}}
        </table>
        {{end}}

        {{if .HasBody}}<h4>Request body</h4><p style="margin-top:0">JSON. Lihat skema di berkas OpenAPI.</p>{{end}}

        {{if .Responses}}
        <h4>Respons</h4>
        <table>
          {{range .Responses}}
          <tr>
            <td class="name status{{if .Bad}} bad{{end}}">{{.Code}}</td>
            <td>{{.Description}}</td>
          </tr>
          {{end}}
        </table>
        {{end}}
      </article>
      {{end}}
    </section>
    {{end}}
  </main>
</div>
</body>
</html>`))
