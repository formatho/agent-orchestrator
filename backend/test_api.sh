#!/bin/bash

# Backend Battle Test Script
# Tests all API endpoints

API="http://localhost:18765"
PASS=0
FAIL=0

test_api() {
    local method=$1
    local endpoint=$2
    local data=$3
    local expected=$4
    
    echo -n "Testing $method $endpoint... "
    
    if [ -z "$data" ]; then
        response=$(curl -s -w "\n%{http_code}" -X $method "$API$endpoint")
    else
        response=$(curl -s -w "\n%{http_code}" -X $method "$API$endpoint" \
            -H "Content-Type: application/json" \
            -d "$data")
    fi
    
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')
    
    if [ "$http_code" == "$expected" ]; then
        echo "✅ PASS ($http_code)"
        ((PASS++))
    else
        echo "❌ FAIL (Expected: $expected, Got: $http_code)"
        echo "   Response: $body"
        ((FAIL++))
    fi
}

echo "=== AGENT ENDPOINTS ==="
test_api "GET" "/api/agents" "" "200"
test_api "POST" "/api/agents" '{"name":"battle-test-agent","model":"gpt-4o","type":"testing"}' "201"
test_api "GET" "/api/agents/battle-test-agent" "" "200"
test_api "PUT" "/api/agents/battle-test-agent" '{"name":"battle-test-agent","model":"gpt-4-turbo"}' "200"
test_api "POST" "/api/agents/battle-test-agent/pause" "" "200"
test_api "POST" "/api/agents/battle-test-agent/resume" "" "200"
test_api "DELETE" "/api/agents/battle-test-agent" "" "200"

echo ""
echo "=== TODO ENDPOINTS ==="
test_api "GET" "/api/todos" "" "200"
test_api "POST" "/api/todos" '{"title":"Battle Test TODO","description":"Testing TODO creation","priority":8}' "201"
test_api "GET" "/api/todos" "" "200"
test_api "POST" "/api/todos/battle-test/start" "" "200"

echo ""
echo "=== CRON ENDPOINTS ==="
test_api "GET" "/api/cron" "" "200"
test_api "POST" "/api/cron" '{"name":"Battle Test Cron","schedule":"*/5 * * * *"}' "201"
test_api "GET" "/api/cron" "" "200"
test_api "POST" "/api/cron/battle-test/pause" "" "200"
test_api "POST" "/api/cron/battle-test/resume" "" "200"

echo ""
echo "=== CONFIG ENDPOINTS ==="
test_api "GET" "/api/config" "" "200"
test_api "PUT" "/api/config" '{"global":{"llm":{"provider":"openai","model":"gpt-4o"}}}' "200"

echo ""
echo "=== SYSTEM ENDPOINTS ==="
test_api "GET" "/health" "" "200"
test_api "GET" "/api/system/status" "" "200"
test_api "GET" "/api/system/health" "" "200"

echo ""
echo "=== RESULTS ==="
echo "✅ Passed: $PASS"
echo "❌ Failed: $FAIL"
echo "📊 Success Rate: $(awk "BEGIN {printf \"%.1f\", ($PASS/($PASS+$FAIL))*100}")%"
