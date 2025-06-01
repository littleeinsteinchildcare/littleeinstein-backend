#!/bin/bash

# Grab Firebase Web API Key from .env and strip out the "FIREBASE_WEB_API_KEY=" section
API_KEY=$(grep -E '^FIREBASE_WEB_API_KEY=' ../.env | cut -d= -f2-)
USER_EMAIL=
USER_PWD=
BACKEND_URL="http://localhost:8080"

# Flags
VERBOSE=false
NUM_TESTS=2
PROD=false
IMG_LIMIT=2

# Option Handling
while getopts "n:pv" opt; do
    case "$opt" in
        n) NUM_TESTS="$OPTARG" ;;
        v) VERBOSE=true ;;
        p) PROD=true ;;


        h) show_help; exit 0 ;;
        ?) show_help; exit 1 ;;
    esac
done

# Color codes
R='\033[0;31m'
G='\033[0;32m'
Y='\033[1;33m'
B='\033[0;36m'
NC='\033[0m' # No Color (reset)

# Color wrappers
# Usage - echo "$(green "$label succeeded -- (status: $status)")"
red()    { echo -e "${R}$1${NC}"; }
green()  { echo -e "${G}$1${NC}"; }
yellow() { echo -e "${Y}$1${NC}"; }
blue()   { echo -e "${B}$1${NC}"; }



setup(){
    if [[ -z "$API_KEY" ]]; then
        echo "$(red "Failed to retrieve Firebase Web API Key from .env file")"
        exit 1
    fi
    if ! command -v jq >/dev/null 2>&1; then
        echo "$(red "Error: jq is not installed. This script requires jq to run.")" >&2
        exit 1
    fi
    if [[ "$PROD" = true ]]; then
        BACKEND_URL="https://lec-api-backend.azurewebsites.net"
    fi
}

declare -a ID_TOKENS
declare -a UIDS

build_tests(){

    # User Setup 
    echo "$(yellow "Setting up Users...")"
    for (( i = 1; i < ("$NUM_TESTS"+1); i++)) do
        USER_EMAIL="User$i@test.com"
        USER_PWD="password$i"
        echo "$(blue "Creating Firebase user: $USER_EMAIL")"
        echo ""
        resp=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -d '{
            "email":"'"$USER_EMAIL"'",
            "password":"'"$USER_PWD"'",
            "displayName":"'"Username$i"'",
            "returnSecureToken":true
        }' \
        "https://identitytoolkit.googleapis.com/v1/accounts:signUp?key=${API_KEY}"
        )
        check_firebase_err "$resp" "Failed to create Firebase user"
        echo "$(blue "Firebase User$i successfully created")"
        echo ""
        idToken=$(jq -r '.idToken' <<< "$resp")
        uid=$(jq -r '.localId' <<< "$resp")

        ID_TOKENS+=("$idToken")
        UIDS+=("$uid")

        create_backend_user $idToken "$i"

    done

    # Image Tests
    echo "$(yellow "Running Image Tests...")"
    for (( i = 1; i <("$NUM_TESTS"+1); i++)); do
        test_post "$i"
        test_get "$i"
        test_get_all
        test_delete "$i"
        test_delete_failure_entity_not_found "$i"
        test_get_failure_entity_not_found "$i"
    done
    test_post_one_user_multiple_images
    test_post_one_user_multiple_images_failure_exceeds_limit


    # User Cleanup
    echo "$(yellow "Cleaning up Users...")"
    for ((i = 1; i < ("$NUM_TESTS"+1); i++)); do
        uid="${UIDS[$i-1]}"
        idToken="${ID_TOKENS[$i-1]}"
        delete_backend_user $idToken "$i" "$uid"
        delete_firebase_user $idToken "$i"
        echo "$(blue "Firebase User$i successfully deleted")"
        echo ""
    done
}


#### FIREBASE ####

check_firebase_err(){
    response=$1
    msg=$2
    if jq -e '.error' <<<"$response" >/dev/null; then
        errMsg=$(jq -r '.error.message' <<<"$response")
        echo "$(red "$msg: $errMsg")" >&2
        exit 1
    fi
    # echo "$response"
}

