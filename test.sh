#!/bin/bash

R='\033[0;31m'
G='\033[0;32m'
Y='\033[1;33m'
B='\033[0;36m'
NC='\033[0m' # No Color (reset)


VERBOSE_MODE=false
PERSISTENCE=false
UPDATE=false
BASE_URL="http://localhost:8080"
ENDPOINT=""
CURL_CMD=""

while getopts "vpu" opt; do
	case $opt in
		v) VERBOSE_MODE=true
		;;
		p) PERSISTENCE=true
		;;
		u) UPDATE=true
		;;
		\?)
			echo "Usage: $0"
			echo "[-v] to show response"
			echo "[-p] to skip deletion (will cause failures on the next test but be fine after that)"

			exit 1
		;;
	esac
done

function test_post(){
	echo -e "${Y}Running POST test...${NC}"

	for (( i = 1; i <6; i++)); do
		ENDPOINT="/users"
		CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}'  -X POST \
		-H 'Content-Type: application/json' \
		-d '{\"username\":\"User $i\", \"email\":\"user$i@example.com\",\"id\":\"$i\", \"role\":\"member\"}' $BASE_URL$ENDPOINT"
		run_test "POST User $i" 201
	done

}

function test_post_failure(){
	echo -e "${Y}Running POST Failure test...${NC}"

	for (( i = 1; i <6; i++)); do
		ENDPOINT="/users"
		CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}'  -X POST \
		-H 'Content-Type: application/json' \
		-d '{\"username\":\"User $i\", \"email\":\"user$i@example.com\",\"id\":\"$i\", \"role\":\"member\"}' $BASE_URL$ENDPOINT"
		run_test "POST User $i Failure" 400
	done

}


function test_get(){
	echo -e "${Y}Running GET test...${NC}"

	for (( i = 1; i <6; i++)); do
		ENDPOINT="/users/$i"
		CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' $BASE_URL$ENDPOINT"
		run_test "Get User by ID $i" 200
	done
}


function test_get_failure(){
	echo -e "${Y}Running GET failure test...${NC}"


	for (( i = 1; i <6; i++)); do
		ENDPOINT="/users/$i"
		CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' $BASE_URL$ENDPOINT"
		run_test "Get User by ID $i Failure" 404
	done
}

function test_update(){
	echo -e "${Y}Running PUT test...${NC}"

	for (( i = 1; i <6; i++)); do
		ENDPOINT="/users/$i"
		CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}'  -X PUT \
		-H 'Content-Type: application/json' \
		-d '{\"name\":\"UPDATED[$i] User $i\", \"email\":\"UPDATEDuser$i@example.com\",\"id\":\"$i\", \"role\":\"member\"}' $BASE_URL$ENDPOINT"
		run_test "PUT User $i" 200
	done

}

function test_delete(){
	echo -e "${Y}Running DELETE failure test...${NC}"


	for (( i = 1; i <6; i++)); do
		ENDPOINT="/users/$i"
		CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' -X DELETE $BASE_URL$ENDPOINT"
		run_test "Delete User by ID $i" 204
	done
}




function run_test(){
	local label=$1
	local expected_status=$2

	echo "Requesting $label..."

	response=$(eval "$CURL_CMD")

	body=$(echo "$response" | sed -e 's/HTTPSTATUS\:.*//g')
	status=$(echo $response | tr -d '\n' | sed -e 's/.*HTTPSTATUS://')


	if [ "$status" -eq "$expected_status" ]; then
		echo -e "${G}$label suceeded -- (status: $status)${NC}"
	else
		echo -e "${R}$label failed -- (status: $status)${NC}"
	fi

	if [ "$VERBOSE_MODE" = true ]; then
		echo
		echo -e "${B}Response:${NC}"
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
	echo "=============================================="

}

function setup(){
	echo -e "${B}Make sure the app is running: ${Y}go run cmd/api/main.go${NC}"
	echo -e "${B}Ensure that azurite is running inside the tmp/ directory: ${Y}cd tmp/ && azurite${NC}"
}

setup
test_get_failure
test_post
test_post_failure
test_get
test_update
test_get
if [ "$PERSISTENCE" = false ]; then
	test_delete
fi
