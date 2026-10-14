![Looking Glass — see exactly what is hitting your service by KloudKIT](https://raw.githubusercontent.com/kloudkit/looking-glass/main/brand/looking-glass-banner.svg)

# KloudKIT • Looking Glass

[![Stars](https://img.shields.io/github/stars/kloudkit/looking-glass?style=for-the-badge&logo=git&logoColor=c6d0f5&labelColor=414559&color=f0c6c6)](https://github.com/kloudkit/looking-glass/stargazers)
[![Docker](https://img.shields.io/badge/latest-a?style=for-the-badge&logo=docker&logoColor=c6d0f5&label=docker&labelColor=414559&color=ef9f76)](https://github.com/kloudkit/looking-glass/pkgs/container/looking-glass)
[![Build](https://img.shields.io/github/actions/workflow/status/kloudkit/looking-glass/ci.yaml?style=for-the-badge&logo=githubactions&logoColor=c6d0f5&label=build&labelColor=414559&color=a6da95)](https://github.com/kloudkit/looking-glass/actions/workflows/ci.yaml)
[![License](https://img.shields.io/github/license/kloudkit/looking-glass?style=for-the-badge&logo=opensourceinitiative&logoColor=c6d0f5&labelColor=414559&color=8caaee)](https://github.com/kloudkit/looking-glass/blob/main/LICENSE)

When a webhook, a proxy or a client misbehaves, the first question is what
request actually arrived. Answering it usually means adding log lines to the
service, or reaching for a packet capture.

Looking Glass is a drop-in HTTP request reflector. Point anything at it and
every method, on every path, is answered with the full request it received:
method, URL, query, headers, client address and body.

## Features

- **Wildcard everything:** every HTTP method on every path returns `200` with
  the full request reflected back.
- **Two response formats:** `html` (default) or `json`, chosen with the
  `X-Glass-Format` header (never the request's `Accept`).
- **Copy as:** the HTML view copies the request to your clipboard as pretty
  `JSON` or a `Raw` HTTP dump, no devtools required.
- **Colored activity log:** stdout gets one readable line per request (time,
  client, method, path, format, size), so `docker logs` stays a clean access
  stream and request contents never leak into the logs.
- **Safe HTML:** reflected values are escaped, so a malicious body can't execute
  in the browser; the logged path is stripped of control characters.
- **Bounded:** request bodies are capped (default `1 MiB`) so a stray upload
  can't take the box down; truncation is shown in every rendering.
- **Tiny image:** a single static Go binary on a shell-less, non-root distroless
  base.

## Installation

Run the multi-arch (`amd64`, `arm64`) image from GHCR:

```sh
docker run --rm -p 8080:8080 ghcr.io/kloudkit/looking-glass:latest
```

## Quick start

The **HTTP response** is `html` by default and `json` on demand. A dedicated
header selects it, so it never collides with the request's own `Accept`; a
`?format=` query param does the same from a browser address bar:

```sh
# HTML (default)
curl 'localhost:8080/api/users?page=2'

# JSON via header
curl -X POST 'localhost:8080/api/users?page=2' -d '{"name":"ada"}' -H 'X-Glass-Format: json'
```

```text
# JSON in a browser: add ?format=json to the URL
http://localhost:8080/api/users?page=2&format=json
```

The HTML view also carries a **Copy as** control in the corner: grab the request
as pretty `JSON` or as a `Raw` HTTP dump (request line, headers, body) without
leaving the page.

Meanwhile the **container logs stay a concise, colored activity stream**, one
line per request (time, client, method, path, response format, body size), and
never the request contents:

```sh
docker logs -f <container>
# 2026/06/10 12:00:00 10.0.0.1:54321  POST  /api/users?page=2  →  html  14b
```

## Configuration

| Flag             | Variable         | Default   | Description                              |
| ---------------- | ---------------- | --------- | ---------------------------------------- |
| `-p`, `--port`   | `PORT`           | `8080`    | Port to listen on.                       |
| `--max-body`     | `MAX_BODY_BYTES` | `1048576` | Max body bytes reflected before cut-off. |

A flag wins over its variable. `looking-glass --help` lists them, and `--version`
prints the version.

| Selector                | Values          | Effect                                      |
| ----------------------- | --------------- | ------------------------------------------- |
| `X-Glass-Format` header | `html` / `json` | Response format. Takes precedence.          |
| `?format=` query param  | `html` / `json` | Browser-friendly fallback. Defaults `html`. |

## Building from source

Run it straight from source while iterating locally:

```sh
go run .
# in another shell
curl -X DELETE 'localhost:8080/anything/here'
```

Or build the image yourself:

```sh
docker build -t looking-glass .
docker run --rm -p 8080:8080 looking-glass
```

> 💡 The final image is [distroless](https://github.com/GoogleContainerTools/distroless)
> and shell-less. Swap the final stage for `scratch` to shave it further, or for
> `ghcr.io/kloudkit/base-image` if you want a shell to exec into.

## License

Released under the [**MIT License**](https://github.com/kloudkit/looking-glass/blob/main/LICENSE),
**except** the [`brand/`](https://github.com/kloudkit/looking-glass/tree/main/brand) directory.

The Looking Glass and KloudKIT logos, icons and other brand assets are proprietary and
**all rights reserved**; see [`brand/LICENSE`](https://github.com/kloudkit/looking-glass/blob/main/brand/LICENSE).