delete_firebase_user(){
    idToken=$1
    i=$2
    echo "$(blue "Deleting Firebase User $i")"
    echo ""

    resp=$(curl -s -X POST \
    -H "Content-Type: application/json" \
    -d '{
      "idToken":"'"$idToken"'"
    }' \
    "https://identitytoolkit.googleapis.com/v1/accounts:delete?key=${API_KEY}"
    )

    check_firebase_err "$resp" "Failed to Delete User $i from Firebase"
}


#-------------------#
#   Backend Tests   #
#-------------------#


#### POST TESTS ####

post(){
    i=$1
    # creator="${UIDS[$i-1]}"
    idToken="${ID_TOKENS[$i-1]}"
    
    ENDPOINT="api/image"
    CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' -X POST \
    -H 'Authorization: Bearer $idToken' \
    -F 'image=@LEC_img$i.jpg' \
    $BACKEND_URL/$ENDPOINT"
}

test_post(){
    i=$1

    post $i
    run_test "POST test <Upload New Image: LEC_img$i.jpg>" 201
}


test_post_one_user_multiple_images(){

    idToken="${ID_TOKENS[0]}"

    for (( i = 1; i < IMG_LIMIT+1; i++)); do
		ENDPOINT="api/image"
		CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' -X POST \
        -H 'Authorization: Bearer $idToken' \
		-F 'image=@LEC_img$i.jpg' \
        $BACKEND_URL/$ENDPOINT"
		run_test "POST Multiple test <Upload New Image On One User: LEC_img$i.jpg>" 201
    done
}

test_post_one_user_multiple_images_failure_exceeds_limit(){

    idToken="${ID_TOKENS[0]}"

    for (( i = IMG_LIMIT+1; i < 5; i++)); do
		ENDPOINT="api/image"
		CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' -X POST \
        -H 'Authorization: Bearer $idToken' \
		-F 'image=@LEC_img$i.jpg' \
        $BACKEND_URL/$ENDPOINT"
		run_test "POST Multiple test <Upload New Image On One User MAX LIMIT EXCEEDED: LEC_img$i.jpg>" 500
    done
}

#### GET TESTS ####

get(){
    i=$1
    creator="${UIDS[$i-1]}"
    idToken="${ID_TOKENS[$i-1]}"

    ENDPOINT="api/image/$creator/LEC_img$i.jpg"
    CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' -X GET $BACKEND_URL/$ENDPOINT \
    -H 'Authorization: Bearer $idToken' -o dl_LEC_img$i.jpg"

}

test_get(){
    i=$1

    get $i
	run_test "GET test <OK: LEC_img$i.jpg>" 200
}


test_get_failure_entity_not_found(){
    i=$1

    get $i
	run_test "GET test <Entity Not Found: LEC_img$i.jpg>" 404
}


test_get_all(){
    idToken="${ID_TOKENS[$i-1]}"
	
    ENDPOINT="api/images"
	CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' -H 'Authorization: Bearer $idToken' \
    $BACKEND_URL/$ENDPOINT"
	run_test "GET test <OK: Get All Image Names>" 200
}

#### DELETE TESTS ####
delete(){
    i=$1
    creator="${UIDS[$i-1]}"
    idToken="${ID_TOKENS[$i-1]}"

    ENDPOINT="api/image/$i"
    CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' -X DELETE \
    -H 'Authorization: Bearer $idToken' \
    -F 'image=@LEC_img$i.jpg' \
    $BACKEND_URL/$ENDPOINT/LEC_img$i.jpg"

}

test_delete(){
    i=$1

    delete "$i"
	run_test "DELETE test <Delete Image: LEC_img$i.jpg>" 200
}

test_delete_failure_entity_not_found(){
    i=$1

    delete "$i"
	run_test "DELETE test <Entity Not Found: LEC_img$i.jpg>" 404
}

#### Users ####

create_backend_user(){
    idToken=$1
    i=$2
    ENDPOINT="api/user"
    CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' -X POST \
    -H 'Authorization: Bearer $idToken' \
    $BACKEND_URL/$ENDPOINT"

    run_test "POST Test <Create & Authorize New Entity: User $i" 201
}

delete_backend_user(){
    idToken=$1
    i=$2
    uid=$3
    ENDPOINT="api/user/$uid"
    CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' -X DELETE \
    -H 'Authorization: Bearer $idToken' \
    $BACKEND_URL/$ENDPOINT"
	run_test "DELETE Test <No Content: User $i>" 204
}

#### Execute Test Command ####
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
				if ($i ~ /"starttime":/) {
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



setup
build_tests