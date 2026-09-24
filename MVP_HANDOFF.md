# UnAlone local MVP handoff

Status recorded 2026-09-11. Implementation is paused at the user's request. Core flows are implemented and tested; final acceptance and cleanup remain. This is the current handoff; `output/UNALONE_PRD_CODEBASE_REVIEW.md` describes the earlier, pre-implementation state.

## Agreed scope

- Finish and test the local MVP first. AWS deployment and production email delivery are deferred.
- Follow `UNALOne FINAL PRD.pdf` v2.1, read in full (22 pages including the final blank page).
- User explicitly approved adjusting the Redis design to enforce strict coordinate expiry: expiring GEO time buckets, expiring user presence, and an expiry-event subscriber for silent hotspot dissolution.
- No friends, chat, RSVP, or other post-MVP additions.

## Implemented

### Backend

- Restored Go dependencies and fixed compilation blockers.
- Development email OTP login: cryptographic six-digit codes, hashed storage with five-minute expiry, atomic consume, email normalization, send/verification limits.
- Stable Mongo user identity; HS256 JWT validation with issuer and expiry; HttpOnly session cookie; session lookup and logout.
- Validated meetup creation and nearby upcoming-meetup retrieval, persistent Mongo records, public WebSocket payloads without creator email.
- Atomic location updates with five-second rate limit; presence expires after 30 seconds. Per-second GEO buckets have fixed expiry at bucket timestamp plus 30 seconds (coordinates live approximately 29–30 seconds). Later writes do not extend older coordinates.
- Nearby lookup deduplicates buckets and verifies current presence/current coordinates. Explicit location pause deletes presence; old GEO coordinates expire within the retention window.
- Hotspot formation at three nearby active users within 50 metres, geohash IDs, formation/dissolution events, overlap suppression, snapshot endpoint, expiry-event-driven dissolution without scheduled cleanup.
- Authenticated WebSockets with bounded queues, ping/pong, origin checks, reconnection support, graceful shutdown.
- API recovery, CORS/credentials, unsafe-request origin checks, request size limit, health endpoint, query-free request logging.
- Redis initialization refuses snapshot/AOF persistence. Production mode intentionally disabled until deployment work.

### Frontend

- Cookie-session bootstrap and OTP login; Discover, Create meetup, Activity and Profile pages; logout.
- Foreground geolocation, pause/resume, hidden-tab suspension and protection against late callbacks.
- Map with own position, hotspot circles, meetup pins and information cards; marker cleanup.
- Local map uses OSM raster tiles without a Mapbox token; optional token enables Mapbox-hosted streets.
- Validated meetup form; newly created meetup appears immediately; persisted meetups reload after refresh.
- WebSocket reconnect/backoff and online/offline handling, hotspot snapshots, bounded in-memory activity feed, deduplication and snapshot/event race protection.
- Responsive styling and status/error UI; same-origin API/WebSocket proxy through Vite and nginx.
- Dependency/type fixes, lockfiles, build/typecheck/browser-test scripts, environment examples and ignore files.

### Local infrastructure

- Mongo and Redis run in Docker, with isolated localhost ports 27018 and 6380.
- Redis persistence disabled; expiry notifications enabled. Mongo uses a persistent volume.
- Compose healthchecks and Dockerfile/proxy fixes added.
- Current tested app mode is native Go + native Vite + Docker databases. Final full-container app startup remains unverified.

## Verification completed

