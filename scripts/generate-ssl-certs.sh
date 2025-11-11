#!/bin/bash
# Generate self-signed SSL certificates for development
# For production, use Let's Encrypt with certbot

set -e

CERT_DIR="./nginx/ssl"
DAYS_VALID=365

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${YELLOW}Generating self-signed SSL certificates for development...${NC}"
echo ""

# Create directory if it doesn't exist
mkdir -p "$CERT_DIR"

# Generate private key
openssl genrsa -out "$CERT_DIR/key.pem" 2048

# Generate certificate
openssl req -new -x509 -sha256 -key "$CERT_DIR/key.pem" -out "$CERT_DIR/cert.pem" -days "$DAYS_VALID" \
    -subj "/C=US/ST=State/L=City/O=Marimo ERP/OU=Development/CN=localhost"

echo ""
echo -e "${GREEN}✓ SSL certificates generated successfully!${NC}"
echo ""
echo "Location: $CERT_DIR"
echo "  - Certificate: cert.pem"
echo "  - Private Key: key.pem"
echo "  - Valid for: $DAYS_VALID days"
echo ""
echo -e "${YELLOW}Note: These are self-signed certificates for development only.${NC}"
echo -e "${YELLOW}For production, use Let's Encrypt (see docs/HTTPS_SETUP.md)${NC}"
