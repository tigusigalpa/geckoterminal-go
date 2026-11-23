# Contributing

Thanks for your interest in contributing to geckoterminal-go.

## Requirements

- Go 1.23 or newer
- `golangci-lint` (optional, for `make lint`)
- `govulncheck` (optional)

## Workflow

1. Fork the repository and create a feature branch from `main`.
2. Make your changes. Keep the library dependency-free: only the Go
   standard library may be imported.
3. Add or update tests. Public API changes need godoc comments,
   golden fixtures under `testdata/` where applicable, and a CHANGELOG
   entry under `Unreleased`.
4. Run `make check` (or at minimum `gofmt -w .`, `go vet ./...` and
   `go test -race ./...`). All tests must pass with `-race` and total
   coverage must stay at 90% or higher (`make cover`).
5. Open a pull request describing the change and its motivation.

## Code style

- `gofmt`-clean, `go vet`-clean.
- Godoc on every exported symbol, starting with the symbol name.
- No panics in library code (except `Must*` helpers), no `fmt.Println`
  or `log.Print` in the package — use `*slog.Logger` via `WithLogger`.
- Monetary values are always `Decimal`, never `float64`.
- Errors wrap sentinel errors so `errors.Is`/`errors.As` keep working.

## Tests

- Unit tests use `net/http/httptest` and golden fixtures in `testdata/`.
- No real network calls in unit tests.
- Integration tests live behind the `integration` build tag and require
  `GECKOTERMINAL_INTEGRATION=1`:

  ```sh
  make integration
  ```

- Fuzz targets:

  ```sh
  make fuzz
  ```

## Commits and pull requests

- [Conventional Commits](https://www.conventionalcommits.org):
  `feat:`, `fix:`, `docs:`, `test:`, `ci:`, `chore:`, `refactor:`,
  `perf:`, `build:`.
- One logical change per commit; imperative English subject lines.
- Never commit secrets, API keys, `.env` files or coverage artifacts.

## Reporting security issues

Please do not open public issues for vulnerabilities. See
[SECURITY.md](SECURITY.md).
