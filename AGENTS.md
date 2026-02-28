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

cmd/        # Entrypoints
domain/     # Business logic (no external deps)
app/        # CQRS commands/queries
ports/      # HTTP handlers, OpenAPI
adapters/   # MySQL, HTTP clients
core/       # errors, logging, tracing

**Rule:** Outer imports inner. Inner MUST NOT import outer.

## BDD/TDD Workflow
- Write Given/When/Then scenarios for all new behaviors. Use natural language. Keep scenarios ignorant of code structure.
- See every scenario fail before implementing it.
- Write a custom parser/runner as glue code connecting scenario language to production code.
- Follow the three laws of TDD to make each scenario pass.
- Never create or allow no-op, pending, or skeleton step definitions. Every step must exercise real production code and be seen to fail first.

## Code Quality
- Keep functions small; cyclomatic complexity no greater than five where practical.
- Decouple tests from production code via a testing API: as tests get more specific, code gets more generic.
- Keep test coverage in the high 90s for line and branch.
- Use available linters.
- Optimize code to minimize token count and reduce context window pressure.

## Git Discipline
- Check test coverage before commit.
- Never push to git without asking first.
