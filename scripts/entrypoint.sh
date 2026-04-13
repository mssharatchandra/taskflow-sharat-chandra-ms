#!/bin/sh
set -e

echo "Waiting for PostgreSQL to be ready..."
until pg_isready -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" 2>/dev/null; do
  echo "PostgreSQL is not ready yet, retrying in 2s..."
  sleep 2
done
echo "PostgreSQL is ready."

echo "Running database migrations..."
migrate -path /migrations -database "$DATABASE_URL" up

echo "Starting TaskFlow API server..."
exec taskflow
