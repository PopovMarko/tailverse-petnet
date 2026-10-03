# Tailverse API

A social network for pets: the pet is the member, and its owner runs the profile. This repo is the Go backend (chi + PostgreSQL/PostGIS + Redis + WebSocket).

## Quick start (local)

Requirements: Go 1.27+, Docker with Compose.

```sh
cp .env.example .env   # or: make env
make up                # PostgreSQL + PostGIS and Redis in Docker, waits for healthchecks
make seed              # applies migrations and loads sample walk spots
make run               # API on http://127.0.0.1:8080 (migrations also run on every start)
```

`make dev` does `up`, `seed` and `run` in one go. `make down` stops the containers, and `make clean-db` also deletes the data volume.

To run everything in containers instead: `docker compose --profile full up --build`.

Check it: `curl http://127.0.0.1:8080/health` → `{"postgres":"ok","redis":"ok"}`.

## Architecture

Every domain follows **transport / service / repository**:

```
cmd/tailverse/main.go                 wiring: config → postgres/redis → services → chi router
internal/core/                        shared infrastructure
  auth/                               JWT issue/verify, owner id in context
  domain/                             domain models (Pet, Owner, WalkSpot, Announcement, Post, ServiceOffer, events)
  errors/                             ErrNotFound / ErrInvalidArgument / ErrConflict / ErrUnauthorized / ErrForbidden
  logger/                             zap logger (stdout + file in out/logs)
  repository/postgres, repository/redis   clients, migration runner, pg error mapping
  server/http/                        http.Server with graceful shutdown
  transport/http/                     middleware (request id, CORS, logger, trace, panic, auth), request decode, response
internal/<domain>/transport/http/     one handler per file + transport.go (service interface, handler, router) + dto_common.go
internal/<domain>/service/            business rules and validation
internal/<domain>/repository/         SQL
internal/presence/                    who is at a walk spot right now (Redis)
internal/ws/                          WebSocket hub for live updates
migrations/                           SQL migrations, embedded into the binary; seed/ has sample data
```

Domains: `auth` (register/login), `owner` (owner profile), `upload` (photos on disk), `pet`, `walkspot`, `announcement` ("Иду гулять"), `feed` (posts), `services` (grooming, dog walking…).

Service errors wrap `core_errors`. `ErrorResponse` maps them to 400/401/403/404/409, and anything else becomes 500 with the details logged but not sent to the client.

## API

Base prefix `/api/v1`, JSON everywhere. Authenticated routes need `Authorization: Bearer <access_token>`.

| Method | Path | Auth | Notes |
|---|---|---|---|
| POST | `/auth/register` | | `{email, password (≥8), nickname, gender?, avatar_url?, visibility?}` → 201 with tokens, 409 if email taken |
| POST | `/auth/login` | | `{email, password}` → tokens, 401 on bad credentials |
| POST | `/auth/refresh` | | `{refresh_token}` → new token pair |
| GET | `/owners/me` | ✓ | own profile `{id, email, nickname, gender, avatar_url, visibility: {gender, avatar_url}, created_at}` |
| PATCH | `/owners/me` | ✓ | any subset of `{nickname, gender, avatar_url, visibility: {gender?, avatar_url?}}`; `""` clears gender/avatar_url |
| GET | `/owners/{id}` | | public profile `{id, nickname, gender, avatar_url, created_at}`; hidden fields are `null`, never the email |
| POST | `/pets` | ✓ | `{name, breed, species?, birth_date? "YYYY-MM-DD", approx_address}`; species is derived from breed when omitted |
| GET | `/pets` | ✓ | `{pets: [...]}` of the current owner |
| GET | `/pets/{id}` | | pet with computed `age` |
| PATCH | `/pets/{id}` | ✓ owner | any subset of the fields |
| DELETE | `/pets/{id}` | ✓ owner | 204 |
| GET | `/walkspots?lat&lng&radius_m` | | `{spots: [{id, name, lat, lng, tags, present_count}]}`, radius defaults to 2000 m |
| GET | `/walkspots/nearby?lat&lng&radius_m` | ✓ | spot picker for "Иду гулять": radius defaults to 500 m and may not exceed 500; spots closest first with `distance_m` |
| GET | `/walkspots/{id}` | | spot + `present: [{pet_id, pet_name, owner_nickname, checked_in_at}]` |
| POST | `/walkspots/{id}/checkin` | ✓ | `{pet_id}` → `{spot_id, pet_id, checked_in_at, expires_at}` |
| DELETE | `/walkspots/{id}/checkin` | ✓ | `{pet_id}` → 204 |
| POST | `/announcements` | ✓ | `{pet_id, spot_id | custom_point {lat,lng}, starts_at, duration_min}` (exactly one place) |
| GET | `/announcements?lat&lng&radius_m&from&to` | | active walks that have not ended; `from`/`to` are RFC 3339 |
| GET | `/announcements/{id}` | | with `participants` |
| POST | `/announcements/{id}/join` | ✓ | `{pet_id}`, 409 if already joined |
| DELETE | `/announcements/{id}/join` | ✓ | `{pet_id}` → 204 |
| POST | `/posts` | ✓ | `{pet_id, spot_id?, text, photo_urls}` (URLs from `POST /uploads`) |
| GET | `/posts?spot_id&cursor&limit` | | `{posts, next_cursor}`, newest first, limit ≤ 50 |
| GET | `/posts/{id}` | | |
| DELETE | `/posts/{id}` | ✓ author | 204 |
| POST | `/services` | ✓ | `{title, description, category, lat, lng, price}` |
| GET | `/services?lat&lng&radius_m&category` | | |
| GET | `/services/{id}` | | |
| POST | `/uploads` | ✓ | multipart/form-data, field `file`: JPEG/PNG/WebP/HEIC ≤ 10 MB → 201 `{url}`; 413 too large, 415 other types |
| GET | `/ws/presence?token=<access_token>` | ✓ | WebSocket (see below) |

