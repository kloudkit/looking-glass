# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- ⚡ HTTP/2 without TLS, for clients and proxies that speak it to a plain backend;
  `proto` then reads `HTTP/2.0`.
- 🏷️ A `server` field with the hostname of the replica that answered, also shown in
  the HTML view.

- 🔗 `glass-status` and `glass-delay` query params, for health checks and browsers
  that can't send `X-Glass-Status` or `X-Glass-Delay`. The header wins when both are
  set.
- 🍪 `--cors-credentials` *(env: `CORS_CREDENTIALS`)*, which lets cross-origin pages
  send cookies and read the reflection.
- 🛡️ `--trusted-proxies` *(env: `TRUSTED_PROXIES`)*, the IPs or CIDRs whose
  `X-Forwarded-For` is believed. It defaults to the private and loopback ranges; set
  it empty to trust none.

### Changed

- 🔒 CORS no longer allows credentials by default. The reflection carries the
  request's cookies, so any site could read the ones scoped to Looking Glass's host.
- 🧭 `client` is the rightmost `X-Forwarded-For` address a trusted proxy vouches for,
  not the first one, and a peer that isn't trusted is the client whatever it
  forwards. A caller can no longer pick the address in the log.
- ⏱️ A request waiting on a delay is answered as soon as Looking Glass shuts down,
  instead of being cut off after the 5-second grace period.
- 📜 The activity line is written once the response is sent, and shows `499` when
  the client hung up during a delay.

### Fixed

- 🔤 A body cut off in the middle of a UTF-8 character drops that character instead
  of being reflected as base64.

## [0.4.0] - 2026-10-19

### Added

- 🚦 `X-Glass-Status` and `X-Glass-Delay` request headers, which answer with another
  status or after a wait, so a proxy's error page, a retry or a timeout can be tested.
  `--max-delay` *(env: `MAX_DELAY`, default `10s`)* caps the wait. A bad value is
  answered `400`.
- 🌍 CORS: every response allows the request's `Origin` with credentials, and a
  preflight is answered `204` with the method and headers it asked for, so a page on
  any origin can read the reflection.
- 🧭 A `client` field with the first `X-Forwarded-For` address, so a request behind a
  proxy shows who sent it. The activity log and the HTML view use it.
- 🧩 `json`, `form` and `files` fields with the body parsed as JSON, as form fields
  *(urlencoded or multipart)*, and as the multipart uploads by name, type and size.
  The HTML view indents a JSON body and lists form fields and files.

### Changed

- 🧱 A body that isn't valid UTF-8 is reflected as base64, flagged by the new
  `bodyEncoding` field, instead of having its invalid bytes replaced with `�`.
- 📏 `bodyBytes` of a cut-off body is its `Content-Length` when the client sent one,
  rather than the size of the part that was kept.
- 📜 The activity log shows the response status, and the client without its port.
- 🔤 **Copy as Raw** lists headers sorted by name instead of in random order.

## [0.3.0] - 2026-10-13

### Added

- 🚩 `--port` and `--max-body` flags, which override `PORT` and `MAX_BODY_BYTES`,
  plus styled `--help` and `--version`.

### Changed

- 🎨 The activity log and startup lines use the logo's colors: lavender for the
  request, yellow for what it is reflected as.
- 🖼️ The HTML view carries the logo and its favicon, in the same colors.
- 🔖 `--version` reports the release tag in published images, and `dev` for any
  other build instead of a Go pseudo-version.

## [0.2.0] - 2026-06-14

### Added

- 📋 A **Copy as** control on the HTML view that copies the reflected request to the
  clipboard as pretty JSON or a raw HTTP dump (request line, headers, body).

### Changed

- 🌐 The default response format is now `html` instead of `json`; request `json`
  explicitly via the `X-Glass-Format: json` header or the `?format=json` query
  param.
- 🔒 Logs are now a single colored activity line per request (time, client, method,
  path, response format, body size) instead of the full request rendering. The
  request body and headers no longer appear in stdout, and the logged path is
  stripped of control characters — closing a terminal-injection vector where a
  crafted request could emit escape sequences into an operator's terminal.

## [0.1.0] - 2026-06-10

### Added

- 🔎 Wildcard HTTP reflector answering every method and path with the request it
  received.
- 🖥️ Always-on colored terminal rendering written to stdout (visible via
  `docker logs`), independent of the request.
- 🔀 Selectable HTTP response format via the `X-Glass-Format` header: `json`
  (default) or `html`, with a `?format=` query-param fallback so HTML can be
  opened directly in a browser.
- 🛡️ HTML rendering with auto-escaped values to prevent reflected content from
  executing in the browser.
- ⚙️ `PORT` and `MAX_BODY_BYTES` environment configuration, with body truncation
  noted in every rendering.
- 🐳 Multi-arch (`amd64`/`arm64`) image published to GHCR on tag push.

[Unreleased]: https://github.com/kloudkit/looking-glass/compare/v0.4.0...HEAD
[0.4.0]: https://github.com/kloudkit/looking-glass/releases/tag/v0.4.0
[0.3.0]: https://github.com/kloudkit/looking-glass/releases/tag/v0.3.0
[0.2.0]: https://github.com/kloudkit/looking-glass/releases/tag/v0.2.0
[0.1.0]: https://github.com/kloudkit/looking-glass/releases/tag/v0.1.0
