#!/bin/bash

# Database backup script for PostgreSQL
# Usage: ./backup.sh [namespace] [backup-dir]

set -e

NAMESPACE=${1:-marimo-erp}
BACKUP_DIR=${2:-/backups}
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_NAME="marimo_backup_${TIMESTAMP}"
RETENTION_DAYS=${RETENTION_DAYS:-30}

echo "=========================================="
echo "PostgreSQL Backup"
echo "Namespace: $NAMESPACE"
echo "Backup directory: $BACKUP_DIR"
echo "Timestamp: $TIMESTAMP"
echo "=========================================="

# Get database credentials from secrets
DB_USER=$(kubectl get secret marimo-secrets -n $NAMESPACE -o jsonpath='{.data.DB_USER}' | base64 -d)
DB_PASSWORD=$(kubectl get secret marimo-secrets -n $NAMESPACE -o jsonpath='{.data.DB_PASSWORD}' | base64 -d)
DB_NAME=$(kubectl get configmap marimo-config -n $NAMESPACE -o jsonpath='{.data.DB_NAME}')
DB_HOST=$(kubectl get configmap marimo-config -n $NAMESPACE -o jsonpath='{.data.DB_HOST}')

# Get postgres pod name
POSTGRES_POD=$(kubectl get pods -n $NAMESPACE -l app=postgres -o jsonpath='{.items[0].metadata.name}')

if [ -z "$POSTGRES_POD" ]; then
    echo "Error: PostgreSQL pod not found in namespace $NAMESPACE"
    exit 1
fi

echo "Using pod: $POSTGRES_POD"

# Create backup directory if it doesn't exist
mkdir -p $BACKUP_DIR

# Perform backup
echo ""
echo "Creating backup..."
kubectl exec -n $NAMESPACE $POSTGRES_POD -- bash -c \
    "PGPASSWORD='$DB_PASSWORD' pg_dump -U $DB_USER -h localhost $DB_NAME -Fc" \
    > ${BACKUP_DIR}/${BACKUP_NAME}.dump

# Check if backup was successful
if [ $? -eq 0 ]; then
    BACKUP_SIZE=$(du -h ${BACKUP_DIR}/${BACKUP_NAME}.dump | cut -f1)
    echo "✓ Backup created successfully: ${BACKUP_NAME}.dump (${BACKUP_SIZE})"
else
    echo "✗ Backup failed"
    exit 1
fi

# Compress backup
echo ""
echo "Compressing backup..."
gzip ${BACKUP_DIR}/${BACKUP_NAME}.dump

if [ $? -eq 0 ]; then
    COMPRESSED_SIZE=$(du -h ${BACKUP_DIR}/${BACKUP_NAME}.dump.gz | cut -f1)
    echo "✓ Backup compressed: ${BACKUP_NAME}.dump.gz (${COMPRESSED_SIZE})"
else
    echo "✗ Compression failed"
    exit 1
fi

# Generate checksum
echo ""
echo "Generating checksum..."
cd $BACKUP_DIR
sha256sum ${BACKUP_NAME}.dump.gz > ${BACKUP_NAME}.sha256
echo "✓ Checksum saved: ${BACKUP_NAME}.sha256"

# Upload to cloud storage (optional - uncomment and configure)
# echo ""
# echo "Uploading to S3..."
# aws s3 cp ${BACKUP_DIR}/${BACKUP_NAME}.dump.gz s3://your-backup-bucket/marimo/postgres/ --storage-class STANDARD_IA
# aws s3 cp ${BACKUP_DIR}/${BACKUP_NAME}.sha256 s3://your-backup-bucket/marimo/postgres/
# echo "✓ Uploaded to S3"

# Clean up old backups
echo ""
echo "Cleaning up old backups (older than $RETENTION_DAYS days)..."
find $BACKUP_DIR -name "marimo_backup_*.dump.gz" -mtime +$RETENTION_DAYS -delete
find $BACKUP_DIR -name "marimo_backup_*.sha256" -mtime +$RETENTION_DAYS -delete
echo "✓ Old backups cleaned up"

# Create backup metadata
cat > ${BACKUP_DIR}/${BACKUP_NAME}.meta <<EOF
{
  "timestamp": "$TIMESTAMP",
  "database": "$DB_NAME",
  "namespace": "$NAMESPACE",
  "pod": "$POSTGRES_POD",
  "backup_file": "${BACKUP_NAME}.dump.gz",
  "checksum_file": "${BACKUP_NAME}.sha256",
  "size": "$(stat -f%z ${BACKUP_DIR}/${BACKUP_NAME}.dump.gz 2>/dev/null || stat -c%s ${BACKUP_DIR}/${BACKUP_NAME}.dump.gz)"
}
EOF

echo ""
echo "=========================================="
echo "✓ Backup completed successfully!"
echo "Backup file: ${BACKUP_DIR}/${BACKUP_NAME}.dump.gz"
echo "=========================================="

# List recent backups
echo ""
echo "Recent backups:"
ls -lh $BACKUP_DIR/marimo_backup_*.dump.gz | tail -5
