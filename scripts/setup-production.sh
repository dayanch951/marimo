#!/bin/bash

# Marimo ERP - Production Setup Script
# This script helps you set up the production environment

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}╔══════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║       Marimo ERP - Production Environment Setup             ║${NC}"
echo -e "${BLUE}╚══════════════════════════════════════════════════════════════╝${NC}"
echo ""

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo -e "${RED}✗ This script must be run as root${NC}"
    echo "Please run: sudo ./setup-production.sh"
    exit 1
fi

echo -e "${GREEN}✓ Running as root${NC}"

# Step 1: Check prerequisites
echo ""
echo -e "${BLUE}Step 1: Checking prerequisites...${NC}"

command -v docker >/dev/null 2>&1 || {
    echo -e "${RED}✗ Docker is not installed${NC}"
    echo "Install Docker: https://docs.docker.com/engine/install/"
    exit 1
}
echo -e "${GREEN}✓ Docker installed${NC}"

command -v docker-compose >/dev/null 2>&1 || {
    echo -e "${RED}✗ Docker Compose is not installed${NC}"
    echo "Install Docker Compose: https://docs.docker.com/compose/install/"
    exit 1
}
echo -e "${GREEN}✓ Docker Compose installed${NC}"

# Step 2: Create production env file
echo ""
echo -e "${BLUE}Step 2: Setting up environment configuration...${NC}"

if [ ! -f .env.production ]; then
    echo -e "${RED}✗ .env.production file not found${NC}"
    exit 1
fi

if [ ! -f .env ]; then
    cp .env.production .env
    echo -e "${YELLOW}⚠ Created .env from .env.production${NC}"
    echo -e "${YELLOW}⚠ IMPORTANT: Edit .env and replace all CHANGE_ME values!${NC}"
    read -p "Press Enter to open .env in editor..." -r
    ${EDITOR:-nano} .env
fi

# Check for CHANGE_ME values
if grep -q "CHANGE_ME" .env; then
    echo -e "${RED}✗ Found CHANGE_ME placeholders in .env${NC}"
    echo -e "${RED}Please replace all CHANGE_ME values before proceeding${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Environment configuration ready${NC}"

# Step 3: Generate secrets
echo ""
echo -e "${BLUE}Step 3: Generating secure secrets...${NC}"

if ! grep -q "^JWT_SECRET=" .env || grep -q "CHANGE_ME" .env | grep JWT_SECRET; then
    JWT_SECRET=$(openssl rand -base64 48)
    sed -i "s|JWT_SECRET=.*|JWT_SECRET=$JWT_SECRET|" .env
    echo -e "${GREEN}✓ Generated JWT_SECRET${NC}"
fi

if ! grep -q "^DB_PASSWORD=" .env || grep -q "CHANGE_ME" .env | grep DB_PASSWORD; then
    DB_PASSWORD=$(openssl rand -base64 32)
    sed -i "s|DB_PASSWORD=.*|DB_PASSWORD=$DB_PASSWORD|" .env
    echo -e "${GREEN}✓ Generated DB_PASSWORD${NC}"
fi

if ! grep -q "^REDIS_PASSWORD=" .env || grep -q "CHANGE_ME" .env | grep REDIS_PASSWORD; then
    REDIS_PASSWORD=$(openssl rand -base64 24)
    sed -i "s|REDIS_PASSWORD=.*|REDIS_PASSWORD=$REDIS_PASSWORD|" .env
    echo -e "${GREEN}✓ Generated REDIS_PASSWORD${NC}"
fi

# Step 4: SSL Certificates
echo ""
echo -e "${BLUE}Step 4: SSL Certificate setup...${NC}"

