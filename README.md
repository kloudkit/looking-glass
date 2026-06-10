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

Then point anything at it — every method, every path is answered with a readable
dump of the request it received:

```sh
curl -X POST 'localhost:8080/api/users?page=2' \
  -H 'Content-Type: application/json' \
  -d '{"name":"ada"}'
```

```text
──────────────────────────────────────────────
POST /api/users?page=2  HTTP/1.1
──────────────────────────────────────────────
Time:    2026-06-10T12:00:00Z
Remote:  10.0.0.1:54321
Host:    localhost:8080

Query:
  page = 2

Headers:
  Accept:       */*
  Content-Type: application/json
  User-Agent:   curl/8.5.0

Body (14 bytes):
  {"name":"ada"}
```

## What's Inside

- **Wildcard everything** — every HTTP method on every path returns `200` with
  the full request reflected back.
- **Human-readable** dump of the method, URL, query params, headers (including
  the `User-Agent`), client address, and body.
- **Logs to stdout** too — the same block lands in `docker logs`.
- **Zero dependencies** — a single static Go binary on a shell-less, non-root
  distroless image.
- **Bounded** — request bodies are capped (default `1 MiB`) so a stray upload
  can't take the box down.

## Configuration

| Variable         | Default   | Description                              |
| ---------------- | --------- | ---------------------------------------- |
| `PORT`           | `8080`    | Port to listen on.                       |
| `MAX_BODY_BYTES` | `1048576` | Max body bytes reflected before cut-off. |

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
