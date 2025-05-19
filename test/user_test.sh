#!/bin/bash

# Flags
POST=false
GET=false
UPDATE=false
DELETE=false
VERBOSE=false
THOROUGH=false
NUM_TESTS=1

# Color codes
R='\033[0;31m'
G='\033[0;32m'
Y='\033[1;33m'
B='\033[0;36m'
NC='\033[0m' # No Color (reset)


###################
#    Utilities    #
###################

# Option Handling
while getopts "n:pgudtvh" opt; do
    case "$opt" in
        n) NUM_TESTS="$OPTARG" ;;
        p) POST=true ;;
        g) GET=true ;;
        u) UPDATE=true ;;
        d) DELETE=true ;;
        t) THOROUGH=true ;;
        v) VERBOSE=true ;;

        h) show_help; exit 0 ;;
        ?) show_help; exit 1 ;;
    esac
done


# Help message
show_help() {
    echo "Usage: $0 -[pgudvh]"
    echo
    echo "Options:"
    echo "  -p              POST request testing"
    echo "  -g              GET request testing"
    echo "  -u              UPDATE request testing"
    echo "  -d              DELETE request testing"
    echo "  -v              Verbose mode: Show full response message" 
    echo "  -t              Thorough test: Test all different failures and correct response codes" 
    echo "  -h              Show this help message"
}

# Color wrappers
red()    { echo -e "${R}$1${NC}"; }
green()  { echo -e "${G}$1${NC}"; }
yellow() { echo -e "${Y}$1${NC}"; }
blue()   { echo -e "${B}$1${NC}"; }


###############
#    TESTS    #
###############

BASE_URL="http://localhost:8080"
ENDPOINT=""
CURL_CMD=""

