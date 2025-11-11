#!/bin/bash
# E2E Test: Complete Authentication Flow
#
# This test covers the full user journey:
# 1. User registration
# 2. Login with credentials
# 3. Access protected resource
# 4. Refresh token
# 5. Logout

set -e

API_URL="${API_URL:-http://localhost:8080}"
TEST_EMAIL="e2e-test-$(date +%s)@example.com"
TEST_PASSWORD="E2ETestPass123!"
TEST_NAME="E2E Test User"

echo "===================="
echo "E2E Authentication Flow Test"
echo "API URL: $API_URL"
echo "===================="
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Step 1: Register
echo -e "${YELLOW}[1/5] Testing user registration...${NC}"
REGISTER_RESPONSE=$(curl -s -X POST "$API_URL/api/users/register" \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"$TEST_EMAIL\",
    \"password\": \"$TEST_PASSWORD\",
    \"name\": \"$TEST_NAME\"
  }")

if echo "$REGISTER_RESPONSE" | grep -q "\"success\":true"; then
  echo -e "${GREEN}✓ Registration successful${NC}"
else
  echo -e "${RED}✗ Registration failed${NC}"
  echo "$REGISTER_RESPONSE"
  exit 1
fi

# Step 2: Login
echo -e "${YELLOW}[2/5] Testing login...${NC}"
LOGIN_RESPONSE=$(curl -s -X POST "$API_URL/api/users/login" \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"$TEST_EMAIL\",
    \"password\": \"$TEST_PASSWORD\"
  }")

ACCESS_TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"access_token":"[^"]*' | cut -d'"' -f4)
REFRESH_TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"refresh_token":"[^"]*' | cut -d'"' -f4)

if [ -n "$ACCESS_TOKEN" ] && [ -n "$REFRESH_TOKEN" ]; then
  echo -e "${GREEN}✓ Login successful${NC}"
  echo "  Access Token: ${ACCESS_TOKEN:0:50}..."
  echo "  Refresh Token: ${REFRESH_TOKEN:0:50}..."
else
  echo -e "${RED}✗ Login failed${NC}"
  echo "$LOGIN_RESPONSE"
  exit 1
fi

# Step 3: Access Protected Resource
echo -e "${YELLOW}[3/5] Testing protected resource access...${NC}"
PROFILE_RESPONSE=$(curl -s -X GET "$API_URL/api/users/profile" \
  -H "Authorization: Bearer $ACCESS_TOKEN")

if echo "$PROFILE_RESPONSE" | grep -q "$TEST_EMAIL"; then
  echo -e "${GREEN}✓ Protected resource access successful${NC}"
else
  echo -e "${RED}✗ Protected resource access failed${NC}"
  echo "$PROFILE_RESPONSE"
  exit 1
fi

# Step 4: Refresh Token
echo -e "${YELLOW}[4/5] Testing token refresh...${NC}"
REFRESH_RESPONSE=$(curl -s -X POST "$API_URL/api/users/refresh" \
  -H "Content-Type: application/json" \
  -d "{
    \"refresh_token\": \"$REFRESH_TOKEN\"
  }")

NEW_ACCESS_TOKEN=$(echo "$REFRESH_RESPONSE" | grep -o '"access_token":"[^"]*' | cut -d'"' -f4)

if [ -n "$NEW_ACCESS_TOKEN" ]; then
  echo -e "${GREEN}✓ Token refresh successful${NC}"
  echo "  New Access Token: ${NEW_ACCESS_TOKEN:0:50}..."
else
  echo -e "${RED}✗ Token refresh failed${NC}"
  echo "$REFRESH_RESPONSE"
  exit 1
fi

# Step 5: Logout
echo -e "${YELLOW}[5/5] Testing logout...${NC}"
LOGOUT_RESPONSE=$(curl -s -X POST "$API_URL/api/users/logout" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $NEW_ACCESS_TOKEN" \
  -d "{
    \"refresh_token\": \"$REFRESH_TOKEN\"
  }")

if echo "$LOGOUT_RESPONSE" | grep -q "\"success\":true"; then
  echo -e "${GREEN}✓ Logout successful${NC}"
else
  echo -e "${RED}✗ Logout failed${NC}"
  echo "$LOGOUT_RESPONSE"
  exit 1
fi

echo ""
echo -e "${GREEN}===================="
echo "All E2E tests passed!"
echo "====================${NC}"
