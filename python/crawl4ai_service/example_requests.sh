#!/bin/bash

BASE_URL="http://localhost:8000"

echo "Health Check:"
curl -X GET "${BASE_URL}/health"
echo -e "\n"

echo "Crawl Request:"
curl -X POST "${BASE_URL}/crawl" \
  -H "Content-Type: application/json" \
  -d '{
    "urls": [
      "https://www.linkedin.com/in/ross-jonathan/"
    ],
    "query": "johnathan"
  }'
