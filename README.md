# UnAlone

UnAlone is a privacy-first, real-time social discovery MVP. It lets signed-in people discover nearby live hotspots and public upcoming meetups without keeping a permanent history of their live location.

The governing product requirements are in [UNALOne FINAL PRD.pdf](./UNALOne%20FINAL%20PRD.pdf). This README documents the current implementation, including changes made after the original project snapshot.

## What works today

- Passwordless email sign-in with one-time six-digit codes.
  - `log` mode prints a code to the backend terminal for local development.
  - `smtp` mode delivers a real email through a configured SMTP provider.
- HttpOnly cookie sessions backed by a signed JWT with a 72-hour lifetime.
- Foreground-only location sharing. The client pauses sharing when the tab is hidden or the user pauses it manually.
- Live hotspots: at least three active users within 50 metres produce an aggregate hotspot, never individual user locations.
- Public upcoming meetups: create, persist, retrieve nearby and show on the map.
- WebSocket updates for hotspot formation, hotspot dissolution and new meetups.
- Automatic reconnect, live-state snapshots and a memory-only activity feed.
- A responsive React map using Mapbox when a token is configured, or OpenStreetMap tiles locally when it is not.

The verified development arrangement is a native Go API and Vite frontend with MongoDB and Redis in Docker. The project also contains Dockerfiles and a full Compose stack; final end-to-end validation of that full application stack remains outstanding.

## Architecture

```text
Browser
  React + TypeScript + Zustand + Mapbox GL
  ├── REST requests through Axios
  ├── WebSocket connection for live events
  └── Browser geolocation while the tab is visible
                 │
                 ▼
Go API (Fiber)
  ├── OTP and JWT/cookie authentication
  ├── Location and hotspot service
  ├── Meetup service
  └── WebSocket hub
       ├── Redis: expiring OTPs, rate limits and live locations
       └── MongoDB: users and public meetups
```

During Vite development, the browser calls `/api/...` on port `5173`. Vite removes the `/api` prefix and proxies the request or WebSocket upgrade to the Go API on port `8080`.

```text
Browser:  /api/auth/me  →  Vite: /auth/me  →  Go API
Browser:  /api/ws       →  Vite: /ws       →  Go WebSocket endpoint
```

## Technology choices

| Technology | Role | Why it fits this MVP |
| --- | --- | --- |
| Go 1.21 | API server | Handles concurrent HTTP requests and long-lived WebSocket connections with lightweight goroutines. |
| Fiber | HTTP framework | Provides routing, JSON parsing, cookies, CORS and recovery with a compact API. |
| Redis 7 | Live state | Fast in-memory GEO lookups, atomic rate limits and automatic expiry for data that must disappear. |
| MongoDB 6 | Persistent data | Stores the simple document-shaped user and meetup records. The code uses DocumentDB-safe basic operations. |
| React 18 + TypeScript | User interface | Components and static types for the map, forms and real-time client state. |
| Vite | Frontend tooling | Fast development server, builds and API/WebSocket proxying. |
| Zustand | Client state | Holds the signed-in user, location, hotspots, meetups, connection status and temporary activity feed. |
| Mapbox GL JS | Map rendering | Interactive map, user marker, hotspot markers and meetup pins. Falls back to OSM tiles when no token is set. |
| WebSockets | Live delivery | Pushes state changes instead of making every browser poll for them. |
| Docker Compose | Local dependencies | Runs isolated MongoDB and Redis with reproducible ports and health checks. |

## Data and privacy model

The project deliberately separates data by its required lifetime.

| Data | Stored in | Lifetime | Notes |
| --- | --- | ---: | --- |
| OTP hash | Redis | 5 minutes | The raw code is never stored. Successful verification deletes it. |
| OTP send/verify counters | Redis | 1 hour | Limits sends and valid-format verification attempts to five per email per hour. |
| Current presence | Redis | 30 seconds | Contains the latest coordinate for one user. |
| GEO coordinate bucket | Redis | At most 30 seconds | A bucket is fixed to its timestamp plus 30 seconds; later writes cannot extend it. |
| Active hotspot | Go process memory | While valid | Only aggregate centre, ID and count are stored. |
| User | MongoDB | Persistent | Opaque user ID, normalized email and creation time. |
| Meetup | MongoDB | Persistent | Public title, description, coordinates, time and creator ID. |
| Activity feed | Browser memory | Current connection | Clears on refresh or WebSocket reconnect. |

