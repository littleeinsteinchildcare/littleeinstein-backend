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

# Outer for loop
# Set Email
# Set Password

#--------------------#
#      Firebase      #
#--------------------#

declare -a ID_TOKENS
declare -a UIDS

create_firebase_users(){


    ## SETUP LOOP TO CREATE USERS
    #! NEED TO SETUP STORING IDTOKENS AND IDS IN ARRAYS SO EVENTS CAN USE THEM
    for (( i = 1; i < ("$NUM_TESTS"+3); i++)) do
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
        echo "$(green "Firebase User$i successfully created")"
        echo ""
        idToken=$(jq -r '.idToken' <<< "$resp")
        uid=$(jq -r '.localId' <<< "$resp")

        ID_TOKENS+=("$idToken")
        UIDS+=("$uid")

        echo "UID: $uid"
        create_backend_user $idToken "$i"

    done

    for (( i = 1; i <("$NUM_TESTS"+1); i++)); do
        echo "EVENT LOOP"
        creator="${UIDS[$i-1]}"
        idToken="${ID_TOKENS[$i-1]}"
        inv1="${UIDS[$((i))]}"
        inv2="${UIDS[$((i+1))]}"

        echo "CREATOR: $creator" 
        echo "INVITEE 1: $inv1" 
        echo "INVITEE 2: $inv2" 

        ENDPOINT="api/event"
        CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' -X POST \
		-H 'Content-Type: application/json' \
        -H 'Authorization: Bearer $idToken' \
		-d '{\"eventname\":\"Event $i\", \"date\":\"1/$i/2025\",\"id\":\"$creator\", \"starttime\":\"$i:00am\",\"endtime\":\"$i:00pm\", \"invitees\":\"$inv1, $inv2\"}' \
        $BACKEND_URL/$ENDPOINT"
		run_test "POST test < Create New Entity: Event $i -- Creator: $creator >" 201

    done

    for ((i = 1; i < ("$NUM_TESTS"+3); i++)); do
        uid="${UIDS[$i-1]}"
        idToken="${ID_TOKENS[$i-1]}"
        delete_backend_user $idToken "$i" "$uid"
        delete_firebase_user $idToken "$i"
        echo "$(green "Firebase User$i successfully deleted")"
        echo ""
    done

    # echo "CALLING EVENT LOOP"
    # for idx in  "${!UIDS[@]}"; do
    #     idToken="${ID_TOKENS[$idx]}"
    #     uid="${UIDS[$idx]}"

    #     echo "ID TOKEN: $idToken"
    #     echo "UID: $uid"
    # done

}

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

create_backend_user(){
    idToken=$1
    i=$2
    ENDPOINT="api/user"
    CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' -X POST \
    -H 'Authorization: Bearer $idToken' \
    $BACKEND_URL/$ENDPOINT"

    run_test "POST Test <Create & Authorize New Entity: User $i" 201
}

test_get(){
    idToken=$1
    i=$2
    uid=$3
    ENDPOINT="api/user/$uid"
	CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' \
    -H 'Authorization: Bearer $idToken' \
    $BACKEND_URL/$ENDPOINT"

	run_test "GET Test <OK: User $i>" 200

}

test_delete(){
    idToken=$1
    i=$2
    uid=$3
    ENDPOINT="api/user/$uid"
    CURL_CMD="curl -s -w 'HTTPSTATUS:%{http_code}' -X DELETE \
    -H 'Authorization: Bearer $idToken' \
    $BACKEND_URL/$ENDPOINT"
	run_test "DELETE Test <No Content: User $i>" 204
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



setup
create_firebase_users