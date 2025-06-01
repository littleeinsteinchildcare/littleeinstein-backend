#!/bin/bash

# Grab Firebase Web API Key from .env and strip out the "FIREBASE_WEB_API_KEY=" section
API_KEY=$(grep -E '^FIREBASE_WEB_API_KEY=' ../.env | cut -d= -f2-)
USER_EMAIL=
USER_PWD=
BACKEND_URL="http://localhost:8080"

# Flags
VERBOSE=false
NUM_TESTS=1
PROD=false

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

build_tests(){
    for (( i = 1; i < ("$NUM_TESTS"+1); i++)) do

        # Create Firebase User
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

        # Run Backend Tests
        test_post $idToken "$i"
        test_post_failure_entity_already_exists $idToken "$i"
        test_get $idToken "$i" "$uid"
        test_get_all $idToken "$uid"
        test_update $idToken "$i" "$uid"
        test_delete $idToken "$i" "$uid"
        test_delete_failure_entity_not_found $idToken "$i" "$uid"
        test_get_failure_entity_not_found $idToken "$i" "$uid"
        test_update_failure_entity_not_found $idToken "$i" "$uid"

        # Firebase Cleanup
        delete_firebase_user $idToken "$i"
        echo "$(blue "Firebase User$i successfully deleted")"
        echo ""
    done
}

check_firebase_err(){
    response=$1
    msg=$2
    if jq -e '.error' <<<"$response" >/dev/null; then
        errMsg=$(jq -r '.error.message' <<<"$response")
        echo "$(red "$msg: $errMsg")" >&2
        exit 1
    fi
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
    idToken=$1
    ENDPOINT="api/user"
    CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' -X POST \
    -H 'Authorization: Bearer $idToken' \
    $BACKEND_URL/$ENDPOINT"
}

test_post(){
    idToken=$1
    i=$2
    post $idToken
    # ENDPOINT="api/user"
    # CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' -X POST \
    # -H 'Authorization: Bearer $idToken' \
    # $BACKEND_URL/$ENDPOINT"

    run_test "POST Test <Create & Authorize New Entity: User $i" 201
}

test_post_failure_entity_already_exists(){
    idToken=$1
    i=$2
    post $idToken
    # ENDPOINT="api/user"
    # CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' -X POST \
    # -H 'Authorization: Bearer $idToken' \
    # $BACKEND_URL/$ENDPOINT"

	run_test "POST test <Entity Already Exists: User $i>" 409
}


#### GET TESTS ####
get(){
    idToken=$1
    uid=$2

    ENDPOINT="api/user/$uid"
	CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' \
    -H 'Authorization: Bearer $idToken' \
    $BACKEND_URL/$ENDPOINT"

}

test_get(){
    idToken=$1
    i=$2
    uid=$3
    # ENDPOINT="api/user/$uid"
	# CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' \
    # -H 'Authorization: Bearer $idToken' \
    # $BACKEND_URL/$ENDPOINT"
    get $idToken "$uid"

	run_test "GET Test <OK: User $i>" 200
}


test_get_failure_entity_not_found(){
    idToken=$1
    i=$2
    uid=$3
    # ENDPOINT="api/user/$uid"
	# CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' \
    # -H 'Authorization: Bearer $idToken' \
    # $BACKEND_URL/$ENDPOINT"

    get $idToken "$uid"
	run_test "GET test <Entity Not Found: User $i>" 404
}

test_get_all(){
    idToken=$1
    uid=$2
    ENDPOINT="api/users"
	CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' \
    -H 'Authorization: Bearer $idToken' \
    $BACKEND_URL/$ENDPOINT"
	run_test "GET test <OK: Get All Users>" 200
}


#### PUT TESTS ####
update(){
    idToken=$1
    i=$2
    uid=$3
    ENDPOINT="api/user/$uid"
    CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}'  -X PUT \
		-H 'Content-Type: application/json' \
        -H 'Authorization: Bearer $idToken' \
		-d '{\"name\":\"UPDATED[$i] User $i\", \"email\":\"UPDATEDuser$i@example.com\",\"id\":\"$uid\", \"role\":\"member\"}' \
        $BACKEND_URL/$ENDPOINT"


}

test_update(){
    idToken=$1
    i=$2
    uid=$3
    # ENDPOINT="api/user/$uid"
    # CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}'  -X PUT \
	# 	-H 'Content-Type: application/json' \
    #     -H 'Authorization: Bearer $idToken' \
	# 	-d '{\"name\":\"UPDATED[$i] User $i\", \"email\":\"UPDATEDuser$i@example.com\",\"id\":\"$uid\", \"role\":\"member\"}' \
    #     $BACKEND_URL/$ENDPOINT"
    update $idToken "$i" "$uid"

	run_test "PUT test <OK: User $i>" 200
}

test_update_failure_entity_not_found(){
    idToken=$1
    i=$2
    uid=$3
    # ENDPOINT="api/user/$uid"
    # CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}'  -X PUT \
	# 	-H 'Content-Type: application/json' \
    #     -H 'Authorization: Bearer $idToken' \
	# 	-d '{\"name\":\"UPDATED[$i] User $i\", \"email\":\"UPDATEDuser$i@example.com\",\"id\":\"$uid\", \"role\":\"member\"}' \
    #     $BACKEND_URL/$ENDPOINT"

    update $idToken "$i" "$uid"

	run_test "PUT test <Entity Not Found: User $i>" 404
}

#### DELETE TESTS ####
delete(){
    idToken=$1
    uid=$2
    ENDPOINT="api/user/$uid"
    CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' -X DELETE \
    -H 'Authorization: Bearer $idToken' \
    $BACKEND_URL/$ENDPOINT"
}


test_delete(){
    idToken=$1
    i=$2
    uid=$3
    # ENDPOINT="api/user/$uid"
    # CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' -X DELETE \
    # -H 'Authorization: Bearer $idToken' \
    # $BACKEND_URL/$ENDPOINT"
    delete $idToken "$uid"
	run_test "DELETE Test <No Content: User $i>" 204
}


test_delete_failure_entity_not_found(){
    idToken=$1
    i=$2
    uid=$3
    # ENDPOINT="api/user/$uid"
    # CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' -X DELETE \
    # -H 'Authorization: Bearer $idToken' \
    # $BACKEND_URL/$ENDPOINT"

    delete $idToken "$uid"
	run_test "DELETE test <Entity Not Found: User $i>" 404
}



### RUN TEST COMMAND ###
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

setup
build_tests