Redis persistence is intentionally disabled for this live-state database. Startup fails if Redis snapshots or append-only persistence are enabled.

### Live-location expiry

Redis GEO members cannot expire individually. UnAlone therefore uses both a per-user presence key and short-lived per-second GEO buckets:

```text
user:<user-id>:presence     → latest position, TTL 30 seconds
geo:local:<unix-second>     → GEO bucket, expiry at bucket second + 30 seconds
```

For a nearby lookup, the backend searches the current bucket plus the preceding 30 buckets, deduplicates user IDs, checks each current presence record and calculates the final distance again in Go. Old bucket membership therefore cannot revive a moved or expired user.

The client sends a location approximately every seven seconds; the backend also enforces a five-second minimum interval. If the user pauses sharing, logs out, hides the tab or stops updating, their presence disappears within the expiry window.

This protects against persistent location history. It does not make activity anonymous: an observer may still infer that someone entered or left an isolated hotspot from the aggregate count and timing.

## Hotspot algorithm

A hotspot is created when the current implementation finds at least three active users within a 50-metre centroid-coverage area.

On each accepted location update, the backend:

1. Writes the position atomically to Redis and refreshes current presence.
2. Reconciles existing hotspots, dissolving any with fewer than three active users nearby.
3. Searches for active users within 50 metres of the updating user.
4. Stops if fewer than three are found.
5. Averages the candidates' coordinates to calculate a centre.
6. Searches again within 50 metres of that centre.
7. Stops if fewer than three are found around the centre.
8. Suppresses a new hotspot if an active hotspot already exists within 50 metres.
9. Creates a seven-character geohash ID, stores only the aggregate result and broadcasts `HOTSPOT_FORMED`.

The implementation does **not** require every pair of users to be within 50 metres. For example, A–B and B–C can each be 49 metres apart while A–C is 98 metres apart; the group can still form a hotspot if all three are within 50 metres of the calculated centre.

The active centre is set when the hotspot forms. Reconciliation updates its count but does not continuously move its centre or rotate its geohash ID. The implementation also does not assign exclusive members to hotspots; an active user can contribute to overlapping searches.

Presence expiry events from Redis trigger hotspot reconciliation. This avoids a scheduled cleanup job. Hotspot state is currently process-local, so production multi-instance deployment needs shared hotspot coordination and event fanout.

## Authentication and email delivery

### OTP flow

1. The browser posts an email to `POST /auth/send-otp`.
2. The backend normalizes and validates the email, generates a cryptographically random six-digit code, stores only its hash in Redis and rate-limits the request.
3. The configured delivery service either logs the code locally or sends it through SMTP.
4. The browser posts the code to `POST /auth/verify-otp`.
5. The backend atomically verifies and consumes the code, finds or creates the user, generates a JWT and sets an HttpOnly cookie.

If SMTP delivery fails, the server removes that specific OTP so a code from a failed delivery attempt cannot be used. The rate-limit attempt remains counted to protect the delivery provider.

### Session model

The JWT includes an opaque user UUID as its subject, the normalized email, issuer `unalone`, issue time and a 72-hour expiry. The browser uses the `unalone_session` HttpOnly, `SameSite=Lax` cookie; it does not store the JWT in localStorage.

The server accepts cookie sessions for the browser and Bearer tokens for integration tests or non-browser clients. Logout clears the cookie and deletes current location presence. There is no server-side JWT revocation list, so a stolen valid JWT remains usable until expiry; this is a known pre-production limitation.

### SMTP configuration

The default is local development logging:

```text
OTP_MODE=log
```

To send real email, configure SMTP before starting the backend:

```text
OTP_MODE=smtp
SMTP_HOST=<provider SMTP host>
SMTP_PORT=<provider SMTP port>
SMTP_USERNAME=<provider username>
SMTP_PASSWORD=<provider password or app password>
SMTP_FROM=UnAlone <verified-sender@example.com>
SMTP_TLS_MODE=starttls
SMTP_TIMEOUT_SECONDS=10
```

`SMTP_TLS_MODE` must be either `starttls` or `implicit`. SMTP mode refuses to start without all required values, a valid sender address and TLS.

