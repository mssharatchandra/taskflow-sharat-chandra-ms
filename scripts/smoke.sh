#!/usr/bin/env bash
set -euo pipefail

PROJECT_NAME="${COMPOSE_PROJECT_NAME:-taskflow_smoke}"
HOST_PORT="${PORT:-18080}"
SEED_PROJECT_ID="b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a22"

cleanup() {
  docker compose -p "$PROJECT_NAME" down -v --remove-orphans >/dev/null 2>&1 || true
}

trap cleanup EXIT

echo "[smoke] Starting stack (project=$PROJECT_NAME, host_port=$HOST_PORT)..."
PORT="$HOST_PORT" docker compose -p "$PROJECT_NAME" up --build -d

echo "[smoke] Waiting for API health..."
for _ in $(seq 1 40); do
  if docker compose -p "$PROJECT_NAME" exec -T api curl -fsS http://localhost:8080/health >/dev/null 2>&1; then
    break
  fi
  sleep 2
done

if ! docker compose -p "$PROJECT_NAME" exec -T api curl -fsS http://localhost:8080/health >/dev/null 2>&1; then
  echo "[smoke] API did not become healthy"
  docker compose -p "$PROJECT_NAME" logs api --tail=80 || true
  exit 1
fi

echo "[smoke] Verifying unauthorized access is blocked..."
unauth_status="$(docker compose -p "$PROJECT_NAME" exec -T api sh -lc 'curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/projects')"
if [ "$unauth_status" != "401" ]; then
  echo "[smoke] Expected 401 from /projects without token, got: $unauth_status"
  exit 1
fi

echo "[smoke] Logging in with seeded credentials and hitting protected endpoints..."
projects_status="$(docker compose -p "$PROJECT_NAME" exec -T api sh -lc 'TOKEN=$(curl -sS -X POST http://localhost:8080/auth/login -H "Content-Type: application/json" -d '\''{"email":"test@example.com","password":"password123"}'\'' | sed -n '\''s/.*"token":"\([^"]*\)".*/\1/p'\''); [ -n "$TOKEN" ] || exit 1; curl -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer $TOKEN" "http://localhost:8080/projects?page=1&limit=2"')"
if [ "$projects_status" != "200" ]; then
  echo "[smoke] Expected 200 from /projects with token, got: $projects_status"
  exit 1
fi

stats_status="$(docker compose -p "$PROJECT_NAME" exec -T api sh -lc 'TOKEN=$(curl -sS -X POST http://localhost:8080/auth/login -H "Content-Type: application/json" -d '\''{"email":"test@example.com","password":"password123"}'\'' | sed -n '\''s/.*"token":"\([^"]*\)".*/\1/p'\''); [ -n "$TOKEN" ] || exit 1; curl -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer $TOKEN" "http://localhost:8080/projects/'"$SEED_PROJECT_ID"'/stats"')"
if [ "$stats_status" != "200" ]; then
  echo "[smoke] Expected 200 from /projects/:id/stats, got: $stats_status"
  exit 1
fi

echo "[smoke] PASS: core auth + protected routes are healthy."
