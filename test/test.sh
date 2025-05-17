#!/bin/bash

# Colors for output
R='\033[0;31m'  # Red
G='\033[0;32m'  # Green
Y='\033[1;33m'  # Yellow
B='\033[0;36m'  # Cyan
P='\033[0;35m'  # Purple
NC='\033[0m'    # No Color (reset)

# Configuration
BASE_URL="http://localhost:8080"
ENDPOINT="/api/banner"
VERBOSE_MODE=false
SAVE_OUTPUT=false
OUTPUT_DIR="./test_outputs"

# Parse command line options
while getopts "vso:" opt; do
    case $opt in
        v) VERBOSE_MODE=true
        ;;
        s) SAVE_OUTPUT=true
           # Create output directory if it doesn't exist
           mkdir -p "$OUTPUT_DIR"
        ;;
        o) OUTPUT_DIR="$OPTARG"
           SAVE_OUTPUT=true
           mkdir -p "$OUTPUT_DIR"
        ;;
        \?)
            echo "Usage: $0 [-v] [-s] [-o directory]"
            echo "  -v  Verbose mode (show full responses)"
            echo "  -s  Save test outputs to files"
            echo "  -o  Specify directory for output files (implies -s)"
            exit 1
        ;;
    esac
done

# Print header
echo -e "${B}=================================================================================${NC}"
echo -e "${B}                      BANNER API TESTING SCRIPT                                  ${NC}"
echo -e "${B}=================================================================================${NC}"

# Ensure server is running
ping_result=$(curl -s -o /dev/null -w "%{http_code}" ${BASE_URL})
if [ "$ping_result" = "000" ]; then
    echo -e "${R}ERROR: Cannot connect to server at ${BASE_URL}${NC}"
    echo -e "${Y}Make sure your server is running before executing this script!${NC}"
    exit 1
else
    echo -e "${G}Server at ${BASE_URL} is reachable.${NC}"
fi

# Test counter
TEST_COUNT=0
PASSED_COUNT=0
FAILED_COUNT=0

# Helper function to run tests
function run_test() {
    local test_num=$((TEST_COUNT + 1))
    local label=$1
    local expected_status=$2
    local curl_command=$3
    local expected_response=$4

    TEST_COUNT=$test_num

    echo -e "\n${P}=================================================================================${NC}"
    echo -e "${Y}TEST #${test_num}: ${label}${NC}"
    echo -e "${P}=================================================================================${NC}"

    # Display the curl command
    echo -e "${B}COMMAND:${NC}"
    # Format the command for better readability
    formatted_command=$(echo "$curl_command" | sed 's/curl -s/curl/g' | sed 's/-w/\\\n  -w/g' | sed 's/-H/\\\n  -H/g' | sed 's/-d/\\\n  -d/g' | sed 's/-X/\\\n  -X/g')
    echo -e "$formatted_command"

    if [ -n "$expected_response" ]; then
        echo -e "\n${B}EXPECTED RESPONSE:${NC}"
        echo -e "$expected_response" | python3 -m json.tool 2>/dev/null || echo -e "$expected_response"
    fi

    echo -e "\n${B}RUNNING TEST...${NC}"

    # Execute the command
    response=$(eval "${curl_command}")

    # Parse status code and body
    body=$(echo "$response" | sed -e 's/HTTPSTATUS\:.*//g')
    status=$(echo $response | tr -d '\n' | sed -e 's/.*HTTPSTATUS://')

    echo -e "${B}ACTUAL STATUS:${NC} ${status}"

    # Compare with expected status
    if [ "$status" -eq "$expected_status" ]; then
        echo -e "${G}✓ STATUS CHECK PASSED: ${status} (expected ${expected_status})${NC}"
        PASSED_COUNT=$((PASSED_COUNT + 1))
    else
        echo -e "${R}✗ STATUS CHECK FAILED: ${status} (expected ${expected_status})${NC}"
        FAILED_COUNT=$((FAILED_COUNT + 1))
    fi

    # Show response body
    if [ "$VERBOSE_MODE" = true ] || [ "$status" -ne "$expected_status" ]; then
        echo -e "\n${B}ACTUAL RESPONSE:${NC}"
        echo "$body" | python3 -m json.tool 2>/dev/null || echo "$body"
    fi

    # Save output to file if requested
    if [ "$SAVE_OUTPUT" = true ]; then
        echo "Test #${test_num}: ${label}" > "${OUTPUT_DIR}/test_${test_num}.txt"
        echo "Command: ${curl_command}" >> "${OUTPUT_DIR}/test_${test_num}.txt"
        echo "Expected Status: ${expected_status}" >> "${OUTPUT_DIR}/test_${test_num}.txt"
        echo "Actual Status: ${status}" >> "${OUTPUT_DIR}/test_${test_num}.txt"
        echo "Response Body:" >> "${OUTPUT_DIR}/test_${test_num}.txt"
        echo "$body" | python3 -m json.tool 2>/dev/null >> "${OUTPUT_DIR}/test_${test_num}.txt" || echo "$body" >> "${OUTPUT_DIR}/test_${test_num}.txt"
        echo -e "${B}Test output saved to: ${OUTPUT_DIR}/test_${test_num}.txt${NC}"
    fi

    # Return the response for any follow-up checks
    echo "$response"
}

