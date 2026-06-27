# ADR-002: Database Choice

## Status
Accepted

## Context
The marketplace needs to enforce several hard consistency rules atomically

These are transactional, relational workloads with strong consistency requirements.

## Decision
PostgreSQL.

The unique partial index on `auctions(item_id) WHERE status = 'active'` enforces the one-active-auction-per-item rule at the database level, no amount of application-level locking can substitute for this under concurrent load.

`SELECT FOR UPDATE` on guild rows gives us row-level locking during balance checks, which prevents the TOCTOU race between "check balance" and "commit reservation."

## Consequences

**Good:**
- Constraint violations (duplicate auction, negative balance) are caught by the DB even if application logic has a bug.
- Full ACID transactions across multiple tables (bid + reserve in one commit).

**Bad:**
- Requires running Postgres locally.
- Horizontal write scaling is harder than with some NoSQL.
