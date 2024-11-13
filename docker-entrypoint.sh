#!/bin/sh
set -e

# Wait for the database to be ready
until PGPASSWORD=$DB_PASSWORD psql -h "$DB_HOST" -U "$DB_USER" -d "$DB_NAME" -c '\q'; do
  >&2 echo "Postgres is unavailable - sleeping"
  sleep 1
done

>&2 echo "Postgres is up - executing migrations"

# Goose variables are already set in the environment from docker-compose
# but we can verify them here
if [ -z "$GOOSE_DRIVER" ] || [ -z "$GOOSE_DBSTRING" ] || [ -z "$GOOSE_MIGRATION_DIR" ]; then
  echo "Error: Missing required Goose environment variables"
  exit 1
fi

# Log migration configuration (without sensitive data)
echo "Migration Configuration:"
echo "Driver: $GOOSE_DRIVER"
echo "Migration Dir: $GOOSE_MIGRATION_DIR"
echo "Database: $DB_NAME on $DB_HOST:$DB_PORT"

if [ "$RESET_MIGRATIONS" = "true" ]; then
  echo "Resetting migrations..."
  goose reset
fi

echo "Running migrations up..."
goose up

# Start the application
exec "./app"