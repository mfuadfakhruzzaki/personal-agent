#!/bin/bash

# Todo Agent Backend API Test Script
# Make sure the server is running before executing this script

BASE_URL="http://localhost:8080"
API_KEY="your-api-key-here"

echo "🚀 Testing Todo Agent Backend API"
echo "================================="

# Test 1: Health Check
echo ""
echo "1. Testing Health Check..."
curl -s -X GET "$BASE_URL/healthz" | jq .

# Test 2: Process Text Input
echo ""
echo "2. Testing Text Processing..."
RESPONSE=$(curl -s -X POST "$BASE_URL/process" \
  -H "X-API-Key: $API_KEY" \
  -F "type=text" \
  -F "content=Besok meeting dengan client jam 10, lalu review code, dan kirim laporan ke manager pada hari Jumat" \
  -F "user_id=test-user-123")

echo "$RESPONSE" | jq .
JOB_ID=$(echo "$RESPONSE" | jq -r '.job_id')

# Test 3: Check Job Status
echo ""
echo "3. Checking Job Status..."
sleep 2  # Wait a bit for processing
curl -s -X GET "$BASE_URL/status/$JOB_ID" \
  -H "X-API-Key: $API_KEY" | jq .

# Test 4: Test Invalid API Key
echo ""
echo "4. Testing Invalid API Key..."
curl -s -X POST "$BASE_URL/process" \
  -H "X-API-Key: invalid-key" \
  -F "type=text" \
  -F "content=Test content" \
  -F "user_id=test-user" | jq .

# Test 5: Test Missing Parameters
echo ""
echo "5. Testing Missing Parameters..."
curl -s -X POST "$BASE_URL/process" \
  -H "X-API-Key: $API_KEY" \
  -F "type=text" \
  -F "user_id=test-user" | jq .

echo ""
echo "✅ API Testing Complete!"