# ===========================================
# Test Case 1: GET when no banner exists
# ===========================================
CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' -X GET ${BASE_URL}${ENDPOINT}"
EXPECTED_RESPONSE='{
  "status": 404,
  "error": "No active banner found"
}'
run_test "GET banner (when none exists)" 404 "$CURL_CMD" "$EXPECTED_RESPONSE"

# ===========================================
# Test Case 2: Create weather banner
# ===========================================
# Create a weather banner expiring in 1 hour
expiration_time=$(date -u -v+1H +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u -d "+1 hour" +"%Y-%m-%dT%H:%M:%SZ")
CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' -X POST -H 'Content-Type: application/json' -d '{\"type\":\"weather\",\"message\":\"Snow warning\",\"expiresAt\":\"${expiration_time}\"}' ${BASE_URL}${ENDPOINT}"
EXPECTED_RESPONSE="{
  \"type\": \"weather\",
  \"message\": \"Snow warning\",
  \"expiresAt\": \"${expiration_time}\"
}"
run_test "Create weather banner" 200 "$CURL_CMD" "$EXPECTED_RESPONSE"

# ===========================================
# Test Case 3: GET the created banner
# ===========================================
CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' -X GET ${BASE_URL}${ENDPOINT}"
EXPECTED_RESPONSE="{
  \"type\": \"weather\",
  \"message\": \"Snow warning\",
  \"expiresAt\": \"${expiration_time}\"
}"
run_test "GET banner (after creation)" 200 "$CURL_CMD" "$EXPECTED_RESPONSE"

# ===========================================
# Test Case 4: Create invalid custom banner (no message)
# ===========================================
CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' -X POST -H 'Content-Type: application/json' -d '{\"type\":\"custom\",\"message\":\"\",\"expiresAt\":\"${expiration_time}\"}' ${BASE_URL}${ENDPOINT}"
EXPECTED_RESPONSE='{
  "status": 400,
  "error": "Message is required for custom banner type"
}'
run_test "Create invalid custom banner (no message)" 400 "$CURL_CMD" "$EXPECTED_RESPONSE"

# ===========================================
# Test Case 5: Create valid custom banner (replacing weather banner)
# ===========================================
CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' -X POST -H 'Content-Type: application/json' -d '{\"type\":\"custom\",\"message\":\"This is a custom message\",\"expiresAt\":\"${expiration_time}\"}' ${BASE_URL}${ENDPOINT}"
EXPECTED_RESPONSE="{
  \"type\": \"custom\",
  \"message\": \"This is a custom message\",
  \"expiresAt\": \"${expiration_time}\"
}"
run_test "Create valid custom banner" 200 "$CURL_CMD" "$EXPECTED_RESPONSE"

