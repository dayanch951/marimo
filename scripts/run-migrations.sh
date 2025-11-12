#!/bin/bash

# Database Migration Runner
# Usage: ./scripts/run-migrations.sh [environment]

set -e

ENVIRONMENT=${1:-development}
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
MIGRATIONS_DIR="$PROJECT_ROOT/migrations"

echo "=========================================="
echo "Running Database Migrations"
echo "Environment: $ENVIRONMENT"
echo "=========================================="

# Load environment variables
if [ -f "$PROJECT_ROOT/.env" ]; then
    export $(cat "$PROJECT_ROOT/.env" | grep -v '^#' | xargs)
fi

# Database connection parameters
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_NAME=${DB_NAME:-marimo_erp}
DB_USER=${DB_USER:-postgres}

echo "Connecting to: $DB_HOST:$DB_PORT/$DB_NAME as $DB_USER"

# Check if psql is available
if ! command -v psql &> /dev/null; then
    echo "Error: psql command not found. Please install PostgreSQL client."
    exit 1
fi

# Check if database is accessible
if ! PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres -c '\q' 2>/dev/null; then
    echo "Error: Cannot connect to database"
    exit 1
fi

echo "✓ Database connection successful"

# Create database if it doesn't exist
echo ""
echo "Checking if database exists..."
DB_EXISTS=$(PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres -tAc "SELECT 1 FROM pg_database WHERE datname='$DB_NAME'")

if [ -z "$DB_EXISTS" ]; then
    echo "Creating database: $DB_NAME"
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres -c "CREATE DATABASE $DB_NAME;"
    echo "✓ Database created"
else
    echo "✓ Database already exists"
fi

# Create migrations table to track applied migrations
echo ""
echo "Setting up migration tracking..."
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
CREATE TABLE IF NOT EXISTS schema_migrations (
    id SERIAL PRIMARY KEY,
    version VARCHAR(255) NOT NULL UNIQUE,
    applied_at TIMESTAMP NOT NULL DEFAULT NOW()
);
"
echo "✓ Migration tracking ready"

# Run migrations in order
echo ""
echo "Applying migrations..."

for migration_file in "$MIGRATIONS_DIR"/*.sql; do
    if [ -f "$migration_file" ]; then
        filename=$(basename "$migration_file")
        version="${filename%.*}"

        # Check if migration has already been applied
        APPLIED=$(PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -tAc "SELECT 1 FROM schema_migrations WHERE version='$version'")

        if [ -z "$APPLIED" ]; then
            echo ""
            echo "Applying: $filename"

            # Run migration in transaction
            if PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -v ON_ERROR_STOP=1 -f "$migration_file"; then
                # Record successful migration
                PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
                    INSERT INTO schema_migrations (version) VALUES ('$version');
                "
                echo "✓ Successfully applied: $filename"
            else
                echo "✗ Failed to apply: $filename"
                exit 1
            fi
        else
            echo "⊘ Already applied: $filename"
        fi
    fi
done

# Run ANALYZE to update statistics
echo ""
echo "Updating database statistics..."
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "ANALYZE;"
echo "✓ Statistics updated"

# Show applied migrations
echo ""
echo "Applied migrations:"
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
    SELECT version, applied_at
    FROM schema_migrations
    ORDER BY applied_at;
"

echo ""
echo "=========================================="
echo "✓ All migrations completed successfully!"
echo "=========================================="
