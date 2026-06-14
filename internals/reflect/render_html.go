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
  .copy { position: fixed; top: 1rem; right: 1rem; display: flex; align-items: center; gap: .4rem; background: #292c3c; border: 1px solid #626880; border-radius: 6px; padding: .35rem .5rem; }
  .copy span { color: #737994; font-size: .8rem; }
  .copy button { font: inherit; font-size: .8rem; color: #c6d0f5; background: #303446; border: 1px solid #626880; border-radius: 4px; padding: .2rem .55rem; cursor: pointer; }
  .copy button:hover { border-color: #8caaee; color: #8caaee; }
</style>
</head>
<body>
<div class="copy">
  <span>Copy as</span>
  <button type="button" data-copy="copy-json">JSON</button>
  <button type="button" data-copy="copy-flat">Raw</button>
</div>
<pre id="copy-json" hidden>{{.JSON}}</pre>
<pre id="copy-flat" hidden>{{.Flat}}</pre>
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
<script>
document.querySelectorAll('.copy button').forEach(function (btn) {
  btn.addEventListener('click', function () {
    var text = document.getElementById(btn.dataset.copy).textContent;
    navigator.clipboard.writeText(text).then(function () {
      var label = btn.textContent;
      btn.textContent = 'Copied';
      setTimeout(function () { btn.textContent = label; }, 1000);
    });
  });
});
</script>
</body>
</html>
`))

func (ref reflection) html() string {
	data := struct {
		reflection
		JSON string
		Flat string
	}{
		reflection: ref,
		JSON:       ref.json(),
		Flat:       ref.flat(),
	}

	var b strings.Builder
	htmlTemplate.Execute(&b, data)

	return b.String()
}
