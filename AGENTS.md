# AGENTS.md

Task Management API with Go using Clean Architecture (Ports & Adapters).

## Commands

```bash
make test          # Run tests
make start         # Start HTTP server
make openapi       # Regenerate OpenAPI
make migration-up  # Apply migrations
```

## Architecture

```
cmd/        # Entrypoints
domain/     # Business logic (no external deps)
app/        # CQRS commands/queries
ports/      # HTTP handlers, OpenAPI
adapters/   # MySQL, HTTP clients
core/       # errors, logging, tracing
```

**Rule:** Outer imports inner. Inner MUST NOT import outer.

## Conventions

- **Imports:** internal, stdlib, external (blank lines between)
- **Interfaces:** `Task`, `User` (nouns)
- **Structs:** lowercase private (`task struct`)
- **Constructors:** `NewTask() (Task, error)`, `From() (Task, error)` for reconstruction
- **Errors:** `errors.E(op, err)` / `errors.E(op, errkind.NotExist, errors.Str("msg"))`
- **Testing:** `package_test`, `t.Parallel()`, `testify/require`
- **Nulls:** `sql.NullTime` for nullable DB fields

## Deps

chi, sqlx, testify, logrus, otel
