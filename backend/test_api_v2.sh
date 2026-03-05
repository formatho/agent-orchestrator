#!/bin/bash

# Backend Battle Test Script v2
# Tests all API endpoints with proper ID handling

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
        echo "$body"
    else
        echo "❌ FAIL (Expected: $expected, Got: $http_code)"
        echo "   Response: $body"
        ((FAIL++))
    fi
    
    echo "$body"
}

echo "=== AGENT ENDPOINTS ==="
echo ""
test_api "GET" "/api/agents" "" "200"

echo ""
echo "Creating agent..."
AGENT_RESPONSE=$(test_api "POST" "/api/agents" '{"name":"battle-test-agent","model":"gpt-4o","type":"testing"}' "201")
AGENT_ID=$(echo "$AGENT_RESPONSE" | python3 -c "import sys, json; print(json.load(sys.stdin)['id'])" 2>/dev/null)

if [ -n "$AGENT_ID" ]; then
    echo "Created agent with ID: $AGENT_ID"
    echo ""
    test_api "GET" "/api/agents/$AGENT_ID" "" "200"
    test_api "PUT" "/api/agents/$AGENT_ID" '{"name":"battle-test-agent","model":"gpt-4-turbo"}' "200"
    test_api "POST" "/api/agents/$AGENT_ID/pause" "" "200"
    test_api "POST" "/api/agents/$AGENT_ID/resume" "" "200"
    test_api "DELETE" "/api/agents/$AGENT_ID" "" "200"
fi

echo ""
echo "=== TODO ENDPOINTS ==="
echo ""
test_api "GET" "/api/todos" "" "200"

echo ""
echo "Creating TODO..."
TODO_RESPONSE=$(test_api "POST" "/api/todos" '{"title":"Battle Test TODO","description":"Testing TODO creation","priority":8}' "201")
TODO_ID=$(echo "$TODO_RESPONSE" | python3 -c "import sys, json; print(json.load(sys.stdin)['id'])" 2>/dev/null)

if [ -n "$TODO_ID" ]; then
    echo "Created TODO with ID: $TODO_ID"
    echo ""
    test_api "GET" "/api/todos/$TODO_ID" "" "200"
    test_api "POST" "/api/todos/$TODO_ID/start" "" "200"
    test_api "DELETE" "/api/todos/$TODO_ID" "" "200"
fi

echo ""
echo "=== CRON ENDPOINTS ==="
echo ""
test_api "GET" "/api/cron" "" "200"

if [ -n "$AGENT_ID" ]; then
    echo ""
    # Recreate agent for cron test
    AGENT2=$(curl -s -X POST "$API/api/agents" -H "Content-Type: application/json" -d '{"name":"cron-agent","model":"gpt-4o"}')
    AGENT2_ID=$(echo "$AGENT2" | python3 -c "import sys, json; print(json.load(sys.stdin)['id'])" 2>/dev/null)
    
    echo "Creating Cron Job..."
    CRON_RESPONSE=$(test_api "POST" "/api/cron" "{\"name\":\"Battle Test Cron\",\"schedule\":\"*/5 * * * *\",\"agent_id\":\"$AGENT2_ID\"}" "201")
    CRON_ID=$(echo "$CRON_RESPONSE" | python3 -c "import sys, json; print(json.load(sys.stdin)['id'])" 2>/dev/null)
    
    if [ -n "$CRON_ID" ]; then
        echo "Created Cron with ID: $CRON_ID"
        echo ""
        test_api "GET" "/api/cron/$CRON_ID" "" "200"
        test_api "POST" "/api/cron/$CRON_ID/pause" "" "200"
        test_api "POST" "/api/cron/$CRON_ID/resume" "" "200"
        test_api "DELETE" "/api/cron/$CRON_ID" "" "200"
        test_api "DELETE" "/api/agents/$AGENT2_ID" "" "200"
    fi
fi

echo ""
echo "=== CONFIG ENDPOINTS ==="
echo ""
test_api "GET" "/api/config" "" "200"
test_api "PUT" "/api/config" '{"global":{"llm":{"provider":"openai","model":"gpt-4o"}}}' "200"

echo ""
echo "=== SYSTEM ENDPOINTS ==="
echo ""
test_api "GET" "/health" "" "200"
test_api "GET" "/api/system/status" "" "200"
test_api "GET" "/api/system/health" "" "200"

echo ""
echo "=== RESULTS ==="
echo "✅ Passed: $PASS"
echo "❌ Failed: $FAIL"
