# UnAlone — MVP

**Problem Statement:**
- People want to discover nearby groups and informal hangouts in real time, without exposing private location history or requiring complex sign-ups. Existing social apps are centered on persistent social graphs, not ephemeral local gatherings.

**Overview:**
- UnAlone is a privacy-minded, real-time social discovery MVP that shows where people are currently hanging out (hotspots) and public meetups happening nearby. It provides a minimal authentication flow (OTP) and lightweight real-time broadcasts for events using WebSockets.
- The system is split into a Go backend (REST + WebSocket) and a React + Vite TypeScript frontend (Mapbox). Data stores are Redis for ephemeral geo positions and MongoDB/DocumentDB for persistent meetups.

**Key Features:**
- OTP-based login (OTP mode = `log` in dev — code printed to backend logs)
- JWT-based authenticated API with 72-hour tokens (subject = email)
- Location updates sent frequently (5–10s) and stored in Redis GEO with 30s TTL
- Hotspot detection (>= 3 active users within 50m) → broadcast `HOTSPOT_UPDATE` via WebSocket
- Create and list public meetups stored in MongoDB/DocumentDB → broadcast `MEETUP_CREATED`
- Frontend: interactive Mapbox map showing live hotspot pulses and meetup pins

**Tech Stack (strict per PRD):**
- Backend
  - Go 1.21
  - Fiber (HTTP REST)
  - Gorilla WebSocket (server-side WebSocket responsibilities fulfilled via Fiber-compatible websocket handler)
  - go-redis v9 (Redis client)
  - MongoDB Go driver (DocumentDB-safe usage)
  - github.com/golang-jwt/jwt/v5 for JWT
- Frontend
  - React + TypeScript
  - Vite
  - Mapbox GL JS
  - Axios
  - Zustand (lightweight state)

**API Endpoints (exact):**
- POST `/auth/send-otp` — body `{ "email": "..." }` — stores OTP in Redis and (dev) logs the code
- POST `/auth/verify-otp` — body `{ "email": "...", "code": "..." }` — verifies OTP and returns `{ "token": "..." }`
- POST `/location/update` — protected (Bearer JWT) — body `{ "lat": <float>, "lon": <float> }` — stores location in Redis GEO, sets TTL (30s), searches neighbors within 50m and broadcasts `HOTSPOT_UPDATE` if >= 3 active users
- POST `/meetups` — protected — body `{ "title":"..","description":"..","lat":..,"lon":.. }` — stores meetup in Mongo (DocumentDB-safe InsertOne) and broadcasts `MEETUP_CREATED`
- GET  `/meetups/near?lat=<>&lon=<>` — protected — returns nearby meetups (server-side simple distance filtering)

**WebSocket Events (broadcasted by backend hub):**
- HOTSPOT_UPDATE: `{ "type": "HOTSPOT_UPDATE", "lat": <float>, "lon": <float>, "activeUsers": <int> }`
- MEETUP_CREATED: `{ "type": "MEETUP_CREATED", "data": { <meetup object> } }`

Note: All clients connect to `/ws` to receive broadcasts.

**Security / Auth Rules:**
- JWT `sub` (subject) is the user's email and tokens expire after 72 hours.
- Middleware extracts the Bearer token and places the `email` in the request context for downstream handlers.

**Data Storage Rules (DocumentDB-safe):**
- Persist meetups with `InsertOne`, `Find`, `FindOne`, `UpdateOne`, `DeleteOne` only. Avoid transactions, sessions, change streams, or advanced aggregations to remain compatible with AWS DocumentDB.
- Use Redis GEO commands (GEOADD / GEOSEARCH) via go-redis v9 for ephemeral location storage; keep per-user TTL keys set to 30 seconds to mark active presence.

**User Flow (text):**
1. User opens the frontend and enters email on Login page.
2. Frontend POSTs `/auth/send-otp`. Backend stores OTP (Redis) and logs the code (dev). Frontend shows OTP input.
3. User enters OTP; frontend POSTs `/auth/verify-otp`. Backend verifies Redis value and returns JWT. Frontend stores JWT in `localStorage` and sets `Authorization` header for Axios.
4. After login, frontend connects to WebSocket `ws://{api}/ws` and starts geolocation tracking (foreground only). The app sends location updates every 5–10 seconds to `POST /location/update`.
5. Backend stores user location in Redis GEO with TTL=30s and checks neighbors via GEOSEARCH (50m radius). If there are >= 3 active users, backend broadcasts a `HOTSPOT_UPDATE` event containing the hotspot center and active user count.
6. Meetups can be created via the Meetup Creator page (POST `/meetups`). Backend stores meetup in Mongo and broadcasts `MEETUP_CREATED` to all connected clients. Clients add meetup pins to the map.

