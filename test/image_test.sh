#!/bin/bash

# Flags
POST=false
GET=false
DELETE=false
VERBOSE=false
THOROUGH=false
CLEANUP=false
# NUM_TESTS=1
NUM_IMAGES=4
IMG_LIMIT=2
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
if [[ $# -eq 0 || "$1" != -* ]]; then
	echo "No options selected, running all tests"
	CLEANUP=true
	POST=true
	GET=true
	DELETE=true
	THOROUGH=true
fi

while getopts "n:pgudtcvh" opt; do
    case "$opt" in
        c) CLEANUP=true ;;
        p) POST=true ;;
        g) GET=true ;;
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
    # echo "  -n <Num Tests>  Specify the Number of requests to make"
    echo "  -p              POST request testing"
    echo "  -g              GET request testing"
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

	for (( i = 1; i <("$NUM_IMAGES"+1); i++)); do
		ENDPOINT="/api/image"
		CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' -X POST \
        -H 'X-User-ID: $i' \
		-F 'image=@LEC_img$i.jpg' \
        $BASE_URL$ENDPOINT"
		run_test "POST test <Upload New Image: LEC_img$i.jpg>" 201 #TODO - Consider switching Status Code
	
	done
}

test_post_one_user_multiple_images(){
 	echo  "$(yellow "Running POST (Single User) 201 (Created) test...")"

	for (( i = 1; i < IMG_LIMIT+1; i++)); do
		ENDPOINT="/api/image"
		CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' -X POST \
        -H 'X-User-ID: 1' \
		-F 'image=@LEC_img$i.jpg' \
        $BASE_URL$ENDPOINT"
		run_test "POST test <Upload New Image: LEC_img$i.jpg>" 201 #TODO - Consider switching Status Code
	
	done
}

test_post_one_user_multiple_images_failure_exceeds_limit(){
 	echo  "$(yellow "Running POST (Max Image Limit Exceeded) 500 (Created) test...")"

	for (( i = IMG_LIMIT+1; i <("$NUM_IMAGES"+1); i++)); do
		ENDPOINT="/api/image"
		CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' -X POST \
        -H 'X-User-ID: 1' \
		-F 'image=@LEC_img$i.jpg' \
        $BASE_URL$ENDPOINT"
		run_test "POST test <Upload New Image: LEC_img$i.jpg>" 500
	
	done
}

test_post_failure_bad_request(){
 	echo  "$(yellow "Running POST 400 (Bad Request) test...")"

	for (( i = 1; i <("$NUM_IMAGES"+1); i++)); do
		ENDPOINT="/api/image"
		CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' -X POST \
        -H 'X-User-I: $i' \
		-F 'image=@LEC_img$i.jpg' \
        $BASE_URL$ENDPOINT"
		run_test "POST test <Upload New Image: LEC_img$i.jpg>" 400 #TODO - Consider switching Status Code
	
	done

}

test_get(){
	echo "$(yellow "Running GET 200 (OK) test...")"

	for (( i = 1; i <("$NUM_IMAGES"+1); i++)); do
		ENDPOINT="/api/image/$i/LEC_img$i.jpg"
		CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' -X GET $BASE_URL$ENDPOINT \
		-H 'X-User-ID: $i' -o dl_LEC_img$i.jpg"
		run_test "GET test <OK: LEC_img$i.jpg>" 200
	done	
}


test_get_failure_entity_not_found(){
	echo "$(yellow "Running GET 404 (Not Found) test...")"

	for (( i = 1; i <("$NUM_IMAGES"+1); i++)); do
		ENDPOINT="/api/image/$i/LEC_img$i.jpg"
		CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' -X GET $BASE_URL$ENDPOINT \
		-H 'X-User-ID: $i' -o dl_LEC_img$i.jpg"
		run_test "GET test <Entity Not Found: LEC_img$i.jpg>" 404
	done	


}

test_get_all(){
	echo "$(yellow "Running GET (ALL) 200 (OK) test...")"

	ENDPOINT="/api/images"
	CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' $BASE_URL$ENDPOINT"
	run_test "GET test <OK: Get All Image Names>" 200
}


test_delete(){
 	echo  "$(yellow "Running DELETE 204 (No Content) test...")"

	for (( i = 1; i <("$NUM_IMAGES"+1); i++)); do
		ENDPOINT="/api/image/$i"
		CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' -X DELETE \
        -H 'X-User-ID: $i' \
		-F 'image=@LEC_img$i.jpg' \
        $BASE_URL$ENDPOINT/LEC_img$i.jpg"
		run_test "POST test <Delete Image: LEC_img$i.jpg>" 200
	done

}

test_delete_failure_entity_not_found(){
 	echo  "$(yellow "Running DELETE 404 (Entity Not Found) test...")"

	for (( i = 1; i <("$NUM_IMAGES"+1); i++)); do
		ENDPOINT="/api/image/$i"
		CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' -X DELETE \
        -H 'X-User-ID: $i' \
		-F 'image=@LEC_img$i.jpg' \
        $BASE_URL$ENDPOINT/LEC_img$i.jpg"
		run_test "POST test <Entity Not Found: LEC_img$i.jpg>" 404
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

	for (( i = 1; i <("$NUM_IMAGES"+1); i++)); do
		ENDPOINT="/api/user"
		CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' -X POST \
		-H 'Content-Type: application/json' \
		-d '{\"username\":\"User $i\", \"email\":\"user$i@example.com\",\"id\":\"$i\", \"role\":\"member\"}' \
        $BASE_URL$ENDPOINT"
		run_test "<Create New Entity: User $i>" 201
	done
}


delete_users(){
 	echo  "$(yellow "Running DELETE 204 (No Content) test...")"

	for (( i = 1; i <("$NUM_IMAGES"+1); i++)); do
		ENDPOINT="/api/user/$i"
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
 
    # CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' http://localhost:8080/users/111"
    CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' -I http://localhost:8080"
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
	if [ "$GET" = true ]; then {
		rm dl_LEC_img1.jpg
		rm dl_LEC_img2.jpg
		rm dl_LEC_img3.jpg
		rm dl_LEC_img4.jpg
	}
	fi
}



############
#   Main   #
############

setup
if [ "$POST" = true ]; then
    test_post
    if [ "$THOROUGH" = true ]; then
		test_post_one_user_multiple_images
		test_post_one_user_multiple_images_failure_exceeds_limit
        test_post_failure_bad_request
    fi
fi
if [ "$GET" = true ]; then
    test_get
	test_get_all
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
fi
if [ "$CLEANUP" = true ]; then
    cleanup
fi