`GET /uploads/{name}` (outside `/api/v1`, no auth) serves the uploaded files.

### WebSocket `/api/v1/ws/presence`

The server pushes:

```json
{"type": "spot_update", "spot_id": "…", "present_count": 3}
{"type": "announcement_created", "announcement": {…}}
```

The client may send `{"type": "location_update", "lat": …, "lng": …}`. It is accepted but not used yet.

### Owner profile visibility

`visibility` controls what `GET /owners/{id}` shows to other users: `gender` (default `false`) and `avatar_url` (default `true`). The nickname is always public. `gender` is one of `male`, `female`, `other`, or unset.

### Photo uploads

`POST /api/v1/uploads` checks the file type by its content (the part's Content-Type is ignored), stores it under a random name in `UPLOADS_DIR` (default `./data/uploads`) and returns an absolute URL. The URL starts with `UPLOADS_PUBLIC_BASE_URL` when it is set, otherwise with the scheme and host the client used, e.g. `http://127.0.0.1:8080/uploads/<name>.jpg`. Use the URL in `photo_urls` of a post or as `avatar_url`.

### Presence in Redis

- `spot:{spot_id}:present` is a sorted set: member is `pet_id`, score is the check-in time. Each member expires on its own after `PRESENCE_TTL` (default 2h), and stale members are trimmed on every read.
- `pet:{pet_id}:location` holds the spot where the pet currently is, with the same TTL. Checking in somewhere else removes the pet from its previous spot.

## Example session

```sh
API=http://127.0.0.1:8080/api/v1
TOKEN=$(curl -s -X POST $API/auth/register -d '{"email":"me@example.com","password":"password123","nickname":"marko"}' | jq -r .access_token)
PET=$(curl -s -X POST $API/pets -H "Authorization: Bearer $TOKEN" \
  -d '{"name":"Rex","breed":"Golden Retriever","birth_date":"2021-04-10","approx_address":"Хамовники"}' | jq -r .id)
curl -s "$API/walkspots?lat=55.7298&lng=37.6010&radius_m=3000" | jq
curl -s -X POST $API/walkspots/11111111-1111-4111-8111-111111111111/checkin -H "Authorization: Bearer $TOKEN" -d "{\"pet_id\":\"$PET\"}" | jq
```

## Not implemented yet

These are open questions in the concept, or features outside the API spec:
- "Куда пойти?" recommendations: no endpoint in the API spec yet. `pets.approx_location` is in the schema for it.
- Changing `is_profile_public` through the API. The column exists; a hidden profile hides the nickname in "who is here" and participant lists, and shows only the nickname on `GET /owners/{id}`.
- Creating walk spots through the API. For now they come from `migrations/seed/walk_spots.sql`.
- Uploads are kept on the local disk, without thumbnails or cleanup of files nobody references.
- Refresh tokens are stateless JWTs with no server-side revocation.