# ===========================================
# Test Case 6: GET the updated banner
# ===========================================
CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' -X GET ${BASE_URL}${ENDPOINT}"
EXPECTED_RESPONSE="{
  \"type\": \"custom\",
  \"message\": \"This is a custom message\",
  \"expiresAt\": \"${expiration_time}\"
}"
run_test "GET banner (after update)" 200 "$CURL_CMD" "$EXPECTED_RESPONSE"

# ===========================================
# Test Case 7: Create closure banner
# ===========================================
CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' -X POST -H 'Content-Type: application/json' -d '{\"type\":\"closure\",\"message\":\"Facility closed today\",\"expiresAt\":\"${expiration_time}\"}' ${BASE_URL}${ENDPOINT}"
EXPECTED_RESPONSE="{
  \"type\": \"closure\",
  \"message\": \"Facility closed today\",
  \"expiresAt\": \"${expiration_time}\"
}"
run_test "Create closure banner" 200 "$CURL_CMD" "$EXPECTED_RESPONSE"

# ===========================================
# Test Case 8: GET the closure banner
# ===========================================
CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' -X GET ${BASE_URL}${ENDPOINT}"
EXPECTED_RESPONSE="{
  \"type\": \"closure\",
  \"message\": \"Facility closed today\",
  \"expiresAt\": \"${expiration_time}\"
}"
run_test "GET banner (closure type)" 200 "$CURL_CMD" "$EXPECTED_RESPONSE"

# ===========================================
# Test Case 9: Create banner with invalid type
# ===========================================
CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' -X POST -H 'Content-Type: application/json' -d '{\"type\":\"invalid\",\"message\":\"This should fail\",\"expiresAt\":\"${expiration_time}\"}' ${BASE_URL}${ENDPOINT}"
EXPECTED_RESPONSE='{
  "status": 400,
  "error": "Invalid banner type: must be weather, closure, or custom"
}'
run_test "Create banner with invalid type" 400 "$CURL_CMD" "$EXPECTED_RESPONSE"

# ===========================================
# Test Case 10: Create banner with past expiration
# ===========================================
past_time=$(date -u -v-1H +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u -d "-1 hour" +"%Y-%m-%dT%H:%M:%SZ")
CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' -X POST -H 'Content-Type: application/json' -d '{\"type\":\"weather\",\"message\":\"This should fail\",\"expiresAt\":\"${past_time}\"}' ${BASE_URL}${ENDPOINT}"
EXPECTED_RESPONSE='{
  "status": 400,
  "error": "Expiration time must be in the future"
}'
run_test "Create banner with past expiration" 400 "$CURL_CMD" "$EXPECTED_RESPONSE"

# ===========================================
# Test Case 11: Try too far in the future (over 72 hours)
# ===========================================
far_future=$(date -u -v+73H +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u -d "+73 hours" +"%Y-%m-%dT%H:%M:%SZ")
CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' -X POST -H 'Content-Type: application/json' -d '{\"type\":\"weather\",\"message\":\"This should fail - too far in future\",\"expiresAt\":\"${far_future}\"}' ${BASE_URL}${ENDPOINT}"
EXPECTED_RESPONSE='{
  "status": 400,
  "error": "Expiration time cannot be more than 72 hours in the future"
}'
run_test "Create banner with expiration too far in future" 400 "$CURL_CMD" "$EXPECTED_RESPONSE"

# ===========================================
# Test Case 12: Delete banner
# ===========================================
CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' -X DELETE ${BASE_URL}${ENDPOINT}"
EXPECTED_RESPONSE='{
  "message": "banner cleared"
}'
run_test "Delete banner" 200 "$CURL_CMD" "$EXPECTED_RESPONSE"

# ===========================================
# Test Case 13: GET banner after deletion
# ===========================================
CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' -X GET ${BASE_URL}${ENDPOINT}"
EXPECTED_RESPONSE='{
  "status": 404,
  "error": "No active banner found"
}'
run_test "GET banner (after deletion)" 404 "$CURL_CMD" "$EXPECTED_RESPONSE"

# ===========================================
# Test Case 14: Delete when no banner exists
# ===========================================
CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' -X DELETE ${BASE_URL}${ENDPOINT}"
EXPECTED_RESPONSE='{
  "message": "banner cleared"
}'
run_test "Delete banner (when none exists)" 200 "$CURL_CMD" "$EXPECTED_RESPONSE"