For Docker Compose, copy [`.env.example`](./.env.example) to `.env`, fill the values and keep `.env` out of version control. For native `go run .`, export the environment variables in the terminal that starts the backend. [backend/.env.example](./backend/.env.example) contains the same reference values, but native Go does not automatically load an `.env` file.

Never commit SMTP passwords, API keys or application passwords. Use a verified sender identity required by the chosen provider.

## API reference

All endpoints below are backend paths. In the Vite frontend and nginx container, prefix them with `/api`.

| Method | Path | Authentication | Purpose |
| --- | --- | --- | --- |
| `GET` | `/health` | No | Returns `{ "status": "ok" }`. |
| `POST` | `/auth/send-otp` | No | Requests an OTP for `{ "email": "..." }`. |
| `POST` | `/auth/verify-otp` | No | Verifies `{ "email": "...", "otp": "123456" }`, sets a session cookie and returns the user. |
| `GET` | `/auth/me` | Cookie or Bearer token | Returns the signed-in user. |
| `POST` | `/auth/logout` | Cookie or Bearer token | Clears session cookie and live location presence. |
| `GET` | `/ws` | Cookie or Bearer token | Upgrades to the authenticated WebSocket. |
| `POST` | `/location/update` | Cookie or Bearer token | Updates `{ "lat": number, "lon": number }`. |
| `DELETE` | `/location` | Cookie or Bearer token | Immediately stops current server-side presence. |
| `GET` | `/hotspots` | Cookie or Bearer token | Returns current hotspot snapshot. |
| `POST` | `/meetups` | Cookie or Bearer token | Creates a public upcoming meetup. |
| `GET` | `/meetups/near` | Cookie or Bearer token | Reads nearby upcoming meetups. |

`POST /meetups` accepts:

```json
{
  "title": "Coffee and a book",
  "description": "Bring something to read.",
  "lat": 12.9716,
  "lon": 77.5946,
  "time": 1790000000
}
```

Title length must be 3–100 characters, description is limited to 1,000 characters, coordinates must be valid and time must be in the future. `GET /meetups/near` requires `lat`, `lon` and an optional `radius` from 1 to 50,000 metres; the default is 1,000 metres.

## WebSocket events

The backend broadcasts public, aggregate events only:

```json
{
  "type": "HOTSPOT_FORMED",
  "hotspotId": "tdr...",
  "lat": 12.9717,
  "lon": 77.5947,
  "activeUsers": 3
}
```

```json
{
  "type": "HOTSPOT_DISSOLVED",
  "hotspotId": "tdr...",
  "lat": 12.9717,
  "lon": 77.5947,
  "activeUsers": 3
}
```

```json
{
  "type": "MEETUP_CREATED",
  "data": { "meetupId": "...", "title": "...", "lat": 12.9716, "lon": 77.5946 }
}
```

The WebSocket hub has a 64-message queue for each client. A slow client is disconnected rather than being allowed to consume unlimited memory. On reconnect, the frontend clears stale live state, loads a hotspot REST snapshot and queues events arriving during that snapshot. Persisted meetups can be reloaded through REST.

## Local development

### Prerequisites

- Docker Desktop with Docker Compose.
- Go 1.21 or newer.
- Node.js 18 or newer with npm.

### 1. Start MongoDB and Redis

From the project root:

```powershell
docker compose up -d mongo redis
```

This starts:

| Service | Host address | Notes |
| --- | --- | --- |
| MongoDB | `127.0.0.1:27018` | Uses a persistent Docker volume. |
| Redis | `127.0.0.1:6380` | Persistence disabled; expiry notifications enabled. |

### 2. Start the API

Open a terminal and run:

```powershell
cd C:\Anika\Projects\UnAlone-Live-Social-Discovery-App-main\backend
go run .
```

Default local settings are:

```text
PORT=8080
MONGO_URI=mongodb://127.0.0.1:27018
MONGO_DB=unalone
REDIS_URL=127.0.0.1:6380
OTP_MODE=log
```

With `OTP_MODE=log`, the API terminal prints codes in this format:

```text
Development OTP for person@example.com: 123456 (expires in 5m0s)
```

