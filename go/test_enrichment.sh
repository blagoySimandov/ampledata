#!/bin/bash

BASE_URL="http://localhost:8080"

GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${BLUE}=== Testing Enrichment API (GCS File-Based Workflow) ===${NC}\n"

CSV_FILE=$(mktemp /tmp/test_enrichment_XXXXXX.csv)
cat >"$CSV_FILE" <<'EOF'
company_name,industry,country
google,technology,USA
apple,technology,USA
qualcomm,semiconductors,USA
mapletree,real estate,Singapore
microsoft,technology,USA
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
echo "$SIGNED_URL_RESPONSE"
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
    "key_column": "company_name",
    "columns_metadata": [
      {
        "name": "founder_picture_url",
        "type": "string",
        "description": "picture of the founder in url format"
      },
      {
        "name": "founder",
        "type": "string",
        "description": "Founder name"
      },
      {
        "name": "website",
        "type": "string",
        "description": "Company website URL"
      },
      {
        "name": "employee_count",
        "type": "number",
        "description": "Number of employees"
      }
    ],
    "entity_type": "company"
  }')

echo "$START_RESPONSE" | jq .

echo -e "\n${GREEN}6. Checking job progress${NC}"
curl -s -X GET "$BASE_URL/api/v1/jobs/$JOB_ID/progress" | jq .

echo -e "\n${GREEN}7. Waiting 10 seconds before checking progress again...${NC}"
sleep 10

echo -e "${GREEN}8. Checking job progress again${NC}"
curl -s -X GET "$BASE_URL/api/v1/jobs/$JOB_ID/progress" | jq .

echo -e "\n${GREEN}9. Getting job results${NC}"
curl -s -X GET "$BASE_URL/api/v1/jobs/$JOB_ID/results" | jq .

echo -e "\n${BLUE}=== Starting another job for cancellation test ===${NC}\n"

CSV_FILE2=$(mktemp /tmp/test_enrichment_XXXXXX.csv)
cat >"$CSV_FILE2" <<'EOF'
person_name,role
person_1,developer
person_2,designer
EOF

CSV_SIZE2=$(wc -c <"$CSV_FILE2" | tr -d ' ')
echo -e "${GREEN}10. Requesting signed URL for second upload${NC}"
SIGNED_URL_RESPONSE2=$(curl -s -X POST "$BASE_URL/api/v1/enrichment-signed-url" \
  -H "Content-Type: application/json" \
  -d "{
    \"contentType\": \"text/csv\",
    \"length\": $CSV_SIZE2
  }")

echo "$SIGNED_URL_RESPONSE2" | jq .
SIGNED_URL2=$(echo "$SIGNED_URL_RESPONSE2" | jq -r '.url')
JOB_ID2=$(echo "$SIGNED_URL_RESPONSE2" | jq -r '.jobId')

echo -e "\n${YELLOW}Job ID: $JOB_ID2${NC}\n"

echo -e "${GREEN}11. Uploading second CSV file to GCS${NC}"
curl -s -X PUT "$SIGNED_URL2" \
  -H "Content-Type: text/csv" \
  --data-binary "@$CSV_FILE2"
rm -f "$CSV_FILE2"

echo -e "\n${GREEN}12. Starting second enrichment job${NC}"
curl -s -X POST "$BASE_URL/api/v1/jobs/$JOB_ID2/start" \
  -H "Content-Type: application/json" \
  -d '{
    "key_column": "person_name",
    "columns_metadata": [
      {
        "name": "full_name",
        "type": "string"
      },
      {
        "name": "email",
        "type": "string"
      }
    ]
  }' | jq .

echo -e "\n${GREEN}13. Cancelling job${NC}"
curl -s -X POST "$BASE_URL/api/v1/jobs/$JOB_ID2/cancel" | jq .

echo -e "\n${GREEN}14. Checking cancelled job progress${NC}"
curl -s -X GET "$BASE_URL/api/v1/jobs/$JOB_ID2/progress" | jq .

echo -e "\n${GREEN}15. Listing all user jobs${NC}"
curl -s -X GET "$BASE_URL/api/v1/jobs" | jq .

echo -e "\n${BLUE}=== Test completed ===${NC}"
