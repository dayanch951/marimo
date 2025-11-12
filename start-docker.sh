#!/bin/bash

# Marimo ERP - Docker Startup Script
# This script helps you easily start the entire Marimo ERP system with Docker Compose

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$SCRIPT_DIR"

echo -e "${BLUE}╔═══════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║           Marimo ERP - Docker Startup Script             ║${NC}"
echo -e "${BLUE}╚═══════════════════════════════════════════════════════════╝${NC}"
echo ""

# Check if .env exists
if [ ! -f .env ]; then
    echo -e "${YELLOW}⚠ .env file not found. Creating from .env.example...${NC}"
    cp .env.example .env
    echo -e "${GREEN}✓ .env file created. Please review and update if needed.${NC}"
fi

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo -e "${RED}✗ Docker is not running. Please start Docker and try again.${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Docker is running${NC}"

# Parse command line arguments
COMMAND=${1:-up}
BUILD_FLAG=""
DETACH_FLAG=""

case "$COMMAND" in
    up)
        echo -e "${BLUE}Starting all services...${NC}"
        BUILD_FLAG="--build"
        DETACH_FLAG="-d"
        ;;
    up-foreground|upf)
        echo -e "${BLUE}Starting all services in foreground...${NC}"
        BUILD_FLAG="--build"
        DETACH_FLAG=""
        ;;
    down)
        echo -e "${YELLOW}Stopping all services...${NC}"
        docker-compose down
        echo -e "${GREEN}✓ All services stopped${NC}"
        exit 0
        ;;
    restart)
        echo -e "${YELLOW}Restarting all services...${NC}"
        docker-compose down
        BUILD_FLAG="--build"
        DETACH_FLAG="-d"
        ;;
    logs)
        echo -e "${BLUE}Showing logs (Ctrl+C to exit)...${NC}"
        docker-compose logs -f ${2}
        exit 0
        ;;
    ps|status)
        echo -e "${BLUE}Service Status:${NC}"
        docker-compose ps
        exit 0
        ;;
    clean)
        echo -e "${YELLOW}⚠ This will remove all containers, volumes, and images!${NC}"
        read -p "Are you sure? (yes/no): " -r
        if [[ $REPLY =~ ^[Yy][Ee][Ss]$ ]]; then
            docker-compose down -v --rmi all
            echo -e "${GREEN}✓ Cleanup complete${NC}"
        else
            echo -e "${BLUE}Cleanup cancelled${NC}"
        fi
        exit 0
        ;;
    help|--help|-h)
        echo "Usage: ./start-docker.sh [COMMAND]"
        echo ""
        echo "Commands:"
        echo "  up               - Start all services in detached mode (default)"
        echo "  upf, up-foreground - Start all services in foreground"
        echo "  down             - Stop all services"
        echo "  restart          - Restart all services"
        echo "  logs [service]   - Show logs (optionally for specific service)"
        echo "  ps, status       - Show status of all services"
        echo "  clean            - Remove all containers, volumes, and images"
        echo "  help             - Show this help message"
        echo ""
        echo "Examples:"
        echo "  ./start-docker.sh up          # Start all services"
        echo "  ./start-docker.sh logs gateway # Show gateway logs"
        echo "  ./start-docker.sh down        # Stop all services"
        exit 0
        ;;
    *)
        echo -e "${RED}✗ Unknown command: $COMMAND${NC}"
        echo "Run './start-docker.sh help' for usage information"
        exit 1
        ;;
esac

# Start services
echo ""
echo -e "${BLUE}Building and starting containers...${NC}"
docker-compose up $BUILD_FLAG $DETACH_FLAG

if [ "$DETACH_FLAG" = "-d" ]; then
    echo ""
    echo -e "${GREEN}✓ All services started successfully!${NC}"
    echo ""
    echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
    echo -e "${GREEN}Services are running:${NC}"
    echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
    echo ""
    echo -e "  ${GREEN}Frontend:${NC}         http://localhost:3000"
    echo -e "  ${GREEN}API Gateway:${NC}      http://localhost:8080"
    echo -e "  ${GREEN}Users Service:${NC}    http://localhost:8081"
    echo -e "  ${GREEN}Config Service:${NC}   http://localhost:8082"
    echo -e "  ${GREEN}Accounting:${NC}       http://localhost:8083"
    echo -e "  ${GREEN}Factory:${NC}          http://localhost:8084"
    echo -e "  ${GREEN}Shop:${NC}             http://localhost:8085"
    echo -e "  ${GREEN}Main Service:${NC}     http://localhost:8086"
    echo ""
    echo -e "  ${YELLOW}Consul UI:${NC}        http://localhost:8500"
    echo -e "  ${YELLOW}RabbitMQ UI:${NC}      http://localhost:15672"
    echo ""
    echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
    echo -e "${GREEN}Useful commands:${NC}"
    echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
    echo ""
    echo -e "  ./start-docker.sh logs           # View all logs"
    echo -e "  ./start-docker.sh logs gateway   # View gateway logs"
    echo -e "  ./start-docker.sh ps             # Check service status"
    echo -e "  ./start-docker.sh down           # Stop all services"
    echo ""
    echo -e "${YELLOW}Default credentials:${NC}"
    echo -e "  Email: admin@example.com"
    echo -e "  Password: admin123"
    echo ""
    echo -e "${YELLOW}RabbitMQ Management:${NC}"
    echo -e "  Username: admin"
    echo -e "  Password: admin"
    echo ""
fi