**Text-based Architecture Diagram:**

```
                     +--------------------+
                     |   Browser Client   |
                     |  (React + Mapbox)  |
                     | - Axios (REST)     |
                     | - WS client        |
                     | - Geolocation hook |
                     +---------+----------+
                               |
             REST (auth,meetups,location) | WS (events)
                               |
                     +---------v----------+
                     |     Backend API     |  (Go 1.21, Fiber)
                     | - JWT auth          |
                     | - OTP (Redis)       |
                     | - WebSocket Hub     |
                     +---+-----+-----+-----+
                         |     |     |
                GEO ops  |     |     |  Meetups
                (redis)  |     |     |  (mongo)
            +----v----+   |     |     |  +--v--+
            | Redis   |<--+     |     +-->Mongo|
            | (geo)   |         |         | DB |
            +---------+         |         +-----+
                                |
                         WebSocket Broadcasts
                                |
                     +----------v-----------+
                     | Connected Clients    |
                     | (receive events)     |
                     +----------------------+
```

**Folder Structure (exact):**
```
backend/
  Dockerfile
  go.mod
  main.go
  config/config.go
  api/routes.go
  api/handlers.go
  api/auth/send_otp.go
  api/auth/verify_otp.go
  api/location/update_location.go
  api/meetups/create_meetup.go
  api/meetups/get_meetups_near.go
  models/meetup.go
  models/user.go
  auth/otp.go
  auth/jwt.go
  auth/middleware.go
  db/mongodb.go
  redis/redis_client.go
  redis/geo_ops.go
  ws/hub.go
  ws/ws_handler.go
  ws/events.go
  services/hotspot.go
  services/meetups_service.go
  utils/response.go
  utils/uuid.go

frontend/
  Dockerfile
  index.html
  package.json
  tsconfig.json
  vite.config.ts
  public/_redirects
  src/
    main.tsx
    App.tsx
    api/api.ts
    pages/
      Login.tsx
      MapView.tsx
      MeetupCreator.tsx
      ProfileSettings.tsx
    components/
      MapboxMap.tsx
      HotspotPulse.tsx
      MeetupCard.tsx
    hooks/
      useLocation.ts
      useWebSocket.ts
    store/
      useStore.ts
    styles/
      main.css
```

**How to run (local dev):**
- Native (fast iteration)
  - Start local MongoDB and Redis (or use the provided `docker-compose` below to start them)
  - Backend (PowerShell):
    ```powershell
    cd C:\Anika\Projects\unalone\backend
    $env:MONGO_URI='mongodb://localhost:27017'
    $env:MONGO_DB='unalone'
    $env:REDIS_URL='localhost:6379'
    $env:JWT_SECRET='change_this_secret'
    $env:OTP_MODE='log'
    go run .
    ```
  - Frontend (PowerShell):
    ```powershell
    cd C:\Anika\Projects\unalone\frontend
    npm install
    $env:VITE_API_BASE='http://localhost:8080'
    $env:VITE_MAPBOX_TOKEN='<YOUR_MAPBOX_TOKEN>'
    npm run dev
    ```
- Docker (recommended reproducible stack):
  - Build and run everything (includes Mongo and Redis):
    ```powershell
    cd C:\Anika\Projects\unalone
    docker-compose up --build
    ```
  - Backend logs show OTP codes in dev mode (OTP_MODE=log). Frontend served on port `5173` by default in the compose file (nginx static site).

**Conclusion:**
- This repository implements the UnAlone MVP according to the provided PRD: ephemeral location via Redis GEO, DocumentDB-safe meetup persistence, JWT auth, OTP login in `log` mode for development, and real-time broadcasts via WebSocket for hotspots and meetups. The frontend is a Vite + React app using Mapbox to render live data.

**Planned future enhancements:**
- Production OTP delivery (SMS or email) with rate limiting and abuse protection.
- Persistent user profiles and opt-in location history (encrypted, user-controlled retention).
- Rate-limiting and authentication improvements (refresh tokens, token revocation).
- Hotspot clustering improvements and server-side geo indexing for meetups (when allowed by DB).
- End-to-end tests, CI pipelines, and infrastructure Terraform modules for AWS deployment (S3 + CloudFront for frontend, ECS for backend, ElastiCache Redis, DocumentDB for Mongo API).

If you want, I can now: (a) run static checks and patch any build errors you hit locally; (b) refine the Docker setup for HMR with Vite; or (c) add a minimal README for running tests and troubleshooting logs. Tell me which next step you prefer.