read -p "Do you want to set up SSL with Let's Encrypt? (y/n): " -r
if [[ $REPLY =~ ^[Yy]$ ]]; then
    read -p "Enter your domain name: " DOMAIN
    read -p "Enter your email for Let's Encrypt: " EMAIL

    # Install certbot
    if ! command -v certbot >/dev/null 2>&1; then
        echo -e "${YELLOW}Installing certbot...${NC}"
        apt-get update
        apt-get install -y certbot python3-certbot-nginx
    fi

    # Create SSL directory
    mkdir -p ./ssl

    # Get certificate
    echo -e "${YELLOW}Obtaining SSL certificate...${NC}"
    certbot certonly --standalone -d $DOMAIN -d www.$DOMAIN -d api.$DOMAIN \
        --email $EMAIL --agree-tos --non-interactive

    # Copy certificates
    cp /etc/letsencrypt/live/$DOMAIN/fullchain.pem ./ssl/
    cp /etc/letsencrypt/live/$DOMAIN/privkey.pem ./ssl/
    cp /etc/letsencrypt/live/$DOMAIN/chain.pem ./ssl/

    # Set permissions
    chmod 600 ./ssl/privkey.pem
    chmod 644 ./ssl/*.pem

    echo -e "${GREEN}✓ SSL certificates obtained${NC}"

    # Update domain in nginx config
    sed -i "s/yourdomain.com/$DOMAIN/g" ./nginx/sites-enabled/marimo.conf

else
    echo -e "${YELLOW}⚠ Skipping SSL setup. You'll need to configure it manually.${NC}"
fi

# Step 5: Database initialization
echo ""
echo -e "${BLUE}Step 5: Database setup...${NC}"

mkdir -p ./backups
mkdir -p ./migrations

echo -e "${GREEN}✓ Database directories created${NC}"

# Step 6: Create directories
echo ""
echo -e "${BLUE}Step 6: Creating required directories...${NC}"

mkdir -p ./logs
mkdir -p ./uploads
mkdir -p ./ssl/postgres

echo -e "${GREEN}✓ Directories created${NC}"

# Step 7: Set permissions
echo ""
echo -e "${BLUE}Step 7: Setting permissions...${NC}"

chown -R 999:999 ./logs
chmod -R 755 ./scripts
chmod +x ./start-docker.sh

echo -e "${GREEN}✓ Permissions set${NC}"

# Step 8: Build and start services
echo ""
echo -e "${BLUE}Step 8: Building and starting services...${NC}"

read -p "Do you want to start the services now? (y/n): " -r
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo -e "${YELLOW}Building Docker images...${NC}"
    docker-compose -f docker-compose.yml -f docker-compose.production.yml build

    echo -e "${YELLOW}Starting services...${NC}"
    docker-compose -f docker-compose.yml -f docker-compose.production.yml up -d

    echo -e "${GREEN}✓ Services started${NC}"

    # Wait for services to be healthy
    echo ""
    echo -e "${YELLOW}Waiting for services to be healthy...${NC}"
    sleep 10

    # Check service status
    docker-compose -f docker-compose.yml -f docker-compose.production.yml ps
else
    echo -e "${YELLOW}⚠ Services not started. Run manually with:${NC}"
    echo "   docker-compose -f docker-compose.yml -f docker-compose.production.yml up -d"
fi

# Step 9: Setup monitoring
echo ""
echo -e "${BLUE}Step 9: Setting up monitoring (optional)...${NC}"

read -p "Do you want to enable Prometheus monitoring? (y/n): " -r
if [[ $REPLY =~ ^[Yy]$ ]]; then
    docker-compose -f docker-compose.yml -f docker-compose.production.yml -f docker-compose.monitoring.yml up -d
    echo -e "${GREEN}✓ Monitoring enabled${NC}"
fi

# Summary
echo ""
echo -e "${BLUE}╔══════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║                     Setup Complete!                          ║${NC}"
echo -e "${BLUE}╚══════════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${GREEN}✓ Production environment is ready!${NC}"
echo ""
echo -e "${YELLOW}Next steps:${NC}"
echo "1. Review .env file and ensure all values are correct"
echo "2. Configure your DNS to point to this server"
echo "3. Set up regular backups (see ./scripts/backup.sh)"
echo "4. Configure monitoring alerts"
echo "5. Review security settings"
echo ""
echo -e "${YELLOW}Useful commands:${NC}"
echo "  View logs:    docker-compose logs -f"
echo "  Stop:         docker-compose -f docker-compose.yml -f docker-compose.production.yml down"
echo "  Restart:      docker-compose -f docker-compose.yml -f docker-compose.production.yml restart"
echo "  Backup DB:    ./scripts/backup.sh"
echo ""
echo -e "${BLUE}Access your application:${NC}"
echo "  Frontend: https://yourdomain.com"
echo "  API:      https://api.yourdomain.com"
echo "  Consul:   http://localhost:8500"
echo ""
echo -e "${RED}IMPORTANT: Change default admin password immediately!${NC}"
echo ""
