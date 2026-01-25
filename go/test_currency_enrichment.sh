#!/bin/bash

BASE_URL="http://localhost:8080"

GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${BLUE}=== Testing Currency Enrichment API ===${NC}\n"

rm -f /tmp/test_currency_*.csv 2>/dev/null
CSV_FILE=$(mktemp /tmp/test_currency_XXXXXX.csv)
cat >"$CSV_FILE" <<'EOF'
code,name,symbol,country
USD,United States Dollar,$,United States
EUR,Euro,€,Eurozone
GBP,British Pound Sterling,£,United Kingdom
JPY,Japanese Yen,¥,Japan
BGN,Bulgarian Lev,лв,Bulgaria
CHF,Swiss Franc,CHF,Switzerland
AUD,Australian Dollar,A$,Australia
CAD,Canadian Dollar,C$,Canada
EOF

echo -e "${GREEN}1. Created test CSV file:${NC}"
cat "$CSV_FILE"
echo ""

CSV_SIZE=$(wc -c <"$CSV_FILE" | tr -d ' ')
echo -e "${GREEN}2. Requesting signed URL for upload (size: $CSV_SIZE bytes)${NC}"
SIGNED_URL_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/enrichment-signed-url" \
  -H "Content-Type: application/json" \
  -d "{
    \"contentType\": \"text/csv\",
    \"length\": $CSV_SIZE
  }")
echo "$SIGNED_URL_RESPONSE" | jq .
SIGNED_URL=$(echo "$SIGNED_URL_RESPONSE" | jq -r '.url')
JOB_ID=$(echo "$SIGNED_URL_RESPONSE" | jq -r '.jobId')

if [ -z "$JOB_ID" ] || [ "$JOB_ID" = "null" ]; then
  echo -e "${RED}Failed to get job ID${NC}"
  rm -f "$CSV_FILE"
  exit 1
fi

echo -e "\n${YELLOW}Job ID: $JOB_ID${NC}\n"

echo -e "${GREEN}3. Uploading CSV file to GCS${NC}"
UPLOAD_RESPONSE=$(curl -s -w "\n%{http_code}" -X PUT "$SIGNED_URL" \
  -H "Content-Type: text/csv" \
  --data-binary "@$CSV_FILE")
UPLOAD_HTTP_CODE=$(echo "$UPLOAD_RESPONSE" | tail -n1)

if [ "$UPLOAD_HTTP_CODE" = "200" ]; then
  echo -e "${GREEN}Upload successful (HTTP $UPLOAD_HTTP_CODE)${NC}"
else
  echo -e "${RED}Upload failed (HTTP $UPLOAD_HTTP_CODE)${NC}"
  echo "$UPLOAD_RESPONSE"
  rm -f "$CSV_FILE"
  exit 1
fi

rm -f "$CSV_FILE"

echo -e "\n${GREEN}4. Listing user jobs${NC}"
curl -s -X GET "$BASE_URL/api/v1/jobs" | jq .

echo -e "\n${GREEN}5. Starting enrichment job${NC}"
START_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/jobs/$JOB_ID/start" \
  -H "Content-Type: application/json" \
  -d '{
    "key_column": "code",
    "columns_metadata": [
      {
        "name": "is_active",
        "type": "boolean",
        "description": "Whether the currency is currently active and in circulation"
      },
      {
        "name": "value_in_usd",
        "type": "number",
        "description": "Current exchange rate value of 1 unit of this currency in USD"
      },
      {
        "name": "value_in_eur",
        "type": "number",
        "description": "Current exchange rate value of 1 unit of this currency in EUR"
      }
    ],
    "entity_type": "currency"
  }')

echo "$START_RESPONSE" | jq .

echo -e "\n${GREEN}6. Checking job progress${NC}"
curl -s -X GET "$BASE_URL/api/v1/jobs/$JOB_ID/progress" | jq .

echo -e "\n${GREEN}7. Waiting 20 seconds before checking progress again...${NC}"
sleep 20

echo -e "${GREEN}8. Checking job progress again${NC}"
curl -s -X GET "$BASE_URL/api/v1/jobs/$JOB_ID/progress" | jq .

echo -e "\n${GREEN}9. Getting job results${NC}"
curl -s -X GET "$BASE_URL/api/v1/jobs/$JOB_ID/results" | jq .

echo -e "\n${BLUE}=== Currency Enrichment Test completed ===${NC}"
