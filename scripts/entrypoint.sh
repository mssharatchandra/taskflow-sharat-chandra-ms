#!/bin/sh
set -e

echo "Waiting for PostgreSQL to be ready..."
until pg_isready -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" 2>/dev/null; do
  echo "PostgreSQL is not ready yet, retrying in 2s..."
  sleep 2
done
echo "PostgreSQL is ready."

echo "Running database migrations..."
MIGRATE_OUTPUT="$(migrate -path /migrations -database "$DATABASE_URL" up 2>&1)" || {
  if echo "$MIGRATE_OUTPUT" | grep -q "no change"; then
    echo "No new migrations to run."
  else
    echo "$MIGRATE_OUTPUT"
    exit 1
  fi
}

if [ -n "$MIGRATE_OUTPUT" ]; then
  echo "$MIGRATE_OUTPUT"
fi

echo "Starting TaskFlow API server..."
exec taskflow