# ===========================================
# Test Case 15: Create short-lived banner to test expiration
# ===========================================
echo -e "\n${Y}Testing banner expiration (create with 2 second expiration)${NC}"
short_expiration=$(date -u -v+2S +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u -d "+2 seconds" +"%Y-%m-%dT%H:%M:%SZ")
CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' -X POST -H 'Content-Type: application/json' -d '{\"type\":\"weather\",\"message\":\"This will expire in 2 seconds\",\"expiresAt\":\"${short_expiration}\"}' ${BASE_URL}${ENDPOINT}"
EXPECTED_RESPONSE="{
  \"type\": \"weather\",
  \"message\": \"This will expire in 2 seconds\",
  \"expiresAt\": \"${short_expiration}\"
}"
run_test "Create short-lived banner" 200 "$CURL_CMD" "$EXPECTED_RESPONSE"

# Verify it exists
CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' -X GET ${BASE_URL}${ENDPOINT}"
EXPECTED_RESPONSE="{
  \"type\": \"weather\",
  \"message\": \"This will expire in 2 seconds\",
  \"expiresAt\": \"${short_expiration}\"
}"
run_test "Verify short-lived banner exists" 200 "$CURL_CMD" "$EXPECTED_RESPONSE"

# Wait for expiration
echo -e "\n${Y}Waiting 3 seconds for banner to expire...${NC}"
sleep 3

# Verify it's gone
CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' -X GET ${BASE_URL}${ENDPOINT}"
EXPECTED_RESPONSE='{
  "status": 404,
  "error": "No active banner found"
}'
run_test "Verify banner expired automatically" 404 "$CURL_CMD" "$EXPECTED_RESPONSE"

# ===========================================
# Test Case 16: Create banner with invalid JSON
# ===========================================
CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' -X POST -H 'Content-Type: application/json' -d '{\"type\":\"weather\",\"message\":\"Missing closing brace\",\"expiresAt\":\"${expiration_time}\"' ${BASE_URL}${ENDPOINT}"
EXPECTED_RESPONSE='{
  "status": 400,
  "error": "Failed to decode JSON request"
}'
run_test "Create banner with invalid JSON" 400 "$CURL_CMD" "$EXPECTED_RESPONSE"

# ===========================================
# Test Case 17: Create banner with missing required field (expiresAt)
# ===========================================
CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' -X POST -H 'Content-Type: application/json' -d '{\"type\":\"weather\",\"message\":\"Missing expiration time\"}' ${BASE_URL}${ENDPOINT}"
EXPECTED_RESPONSE='{
  "status": 400,
  "error": "Expiration time is required"
}'
run_test "Create banner missing required field" 400 "$CURL_CMD" "$EXPECTED_RESPONSE"

# ===========================================
# Test Case 18: Create banner with invalid date format
# ===========================================
CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' -X POST -H 'Content-Type: application/json' -d '{\"type\":\"weather\",\"message\":\"Invalid date format\",\"expiresAt\":\"tomorrow\"}' ${BASE_URL}${ENDPOINT}"
EXPECTED_RESPONSE='{
  "status": 400,
  "error": "Invalid expiration time format, use ISO 8601"
}'
run_test "Create banner with invalid date format" 400 "$CURL_CMD" "$EXPECTED_RESPONSE"

# ===========================================
# Test Summary
# ===========================================
echo -e "\n${P}=================================================================================${NC}"
echo -e "${B}                                TEST SUMMARY                                     ${NC}"
echo -e "${P}=================================================================================${NC}"
echo -e "${B}Total Tests:${NC} $TEST_COUNT"
echo -e "${G}Tests Passed:${NC} $PASSED_COUNT"
echo -e "${R}Tests Failed:${NC} $FAILED_COUNT"

if [ $FAILED_COUNT -eq 0 ]; then
    echo -e "\n${G}✓ ALL TESTS PASSED!${NC}"
else
    echo -e "\n${R}✗ SOME TESTS FAILED!${NC}"
    exit 1
fi

echo -e "\n${B}=================================================================================${NC}"