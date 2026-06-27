# ADR-001: Separate domain logic from infrastructure

## Status
Accepted

## Context
The project manages guild wallets (reserve, deduct, release gold). As persistence and business logic grow, mixing SQL queries with business rules in the same file makes both harder to test and change independently.
We needed to decide how to structure `internal/` flat (one package per concern) vs layered (domain / infrastructure split).

## Decision
We split `internal/` into two layers:

- **`internal/guild/`** owns the domain: the `Guild` entity, the `Repository` interface, and `WalletService`. No import of any database driver or third-party infrastructure library is allowed here.
- **`internal/postgres/`** owns the Postgres implementation of `guild.Repository`. It is the only package that imports `database/sql` or a Postgres driver.

Dependency direction is enforced by Go's import rules: `internal/postgres` imports `internal/guild`, never the reverse.

## Consequences

**Good:**
- `WalletService` can be unit-tested with a simple in-memory mock (`MockGuildRepo`) no database or Docker required.
- Swapping the storage only requires a new implementation of `guild.Repository`; service logic is untouched.
- Business rules live in one place and are readable without SQL noise.

**Bad:**
- More files and packages than a flat structure .
