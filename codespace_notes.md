# Database Migration

## Running Migrations

To execute database migrations, first ensure your environment variables are properly loaded from your `.env` file, then run the migrations using goose.

```bash
export $(cat .env | grep -v '#' | xargs) && goose up
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