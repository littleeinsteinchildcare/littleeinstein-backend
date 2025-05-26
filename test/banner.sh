#!/bin/bash

# Colors for output
R='\033[0;31m'  # Red
G='\033[0;32m'  # Green
Y='\033[1;33m'  # Yellow
B='\033[0;36m'  # Cyan
NC='\033[0m'    # No Color (reset)

# Configuration
VERBOSE_MODE=true
BASE_URL="http://localhost:8080"
ENDPOINT="/api/banner"
CURL_CMD=""

# Parse command line options
while getopts "v" opt; do
    case $opt in
        v) VERBOSE_MODE=true
        ;;
        \?)
            echo "Usage: $0 [-v]"
            echo "  -v  Show full response bodies"
            exit 1
        ;;
    esac
done

# Print header
echo -e "${B}===================================${NC}"
echo -e "${B}    Banner API Testing Script      ${NC}"
echo -e "${B}===================================${NC}"

# Ensure server is running
echo -e "${Y}Ensure your server is running at ${BASE_URL}${NC}"
echo -e "${Y}Press Enter to continue or Ctrl+C to exit...${NC}"
read

# Helper function to run tests
function run_test() {
    local label=$1
    local expected_status=$2
    local curl_command=$3

    echo -e "\n${Y}Running test: ${label}${NC}"
    echo "Command: ${curl_command}"

    response=$(eval "${curl_command}")

    body=$(echo "$response" | sed -e 's/HTTPSTATUS\:.*//g')
    status=$(echo $response | tr -d '\n' | sed -e 's/.*HTTPSTATUS://')

    if [ "$status" -eq "$expected_status" ]; then
        echo -e "${G}✓ Test passed (status: ${status})${NC}"
    else
        echo -e "${R}✗ Test failed - Expected: ${expected_status}, Got: ${status}${NC}"
    fi

    if [ "$VERBOSE_MODE" = true ]; then
        echo -e "${B}Response:${NC}"
        echo "$body" | python -m json.tool 2>/dev/null || echo "$body"
    fi
    
    echo -e "${B}----------------------------------${NC}"
}

# ===========================================
# Test Case 1: GET when no banner exists
# ===========================================
CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' -X GET ${BASE_URL}${ENDPOINT}"
run_test "GET banner (when none exists)" 404 "$CURL_CMD"

# ===========================================
# Test Case 2: Create weather banner
# ===========================================
# Create a weather banner expiring in 1 hour
expiration_time=$(date -u -v+1H +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u -d "+1 hour" +"%Y-%m-%dT%H:%M:%SZ")
CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' -X POST \
-H 'Content-Type: application/json' \
-d '{\"type\":\"weather\",\"message\":\"Snow warning\",\"expiresAt\":\"${expiration_time}\"}' \
${BASE_URL}${ENDPOINT}"
run_test "Create weather banner" 200 "$CURL_CMD"

# ===========================================
# Test Case 3: GET the created banner
# ===========================================
CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' -X GET ${BASE_URL}${ENDPOINT}"
run_test "GET banner (after creation)" 200 "$CURL_CMD"

# ===========================================
# Test Case 4: Create invalid custom banner (no message)
# ===========================================
CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' -X POST \
-H 'Content-Type: application/json' \
-d '{\"type\":\"custom\",\"message\":\"\",\"expiresAt\":\"${expiration_time}\"}' \
${BASE_URL}${ENDPOINT}"
run_test "Create invalid custom banner (no message)" 400 "$CURL_CMD"

# ===========================================
# Test Case 5: Create valid custom banner
# ===========================================
CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' -X POST \
-H 'Content-Type: application/json' \
-d '{\"type\":\"custom\",\"message\":\"This is a custom message\",\"expiresAt\":\"${expiration_time}\"}' \
${BASE_URL}${ENDPOINT}"
run_test "Create valid custom banner" 200 "$CURL_CMD"

# ===========================================
# Test Case 6: GET the updated banner
# ===========================================
CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' -X GET ${BASE_URL}${ENDPOINT}"
run_test "GET banner (after update)" 200 "$CURL_CMD"

# ===========================================
# Test Case.7: Create closure banner
# ===========================================
CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' -X POST \
-H 'Content-Type: application/json' \
-d '{\"type\":\"closure\",\"message\":\"Facility closed today\",\"expiresAt\":\"${expiration_time}\"}' \
${BASE_URL}${ENDPOINT}"
run_test "Create closure banner" 200 "$CURL_CMD"

# ===========================================
# Test Case 8: GET the closure banner
# ===========================================
CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' -X GET ${BASE_URL}${ENDPOINT}"
run_test "GET banner (closure type)" 200 "$CURL_CMD"

# ===========================================
# Test Case 9: Create banner with invalid type
# ===========================================
CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' -X POST \
-H 'Content-Type: application/json' \
-d '{\"type\":\"invalid\",\"message\":\"This should fail\",\"expiresAt\":\"${expiration_time}\"}' \
${BASE_URL}${ENDPOINT}"
run_test "Create banner with invalid type" 400 "$CURL_CMD"

# ===========================================
# Test Case 10: Create banner with past expiration
# ===========================================
past_time=$(date -u -v-1H +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u -d "-1 hour" +"%Y-%m-%dT%H:%M:%SZ")
CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' -X POST \
-H 'Content-Type: application/json' \
-d '{\"type\":\"weather\",\"message\":\"This should fail\",\"expiresAt\":\"${past_time}\"}' \
${BASE_URL}${ENDPOINT}"
run_test "Create banner with past expiration" 400 "$CURL_CMD"

# ===========================================
# Test Case 11: Delete banner
# ===========================================
CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' -X DELETE ${BASE_URL}${ENDPOINT}"
run_test "Delete banner" 200 "$CURL_CMD"

# ===========================================
# Test Case 12: GET banner after deletion
# ===========================================
CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' -X GET ${BASE_URL}${ENDPOINT}"
run_test "GET banner (after deletion)" 404 "$CURL_CMD"

# ===========================================
# Test Case 13: Create short-lived banner to test expiration
# ===========================================
echo -e "\n${Y}Testing banner expiration (create with 2 second expiration)${NC}"
short_expiration=$(date -u -v+2S +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u -d "+2 seconds" +"%Y-%m-%dT%H:%M:%SZ")
CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' -X POST \
-H 'Content-Type: application/json' \
-d '{\"type\":\"weather\",\"message\":\"This will expire in 2 seconds\",\"expiresAt\":\"${short_expiration}\"}' \
${BASE_URL}${ENDPOINT}"
run_test "Create short-lived banner" 200 "$CURL_CMD"

# Verify it exists
CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' -X GET ${BASE_URL}${ENDPOINT}"
run_test "Verify short-lived banner exists" 200 "$CURL_CMD"

# Wait for expiration
echo -e "${Y}Waiting 3 seconds for banner to expire...${NC}"
sleep 3

# Verify it's gone
CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' -X GET ${BASE_URL}${ENDPOINT}"
run_test "Verify banner expired automatically" 404 "$CURL_CMD"

# Summary
echo -e "\n${G}=====================================${NC}"
echo -e "${G}    Banner API Tests Complete         ${NC}"
echo -e "${G}=====================================${NC}"
