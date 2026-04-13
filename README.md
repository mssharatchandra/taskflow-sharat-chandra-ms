# TaskFlow Backend (Go + PostgreSQL)

## 1. Overview
TaskFlow is a task management backend API with JWT auth, project management, and task assignment workflows.

This implementation is built for the **Backend Engineer** track and includes:
- Go REST API (`gin`)
- PostgreSQL with SQL migrations (up/down)
- JWT authentication + bcrypt password hashing
- Docker multi-stage build + `docker compose` workflow
- Seed data for immediate review
- Integration tests for auth endpoints
- Postman collection for end-to-end API testing

Tech stack:
- Go 1.24
- Gin
- PostgreSQL 16
- `database/sql` + `lib/pq`
- `golang-jwt/jwt/v5`
- `golang.org/x/crypto/bcrypt`
- `golang-migrate` (inside container entrypoint)

---

## 2. Architecture Decisions
### Layered backend structure
I separated the code into `handler`, `service`, and `repository` layers under `internal/`.
- `handler`: HTTP parsing/validation/response mapping
- `service`: business logic and authorization checks
- `repository`: SQL and database access

Why: this keeps handlers thin, makes business rules testable, and avoids "god functions".

### Raw SQL over ORM
I intentionally used `database/sql` and handwritten SQL instead of an ORM.

Why:
- Explicit control over joins, filters, aggregates, and pagination queries
- Easier to reason about performance and indexes in a take-home where SQL quality is evaluated
- Better alignment with Go backend conventions for small/medium APIs

Tradeoff:
- More boilerplate mapping `Scan(...)` into structs
- No compile-time query generation (could be improved later with `sqlc`)

### Auth and security
- Passwords are hashed with bcrypt (configurable, default cost `12`)
- JWT includes `user_id` and `email` claims and expires in 24 hours
- Protected routes use Bearer token middleware
- 401 and 403 are handled distinctly

### Data modeling and constraints
Schema is migration-driven with:
- UUID primary keys
- Foreign keys and cascades
- Enum-like checks for `status` and `priority`
- Indexes on common lookup/filter columns (`users.email`, `projects.owner_id`, `tasks.project_id`, `tasks.assignee_id`, `tasks.status`)

### Intentionally left out
To stay within scope and keep implementation quality high:
- No background jobs/queues
- No rate limiting yet
- No OpenAPI generation yet
- Limited integration tests (focused on auth)

---

## 3. Running Locally
Assumption: Docker is installed and running.

```bash
git clone https://github.com/<your-username>/taskflow-<your-name>.git
cd taskflow-<your-name>
cp .env.example .env
docker compose up --build
```

API is available at:
- `http://localhost:8080`

Health check:
- `GET http://localhost:8080/health`

Notes:
- PostgreSQL and API start from one command.
- Migrations are automatically run on API container startup.

---

## 4. Running Migrations
Migrations run automatically in `scripts/entrypoint.sh` when the API container starts.

If you need to run manually from inside the API container:

```bash
docker compose exec api migrate -path /migrations -database "$DATABASE_URL" up
```

Run down migration (one step):

```bash
docker compose exec api migrate -path /migrations -database "$DATABASE_URL" down 1
```

---

## 5. Test Credentials
Seeded credentials:

- **Email:** `test@example.com`
- **Password:** `password123`

Seed data also includes:
- 1 project
- 3 tasks (`todo`, `in_progress`, `done`)

---

## 6. API Reference
### Base URL
- `http://localhost:8080`

### Authentication
- `POST /auth/register`
- `POST /auth/login`

### Projects
- `GET /projects` (supports `page`, `limit`)
- `POST /projects`
- `GET /projects/:id` (project details + tasks)
- `PATCH /projects/:id`
- `DELETE /projects/:id`
- `GET /projects/:id/stats` (bonus)

### Tasks
- `GET /projects/:id/tasks` (supports `status`, `assignee`, `page`, `limit`)
- `POST /projects/:id/tasks`
- `PATCH /tasks/:id`
- `DELETE /tasks/:id`

### Error response shapes
Validation error:

```json
{ "error": "validation failed", "fields": { "email": "is required" } }
```

Not found:

```json
{ "error": "not found" }
```

Forbidden:

```json
{ "error": "forbidden" }
```

Unauthorized:

```json
{ "error": "unauthorized" }
```

### Postman assets
- Collection: `docs/taskflow.postman_collection.json`
- Environment: `docs/taskflow.postman_environment.json`

These include all endpoints with sample requests and token auto-capture after login/register.

---

## 7. What I'd Do With More Time
1. Add fuller integration coverage for project/task authorization edge cases.
2. Add request ID middleware and correlation-friendly structured logs.
3. Add rate limiting and brute-force protection on auth endpoints.
4. Improve API documentation with OpenAPI/Swagger generation.
5. Add Makefile targets (`test`, `lint`, `compose-up`, `compose-down`) for smoother developer UX.
6. Consider `sqlc` for type-safe query generation and less manual scan boilerplate.

---

## Additional Notes
### Running tests
```bash
GOCACHE=$(pwd)/.cache/go-build go test ./...
```

### Environment variables
See `.env.example` for required settings:
- `PORT`
- `DATABASE_URL`
- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`
- `JWT_SECRET`
- `BCRYPT_COST`
