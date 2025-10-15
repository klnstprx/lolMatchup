# Repository Guidelines

## Project Structure & Modules

- `main.go`: App entrypoint; starts Gin server and routes.
- `components/`: Templ templates (`*.templ`) and generated `*_templ.go`.
- `handlers/`, `router/`, `middleware/`: HTTP routing, handlers, middleware.
- `client/`, `data/`, `models/`: Riot API client, loaders, domain models.
- `config/`: App config loader; see `config.toml` and `config.toml.example`.
- `static/`: Embedded static assets; `static/embed.go`.
- `lolmatchup_testing/`: Optional integration/bruno artifacts.
- `cache/`: Caching logic and tests.

## Build, Test, and Development

- `go mod download`: Fetch dependencies.
- `templ generate`: Generate Go from `*.templ` files (requires `templ`).
- `make templ`: Runs `templ generate`.
- `make build` or `go build -o lolmatchup.bin`: Build binary.
- `go run .`: Run locally (ensure templates are generated first).
- `air`: Live-reload dev server (uses `.air.toml`).
- `go test ./... -cover`: Run unit tests with coverage.

## Coding Style & Naming

- Formatting: run `gofmt -s -w .`; vet with `go vet ./...`.
- Go naming: exported `CamelCase`, unexported `camelCase`; packages lowercase.
- Templates: source files `*.templ` produce `*_templ.go`. Keep components small and composable.
- Errors: wrap with context; prefer `fmt.Errorf("...: %w", err)`.

## Testing Guidelines

- Framework: standard `testing` package; files end with `*_test.go`.
- Structure: table-driven tests where practical (see `cache/`, `client/`, `components/`).
- Coverage: target critical packages; check with `-cover`.
- Run: `go test ./...` before pushing; add tests for new handlers and clients.

## Commit & Pull Requests

- Commits: short imperative subject; optional Conventional Commits (`feat:`, `fix:`, `refactor:`). Examples: `feat: add live game support`, `fix: cache eviction on update`.
- PRs: include a clear description, linked issues, test plan (`go test` output), and screenshots/GIFs for UI changes.
- Generated code: commit both `*.templ` and `*_templ.go` to keep builds reproducible.

## Security & Configuration

- API keys: do not commit real Riot keys; use `config.toml` or env-injected values locally.
- Defaults: see `config.toml.example`; confirm `port`/`listen_addr` before sharing screenshots or logs.
