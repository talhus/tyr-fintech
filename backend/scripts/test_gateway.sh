#!/usr/bin/env bash

# ==============================================================================
# Foodeli <-> TyrFintech B2B Payment Gateway End-to-End Test Suite
# Tests:
#   1. Successful Charge (200 OK)
#   2. Idempotency Replay (Cached Response, No Double-Charge)
#   3. Card Decline: Insufficient Funds (402 Payment Required)
#   4. Card Decline: Card Expired (400 Bad Request)
#   5. Authentication Rejection: Invalid/Missing API Key (401 Unauthorized)
# ==============================================================================

BASE_URL="${TYR_BASE_URL:-http://localhost:8081}"
API_KEY="${TYR_API_KEY:-tyr_live_foodeli_secret_key_12345}"

GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${BLUE}================================================================${NC}"
echo -e "${BLUE} Starting TyrFintech Gateway Verification for Merchant Foodeli  ${NC}"
echo -e "${BLUE} Base URL : ${BASE_URL}${NC}"
echo -e "${BLUE} API Key  : ${API_KEY}${NC}"
echo -e "${BLUE}================================================================${NC}\n"

# ------------------------------------------------------------------------------
# Test 1: Successful Payment (100.00 TRY -> 10000 minor units)
# ------------------------------------------------------------------------------
IDEMP_KEY="idemp_$(date +%s)_$RANDOM"
echo -e "${YELLOW}[TEST 1] Testing Successful Charge (Card ending in 0000)...${NC}"
RESP1=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X POST "${BASE_URL}/api/v1/charges" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${API_KEY}" \
  -H "Idempotency-Key: ${IDEMP_KEY}" \
  -d '{
    "amount": 10000,
    "currency": "TRY",
    "card_number": "4111111111110000",
    "card_holder_name": "John Doe",
    "expire_month": "12",
    "expire_year": "28",
    "cvv": "123"
  }')

BODY1=$(echo "$RESP1" | sed -e '$d')
STATUS1=$(echo "$RESP1" | tail -n1 | cut -d':' -f2)

echo "Response Body : $BODY1"
echo "HTTP Status   : $STATUS1"

if [ "$STATUS1" -eq 200 ] && [[ "$BODY1" == *"SUCCEEDED"* ]]; then
  echo -e "${GREEN}✓ Test 1 PASSED: Payment succeeded with HTTP 200${NC}\n"
else
  echo -e "${RED}✗ Test 1 FAILED${NC}\n"
fi

# ------------------------------------------------------------------------------
# Test 2: Idempotent Replay (Re-send exact same Idempotency-Key)
# ------------------------------------------------------------------------------
echo -e "${YELLOW}[TEST 2] Testing Idempotency Replay with same key (${IDEMP_KEY})...${NC}"
RESP2=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X POST "${BASE_URL}/api/v1/charges" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${API_KEY}" \
  -H "Idempotency-Key: ${IDEMP_KEY}" \
  -d '{
    "amount": 10000,
    "currency": "TRY",
    "card_number": "4111111111110000",
    "card_holder_name": "John Doe",
    "expire_month": "12",
    "expire_year": "28",
    "cvv": "123"
  }')

BODY2=$(echo "$RESP2" | sed -e '$d')
STATUS2=$(echo "$RESP2" | tail -n1 | cut -d':' -f2)

echo "Response Body : $BODY2"
echo "HTTP Status   : $STATUS2"

if [ "$STATUS2" -eq 200 ] && [ "$BODY1" == "$BODY2" ]; then
  echo -e "${GREEN}✓ Test 2 PASSED: Idempotent cached response returned identical transaction_id${NC}\n"
else
  echo -e "${RED}✗ Test 2 FAILED: Response was not idempotent${NC}\n"
fi

# ------------------------------------------------------------------------------
# Test 3: Card Decline - Insufficient Funds (Card ending in 5001)
# ------------------------------------------------------------------------------
echo -e "${YELLOW}[TEST 3] Testing Card Decline: Insufficient Funds (Card ending in 5001)...${NC}"
RESP3=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X POST "${BASE_URL}/api/v1/charges" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${API_KEY}" \
  -d '{
    "amount": 50000,
    "currency": "TRY",
    "card_number": "4111111111115001",
    "card_holder_name": "Jane Doe",
    "expire_month": "11",
    "expire_year": "27",
    "cvv": "456"
  }')

BODY3=$(echo "$RESP3" | sed -e '$d')
STATUS3=$(echo "$RESP3" | tail -n1 | cut -d':' -f2)

echo "Response Body : $BODY3"
echo "HTTP Status   : $STATUS3"

if [ "$STATUS3" -eq 402 ] && [[ "$BODY3" == *"Insufficient funds"* ]]; then
  echo -e "${GREEN}✓ Test 3 PASSED: Expected HTTP 402 with 'Insufficient funds' failure${NC}\n"
else
  echo -e "${RED}✗ Test 3 FAILED${NC}\n"
fi

# ------------------------------------------------------------------------------
# Test 4: Card Decline - Card Expired (Card ending in 5002)
# ------------------------------------------------------------------------------
echo -e "${YELLOW}[TEST 4] Testing Card Decline: Card Expired (Card ending in 5002)...${NC}"
RESP4=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X POST "${BASE_URL}/api/v1/charges" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${API_KEY}" \
  -d '{
    "amount": 10000,
    "currency": "TRY",
    "card_number": "4111111111115002",
    "card_holder_name": "Old Card Holder",
    "expire_month": "12",
    "expire_year": "28",
    "cvv": "789"
  }')

BODY4=$(echo "$RESP4" | sed -e '$d')
STATUS4=$(echo "$RESP4" | tail -n1 | cut -d':' -f2)

echo "Response Body : $BODY4"
echo "HTTP Status   : $STATUS4"

if [ "$STATUS4" -eq 400 ] && [[ "$BODY4" == *"Card expired"* ]]; then
  echo -e "${GREEN}✓ Test 4 PASSED: Expected HTTP 400 with 'Card expired' failure${NC}\n"
else
  echo -e "${RED}✗ Test 4 FAILED${NC}\n"
fi

# ------------------------------------------------------------------------------
# Test 5: Authentication Rejection - Invalid API Key
# ------------------------------------------------------------------------------
echo -e "${YELLOW}[TEST 5] Testing Authentication Failure (Invalid API Key)...${NC}"
RESP5=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X POST "${BASE_URL}/api/v1/charges" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer tyr_live_invalid_unauthorized_key" \
  -d '{
    "amount": 10000,
    "currency": "TRY",
    "card_number": "4111111111110000",
    "card_holder_name": "Intruder",
    "expire_month": "12",
    "expire_year": "28",
    "cvv": "123"
  }')

BODY5=$(echo "$RESP5" | sed -e '$d')
STATUS5=$(echo "$RESP5" | tail -n1 | cut -d':' -f2)

echo "Response Body : $BODY5"
echo "HTTP Status   : $STATUS5"

if [ "$STATUS5" -eq 401 ]; then
  echo -e "${GREEN}✓ Test 5 PASSED: Expected HTTP 401 Unauthorized${NC}\n"
else
  echo -e "${RED}✗ Test 5 FAILED${NC}\n"
fi

echo -e "${BLUE}================================================================${NC}"
echo -e "${BLUE}             All Gateway Tests Executed                         ${NC}"
echo -e "${BLUE}================================================================${NC}"
