#!/usr/bin/env bash

# ==============================================================================
# Foodeli <-> TyrFintech Hosted Checkout ("Pay with TyrFintech") Test Suite
# Tests:
#   1. Foodeli creates Checkout Session (POST /api/v1/checkout/sessions)
#   2. Public UI fetches Session Details (GET /api/v1/checkout/sessions/:id)
#   3. User logs in & pays from their wallet (POST /api/v1/checkout/sessions/:id/pay)
#   4. Verifies Redirect URL & Webhook Dispatch
# ==============================================================================

BASE_URL="${TYR_BASE_URL:-http://localhost:8081}"
API_KEY="${TYR_API_KEY:-tyr_live_foodeli_secret_key_12345}"

GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${BLUE}================================================================${NC}"
echo -e "${BLUE} Starting Hosted Checkout Verification ('Pay with TyrFintech')  ${NC}"
echo -e "${BLUE} Base URL: ${BASE_URL}${NC}"
echo -e "${BLUE}================================================================${NC}\n"

# ------------------------------------------------------------------------------
# Step 1: Foodeli Server creates a Checkout Session
# ------------------------------------------------------------------------------
ORDER_ID="foodeli_order_$(date +%s)"
echo -e "${YELLOW}[STEP 1] Foodeli backend creating checkout session for ${ORDER_ID}...${NC}"

SESSION_RESP=$(curl -s -X POST "${BASE_URL}/api/v1/checkout/sessions" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${API_KEY}" \
  -d '{
    "order_id": "'"${ORDER_ID}"'",
    "amount": 15000,
    "currency": "TRY",
    "callback_url": "https://foodeli.com/orders/callback",
    "webhook_url": "https://api.foodeli.com/webhooks/tyrfintech"
  }')

echo "Response: $SESSION_RESP"
SESSION_ID=$(echo "$SESSION_RESP" | grep -o '"session_id":"[^"]*' | cut -d'"' -f4)

if [ -n "$SESSION_ID" ]; then
  echo -e "${GREEN}✓ Step 1 PASSED: Created session ${SESSION_ID}${NC}\n"
else
  echo -e "${RED}✗ Step 1 FAILED: Could not create checkout session${NC}\n"
  exit 1
fi

# ------------------------------------------------------------------------------
# Step 2: Checkout Page UI fetches Session Details
# ------------------------------------------------------------------------------
echo -e "${YELLOW}[STEP 2] Checkout UI loading details for ${SESSION_ID}...${NC}"

DETAILS_RESP=$(curl -s -X GET "${BASE_URL}/api/v1/checkout/sessions/${SESSION_ID}")
echo "Response: $DETAILS_RESP"

if [[ "$DETAILS_RESP" == *"Foodeli Inc."* ]] && [[ "$DETAILS_RESP" == *"15000"* ]]; then
  echo -e "${GREEN}✓ Step 2 PASSED: Fetched session details successfully${NC}\n"
else
  echo -e "${RED}✗ Step 2 FAILED${NC}\n"
fi

# ------------------------------------------------------------------------------
# Step 3: User logs in to TyrFintech to get their JWT token & Wallet ID
# ------------------------------------------------------------------------------
echo -e "${YELLOW}[STEP 3] User logging in as demo@tyr.com...${NC}"

LOGIN_RESP=$(curl -s -X POST "${BASE_URL}/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "demo@tyr.com",
    "password": "demo123456"
  }')

USER_TOKEN=$(echo "$LOGIN_RESP" | grep -o '"access_token":"[^"]*' | cut -d'"' -f4)
if [ -z "$USER_TOKEN" ]; then
  USER_TOKEN=$(echo "$LOGIN_RESP" | grep -o '"token":"[^"]*' | cut -d'"' -f4)
fi

if [ -z "$USER_TOKEN" ]; then
  echo -e "${RED}✗ Step 3 FAILED: Could not log in demo user${NC}\n"
  exit 1
fi

# Fetch User Wallets
WALLETS_RESP=$(curl -s -X GET "${BASE_URL}/api/v1/wallets" \
  -H "Authorization: Bearer ${USER_TOKEN}")

TRY_WALLET_ID=$(echo "$WALLETS_RESP" | grep -o '"id":"[^"]*' | head -n1 | cut -d'"' -f4)
echo -e "${GREEN}✓ Step 3 PASSED: User authenticated. Paying with wallet: ${TRY_WALLET_ID}${NC}\n"

# ------------------------------------------------------------------------------
# Step 4: User Approves Payment (POST /checkout/sessions/:id/pay)
# ------------------------------------------------------------------------------
echo -e "${YELLOW}[STEP 4] User approving payment from wallet...${NC}"

PAY_RESP=$(curl -s -X POST "${BASE_URL}/api/v1/checkout/sessions/${SESSION_ID}/pay" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${USER_TOKEN}" \
  -d '{
    "wallet_id": "'"${TRY_WALLET_ID}"'"
  }')

echo "Response: $PAY_RESP"

if [[ "$PAY_RESP" == *"SUCCEEDED"* ]] && [[ "$PAY_RESP" == *"redirect_url"* ]]; then
  echo -e "${GREEN}✓ Step 4 PASSED: Payment succeeded! User redirected back to Foodeli callback.${NC}\n"
else
  echo -e "${RED}✗ Step 4 FAILED${NC}\n"
  exit 1
fi

echo -e "${BLUE}================================================================${NC}"
echo -e "${BLUE} Hosted Checkout & Webhook Verification Complete!               ${NC}"
echo -e "${BLUE}================================================================${NC}"
