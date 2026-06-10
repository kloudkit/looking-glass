package reflect

import (
	"html/template"
	"strings"
)

var htmlTemplate = template.Must(template.New("reflect").Parse(`<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<title>{{.Method}} {{.URI}}</title>
<style>
  :root { color-scheme: dark; }
  body { background: #303446; color: #c6d0f5; font: 14px/1.5 ui-monospace, SFMono-Regular, Menlo, monospace; margin: 0; padding: 2rem; }
  h1 { color: #ca9ee6; font-size: 1.1rem; margin: 0 0 .25rem; word-break: break-all; }
  .meta { color: #737994; margin-bottom: 1.5rem; }
  h2 { color: #8caaee; font-size: .95rem; margin: 1.5rem 0 .5rem; }
  table { border-collapse: collapse; width: 100%; }
  th, td { border: 1px solid #626880; padding: .35rem .6rem; text-align: left; vertical-align: top; }
  th { color: #81c8be; white-space: nowrap; width: 1%; }
  td { color: #c6d0f5; word-break: break-all; }
  pre { background: #292c3c; border: 1px solid #626880; border-radius: 6px; padding: .75rem 1rem; overflow-x: auto; white-space: pre-wrap; word-break: break-all; }
  .truncated { color: #e5c890; }
</style>
</head>
<body>
<h1>{{.Method}} {{.URI}} <small>{{.Proto}}</small></h1>
<div class="meta">{{.Time}} &middot; {{.Remote}} &middot; {{.Host}}</div>
{{if .Query}}
<h2>Query</h2>
<table>{{range $k, $vs := .Query}}{{range $vs}}<tr><th>{{$k}}</th><td>{{.}}</td></tr>{{end}}{{end}}</table>
{{end}}
{{if .Headers}}
<h2>Headers</h2>
<table>{{range $k, $vs := .Headers}}{{range $vs}}<tr><th>{{$k}}</th><td>{{.}}</td></tr>{{end}}{{end}}</table>
{{end}}
{{if .BodyBytes}}
<h2>Body ({{.BodyBytes}} bytes{{if .Truncated}}<span class="truncated">, truncated</span>{{end}})</h2>
<pre>{{.Body}}</pre>
{{end}}
</body>
</html>
`))

func (ref reflection) html() string {
	var b strings.Builder
	htmlTemplate.Execute(&b, ref)

	return b.String()
}
