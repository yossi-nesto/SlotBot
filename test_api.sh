#!/bin/bash

# SlotBot API Test Script
# This script tests the SlotBot endpoints locally

BASE_URL="http://localhost:8080"

echo "🧪 SlotBot API Test Script"
echo "=========================="
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test 1: Health Check
echo -e "${YELLOW}Test 1: Health Check${NC}"
response=$(curl -s -w "\n%{http_code}" "$BASE_URL/health")
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

if [ "$http_code" = "200" ]; then
    echo -e "${GREEN}✓ Health check passed${NC}"
    echo "Response: $body"
else
    echo -e "${RED}✗ Health check failed (HTTP $http_code)${NC}"
fi
echo ""

# Test 2: Book Environment (without signature - will fail with 401)
echo -e "${YELLOW}Test 2: Book Environment (Expected to fail - no signature)${NC}"
response=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/slack/book" \
    -d "text=staging auth PROJ-123" \
    -d "user_name=testuser")
http_code=$(echo "$response" | tail -n1)

if [ "$http_code" = "401" ]; then
    echo -e "${GREEN}✓ Correctly rejected request without signature${NC}"
else
    echo -e "${RED}✗ Unexpected response (HTTP $http_code)${NC}"
fi
echo ""

# Test 3: Find Next Slot (without signature - will fail with 401)
echo -e "${YELLOW}Test 3: Find Next Slot (Expected to fail - no signature)${NC}"
response=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/slack/next" \
    -d "text=staging auth")
http_code=$(echo "$response" | tail -n1)

if [ "$http_code" = "401" ]; then
    echo -e "${GREEN}✓ Correctly rejected request without signature${NC}"
else
    echo -e "${RED}✗ Unexpected response (HTTP $http_code)${NC}"
fi
echo ""

echo "=========================="
echo -e "${YELLOW}Note:${NC} Slack endpoints require valid signatures."
echo "To test with real Slack, use ngrok and configure your Slack app."
echo ""
echo "Example ngrok setup:"
echo "  1. Run: ngrok http 8080"
echo "  2. Copy the HTTPS URL"
echo "  3. Update Slack App slash commands to point to:"
echo "     - /env-book → https://YOUR_NGROK_URL/slack/book"
echo "     - /env-next → https://YOUR_NGROK_URL/slack/next"
echo "     - /env-bookings → https://YOUR_NGROK_URL/slack/bookings"
