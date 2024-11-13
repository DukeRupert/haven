# Database Migration

## Running Migrations

To execute database migrations, first ensure your environment variables are properly loaded from your `.env` file, then run the migrations using goose.

```bash
export $(cat .env | grep -v '#' | xargs) && goose up
```
```bash
GOOSE_DBSTRING=postgresql://postgres.kovrchxpqjntxeicmgtc:[YOUR-PASSWORD]@aws-0-us-east-2.pooler.supabase.com:5432/postgres?sslmode=disable GOOSE_DRIVER=postgres goose reset
```

This command:
1. Reads the `.env` file (`cat .env`)
2. Filters out commented lines (`grep -v '#'`)
3. Converts the variables to a format suitable for export (`xargs`)
4. Exports them to the current shell session (`export`)
5. Runs the database migrations (`goose up`)

### Prerequisites
- A properly configured `.env` file
- [goose](https://github.com/pressly/goose) installed on your system
- Database connection details set in your environment variables

### Important Note for Supabase
When running goose migrations against a Supabase database, you must use session mode (port 5432) rather than transaction mode (port 6543). Ensure your database connection string uses port 5432 in your .env file:

```bash
# Correct - Session mode
DATABASE_URL=postgresql://postgres:[YOUR-PASSWORD]@localhost:5432/postgres

# Wrong - Transaction mode
DATABASE_URL=postgresql://postgres:[YOUR-PASSWORD]@localhost:6543/postgres
```

### Note
Make sure your `.env` file contains all necessary database connection details required by your migration configuration.

Would you like me to add any additional sections like troubleshooting or common environment variables?

## Docker 

To use this configuration:

1. Local development:
```bash
# Copy the example env file
cp .env.example .env

# Edit with your values
nano .env

# Start the services
docker compose up --build
```

2. To run migrations manually (if needed):
```bash
# Enter the container
docker compose exec web sh

# Run goose commands
goose up
goose status
goose reset
```

3. To test migrations:
```bash
# Reset and rerun migrations
RESET_MIGRATIONS=true docker compose up --build
```

Key points about this setup:

1. `GOOSE_DBSTRING` is constructed in docker-compose.yml using the individual DB_* variables. This:
   - Makes it easier to manage
   - Avoids duplicating sensitive information
   - Maintains consistency with the application's database configuration

2. `GOOSE_MIGRATION_DIR` points to `/app/migrations` in the container because:
   - The Dockerfile copies migrations to this location
   - This ensures consistency between development and production

3. `GOOSE_DRIVER` is set to 'postgres' directly in docker-compose.yml since it's unlikely to change