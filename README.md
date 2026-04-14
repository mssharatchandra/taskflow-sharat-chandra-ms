# TaskFlow Backend (Backend Engineer Take-Home)

## 1. Overview
TaskFlow is a production-style task management backend that supports:
- user registration/login
- project creation and management
- task creation, assignment, update, filtering, and deletion
- JWT-protected access control

This submission targets the **Backend Engineer** track and is intentionally backend-focused (no frontend app).

### Stack
- Go 1.25
- Gin (`github.com/gin-gonic/gin`)
- PostgreSQL 16
- `database/sql` + `lib/pq`
- `golang-migrate` for SQL migrations
- JWT (`golang-jwt/jwt/v5`)
- bcrypt (`golang.org/x/crypto/bcrypt`)
- Docker + Docker Compose

---

## 2. Architecture Decisions
### Layering and boundaries
I used a layered architecture under `internal/`:
- `handler/`: HTTP-only concerns (binding, validation, response mapping)
- `service/`: business logic + authorization rules
- `repository/`: SQL queries and persistence

This keeps HTTP framework details out of business logic and makes each layer independently testable.

### Why raw SQL (not ORM)
I chose `database/sql` + explicit SQL instead of an ORM:
- better query control for ownership/access filters and stats aggregation
- easier to reason about indexes and query plans
- less hidden behavior for a take-home where correctness and clarity matter

Tradeoff accepted: more manual mapping (`Scan`) and boilerplate.

### Auth and authorization design
- Passwords are hashed with bcrypt (`BCRYPT_COST`, default `12`, validated range `10..20`)
- JWT expires in 24h and includes `user_id` + `email`
- Middleware enforces authentication on non-auth routes
- 401 vs 403 semantics are separated:
  - `401 unauthorized`: missing/invalid token
  - `403 forbidden`: authenticated but lacks permission

### Data and migration strategy
- SQL migrations are versioned with both `up` and `down`
- schema includes FK constraints, cascade behavior, enum-like checks, and indexes
- seed migration creates the required review data (user/project/3 tasks)

### Error contract
Validation failures return structured payloads:
```json
{ "error": "validation failed", "fields": { "email": "is required" } }
```
Not found / forbidden / unauthorized are consistently shaped.

### Extra-mile ownership effort
Beyond baseline requirements, I added:
- pagination on list endpoints (`page`, `limit`, metadata)
- `GET /projects/:id/stats` (counts by status + assignee)
- integration tests for auth and authorization edge cases
- CI workflow (`gofmt` check + `go vet` + `go test`)
- `Makefile` for common workflows
- dockerized smoke test script (`scripts/smoke.sh`) for fast sanity verification

---

## 3. Running Locally
Assumption: reviewer has Docker installed and running.

```bash
git clone <your-public-repo-url>
cd <your-repo-folder>
cp .env.example .env
docker compose up --build
```

API URL:
- `http://localhost:8080`

Health check:
- `GET http://localhost:8080/health`

If port `8080` is already occupied locally:
```bash
PORT=18080 docker compose up --build
```

---

## 4. Running Migrations
Migrations run automatically on container startup via `scripts/entrypoint.sh`.

Manual commands (inside API container):
```bash
docker compose exec api migrate -path /migrations -database "$DATABASE_URL" up
```

Run one down migration:
```bash
docker compose exec api migrate -path /migrations -database "$DATABASE_URL" down 1
```

---

## 5. Test Credentials
Seed credentials (from migration):
- Email: `test@example.com`
- Password: `password123`

Seeded data includes:
- 1 project
- 3 tasks (`todo`, `in_progress`, `done`)

---

## 6. API Reference
Base URL: `http://localhost:8080`

### Auth
- `POST /auth/register`
- `POST /auth/login`

### Projects
- `GET /projects` (supports `page`, `limit`)
- `POST /projects`
- `GET /projects/:id` (returns project + tasks)
- `PATCH /projects/:id` (owner only)
- `DELETE /projects/:id` (owner only)
- `GET /projects/:id/stats` (bonus)

### Tasks
- `GET /projects/:id/tasks` (supports `status`, `assignee`, `page`, `limit`)
- `POST /projects/:id/tasks`
- `PATCH /tasks/:id`
- `DELETE /tasks/:id` (project owner or task creator)

### Error responses
Validation:
```json
{ "error": "validation failed", "fields": { "email": "is required" } }
```

Unauthorized:
```json
{ "error": "unauthorized" }
```

Forbidden:
```json
{ "error": "forbidden" }
```

Not found:
```json
{ "error": "not found" }
```

### API artifacts
- Postman collection: `docs/taskflow.postman_collection.json`
- Postman environment: `docs/taskflow.postman_environment.json`

---

## 7. What I'd Do With More Time
1. Add broader integration coverage around cross-user authorization and negative cases.
2. Add request IDs + correlation fields in logs for easier production debugging.
3. Add auth hardening: rate limiting and brute-force protection.
4. Generate OpenAPI docs from source and publish API docs.
5. Add race checks and richer CI quality gates.
6. Consider `sqlc` for type-safe query generation while keeping SQL explicit.

---

## Useful Commands
```bash
make ci          # gofmt check + go vet + go test
make smoke       # dockerized auth/protected-route smoke test
make compose-up
make compose-down
```

Required env vars are documented in `.env.example`.
