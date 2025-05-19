#!/bin/bash

# Flags
POST=false
GET=false
UPDATE=false
DELETE=false
VERBOSE=false
THOROUGH=false
CLEANUP=false
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
while getopts "n:pgudtcvh" opt; do
    case "$opt" in
        c) CLEANUP=true ;;
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
    echo "  -n <Num Tests>  Specify the Number of requests to make"
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


#####################
#    EVENT TESTS    #
#####################

BASE_URL="http://localhost:8080"
ENDPOINT=""
CURL_CMD=""

test_post(){
 	echo  "$(yellow "Running POST 201 (Created) test...")"

	for (( i = 1; i <("$NUM_TESTS"+1); i++)); do
        inv1=$((i+1))
        inv2=$((i+2))
		ENDPOINT="/events"
		CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' -X POST \
		-H 'Content-Type: application/json' \
        -H 'X-User-ID: $i' \
		-d '{\"eventname\":\"Event $i\", \"date\":\"1/$i/2025\",\"id\":\"$i\", \"starttime\":\"$i:00am\",\"endtime\":\"$i:00pm\", \"invitees\":\"$inv1, $inv2\"}' \
        $BASE_URL$ENDPOINT"
		run_test "POST test <Create New Entity: Event $i>" 201
	done
}

test_post_failure_bad_request(){
 	echo  "$(yellow "Running POST 400 (Bad Request) test...")"

	for (( i = 1; i <("$NUM_TESTS"+1); i++)); do
        inv1=$((i+1))
        inv2=$((i+2))
		ENDPOINT="/events"
		CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' -X POST \
		-H 'Content-Type: application/json' \
        -H 'X-User-ID: $i' \
		-d '{\"eventname\":\"Event $i\", \"date\":\"1/$i/2025\",\"id\":\"$i\", \"starttime\":\"$i:00am\",\"endtime\":\"$i:00pm\", \"invitees\":\"$inv1, $inv2\",}' \
        $BASE_URL$ENDPOINT"
		run_test "POST test <Bad Request: User $i>" 400
	done
	
}

test_post_failure_entity_already_exists(){
 	echo  "$(yellow "Running POST 409 (Entity Exists) test...")"

	for (( i = 1; i <("$NUM_TESTS"+1); i++)); do
        inv1=$((i+1))
        inv2=$((i+2))
		ENDPOINT="/events"
		CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' -X POST \
		-H 'Content-Type: application/json' \
        -H 'X-User-ID: $i' \
		-d '{\"eventname\":\"Event $i\", \"date\":\"1/$i/2025\",\"id\":\"$i\", \"starttime\":\"$i:00am\",\"endtime\":\"$i:00pm\", \"invitees\":\"$inv1, $inv2\"}' \
        $BASE_URL$ENDPOINT"

		run_test "POST test <Entity Already Exists: Event $i>" 409
	done
}


test_get(){
	echo "$(yellow "Running GET 200 (OK) test...")"

	for (( i = 1; i <("$NUM_TESTS"+1); i++)); do
		ENDPOINT="/events/$i"
		CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' $BASE_URL$ENDPOINT"
		run_test "GET test <OK: Event $i>" 200
	done	
}


test_get_failure_entity_not_found(){
	echo "$(yellow "Running GET 404 (Not Found) test...")"

	for (( i = 1; i <("$NUM_TESTS"+1); i++)); do
		ENDPOINT="/events/$i"
		CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' $BASE_URL$ENDPOINT"
		run_test "GET test <Entity Not Found: Events $i>" 404
	done	
}


test_update(){
    echo "$(yellow "Running PUT 200 (OK) test...")"

	for (( i = 1; i <("$NUM_TESTS"+1); i++)); do
        inv1=$((i+3))
        inv2=$((i+4))
		ENDPOINT="/events/$i"
		CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}'  -X PUT \
		-H 'Content-Type: application/json' \
    	-d '{\"eventname\":\"UPDATED Event $i\", \"date\":\"9/$i/1991\",\"id\":\"$i\", \"starttime\":\"$i:30am\",\"endtime\":\"$i:30pm\", \"invitees\":\"$inv1, $inv2\"}' \
        $BASE_URL$ENDPOINT"
		run_test "PUT test <OK: Event $i>" 200
	done
	
}

test_update_failure_bad_request(){
    echo "$(yellow "Running PUT 400 (Bad Request) test...")"

	for (( i = 1; i <("$NUM_TESTS"+1); i++)); do
        inv1=$((i+3))
        inv2=$((i+4))
		ENDPOINT="/events/$i"
		CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}'  -X PUT \
		-H 'Content-Type: application/json' \
    	-d '{\"eventname\":\"Event $i\", \"date\":\"1/$i/2025\",\"id\":\"$i\", \"starttime\":\"$i:00am\",\"endtime\":\"$i:00pm\", \"invitees\":\"$inv1, $inv2\",}' \
        $BASE_URL$ENDPOINT"
		run_test "PUT test <Bad Request: Event $i>" 400
	done
	
}


test_update_failure_entity_not_found(){
    echo "$(yellow "Running PUT 409 (Entity Not Found) test...")"

	for (( i = 1; i <("$NUM_TESTS"+1); i++)); do
        inv1=$((i+3))
        inv2=$((i+4))
		ENDPOINT="/events/$i"
		CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}'  -X PUT \
		-H 'Content-Type: application/json' \
    	-d '{\"eventname\":\"Event $i\", \"date\":\"1/$i/2025\",\"id\":\"$i\", \"starttime\":\"$i:00am\",\"endtime\":\"$i:00pm\", \"invitees\":\"$inv1, $inv2\"}' \
        $BASE_URL$ENDPOINT"
		run_test "PUT test <Entity Not Found: Event $i>" 404
	done
	
}


test_delete(){
 	echo  "$(yellow "Running DELETE 204 (No Content) test...")"

	for (( i = 1; i <("$NUM_TESTS"+1); i++)); do
		ENDPOINT="/events/$i"
		CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' -X DELETE \
		-H 'Content-Type: application/json' \
        $BASE_URL$ENDPOINT"
		run_test "DELETE test <No Content: Event $i>" 204
    done
}


test_delete_failure_entity_not_found(){
 	echo  "$(yellow "Running DELETE 404 (Entity Not Found) test...")"

	for (( i = 1; i <("$NUM_TESTS"+1); i++)); do
        ENDPOINT="/events/$i"
		CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' -X DELETE \
		-H 'Content-Type: application/json' \
        $BASE_URL$ENDPOINT"

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
				if ($i ~ /"EventName":/) {
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

#####################
#   User Creation   #
##################### 
create_users(){
 	echo  "$(yellow "Creating Users...")"

	for (( i = 1; i <("$NUM_TESTS"+5); i++)); do
		ENDPOINT="/users"
		CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' -X POST \
		-H 'Content-Type: application/json' \
		-d '{\"username\":\"User $i\", \"email\":\"user$i@example.com\",\"id\":\"$i\", \"role\":\"member\"}' \
        $BASE_URL$ENDPOINT"
		run_test "<Create New Entity: User $i>" 201
	done
}


delete_users(){
 	echo  "$(yellow "Running DELETE 204 (No Content) test...")"

	for (( i = 1; i <("$NUM_TESTS"+5); i++)); do
		ENDPOINT="/users/$i"
		CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' -X DELETE $BASE_URL$ENDPOINT"
		run_test "<No Content: User $i>" 204
	done	
}



#######################
#   Setup & Cleanup   #
#######################

setup(){
	echo -e "${B}Make sure the app is running: ${Y}go run cmd/api/main.go${NC}"
	echo -e "${B}Ensure that azurite is running inside the tmp/ directory: ${Y}cd tmp/ && azurite${NC}"
 
    CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' http://localhost:8080/users/111"
    response=$(eval "$CURL_CMD")
	status=$(echo $response | tr -d '\n' | sed -e 's/.*HTTPSTATUS://')

	if [ "$status" -eq "000" ]; then
        echo "$(red "GO API IS NOT CURRENTLY RUNNING")"
        exit 1
    fi
    create_users
}

cleanup(){
    echo "$(yellow "Deleting Users...")"
    delete_users
}



############
#   Main   #
############

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
if [ "$CLEANUP" = true ]; then
    cleanup
fi