- `go test ./...` passed: compile and unit tests for JWT, email, coordinates and distance.
- `go test -tags integration ./api -run TestLocalMVP -count=1 -v -timeout 150s` passed in 65.35 seconds against real Mongo/Redis. Covered authentication/OTP limits and replay, stable identity, meetup validation/radius/privacy, hotspot formation/rate limiting/manual and silent dissolution, TTL behaviour, WebSocket connection beyond 60 seconds, cookies and cross-origin logout protection. Integration tests cleaned their isolated Mongo database and Redis DB 15.
- Frontend typecheck passed during implementation. A baseline build passed before all changes. Final build/typecheck after the last edits still need confirmation.
- Final browser run passed all eight checks in `output/browser-check/report.json`: OTP/location; pause/resume/hidden-tab lifecycle; immediate meetup display; pin card; session/meetup persistence; refresh/feed events; network reconnect; identity/logout.
- Browser checks used mocked geolocation and map tiles. Hidden-tab behaviour was simulated. Real GPS and live external basemap need manual verification. Mobile login overflow and browser JavaScript errors were checked.
- Last status check: frontend `http://127.0.0.1:5173` returned 200; API `http://127.0.0.1:8080/health` returned `{"status":"ok"}`.
- A Docker build was started near interruption, but no final result was collected. `docker compose images` showed running database container images only; that does not establish backend/frontend build success or failure.

## Remaining before closing the local milestone

1. Review final source changes against PRD acceptance criteria; fix any remaining gaps uncovered.
2. Run final frontend production build/typecheck and appropriate final backend checks after any further changes.
3. Confirm Docker builds and full Compose app startup. Avoid competing with the native servers on ports 8080/5173.
4. Manually check real browser GPS permissions, live basemap and complete user flow; review final screenshots.
5. Rewrite the stale README with current setup, API contracts, privacy design and testing instructions.
6. Remove only our browser-test users/meetups from the app database. Integration-test data is already cleaned. Exact browser-test records are below.
7. Record final acceptance results and any limitations. No load/concurrency or multi-instance acceptance has been established.

Deferred: AWS, SES/production email, HTTPS/deployment configuration, production hardening and scaling validation. Hotspot state and WebSocket fanout currently assume one backend process and one local namespace. Nearby meetup filtering currently happens in Go after fetching future meetups, so production-scale query work remains.

## Open and sign in

Open `http://127.0.0.1:5173`, enter an email and request a sign-in code. Local mode does not send email: the backend console prints `Development OTP for <email>: <six digits> (expires in 5m0s)`. Enter that code. Codes expire after five minutes; send limit is five per email per hour. Allow browser location permission to exercise discovery.

At handoff the backend and frontend were started in Codex command sessions, so their logs may not appear in the user's own terminal. The assistant can retrieve the active backend output while that session exists. Backend exec session was 12337; frontend exec session was 4804. These IDs are temporary, not durable setup instructions.

For the next restart, use these commands after stopping any existing instance using the same ports.

Project-root terminal:

```powershell
cd C:\Anika\Projects\UnAlone-Live-Social-Discovery-App-main
docker compose up -d mongo redis
```

Backend terminal (leave open; sign-in codes appear here):

```powershell
cd C:\Anika\Projects\UnAlone-Live-Social-Discovery-App-main\backend
go run .
```

Separate frontend terminal:

```powershell
cd C:\Anika\Projects\UnAlone-Live-Social-Discovery-App-main\frontend
npm run dev -- --host 127.0.0.1
```

Dependencies are already installed on this machine. Browser test: run `npm run test:browser` from frontend with both servers running; supply the requested OTP from the backend console. It writes screenshots and a report under `output/browser-check`.

## Browser-test records pending cleanup

Delete only these exact test emails and meetup IDs; do not drop the application database.

| Test email | Meetup IDs |
| --- | --- |
| browser-1789025437341@example.com | 23c7c47f-a0a7-4099-9425-6f8bbdc23eca; de533f38-e52e-4118-ac82-ebdd51c45295 |
| browser-1789025648044@example.com | 1a12a329-7149-4d1f-9ad6-c4b1136fe7ce; 0a2d257d-037c-46bc-b27a-f5f8631a8c87 |
| browser-1789025794951@example.com | 1636d9d5-03d1-48b8-85c4-4ec5b5d1e4b6; af0575ab-e3c8-4295-bd83-bae4a969ece0 |
| browser-1789025933609@example.com | 50503a1d-4959-4408-92ee-c16fa5f40ab3; 7201afa4-09b2-4b5d-8794-6a0318cfb435 |

## Repository state

This folder is a downloaded snapshot with no Git metadata. Changes are saved on disk; no commits were created. Continue from these files rather than reconstructing work from the historical review.