For SMTP mode, set the variables described in [SMTP configuration](#smtp-configuration) in this same terminal before running `go run .`.

### 3. Start the frontend

In a separate terminal:

```powershell
cd C:\Anika\Projects\UnAlone-Live-Social-Discovery-App-main\frontend
npm ci
npm run dev -- --host 127.0.0.1
```

Open [http://127.0.0.1:5173](http://127.0.0.1:5173).

Allow browser location permission after signing in. The map can use OpenStreetMap tiles without further configuration. To use Mapbox’s hosted style, create `frontend/.env` from [frontend/.env.example](./frontend/.env.example) and set `VITE_MAPBOX_TOKEN` before starting Vite.

### Full Docker Compose stack

The project contains backend and frontend container definitions as well:

```powershell
docker compose up --build
```

This exposes the frontend on `http://127.0.0.1:5173` and API on `http://127.0.0.1:8080`. Do not run this alongside the native API/Vite setup because both use the same host ports. The database-only Compose flow above is the verified development path; complete app-container startup still needs final acceptance validation.

## Testing

Run backend unit and package tests:

```powershell
cd backend
go test ./...
```

Run the integration test against local Docker MongoDB and Redis:

```powershell
cd backend
go test -tags integration ./api -run TestLocalMVP -count=1 -v -timeout 150s
```

The integration test creates a randomly named MongoDB database and uses Redis database 15. It requires Redis DB 15 to be empty, then cleans up its own test database and Redis keys.

Run frontend checks:

```powershell
cd frontend
npm run typecheck
npm run build
```

Run the browser smoke test while MongoDB, Redis, API and frontend are running:

```powershell
cd frontend
npm run test:browser
```

The browser test uses mocked geolocation and map tiles. It verifies OTP login, foreground/hidden-tab location behaviour, meetup creation and map cards, refresh persistence, WebSocket events, reconnect and logout. Screenshots and the JSON report are written to `output/browser-check/`.

## Current verification

- Backend package tests pass, including JWT validation, coordinate/email validation, SMTP configuration and email message rendering.
- The local MongoDB/Redis integration suite has passed authentication, OTP replay/rate limits, meetup validation and radius filtering, hotspot formation/dissolution/expiry, WebSocket keepalive, cookie session handling and origin protection.
- Frontend type checking passes.
- The latest browser smoke report records eight successful checks, including reconnect and logout.

Manual validation still needed before the MVP is closed:

- Real browser GPS permission and movement behaviour.
- Live external map-tile behaviour.
- Final frontend production build after any further changes.
- Full Docker Compose application startup.
- Cleanup of browser-test users and meetups from the application database.

## Repository guide

```text
backend/
  api/          HTTP handlers and routes
  auth/         OTP, JWT and auth middleware
  config/       Environment configuration and validation
  db/           MongoDB setup and indexes
  mailer/       Development log and TLS SMTP OTP delivery
  models/       User and meetup documents
  redis/        Redis client, TTL and GEO operations
  services/     User, meetup and hotspot business logic
  ws/           WebSocket hub, events and connection lifecycle

frontend/
  src/api/      Axios client and session-expiry handling
  src/components/ Map and meetup UI components
  src/hooks/    Location and WebSocket lifecycle hooks
  src/pages/    Login, map, activity, profile and meetup creation pages
  src/store/    Zustand application state
  tests/        Browser smoke test

docker-compose.yml  Local MongoDB, Redis, API and frontend stack
```

## Known limitations and next work

This is a local MVP, not a production deployment. The main follow-up work is:

- AWS deployment, HTTPS, secrets management and production monitoring.
- Production-grade sender domain setup and durable SMTP/provider configuration.
- Session revocation or shorter-lived access-token design.
- Multi-instance hotspot coordination and WebSocket fanout through shared Pub/Sub or an event bus.
- Load testing and performance work. The current 31 GEO-bucket searches per nearby lookup favour strict expiry over high-scale efficiency.
- A scalable meetup geo query. The current code loads future meetups then applies exact radius filtering in Go, which is correct for the MVP but unsuitable for a very large dataset.
- A more formal hotspot clustering model if the product needs exclusive membership, continuously moving centres or stricter pairwise-distance semantics.
- CI, deployment automation and complete container-stack acceptance testing.

The current implementation is intentionally focused: no social graph, chat, RSVP flow, private messaging or persistent live-location history.
