#!/bin/bash

BASE_URL="http://localhost:8080"

GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${BLUE}=== Testing Enrichment API ===${NC}\n"

echo -e "${GREEN}1. Starting enrichment job with sample data${NC}"
RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/enrich" \
  -H "Content-Type: application/json" \
  -d '{
    "row_keys": ["google", "apple", "qualcomm","mapletree","microsoft"],
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

echo "$RESPONSE"
echo "$RESPONSE" | jq .
JOB_ID=$(echo "$RESPONSE" | jq -r '.job_id')

if [ -z "$JOB_ID" ] || [ "$JOB_ID" = "null" ]; then
  echo -e "${RED}Failed to get job ID${NC}"
  exit 1
fi

echo -e "\n${YELLOW}Job ID: $JOB_ID${NC}\n"

echo -e "${GREEN}2. Checking job progress${NC}"
curl -s -X GET "$BASE_URL/api/v1/jobs/$JOB_ID/progress" | jq .

echo -e "\n${GREEN}3. Waiting 4 seconds before checking progress again...${NC}"
sleep 10

echo -e "${GREEN}4. Checking job progress again${NC}"
curl -s -X GET "$BASE_URL/api/v1/jobs/$JOB_ID/progress" | jq .

echo -e "\n${GREEN}5. Getting job results${NC}"
curl -s -X GET "$BASE_URL/api/v1/jobs/$JOB_ID/results" | jq .

echo -e "\n${BLUE}=== Starting another job for cancellation test ===${NC}\n"

echo -e "${GREEN}6. Starting new enrichment job${NC}"
RESPONSE2=$(curl -s -X POST "$BASE_URL/api/v1/enrich" \
  -H "Content-Type: application/json" \
  -d '{
    "row_keys": ["person_1", "person_2"],
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
  }')

echo "$RESPONSE2" | jq .
JOB_ID2=$(echo "$RESPONSE2" | jq -r '.job_id')

echo -e "\n${YELLOW}Job ID: $JOB_ID2${NC}\n"

echo -e "${GREEN}7. Cancelling job${NC}"
curl -s -X POST "$BASE_URL/api/v1/jobs/$JOB_ID2/cancel" | jq .

echo -e "\n${GREEN}8. Checking cancelled job progress${NC}"
curl -s -X GET "$BASE_URL/api/v1/jobs/$JOB_ID2/progress" | jq .

echo -e "\n${BLUE}=== Test completed ===${NC}"
