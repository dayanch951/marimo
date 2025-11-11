#!/bin/bash
# Database Reset Script for Marimo ERP
# This script drops and recreates the database (USE WITH CAUTION!)

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Load environment variables
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
elif [ -f .env.development ]; then
    export $(cat .env.development | grep -v '^#' | xargs)
fi

# Database connection parameters
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-postgres}"
DB_PASSWORD="${DB_PASSWORD:-postgres}"
DB_NAME="${DB_NAME:-marimo_dev}"

echo -e "${RED}================================${NC}"
echo -e "${RED}⚠️  DATABASE RESET WARNING ⚠️${NC}"
echo -e "${RED}================================${NC}"
echo ""
echo "This will DROP and RECREATE the database:"
echo "Database: ${DB_NAME}"
echo "Host: ${DB_HOST}:${DB_PORT}"
echo ""
echo -e "${YELLOW}ALL DATA WILL BE LOST!${NC}"
echo ""
read -p "Are you sure you want to continue? (type 'yes' to proceed): " -r
echo

if [[ ! $REPLY =~ ^yes$ ]]; then
    echo -e "${GREEN}Reset cancelled.${NC}"
    exit 0
fi

echo -e "${YELLOW}Dropping database ${DB_NAME}...${NC}"
PGPASSWORD=$DB_PASSWORD psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -c "DROP DATABASE IF EXISTS $DB_NAME;" || {
    echo -e "${RED}Failed to drop database${NC}"
    exit 1
}
echo -e "${GREEN}✓ Database dropped${NC}"

echo -e "${YELLOW}Creating database ${DB_NAME}...${NC}"
PGPASSWORD=$DB_PASSWORD psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -c "CREATE DATABASE $DB_NAME;" || {
    echo -e "${RED}Failed to create database${NC}"
    exit 1
}
echo -e "${GREEN}✓ Database created${NC}"

echo ""
echo -e "${GREEN}Running initialization script...${NC}"
./scripts/init-db.sh

echo ""
echo -e "${GREEN}Database reset completed!${NC}"