test_post(){
 	echo  "$(yellow "Running POST 201 (Created) test...")"

	for (( i = 1; i <("$NUM_TESTS"+1); i++)); do
		ENDPOINT="/users"
		CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' -X POST \
		-H 'Content-Type: application/json' \
		-d '{\"username\":\"User $i\", \"email\":\"user$i@example.com\",\"id\":\"$i\", \"role\":\"member\"}' \
        $BASE_URL$ENDPOINT"
		run_test "POST test <Create New Entity: User $i>" 201
	done
}

test_post_failure_bad_request(){
 	echo  "$(yellow "Running POST 400 (Bad Request) test...")"

	for (( i = 1; i <("$NUM_TESTS"+1); i++)); do
		ENDPOINT="/users"
		CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}'  -X POST \
		-H 'Content-Type: application/json' \
		-d '{\"username\":\"User $i\", \"email\":\"user$i@example.com\",\"id\":\"$i\", \"role\":\"member\",}'  \
        $BASE_URL$ENDPOINT"
		run_test "POST test <Bad Request: User $i>" 400
	done
	
}

test_post_failure_entity_already_exists(){
 	echo  "$(yellow "Running POST 409 (Entity Exists) test...")"

	for (( i = 1; i <("$NUM_TESTS"+1); i++)); do
		ENDPOINT="/users"
		CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}'  -X POST \
		-H 'Content-Type: application/json' \
		-d '{\"username\":\"User $i\", \"email\":\"user$i@example.com\",\"id\":\"$i\", \"role\":\"member\"}'  \
        $BASE_URL$ENDPOINT"
		run_test "POST test <Entity Already Exists: User $i>" 409
	done
}


test_get(){
	echo "$(yellow "Running GET 200 (OK) test...")"

	for (( i = 1; i <("$NUM_TESTS"+1); i++)); do
		ENDPOINT="/users/$i"
		CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' $BASE_URL$ENDPOINT"
		run_test "GET test <OK: User $i>" 200
	done	
}


test_get_failure_entity_not_found(){
	echo "$(yellow "Running GET 404 (Not Found) test...")"

	for (( i = 1; i <("$NUM_TESTS"+1); i++)); do
		ENDPOINT="/users/$i"
		CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' $BASE_URL$ENDPOINT"
		run_test "GET test <Entity Not Found: User $i>" 404
	done	
}


test_update(){
    echo "$(yellow "Running PUT 200 (OK) test...")"

	for (( i = 1; i <("$NUM_TESTS"+1); i++)); do
		ENDPOINT="/users/$i"
		CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}'  -X PUT \
		-H 'Content-Type: application/json' \
		-d '{\"name\":\"UPDATED[$i] User $i\", \"email\":\"UPDATEDuser$i@example.com\",\"id\":\"$i\", \"role\":\"member\"}' $BASE_URL$ENDPOINT"
		run_test "PUT test <OK: User $i>" 200
	done
	
}

test_update_failure_bad_request(){
    echo "$(yellow "Running PUT 400 (Bad Request) test...")"

	for (( i = 1; i <("$NUM_TESTS"+1); i++)); do
		ENDPOINT="/users/$i"
		CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}'  -X PUT \
		-H 'Content-Type: application/json' \
		-d '{\"name\":\"UPDATED[$i] User $i\", \"email\":\"UPDATEDuser$i@example.com\",\"id\":\"$i\", \"role\":\"member\",}' $BASE_URL$ENDPOINT"
		run_test "PUT test <Bad Request: User $i>" 400
	done
	
}


test_update_failure_entity_not_found(){
    echo "$(yellow "Running PUT 409 (Entity Not Found) test...")"

	for (( i = 1; i <("$NUM_TESTS"+1); i++)); do
		ENDPOINT="/users/$i"
		CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}'  -X PUT \
		-H 'Content-Type: application/json' \
		-d '{\"name\":\"UPDATED[$i] User $i\", \"email\":\"UPDATEDuser$i@example.com\",\"id\":\"$i\", \"role\":\"member\"}' $BASE_URL$ENDPOINT"
		run_test "PUT test <Entity Not Found: User $i>" 404
	done
	
}


test_delete(){
 	echo  "$(yellow "Running DELETE 204 (No Content) test...")"

	for (( i = 1; i <("$NUM_TESTS"+1); i++)); do
		ENDPOINT="/users/$i"
		CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' -X DELETE $BASE_URL$ENDPOINT"
		run_test "DELETE test <No Content: User $i>" 204
	done	
}


test_delete_failure_entity_not_found(){
 	echo  "$(yellow "Running DELETE 404 (Entity Not Found) test...")"

	for (( i = 1; i <("$NUM_TESTS"+1); i++)); do
		ENDPOINT="/users/$i"
		CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' -X DELETE $BASE_URL$ENDPOINT"
		run_test "DELETE test <Entity Not Found: User $i>" 404
	done	
}

run_test(){
	label=$1
	expected_status=$2

	response=$(eval "$CURL_CMD")

	body=$(echo "$response" | sed -e 's/HTTPSTATUS\:.*//g')
	status=$(echo $response | tr -d '\n' | sed -e 's/.*HTTPSTATUS://')


	if [ "$status" -eq "$expected_status" ]; then
        echo "$(green "$label succeeded -- (status: $status)")"
	else
        echo "$(red "$label red -- (status: $status)")"
	fi

	if [ "$VERBOSE" = true ]; then
		echo 
		echo "$(blue "Response: ")"
		echo "$body" | tr -d '{}' | awk -F, '
		{ 
			for (i = 1; i <= NF; i++) {
				if ($i ~ /"Username":/) {
					print $i
					print ""	
				} else {
					print $i
				}
			}
		}'
	fi
    echo ""
}




setup(){
	echo -e "${B}Make sure the app is running: ${Y}go run cmd/api/main.go${NC}"
	echo -e "${B}Ensure that azurite is running inside the tmp/ directory: ${Y}cd tmp/ && azurite${NC}"
}

setup
if [ "$POST" = true ]; then
    test_post
   
    if [ "$THOROUGH" = true ]; then
        test_post_failure_bad_request
        test_post_failure_entity_already_exists
    fi
fi
if [ "$GET" = true ]; then
    test_get
fi
if [ "$UPDATE" = true ]; then
    test_update
fi
if [ "$DELETE" = true ]; then
    test_delete
    
    if [ "$THOROUGH" = true ]; then
      test_delete_failure_entity_not_found
    fi
fi

if [ "$THOROUGH" = true ]; then
    if [[ "$GET" = true && "$DELETE" = true ]]; then
        test_get_failure_entity_not_found
    fi
    if [ "$UPDATE" = true ]; then
        test_update_failure_bad_request
        if [ "$DELETE" = true ]; then
            test_update_failure_entity_not_found
        fi
    fi
fi