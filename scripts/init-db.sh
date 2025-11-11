#!/bin/bash
# Database Initialization Script for Marimo ERP
# This script initializes the PostgreSQL database with necessary tables and seed data

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Load environment variables
if [ -f .env ]; then
    echo -e "${GREEN}Loading environment variables from .env...${NC}"
    export $(cat .env | grep -v '^#' | xargs)
elif [ -f .env.development ]; then
    echo -e "${YELLOW}Loading environment variables from .env.development...${NC}"
    export $(cat .env.development | grep -v '^#' | xargs)
else
    echo -e "${RED}No .env file found. Using default values...${NC}"
fi

# Database connection parameters
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-postgres}"
DB_PASSWORD="${DB_PASSWORD:-postgres}"
DB_NAME="${DB_NAME:-marimo_dev}"

echo -e "${GREEN}================================${NC}"
echo -e "${GREEN}Marimo ERP - Database Initialization${NC}"
echo -e "${GREEN}================================${NC}"
echo ""
echo "Database: ${DB_NAME}"
echo "Host: ${DB_HOST}:${DB_PORT}"
echo "User: ${DB_USER}"
echo ""

# Check if PostgreSQL is running
echo -e "${YELLOW}Checking PostgreSQL connection...${NC}"
if ! PGPASSWORD=$DB_PASSWORD psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -lqt | cut -d \| -f 1 | grep -qw "$DB_NAME"; then
    echo -e "${RED}Error: Cannot connect to PostgreSQL or database does not exist${NC}"
    echo -e "${YELLOW}Creating database ${DB_NAME}...${NC}"
    PGPASSWORD=$DB_PASSWORD psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -c "CREATE DATABASE $DB_NAME;" || {
        echo -e "${RED}Failed to create database${NC}"
        exit 1
    }
fi
echo -e "${GREEN}✓ PostgreSQL connection successful${NC}"
echo ""

# Run migrations
echo -e "${YELLOW}Running database migrations...${NC}"
MIGRATIONS_DIR="./migrations"

if [ ! -d "$MIGRATIONS_DIR" ]; then
    echo -e "${RED}Error: Migrations directory not found at $MIGRATIONS_DIR${NC}"
    exit 1
fi

# Execute all .up.sql migration files in order
for migration in $(ls -1 "$MIGRATIONS_DIR"/*.up.sql 2>/dev/null | sort); do
    echo -e "${YELLOW}Applying migration: $(basename $migration)${NC}"
    PGPASSWORD=$DB_PASSWORD psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f "$migration" || {
        echo -e "${RED}Failed to apply migration: $migration${NC}"
        exit 1
    }
    echo -e "${GREEN}✓ Migration applied successfully${NC}"
done

echo ""
echo -e "${GREEN}================================${NC}"
echo -e "${GREEN}Database initialization completed!${NC}"
echo -e "${GREEN}================================${NC}"
echo ""
echo -e "${YELLOW}Default admin user credentials:${NC}"
echo "Email: admin@marimo.com"
echo "Password: admin123"
echo ""
echo -e "${RED}⚠️  IMPORTANT: Change the admin password immediately in production!${NC}"
