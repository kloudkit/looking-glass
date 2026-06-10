# Kloud • Looking Glass

> 🔎 **Drop-in HTTP request reflector** — see *exactly* what is hitting your service

[![Stars](https://img.shields.io/github/stars/kloudkit/looking-glass?style=for-the-badge&logo=git&logoColor=c6d0f5&labelColor=414559&color=f0c6c6)](https://github.com/kloudkit/looking-glass/stargazers)
[![Docker](https://img.shields.io/badge/latest-a?style=for-the-badge&logo=docker&logoColor=c6d0f5&label=docker&labelColor=414559&color=ef9f76)](https://github.com/kloudkit/looking-glass/pkgs/container/looking-glass)
[![Build](https://img.shields.io/github/actions/workflow/status/kloudkit/looking-glass/build.yaml?style=for-the-badge&logo=githubactions&logoColor=c6d0f5&label=build&labelColor=414559&color=a6da95)](https://github.com/kloudkit/looking-glass/actions/workflows/build.yaml)
[![License](https://img.shields.io/github/license/kloudkit/looking-glass?style=for-the-badge&logo=opensourceinitiative&logoColor=c6d0f5&labelColor=414559&color=8caaee)](https://github.com/kloudkit/looking-glass/blob/main/LICENSE)

---

```sh
# TL;DR  🔎 run the reflector on :8080
docker run --rm -p 8080:8080 ghcr.io/kloudkit/looking-glass:latest
```

Then point anything at it — every method, every path is answered with the full
request it received (method, URL, query, headers, client address, body).

The **HTTP response** is `json` by default and `html` on demand, selected by a
dedicated header so it never collides with the request's own `Accept` — or by a
`?format=` query param so it works straight from a browser address bar:

```sh
# JSON (default)
curl -X POST 'localhost:8080/api/users?page=2' -d '{"name":"ada"}'

# HTML via header
curl 'localhost:8080/api/users?page=2' -H 'X-Glass-Format: html'
```

```text
# HTML in a browser — just add ?format=html to the URL
http://localhost:8080/api/users?page=2&format=html
```

Meanwhile the **container logs stay a concise, colored activity stream** — one
line per request (time, client, method, path, response format, body size), never
the request contents:

```sh
docker logs -f <container>
# 2026/06/10 12:00:00 10.0.0.1:54321  POST  /api/users?page=2  →  json  14b
```

## What's Inside

- **Wildcard everything** — every HTTP method on every path returns `200` with
  the full request reflected back.
- **Two response formats** — `json` (default) or `html`, chosen with the
  `X-Glass-Format` header (never the request's `Accept`).
- **Colored activity log** — stdout gets one readable line per request (time,
  client, method, path, format, size), so `docker logs` stays a clean access
  stream and request contents never leak into the logs.
- **Safe HTML** — reflected values are escaped, so a malicious body can't execute
  in the browser; the logged path is stripped of control characters.
- **Bounded** — request bodies are capped (default `1 MiB`) so a stray upload
  can't take the box down; truncation is shown in every rendering.
- **Tiny image** — a single static Go binary on a shell-less, non-root distroless
  base.

## Configuration

| Variable         | Default   | Description                              |
| ---------------- | --------- | ---------------------------------------- |
| `PORT`           | `8080`    | Port to listen on.                       |
| `MAX_BODY_BYTES` | `1048576` | Max body bytes reflected before cut-off. |

| Selector                | Values          | Effect                                      |
| ----------------------- | --------------- | ------------------------------------------- |
| `X-Glass-Format` header | `json` / `html` | Response format. Takes precedence.          |
| `?format=` query param  | `json` / `html` | Browser-friendly fallback. Defaults `json`. |

## Getting Started

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

Released under the [**MIT License**](https://github.com/kloudkit/looking-glass/blob/main/LICENSE)
