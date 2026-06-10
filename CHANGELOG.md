# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

[Unreleased]: https://github.com/kloudkit/looking-glass/commits/main
