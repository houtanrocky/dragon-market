# Dragon Market - Go Marketplace Backend Example

[![Go](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go)](https://go.dev/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-4169E1?logo=postgresql&logoColor=white)](https://www.postgresql.org/)
[![CI](https://github.com/houtanrocky/dragon-market/actions/workflows/ci.yml/badge.svg)](https://github.com/houtanrocky/dragon-market/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

Dragon Market is an educational **marketplace backend written in Go** with a REST API, PostgreSQL, Docker Compose, fixed-price orders, timed auctions, wallet reservations, idempotent payments, and concurrency-safe transactions. It is a practical reference for developers looking for a **Golang clean architecture example**, a **PostgreSQL transaction example**, or patterns for preventing double-selling in an e-commerce or trading system.

The fantasy-item domain keeps the example approachable: common and rare items use limit orders, while one-of-a-kind legendary items use auctions.

> This repository is designed as a compact backend engineering example. Authentication, production observability, and deployment infrastructure are intentionally outside its scope.

## What this project demonstrates

- Clean separation between Go domain logic, HTTP transport, and PostgreSQL infrastructure
- Transactional order purchases that prevent concurrent double-selling
- Auction bidding with row locks, a 5% minimum increment, and deadline extension
- Wallet balance reservations, daily spending limits, and an append-only audit trail
- Idempotency keys for safely retrying financial HTTP requests
- Background settlement of expired auctions
- Resilient external price-oracle integration with last-known-good fallback
- PostgreSQL constraints and partial unique indexes for business invariants
- Unit, repository, HTTP, race, and concurrency tests using Testcontainers
- Architecture Decision Records explaining important design choices

## Technology stack

| Component | Choice |
|---|---|
| Language | Go 1.25 |
| HTTP router | chi |
| Database | PostgreSQL 16 |
| Database driver | pgx through `database/sql` |
| Integration testing | Testcontainers for Go |
| Local environment | Docker Compose |
| API client examples | Importable Postman collection |

## Architecture

```text
HTTP request
    |
    v
internal/api          REST handlers and status-code mapping
    |
    v
internal/{domain}     entities, business rules, repository interfaces
    |
    v
internal/postgres     SQL repositories, transactions, row locks
    |
    v
PostgreSQL
```

Domain services live in `internal/guild`, `internal/item`, `internal/order`, and `internal/auction`. They depend on small interfaces rather than database or HTTP packages. PostgreSQL implementations live in `internal/postgres`; HTTP handlers live in `internal/api`.

PostgreSQL provides the strong consistency needed for purchases, bids, wallet reservations, and ownership transfers. `SELECT ... FOR UPDATE` serializes competing operations, while a partial unique index guarantees one active auction per legendary item. See the [Architecture Decision Records](docs/adr) for the reasoning and tradeoffs.

## Quick start

Requirements: Go 1.25+, Docker, and Docker Compose.

Start PostgreSQL and the included mock price oracle:

```sh
docker compose up -d postgres oracle
```

On a fresh database, Compose runs [`dev/seed.sql`](dev/seed.sql) and creates these manual-testing guilds:

| Guild | Gold | Daily limit |
|---|---:|---:|
| `guild-seller` | 10,000 | 5,000 |
| `guild-buyer` | 10,000 | 5,000 |
| `guild-bidder` | 10,000 | 5,000 |

Run the API:

```sh
DATABASE_URL='postgres://market:market@localhost:5432/market?sslmode=disable' \
PRICE_ORACLE_URL='http://localhost:8090' \
go run ./cmd/server
```

PowerShell:

```powershell
$env:DATABASE_URL = "postgres://market:market@localhost:5432/market?sslmode=disable"
$env:PRICE_ORACLE_URL = "http://localhost:8090"
go run ./cmd/server
```

The API listens on `http://localhost:8080`. Verify it with:

```sh
curl http://localhost:8080/healthz
```

If the database was initialized before the seed was added, apply it once:

```sh
docker compose exec -T postgres psql -U market -d market -f /docker-entrypoint-initdb.d/999_seed.sql
```

## Try the API

Import [`Dragon_Market.postman_collection.json`](Dragon_Market.postman_collection.json) into Postman. The collection contains every endpoint, reusable variables, generated idempotency keys, and scripts that capture created item, order, auction, and bid IDs.

Create a guild:

```sh
curl -X POST http://localhost:8080/guilds \
  -H 'Content-Type: application/json' \
  -d '{"id":"guild-demo","gold":10000,"daily_limit":5000}'
```

Create an item:

```sh
curl -X POST http://localhost:8080/items \
  -H 'Content-Type: application/json' \
  -d '{"name":"Arcane Wand","type":"rare","owner_id":"guild-seller","base_price":500}'
```

## REST API

| Method | Route | Purpose |
|---|---|---|
| `GET` | `/healthz` | Process health check |
| `POST` | `/guilds` | Create a guild and wallet |
| `GET` | `/guilds/{id}/wallet` | Read balance, reservations, and daily spend |
| `POST` | `/items` | Create a common, rare, or legendary item |
| `GET` | `/items` | List free items |
| `GET` | `/items/{id}` | Get an item |
| `POST` | `/orders` | List a common or rare item at a fixed price |
| `POST` | `/orders/{id}/buy` | Buy an order; requires `Idempotency-Key` |
| `DELETE` | `/orders/{id}` | Cancel an order |
| `POST` | `/auctions` | Start a legendary-item auction |
| `GET` | `/auctions/{id}` | Get an auction |
| `POST` | `/auctions/{id}/bids` | Place a bid; requires `Idempotency-Key` |
| `GET` | `/bids/{id}` | Get a bid |
| `DELETE` | `/auctions/{auctionID}/bids/{bidID}` | Cancel the current highest bid |

## Concurrency and reliability

- Purchases lock both the item and order before transferring funds and ownership.
- Auction mutations lock the auction row, producing one serialized bid order.
- Available wallet balance is total gold minus reserved gold.
- Daily spending resets on the next UTC calendar day.
- Wallet mutations append an audit record inside the same transaction.
- Bids reserve funds; an outbid bidder's reservation is released atomically.
- Bids must exceed the current bid by at least 5%.
- Bids placed in the final five minutes extend the auction by five minutes.
- A background worker settles expired auctions; database locks make concurrent workers safe.
- Repeating a completed mutation with the same idempotency key returns its original result.

## Price oracle

The external oracle contract is:

```http
GET /prices/{itemID}
```

```json
{"base_price": 1250}
```

Prices refresh immediately at startup and every 30 seconds. Zero, negative, malformed, timed-out, and non-200 responses are rejected. The last valid database value is retained, and one failed item does not stop the rest of the refresh batch. Docker Compose includes a deterministic mock oracle for local development.

## Project structure

```text
cmd/server/             API entry point and dependency wiring
cmd/mock-oracle/        local deterministic price service
internal/api/           HTTP router and handlers
internal/auction/       auction and bidding rules
internal/guild/         wallet and daily-limit rules
internal/item/          item lifecycle
internal/order/         fixed-price order rules
internal/oracle/        HTTP oracle, fallback, and updater
internal/postgres/      PostgreSQL repository implementations
migrations/             ordered schema migrations
docs/adr/               architecture decision records
dev/seed.sql            development fixtures
```

## Testing

Docker must be running because repository tests start PostgreSQL with Testcontainers.

```sh
go build ./...
go test ./...
go test -race ./...
go vet ./...
```

The integration suite covers competing purchases, concurrent bids, duplicate auction creation, repository behavior, and transaction rollback.

## Assumptions and scope

- Gold uses an integer smallest unit; floating-point arithmetic is not used.
- Auction deadlines and daily-limit dates use UTC.
- Authentication is outside this example, so guild IDs are treated as trusted input.
- A highest bidder may cancel only while its bid remains active.
- Limit-order prices are locked at listing time and do not change with later oracle updates.

## Contributing

Issues and pull requests are welcome. Read [CONTRIBUTING.md](CONTRIBUTING.md) for setup, testing, and architectural constraints.

## License

Released under the [MIT License](LICENSE). You may use, modify, and redistribute this example with attribution.
