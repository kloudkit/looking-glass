# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed

- Logs are now a single colored activity line per request (time, client, method,
  path, response format, body size) instead of the full request rendering. The
  request body and headers no longer appear in stdout, and the logged path is
  stripped of control characters — closing a terminal-injection vector where a
  crafted request could emit escape sequences into an operator's terminal.

## [0.1.0] - 2026-06-10

### Added

- Wildcard HTTP reflector answering every method and path with the request it
  received.
- Always-on colored terminal rendering written to stdout (visible via
  `docker logs`), independent of the request.
- Selectable HTTP response format via the `X-Glass-Format` header: `json`
  (default) or `html`, with a `?format=` query-param fallback so HTML can be
  opened directly in a browser.
- HTML rendering with auto-escaped values to prevent reflected content from
  executing in the browser.
- `PORT` and `MAX_BODY_BYTES` environment configuration, with body truncation
  noted in every rendering.
- Multi-arch (`amd64`/`arm64`) image published to GHCR on tag push.

[Unreleased]: https://github.com/kloudkit/looking-glass/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/kloudkit/looking-glass/releases/tag/v0.1.0
