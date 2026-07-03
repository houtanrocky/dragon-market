package adr
# ADR-003: HTTP Router

## Status

Accepted

## Context

The project exposes a small REST API. We needed routing, URL parameters, and middleware without introducing a large framework.

## Decision

Use **chi** as the HTTP router.

`chi` builds on Go's standard `net/http`, keeping handlers framework-agnostic while providing convenient routing and middleware support.

## Consequences

**Good:**

* Lightweight with minimal dependencies.
* Uses standard `http.Handler`, making handlers easy to test.
* Easy to replace with another router if needed.

**Bad:**

* Less built-in functionality (binding, validation, etc.) than full frameworks like Gin.
