#!/usr/bin/env bash
# End-to-end smoke test of the REST API against a running local server.
# Usage: make up && make seed && make run   (in another terminal)  then: ./scripts/smoke.sh
# Requires curl and jq. Creates two throwaway owners with random emails (alice-<run>@example.com, bob-<run>@example.com).
# CLEANUP=1 deletes them (and their pets, posts, walks) and the uploaded test photo afterwards
# through `docker compose exec postgres` — local development only; there is no delete-account endpoint.
set -euo pipefail
cd "$(dirname "$0")/.."

API=${API:-http://127.0.0.1:8080/api/v1}
SPOT=11111111-1111-4111-8111-111111111111 # from migrations/seed/walk_spots.sql
RUN=$(date +%s)-$RANDOM
failures=0

# call METHOD PATH [BODY] [TOKEN] -> sets $status and $body
call() {
  local method=$1 path=$2 data=${3:-} token=${4:-}
  local args=(-s -o /tmp/tailverse-smoke-body -w '%{http_code}' -X "$method" "$API$path" -H 'Content-Type: application/json')
  [[ -n $token ]] && args+=(-H "Authorization: Bearer $token")
  [[ -n $data ]] && args+=(-d "$data")
  status=$(curl "${args[@]}")
  body=$(cat /tmp/tailverse-smoke-body)
}

expect() {
  local want=$1 name=$2
  if [[ $status == "$want" ]]; then
    printf '  ok   %-58s %s\n' "$name" "$status"
  else
    printf '  FAIL %-58s got %s, want %s: %s\n' "$name" "$status" "$want" "$body"
    failures=$((failures + 1))
  fi
}

echo "auth"
call POST /auth/register "{\"email\":\"alice-$RUN@example.com\",\"password\":\"password123\",\"nickname\":\"alice\"}"
expect 201 "register alice"
ALICE=$(jq -r .access_token <<<"$body")
ALICE_REFRESH=$(jq -r .refresh_token <<<"$body")
call POST /auth/register "{\"email\":\"alice-$RUN@example.com\",\"password\":\"password123\",\"nickname\":\"dup\"}"
expect 409 "register duplicate email"
call POST /auth/register '{"email":"not-an-email","password":"short","nickname":""}'
expect 400 "register invalid body"
call POST /auth/register "{\"email\":\"bob-$RUN@example.com\",\"password\":\"password123\",\"nickname\":\"bob\"}"
expect 201 "register bob"
BOB=$(jq -r .access_token <<<"$body")
call POST /auth/login "{\"email\":\"ALICE-$RUN@example.com\",\"password\":\"password123\"}"
expect 200 "login (email is case-insensitive)"
call POST /auth/login "{\"email\":\"alice-$RUN@example.com\",\"password\":\"wrong-password\"}"
expect 401 "login wrong password"
call POST /auth/refresh "{\"refresh_token\":\"$ALICE_REFRESH\"}"
expect 200 "refresh"
call POST /auth/refresh "{\"refresh_token\":\"$ALICE\"}"
expect 401 "refresh with an access token"

echo "owner profile"
call GET /owners/me
expect 401 "own profile without token"
call GET /owners/me "" "$ALICE"
expect 200 "own profile"
ALICE_ID=$(jq -r .id <<<"$body")
[[ $(jq -r .email <<<"$body") == "alice-$RUN@example.com" && $(jq -r .visibility.gender <<<"$body") == false ]] || { echo "  FAIL own profile: $body"; failures=$((failures + 1)); }
call PATCH /owners/me '{"gender":"female","visibility":{"gender":false}}' "$ALICE"
expect 200 "set gender (hidden)"
call GET "/owners/$ALICE_ID"
expect 200 "public profile"
[[ $(jq -r .gender <<<"$body") == null && $(jq 'has("email")' <<<"$body") == false && $(jq -r .nickname <<<"$body") == alice ]] || { echo "  FAIL public profile must hide gender and email: $body"; failures=$((failures + 1)); }
call PATCH /owners/me '{"visibility":{"gender":true}}' "$ALICE"
expect 200 "show gender"
call GET "/owners/$ALICE_ID"
[[ $(jq -r .gender <<<"$body") == female ]] || { echo "  FAIL public profile should show gender now: $body"; failures=$((failures + 1)); }
call PATCH /owners/me '{"gender":"unicorn"}' "$ALICE"
expect 400 "unknown gender"

echo "pets"
call POST /pets '{"name":"Rex","breed":"Golden Retriever","approx_address":"x"}'
expect 401 "create pet without token"
call POST /pets '{"name":"Rex","breed":"Golden Retriever","birth_date":"2021-04-10","approx_address":"Хамовники"}' "$ALICE"
expect 201 "create pet (species from breed)"
REX=$(jq -r .id <<<"$body")
[[ $(jq -r .species <<<"$body") == dog && $(jq -r .age <<<"$body") != null ]] || { echo "  FAIL species/age: $body"; failures=$((failures + 1)); }
call POST /pets '{"name":"Kesha","breed":"Budgerigar","approx_address":"x"}' "$ALICE"
expect 400 "create pet with unknown breed and no species"
call POST /pets '{"name":"Rex","breed":"x","birth_date":"10.04.2021","approx_address":"x"}' "$ALICE"
expect 400 "create pet with bad birth_date"
call POST /pets '{"name":"Murka","breed":"Maine Coon","approx_address":"Сокольники"}' "$BOB"
expect 201 "create bob's pet"
MURKA=$(jq -r .id <<<"$body")
call GET "/pets/$REX"
expect 200 "get pet (public)"
call GET /pets/00000000-0000-4000-8000-000000000000
expect 404 "get missing pet"
call GET /pets/not-a-uuid
expect 400 "get pet with invalid id"
call GET /pets "" "$ALICE"
expect 200 "list own pets"
[[ $(jq '.pets | length' <<<"$body") == 1 ]] || { echo "  FAIL list should contain only alice's pet: $body"; failures=$((failures + 1)); }
call PATCH "/pets/$REX" '{"name":"Rex II"}' "$ALICE"
expect 200 "patch own pet"
call PATCH "/pets/$REX" '{"name":"Stolen"}' "$BOB"
expect 403 "patch someone else's pet"
call DELETE "/pets/$REX" "" "$BOB"
expect 403 "delete someone else's pet"
call PATCH "/pets/$REX" '{"breed":""}' "$ALICE"
expect 400 "patch breed to empty"
call PATCH "/pets/$REX" '{"approx_address":"   "}' "$ALICE"
expect 400 "patch area to blank"
call PATCH "/pets/$REX" '{"birth_date":null}' "$ALICE"
expect 200 "patch birth_date null (unchanged)"
[[ $(jq -r .birth_date <<<"$body") == 2021-04-10 ]] || { echo "  FAIL null must keep the birth date: $body"; failures=$((failures + 1)); }
call PATCH "/pets/$REX" '{"birth_date":""}' "$ALICE"
expect 200 "patch birth_date \"\" (clear)"
[[ $(jq -r .birth_date <<<"$body") == null && $(jq -r .age <<<"$body") == null ]] || { echo "  FAIL birth date should be cleared: $body"; failures=$((failures + 1)); }
call PATCH "/pets/$REX" '{"birth_date":"2021-04-10"}' "$ALICE"
expect 200 "patch birth_date back"

echo "walk spots + presence"
call GET "/walkspots?lat=55.7298&lng=37.6010&radius_m=3000"
expect 200 "list nearby spots"
[[ $(jq '.spots | length' <<<"$body") -ge 2 ]] || { echo "  FAIL expected seeded spots nearby: $body"; failures=$((failures + 1)); }
call GET "/walkspots?lat=55.7298"
expect 400 "list spots without lng"
call POST "/walkspots/$SPOT/checkin" "{\"pet_id\":\"$REX\"}" "$ALICE"
expect 200 "check in"
call POST "/walkspots/$SPOT/checkin" "{\"pet_id\":\"$MURKA\"}" "$ALICE"
expect 403 "check in someone else's pet"
call POST /walkspots/00000000-0000-4000-8000-000000000000/checkin "{\"pet_id\":\"$REX\"}" "$ALICE"
expect 404 "check in at missing spot"
call GET "/walkspots/$SPOT"
expect 200 "spot details"
[[ $(jq -r '.present[0].pet_name' <<<"$body") == "Rex II" ]] || { echo "  FAIL Rex should be present: $body"; failures=$((failures + 1)); }
call GET "/walkspots?lat=55.7298&lng=37.6010&radius_m=100"
[[ $(jq ".spots[] | select(.id == \"$SPOT\") | .present_count" <<<"$body") == 1 ]] || { echo "  FAIL present_count should be 1: $body"; failures=$((failures + 1)); }
call DELETE "/walkspots/$SPOT/checkin" "{\"pet_id\":\"$REX\"}" "$ALICE"
expect 204 "check out"
call DELETE "/walkspots/$SPOT/checkin" "{\"pet_id\":\"$REX\"}" "$ALICE"
expect 404 "check out again"
call GET "/walkspots?lat=55.7298&lng=37.6010&radius_m=3000"
[[ $(jq '[.spots[] | has("distance_m")] | any' <<<"$body") == false ]] || { echo "  FAIL GET /walkspots must stay unchanged (no distance_m): $body"; failures=$((failures + 1)); }

echo "walk spot picker (nearby)"
call GET "/walkspots/nearby?lat=55.7298&lng=37.6010"
expect 401 "nearby without token"
call GET "/walkspots/nearby?lat=55.7298&lng=37.6010" "" "$ALICE"
expect 200 "nearby (default 500 m)"
[[ $(jq '[.spots[].distance_m] | (all(. <= 500)) and (. == sort)' <<<"$body") == true ]] || { echo "  FAIL nearby must be within 500 m, closest first: $body"; failures=$((failures + 1)); }
call GET "/walkspots/nearby?lat=55.7298&lng=37.6010&radius_m=501" "" "$ALICE"
expect 400 "nearby radius over 500 m"
call GET "/walkspots/nearby?lat=55.7298" "" "$ALICE"
expect 400 "nearby without lng"

echo "uploads"
PNG=$(mktemp -t tailverse-smoke).png
base64 -d <<<'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNkYPhfDwAChwGA60e6kgAAAABJRU5ErkJggg==' >"$PNG"
upload() {
  local file=$1 token=${2:-}
  local args=(-s -o /tmp/tailverse-smoke-body -w '%{http_code}' -X POST "$API/uploads" -F "file=@$file")
  [[ -n $token ]] && args+=(-H "Authorization: Bearer $token")
  status=$(curl "${args[@]}")
  body=$(cat /tmp/tailverse-smoke-body)
}
upload "$PNG"
expect 401 "upload without token"
upload "$PNG" "$ALICE"
expect 201 "upload a png"
PHOTO=$(jq -r .url <<<"$body")
[[ $PHOTO == "${API%/api/v1}/uploads/"*.png ]] || { echo "  FAIL upload url should use the host the client called: $PHOTO"; failures=$((failures + 1)); }
[[ $(curl -s -o /dev/null -w '%{http_code} %{content_type}' "$PHOTO") == "200 image/png" ]] || { echo "  FAIL uploaded file is not served: $PHOTO"; failures=$((failures + 1)); }
echo "plain text" >"$PNG.txt"
upload "$PNG.txt" "$ALICE"
expect 415 "upload a text file"
rm -f "$PNG" "$PNG.txt"

echo "announcements"
STARTS=$(date -u -v+1H +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || date -u -d '+1 hour' +%Y-%m-%dT%H:%M:%SZ)
call POST /announcements "{\"pet_id\":\"$REX\",\"spot_id\":\"$SPOT\",\"starts_at\":\"$STARTS\",\"duration_min\":60}" "$ALICE"
expect 201 "create announcement at spot"
ANN=$(jq -r .id <<<"$body")
call POST /announcements "{\"pet_id\":\"$REX\",\"custom_point\":{\"lat\":55.75,\"lng\":37.62},\"starts_at\":\"$STARTS\",\"duration_min\":30}" "$ALICE"
expect 201 "create announcement at custom point"
call POST /announcements "{\"pet_id\":\"$REX\",\"starts_at\":\"$STARTS\",\"duration_min\":30}" "$ALICE"
expect 400 "create announcement without place"
call POST /announcements "{\"pet_id\":\"$REX\",\"spot_id\":\"00000000-0000-4000-8000-000000000000\",\"starts_at\":\"$STARTS\",\"duration_min\":30}" "$ALICE"
expect 400 "create announcement at missing spot"
call GET "/announcements?lat=55.7298&lng=37.6010&radius_m=5000"
expect 200 "list nearby announcements"
[[ $(jq '.announcements | length' <<<"$body") -ge 2 ]] || { echo "  FAIL expected both announcements: $body"; failures=$((failures + 1)); }
call POST "/announcements/$ANN/join" "{\"pet_id\":\"$MURKA\"}" "$BOB"
expect 200 "join announcement"
call POST "/announcements/$ANN/join" "{\"pet_id\":\"$MURKA\"}" "$BOB"
expect 409 "join twice"
call POST "/announcements/$ANN/join" "{\"pet_id\":\"$REX\"}" "$ALICE"
expect 400 "author's pet joins own walk"
call GET "/announcements/$ANN"
expect 200 "announcement details"
[[ $(jq -r '.participants[0].owner_nickname' <<<"$body") == bob ]] || { echo "  FAIL bob should participate: $body"; failures=$((failures + 1)); }
call DELETE "/announcements/$ANN/join" "{\"pet_id\":\"$MURKA\"}" "$BOB"
expect 204 "leave announcement"

echo "feed"
for i in 1 2 3; do
  call POST /posts "{\"pet_id\":\"$REX\",\"spot_id\":\"$SPOT\",\"text\":\"Прогулка $i\",\"photo_urls\":[\"https://cdn.example.com/$i.jpg\"]}" "$ALICE"
  expect 201 "create post $i"
done
POST_ID=$(jq -r .id <<<"$body")
call POST /posts "{\"pet_id\":\"$MURKA\",\"text\":\"hi\"}" "$ALICE"
expect 403 "post as someone else's pet"
call POST /posts "{\"pet_id\":\"$REX\",\"spot_id\":\"00000000-0000-4000-8000-000000000000\",\"text\":\"hi\"}" "$ALICE"
expect 400 "post tagged with a missing spot"
call POST /posts "{\"pet_id\":\"$REX\",\"text\":\"С фото\",\"photo_urls\":[\"$PHOTO\"]}" "$ALICE"
expect 201 "post with an uploaded photo, no place"
[[ $(jq -r .spot_id <<<"$body") == null && $(jq -r '.photo_urls[0]' <<<"$body") == "$PHOTO" ]] || { echo "  FAIL post with photo: $body"; failures=$((failures + 1)); }
call GET /posts?limit=1
[[ $(jq -r '.posts[0].photo_urls[0]' <<<"$body") == "$PHOTO" ]] || { echo "  FAIL the global feed should start with the newest post: $body"; failures=$((failures + 1)); }
call GET "/posts?spot_id=$SPOT&limit=2"
expect 200 "list posts page 1"
CURSOR=$(jq -r .next_cursor <<<"$body")
[[ $CURSOR != null && $(jq '.posts | length' <<<"$body") == 2 ]] || { echo "  FAIL page 1 should have 2 posts and a cursor: $body"; failures=$((failures + 1)); }
call GET "/posts?spot_id=$SPOT&limit=2&cursor=$CURSOR"
expect 200 "list posts page 2"
call GET "/posts?cursor=garbage"
expect 400 "list posts with bad cursor"
call GET "/posts/$POST_ID"
expect 200 "get post"
call DELETE "/posts/$POST_ID" "" "$BOB"
expect 403 "delete someone else's post"
call DELETE "/posts/$POST_ID" "" "$ALICE"
expect 204 "delete own post"
call GET "/posts/$POST_ID"
expect 404 "get deleted post"

echo "services"
call POST /services '{"title":"Груминг","description":"Стрижка","category":"grooming","lat":55.73,"lng":37.60,"price":1500}' "$BOB"
expect 201 "create service"
SERVICE=$(jq -r .id <<<"$body")
call POST /services '{"title":"x","category":"Bad Category!","lat":55.73,"lng":37.60,"price":1}' "$BOB"
expect 400 "create service with bad category"
call GET "/services?lat=55.7298&lng=37.6010&radius_m=3000&category=grooming"
expect 200 "list services by category"
call GET "/services/$SERVICE"
expect 200 "get service"

echo "cleanup"
call DELETE "/pets/$REX" "" "$ALICE"
expect 204 "delete own pet"
call GET "/pets/$REX"
expect 404 "get deleted pet"

if [[ ${CLEANUP:-0} == 1 ]]; then
  docker compose exec -T postgres psql -q -U "${POSTGRES_USER:-postgres_user}" -d "${POSTGRES_DB:-postgres_db}" \
    -c "DELETE FROM owners WHERE email IN ('alice-$RUN@example.com', 'bob-$RUN@example.com')" >/dev/null &&
    echo "  removed the test owners" || echo "  could not remove the test owners"
  rm -f "${UPLOADS_DIR:-./data/uploads}/${PHOTO##*/}"
fi

echo
if ((failures > 0)); then
  echo "$failures check(s) failed"
  exit 1
fi
echo "all checks passed"
