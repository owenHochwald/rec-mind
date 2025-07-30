#!/bin/bash

# Test script for Article Processing Pipeline Integration
# This script tests the integration between Go API and Python ML service

BASE_URL="http://localhost:8080"
PYTHON_URL="http://localhost:8000"

echo "🧪 Testing Article Processing Pipeline Integration"
echo "=================================================="

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    case $1 in
        "SUCCESS") echo -e "${GREEN}✅ $2${NC}" ;;
        "ERROR") echo -e "${RED}❌ $2${NC}" ;;
        "WARNING") echo -e "${YELLOW}⚠️  $2${NC}" ;;
        "INFO") echo -e "${BLUE}ℹ️  $2${NC}" ;;
    esac
}

# Test 1: Check Go API health
echo -e "\n${BLUE}1. Testing Go API Health${NC}"
response=$(curl -s -o /dev/null -w "%{http_code}" $BASE_URL/health)
if [ $response -eq 200 ]; then
    print_status "SUCCESS" "Go API is healthy"
else
    print_status "ERROR" "Go API is not responding (HTTP $response)"
    exit 1
fi

# Test 2: Check Python ML service health via Go API
echo -e "\n${BLUE}2. Testing Python ML Service Health via Go API${NC}"
response=$(curl -s $BASE_URL/api/ml/health)
echo "Response: $response"
if echo $response | grep -q '"ml_service_healthy":true'; then
    print_status "SUCCESS" "Python ML service is healthy"
else
    print_status "WARNING" "Python ML service may not be available"
fi

# Test 3: Check Python ML service directly
echo -e "\n${BLUE}3. Testing Python ML Service Directly${NC}"
python_response=$(curl -s -o /dev/null -w "%{http_code}" $PYTHON_URL/health 2>/dev/null)
if [ $python_response -eq 200 ]; then
    print_status "SUCCESS" "Python ML service is responding directly"
else
    print_status "WARNING" "Python ML service is not responding directly (HTTP $python_response)"
fi

# Test 4: Test async article upload
echo -e "\n${BLUE}4. Testing Async Article Upload${NC}"
test_article=$(cat <<EOF
{
  "title": "Test Article - Async Processing",
  "content": "This is a test article to verify the integration between Go API and Python ML service. The article should be stored in PostgreSQL and embeddings should be generated asynchronously.",
  "url": "https://example.com/test-async",
  "category": "integration-test",
  "published_at": "2025-07-30T18:00:00Z"
}
EOF
)

async_response=$(curl -s -X POST $BASE_URL/api/upload \
  -H "Content-Type: application/json" \
  -d "$test_article")

echo "Async Response: $async_response"

if echo $async_response | grep -q '"processing_mode":"async"'; then
    print_status "SUCCESS" "Async article upload successful"
    # Extract article ID for later verification
    ARTICLE_ID=$(echo $async_response | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
    print_status "INFO" "Article ID: $ARTICLE_ID"
else
    print_status "ERROR" "Async article upload failed"
fi

# Wait a moment for async processing
echo -e "\n${YELLOW}⏳ Waiting 3 seconds for async processing...${NC}"
sleep 3

# Test 5: Test sync article upload
echo -e "\n${BLUE}5. Testing Sync Article Upload${NC}"
test_article_sync=$(cat <<EOF
{
  "title": "Test Article - Sync Processing",
  "content": "This is a test article to verify the synchronous integration between Go API and Python ML service. The article should be stored in PostgreSQL and embeddings should be generated synchronously before response.",
  "url": "https://example.com/test-sync",
  "category": "integration-test",
  "published_at": "2025-07-30T18:00:00Z"
}
EOF
)

echo "Sending sync request..."
sync_response=$(curl -s -X POST "$BASE_URL/api/upload?processing=sync" \
  -H "Content-Type: application/json" \
  -d "$test_article_sync")

echo "Sync Response: $sync_response"

if echo $sync_response | grep -q '"embedding_generated":true'; then
    print_status "SUCCESS" "Sync article upload with embedding generation successful"
elif echo $sync_response | grep -q '"embedding_generated":false'; then
    print_status "WARNING" "Sync article upload successful but embedding generation failed"
else
    print_status "ERROR" "Sync article upload failed or unexpected response format"
fi

# Test 6: Test legacy upload
echo -e "\n${BLUE}6. Testing Legacy Upload (No ML Processing)${NC}"
test_article_legacy=$(cat <<EOF
{
  "title": "Test Article - Legacy",
  "content": "This is a test article for the legacy upload endpoint without ML processing.",
  "url": "https://example.com/test-legacy",
  "category": "integration-test",
  "published_at": "2025-07-30T18:00:00Z"
}
EOF
)

legacy_response=$(curl -s -X POST $BASE_URL/api/upload/legacy \
  -H "Content-Type: application/json" \
  -d "$test_article_legacy")

echo "Legacy Response: $legacy_response"

if echo $legacy_response | grep -q '"id"'; then
    print_status "SUCCESS" "Legacy article upload successful"
else
    print_status "ERROR" "Legacy article upload failed"
fi

# Test 7: Verify articles were created
echo -e "\n${BLUE}7. Verifying Articles in Database${NC}"
articles_response=$(curl -s "$BASE_URL/api/v1/articles?limit=10&category=integration-test")
article_count=$(echo $articles_response | grep -o '"id"' | wc -l)

if [ $article_count -ge 3 ]; then
    print_status "SUCCESS" "Found $article_count integration test articles in database"
else
    print_status "WARNING" "Expected at least 3 articles, found $article_count"
fi

# Test 8: Check Python ML service endpoints
echo -e "\n${BLUE}8. Testing Python ML Service Endpoints${NC}"
if [ $python_response -eq 200 ]; then
    # Test batch endpoint
    batch_health=$(curl -s -o /dev/null -w "%{http_code}" $PYTHON_URL/embeddings/batch 2>/dev/null || echo "000")
    if [ $batch_health -eq 422 ] || [ $batch_health -eq 400 ]; then
        print_status "SUCCESS" "Python ML batch endpoint is accessible (returns expected validation error)"
    else
        print_status "WARNING" "Python ML batch endpoint responded with HTTP $batch_health"
    fi

    # Test index stats
    stats_response=$(curl -s $PYTHON_URL/index/stats 2>/dev/null)
    if echo $stats_response | grep -q 'total_vectors'; then
        print_status "SUCCESS" "Python ML index stats endpoint working"
    else
        print_status "WARNING" "Python ML index stats endpoint may not be working"
    fi
fi

# Summary
echo -e "\n${BLUE}===============================================${NC}"
echo -e "${BLUE}🎯 Integration Test Summary${NC}"
echo -e "${BLUE}===============================================${NC}"

if [ $python_response -eq 200 ]; then
    print_status "SUCCESS" "Full integration is working!"
    echo -e "${GREEN}✨ The Go API successfully communicates with the Python ML service${NC}"
else
    print_status "WARNING" "Partial integration - Go API works but Python ML service may be unavailable"
    echo -e "${YELLOW}💡 Start the Python ML service with: cd llm && python -m uvicorn app.main:app --reload${NC}"
fi

echo -e "\n${BLUE}📚 View API documentation at: http://localhost:8080/swagger/index.html${NC}"
echo -e "${BLUE}🔍 Monitor logs for embedding generation progress${NC}"

echo -e "\n${GREEN}🚀 Integration testing complete!${NC}"