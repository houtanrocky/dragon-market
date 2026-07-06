# Contributing to Dragon Market

Contributions that improve the project as a clear, focused Go marketplace example are welcome.

## Development setup

1. Install Go 1.25+ and Docker.
2. Start local dependencies with `docker compose up -d postgres oracle`.
3. Run the API with `go run ./cmd/server`.
4. Import `Dragon_Market.postman_collection.json` for manual API testing.

## Before opening a pull request

Run:

```sh
gofmt -w cmd internal
go test ./...
go test -race ./...
go vet ./...
```

Keep changes focused and include tests for new behavior or regressions. Use concise Conventional Commit messages such as `feat: add order lookup` or `fix: preserve bid reservation`.

## Architecture constraints

- Keep business entities, rules, and repository interfaces in their domain package.
- Domain packages must not import `internal/postgres` or HTTP concerns.
- Infrastructure may depend on domain packages, never the reverse.
- Preserve transaction boundaries for wallet, ownership, order, and auction mutations.
- Add schema changes as new ordered migrations; do not rewrite applied migrations.

For architectural context, read the decisions in `docs/adr` before proposing a cross-package change.
