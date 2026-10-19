package reflect

import (
	"bytes"
	"encoding/json"
	"html/template"
	"net"
	"strings"
)

var htmlTemplate = template.Must(template.New("reflect").Parse(`<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<title>{{.Method}} {{.URI}}</title>
<link rel="icon" href="data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 32 32'%3E%3Cpath d='M13.60,1.40 A14.80,14.80 0 0 0 13.60,30.60' fill='none' stroke='%23babbf1' stroke-width='2.40' stroke-linecap='round'/%3E%3Cpath d='M18.40,1.40 A14.80,14.80 0 0 1 18.40,30.60' fill='none' stroke='%23e5c890' stroke-width='2.40' stroke-linecap='round'/%3E%3Cpath d='M8.24,10.56 L12.40,16.00 L8.24,21.44' fill='none' stroke='%23babbf1' stroke-width='2.40' stroke-linecap='round' stroke-linejoin='round'/%3E%3Cpath d='M23.76,10.56 L19.60,16.00 L23.76,21.44' fill='none' stroke='%23e5c890' stroke-width='2.40' stroke-linecap='round' stroke-linejoin='round'/%3E%3C/svg%3E">
<style>
  :root { color-scheme: dark; }
  body { background: #303446; color: #c6d0f5; font: 14px/1.5 ui-monospace, SFMono-Regular, Menlo, monospace; margin: 0; padding: 2rem; }
  .brand { display: flex; align-items: center; gap: .55rem; color: #c6d0f5; font-size: 1rem; margin-bottom: 1.5rem; }
  .brand .logo { width: 28px; height: 28px; }
  .brand .l { color: #babbf1; } .brand .y { color: #e5c890; }
  .lg-a, .lg-b { animation: 6s ease-in-out infinite; }
  .lg-a { animation-name: lg-a; } .lg-b { animation-name: lg-b; }
  .lg-cl, .lg-cr { transform-box: fill-box; animation-name: lg-a, lg-cl; animation-duration: 6s, 2s; }
  .lg-cr { animation-name: lg-b, lg-cr; }
  @keyframes lg-a { 0%, 100% { stroke: #babbf1; } 50% { stroke: #e5c890; } }
  @keyframes lg-b { 0%, 100% { stroke: #e5c890; } 50% { stroke: #babbf1; } }
  @keyframes lg-cl { 0%, 100% { transform: translateX(0); } 50% { transform: translateX(22%); } }
  @keyframes lg-cr { 0%, 100% { transform: translateX(0); } 50% { transform: translateX(-22%); } }
  @media (prefers-reduced-motion: reduce) { .brand * { animation: none !important; } }
  h1 { color: #babbf1; font-size: 1.1rem; margin: 0 0 .25rem; word-break: break-all; }
  .meta { color: #737994; margin-bottom: 1.5rem; }
  h2 { color: #e5c890; font-size: .95rem; margin: 1.5rem 0 .5rem; }
  table { border-collapse: collapse; width: 100%; }
  th, td { border: 1px solid #626880; padding: .35rem .6rem; text-align: left; vertical-align: top; }
  th { color: #babbf1; white-space: nowrap; width: 1%; }
  td { color: #c6d0f5; word-break: break-all; }
  pre { background: #292c3c; border: 1px solid #626880; border-radius: 6px; padding: .75rem 1rem; overflow-x: auto; white-space: pre-wrap; word-break: break-all; }
  .truncated { color: #ef9f76; }
  .copy { position: fixed; top: 1rem; right: 1rem; display: flex; align-items: center; gap: .4rem; background: #292c3c; border: 1px solid #626880; border-radius: 6px; padding: .35rem .5rem; }
  .copy span { color: #737994; font-size: .8rem; }
  .copy button { font: inherit; font-size: .8rem; color: #c6d0f5; background: #303446; border: 1px solid #626880; border-radius: 4px; padding: .2rem .55rem; cursor: pointer; }
  .copy button:hover { border-color: #e5c890; color: #e5c890; }
</style>
</head>
<body>
<div class="copy">
  <span>Copy as</span>
  <button type="button" data-copy="copy-json">JSON</button>
  <button type="button" data-copy="copy-flat">Raw</button>
</div>
<pre id="copy-json" hidden>{{.CopyJSON}}</pre>
<pre id="copy-flat" hidden>{{.CopyFlat}}</pre>
<div class="brand"><svg class="logo" viewBox="0 0 32 32" aria-hidden="true"><path d="M13.6,1.4 A14.8,14.8 0 0 0 13.6,30.6" fill="none" stroke="#babbf1" stroke-width="2.4" stroke-linecap="round" class="lg-a"/><path d="M18.4,1.4 A14.8,14.8 0 0 1 18.4,30.6" fill="none" stroke="#e5c890" stroke-width="2.4" stroke-linecap="round" class="lg-b"/><path d="M8.2,10.5 L12.4,16.0 L8.2,21.4" fill="none" stroke="#babbf1" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round" class="lg-a lg-cl"/><path d="M23.7,10.5 L19.6,16.0 L23.7,21.4" fill="none" stroke="#e5c890" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round" class="lg-b lg-cr"/></svg><span>l<span class="l">o</span><span class="y">o</span>king glass</span></div>
<h1>{{.Method}} {{.URI}} <small>{{.Proto}}</small></h1>
<div class="meta">{{.Time}} &middot; {{.Client}}{{if .Proxy}} via {{.Proxy}}{{end}} &middot; {{.Host}}{{if .Server}} &middot; served by {{.Server}}{{end}}</div>
{{if .Query}}
<h2>Query</h2>
<table>{{range $k, $vs := .Query}}{{range $vs}}<tr><th>{{$k}}</th><td>{{.}}</td></tr>{{end}}{{end}}</table>
{{end}}
{{if .Headers}}
<h2>Headers</h2>
<table>{{range $k, $vs := .Headers}}{{range $vs}}<tr><th>{{$k}}</th><td>{{.}}</td></tr>{{end}}{{end}}</table>
{{end}}
{{if .Form}}
<h2>Form</h2>
<table>{{range $k, $vs := .Form}}{{range $vs}}<tr><th>{{$k}}</th><td>{{.}}</td></tr>{{end}}{{end}}</table>
{{end}}
{{if .Files}}
<h2>Files</h2>
<table>{{range $k, $fs := .Files}}{{range $fs}}<tr><th>{{$k}}</th><td>{{.Filename}} &middot; {{.ContentType}} &middot; {{.Bytes}} bytes</td></tr>{{end}}{{end}}</table>
{{end}}
{{if .BodyBytes}}
<h2>Body ({{.BodyBytes}} bytes{{if eq .BodyEncoding "base64"}}, base64{{end}}{{if .Truncated}}<span class="truncated">, truncated</span>{{end}})</h2>
<pre>{{.Pretty}}</pre>
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
		CopyJSON string
		CopyFlat string
		Proxy    string
		Pretty   string
	}{
		reflection: ref,
		CopyJSON:   ref.json(),
		CopyFlat:   ref.flat(),
		Pretty:     ref.Body,
	}

	// The client came through a proxy, so the direct peer is shown too.
	if peer, _, _ := net.SplitHostPort(ref.Remote); peer != ref.Client {
		data.Proxy = ref.Remote
	}

	var pretty bytes.Buffer
	if json.Indent(&pretty, ref.JSON, "", "  ") == nil {
		data.Pretty = pretty.String()
	}

	var b strings.Builder
	htmlTemplate.Execute(&b, data)

	return b.String()
}
