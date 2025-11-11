#!/bin/bash

# Database restore script for PostgreSQL
# Usage: ./restore.sh <backup-file> [namespace]

set -e

BACKUP_FILE=$1
NAMESPACE=${2:-marimo-erp}

if [ -z "$BACKUP_FILE" ]; then
    echo "Usage: $0 <backup-file> [namespace]"
    echo "Example: $0 /backups/marimo_backup_20240115_120000.dump.gz marimo-erp"
    exit 1
fi

if [ ! -f "$BACKUP_FILE" ]; then
    echo "Error: Backup file not found: $BACKUP_FILE"
    exit 1
fi

echo "=========================================="
echo "PostgreSQL Restore"
echo "Backup file: $BACKUP_FILE"
echo "Namespace: $NAMESPACE"
echo "=========================================="

# Warning
echo ""
echo "⚠️  WARNING: This will restore the database from backup!"
echo "⚠️  All current data will be REPLACED!"
echo ""
read -p "Are you sure you want to continue? (type 'yes' to confirm): " CONFIRM

if [ "$CONFIRM" != "yes" ]; then
    echo "Restore cancelled."
    exit 0
fi

# Verify checksum if available
CHECKSUM_FILE="${BACKUP_FILE%.gz}.sha256"
if [ -f "$CHECKSUM_FILE" ]; then
    echo ""
    echo "Verifying backup integrity..."
    if cd "$(dirname "$BACKUP_FILE")" && sha256sum -c "$(basename "$CHECKSUM_FILE")" > /dev/null 2>&1; then
        echo "✓ Backup integrity verified"
    else
        echo "✗ Backup integrity check failed!"
        read -p "Continue anyway? (yes/no): " FORCE
        if [ "$FORCE" != "yes" ]; then
            exit 1
        fi
    fi
fi

# Get database credentials
DB_USER=$(kubectl get secret marimo-secrets -n $NAMESPACE -o jsonpath='{.data.DB_USER}' | base64 -d)
DB_PASSWORD=$(kubectl get secret marimo-secrets -n $NAMESPACE -o jsonpath='{.data.DB_PASSWORD}' | base64 -d)
DB_NAME=$(kubectl get configmap marimo-config -n $NAMESPACE -o jsonpath='{.data.DB_NAME}')

# Get postgres pod name
POSTGRES_POD=$(kubectl get pods -n $NAMESPACE -l app=postgres -o jsonpath='{.items[0].metadata.name}')

if [ -z "$POSTGRES_POD" ]; then
    echo "Error: PostgreSQL pod not found in namespace $NAMESPACE"
    exit 1
fi

echo "Using pod: $POSTGRES_POD"

# Scale down services to prevent connections
echo ""
echo "Scaling down services..."
kubectl scale deployment --all --replicas=0 -n $NAMESPACE --selector='app!=postgres'
echo "✓ Services scaled down"

# Wait a moment for connections to close
sleep 5

# Terminate existing connections
echo ""
echo "Terminating existing database connections..."
kubectl exec -n $NAMESPACE $POSTGRES_POD -- bash -c \
    "PGPASSWORD='$DB_PASSWORD' psql -U $DB_USER -d postgres -c \"SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = '$DB_NAME' AND pid <> pg_backend_pid();\""

# Drop and recreate database
echo ""
echo "Recreating database..."
kubectl exec -n $NAMESPACE $POSTGRES_POD -- bash -c \
    "PGPASSWORD='$DB_PASSWORD' psql -U $DB_USER -d postgres -c 'DROP DATABASE IF EXISTS $DB_NAME;'"
kubectl exec -n $NAMESPACE $POSTGRES_POD -- bash -c \
    "PGPASSWORD='$DB_PASSWORD' psql -U $DB_USER -d postgres -c 'CREATE DATABASE $DB_NAME;'"
echo "✓ Database recreated"

# Decompress and restore
echo ""
echo "Restoring backup..."
if [[ $BACKUP_FILE == *.gz ]]; then
    gunzip -c $BACKUP_FILE | kubectl exec -i -n $NAMESPACE $POSTGRES_POD -- bash -c \
        "PGPASSWORD='$DB_PASSWORD' pg_restore -U $DB_USER -d $DB_NAME --no-owner --no-acl"
else
    kubectl exec -i -n $NAMESPACE $POSTGRES_POD -- bash -c \
        "PGPASSWORD='$DB_PASSWORD' pg_restore -U $DB_USER -d $DB_NAME --no-owner --no-acl" < $BACKUP_FILE
fi

if [ $? -eq 0 ]; then
    echo "✓ Backup restored successfully"
else
    echo "✗ Restore failed"
    echo "Database may be in an inconsistent state!"
    exit 1
fi

# Verify restore
echo ""
echo "Verifying restore..."
TABLE_COUNT=$(kubectl exec -n $NAMESPACE $POSTGRES_POD -- bash -c \
    "PGPASSWORD='$DB_PASSWORD' psql -U $DB_USER -d $DB_NAME -t -c \"SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public';\"" | tr -d ' ')

echo "Tables restored: $TABLE_COUNT"

# Scale up services
echo ""
echo "Scaling up services..."
kubectl scale deployment api-gateway --replicas=3 -n $NAMESPACE
kubectl scale deployment auth-service --replicas=2 -n $NAMESPACE
echo "✓ Services scaled up"

# Wait for services to be ready
echo ""
echo "Waiting for services to be ready..."
kubectl wait --for=condition=available --timeout=180s deployment/api-gateway -n $NAMESPACE
kubectl wait --for=condition=available --timeout=180s deployment/auth-service -n $NAMESPACE

echo ""
echo "=========================================="
echo "✓ Restore completed successfully!"
echo "Tables restored: $TABLE_COUNT"
echo "=========================================="
