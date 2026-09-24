//go:build integration

package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	websocket "github.com/fasthttp/websocket"
	"github.com/google/uuid"
	"unalone/backend/auth"
	"unalone/backend/config"
	"unalone/backend/db"
	"unalone/backend/mailer"
	"unalone/backend/redis"
	"unalone/backend/services"
	"unalone/backend/ws"
)

// Integration tests use an empty Redis DB 15 and a new, uniquely named MongoDB
// database. They never flush the application's default Redis database or MongoDB.
func TestLocalMVP(t *testing.T) {
	cfg := config.Load()
	cfg.MongoDB = "unalone_test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	cfg.RedisURL = "redis://127.0.0.1:6380/15"
	cfg.JWTSecret = "integration-only-secret"
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := db.Init(ctx, cfg); err != nil {
		t.Fatal(err)
	}
	defer db.Disconnect(context.Background())
	defer db.DB.Drop(context.Background())
	if err := redis.Init(cfg); err != nil {
		t.Fatal(err)
	}
	defer redis.Client.Close()
	if size, err := redis.Client.DBSize(ctx).Result(); err != nil || size != 0 {
		t.Fatalf("integration tests require an empty Redis DB 15; found %d keys: %v", size, err)
	}
	defer redis.Client.FlushDB(context.Background())
	hub := ws.NewHub()
	defer hub.Close()
	hotspots := services.NewHotspotService(hub)
	if err := hotspots.StartExpiryListener(ctx); err != nil {
		t.Fatal(err)
	}
	app := NewApp(cfg, hub, hotspots)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	base := "http://" + listener.Addr().String()
	go app.Listener(listener)
	defer app.Shutdown()
	client := &http.Client{Timeout: 10 * time.Second}
	request := func(method, path, token string, body interface{}) (int, map[string]interface{}) {
		t.Helper()
		data, _ := json.Marshal(body)
		req, _ := http.NewRequest(method, base+path, bytes.NewReader(data))
		req.Header.Set("Content-Type", "application/json")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		response, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer response.Body.Close()
		var result map[string]interface{}
		_ = json.NewDecoder(response.Body).Decode(&result)
		return response.StatusCode, result
	}
	account := func() (string, string, string) {
		t.Helper()
		email := uuid.NewString() + "@example.com"
		code, err := auth.SendOTP(ctx, email, mailer.LogSender{})
		if err != nil {
			t.Fatal(err)
		}
		status, result := request("POST", "/auth/verify-otp", "", map[string]interface{}{"email": email, "otp": code})
		if status != 200 {
			t.Fatalf("verify returned %d: %v", status, result)
		}
		user := result["user"].(map[string]interface{})
		return result["token"].(string), user["userId"].(string), email
	}
	connect := func(token string) (*websocket.Conn, chan map[string]interface{}, chan error) {
		t.Helper()
		headers := http.Header{"Authorization": []string{"Bearer " + token}, "Origin": []string{"http://localhost:5173"}}
		conn, _, err := websocket.DefaultDialer.Dial(strings.Replace(base, "http", "ws", 1)+"/ws", headers)
		if err != nil {
			t.Fatal(err)
		}
		messages := make(chan map[string]interface{}, 100)
		failures := make(chan error, 1)
		go func() {
			for {
				var event map[string]interface{}
				if err := conn.ReadJSON(&event); err != nil {
					failures <- err
					return
				}
				messages <- event
			}
		}()
		return conn, messages, failures
	}
	token, _, _ := account()
	heartbeat, heartbeatEvents, heartbeatErrors := connect(token)
	defer heartbeat.Close()
	connectedAt := time.Now()

	t.Run("authentication boundaries and persisted identity", func(t *testing.T) {
		if status, _ := request("GET", "/hotspots", "", nil); status != 401 {
			t.Fatalf("public hotspot endpoint: %d", status)
		}
		if status, _ := request("POST", "/location/update", token, map[string]interface{}{"lat": 999, "lon": 0}); status != 400 {
			t.Fatalf("invalid coordinates accepted: %d", status)
		}
		if status, _ := request("POST", "/location/update", token, map[string]interface{}{}); status != 400 {
			t.Fatal("missing coordinates accepted")
		}
		email := uuid.NewString() + "@example.com"
		code, err := auth.SendOTP(ctx, email, mailer.LogSender{})
		if err != nil {
			t.Fatal(err)
		}
		status, first := request("POST", "/auth/verify-otp", "", map[string]interface{}{"email": email, "otp": code})
		if status != 200 {
			t.Fatal(first)
		}
		if status, _ := request("POST", "/auth/verify-otp", "", map[string]interface{}{"email": email, "otp": code}); status != 401 {
			t.Fatal("reused OTP accepted")
		}
		code, _ = auth.SendOTP(ctx, email, mailer.LogSender{})
		_, second := request("POST", "/auth/verify-otp", "", map[string]interface{}{"email": strings.ToUpper(email), "otp": code})
		if first["user"].(map[string]interface{})["userId"] != second["user"].(map[string]interface{})["userId"] {
			t.Fatal("repeat sign-in created a second user")
		}
		limitEmail := uuid.NewString() + "@example.com"
		for i := 0; i < 6; i++ {
			status, _ := request("POST", "/auth/send-otp", "", map[string]interface{}{"email": limitEmail})
			want := 200
			if i == 5 {
				want = 429
			}
			if status != want {
				t.Fatalf("OTP send attempt %d: %d", i, status)
			}
		}
		for i := 0; i < 6; i++ {
			status, _ := request("POST", "/auth/verify-otp", "", map[string]interface{}{"email": limitEmail, "otp": "xxxxxx"})
			if status != 400 {
				t.Fatal("malformed OTP accepted")
			}
		}
		for i := 0; i < 6; i++ {
			status, _ := request("POST", "/auth/verify-otp", "", map[string]interface{}{"email": uuid.Nil.String() + "@example.com", "otp": "000000"})
			want := 401
			if i == 5 {
				want = 429
			}
			if status != want {
				t.Fatalf("OTP verify attempt %d: %d", i, status)
			}
		}
	})

	t.Run("meetup creation retrieval radius and privacy", func(t *testing.T) {
		future := time.Now().Add(time.Hour).Unix()
		body := map[string]interface{}{"title": "Integration meetup", "description": "Test only", "lat": 12.97, "lon": 77.59, "time": future}
		status, result := request("POST", "/meetups", token, body)
		if status != 201 {
			t.Fatalf("create %d: %v", status, result)
		}
		meetup := result["meetup"].(map[string]interface{})
		if strings.Contains(meetup["createdBy"].(string), "@") || meetup["time"].(float64) != float64(future) {
			t.Fatal("meetup contract/privacy mismatch")
		}
		status, near := request("GET", "/meetups/near?lat=12.97&lon=77.59&radius=1000", token, nil)
		if status != 200 || len(near["meetups"].([]interface{})) != 1 {
			t.Fatal("saved meetup not retrieved")
		}
		_, far := request("GET", "/meetups/near?lat=12.97&lon=77.61&radius=100", token, nil)
		if len(far["meetups"].([]interface{})) != 0 {
			t.Fatal("radius ignored")
		}
		_, wide := request("GET", "/meetups/near?lat=12.97&lon=77.61&radius=5000", token, nil)
		if len(wide["meetups"].([]interface{})) != 1 {
			t.Fatal("larger radius ignored")
		}
		body["time"] = time.Now().Add(-time.Hour).Unix()
		if status, _ := request("POST", "/meetups", token, body); status != 400 {
			t.Fatal("past meetup accepted")
		}
	})

	t.Run("hotspot forms once dissolves and expires silently", func(t *testing.T) {
		connection, events, failures := connect(token)
		defer connection.Close()
		next := func(kind string, wait time.Duration) {
			t.Helper()
			select {
			case e := <-events:
				if e["type"] != kind {
					t.Fatalf("wanted %s, got %v", kind, e)
				}
			case err := <-failures:
				t.Fatal(err)
			case <-time.After(wait):
				t.Fatalf("no %s event", kind)
			}
		}
		tokens := make([]string, 3)
		ids := make([]string, 3)
		for i := range tokens {
			tokens[i], ids[i], _ = account()
			status, result := request("POST", "/location/update", tokens[i], map[string]interface{}{"lat": 13.1, "lon": 77.6})
			if status != 200 {
				t.Fatalf("location %d: %v", status, result)
			}
		}
		next("HOTSPOT_FORMED", 3*time.Second)
		if status, _ := request("POST", "/location/update", tokens[0], map[string]interface{}{"lat": 13.1, "lon": 77.6}); status != 429 {
			t.Fatal("location limit not enforced")
		}
		if err := hotspots.Reconcile(ctx); err != nil {
			t.Fatal(err)
		}
		select {
		case e := <-events:
			t.Fatalf("duplicate lifecycle event: %v", e)
		case <-time.After(100 * time.Millisecond):
		}
		if status, _ := request("DELETE", "/location", tokens[2], nil); status != 200 {
			t.Fatal("could not stop sharing")
		}
		next("HOTSPOT_DISSOLVED", 3*time.Second)
		for i := range tokens {
			tokens[i], ids[i], _ = account()
			if status, result := request("POST", "/location/update", tokens[i], map[string]interface{}{"lat": 13.3, "lon": 77.8}); status != 200 {
				t.Fatal(result)
			}
		}
		next("HOTSPOT_FORMED", 3*time.Second)
		keys, err := redis.Client.Keys(ctx, "geo:local:*").Result()
		if err != nil {
			t.Fatal(err)
		}
		for _, key := range keys {
			ttl, err := redis.Client.PTTL(ctx, key).Result()
			if err != nil || ttl <= 0 || ttl > 30*time.Second {
				t.Fatalf("unsafe coordinate TTL %s: %v %v", key, ttl, err)
			}
		}
		for _, id := range ids {
			ttl := redis.Client.PTTL(ctx, redis.PresenceKey(id)).Val()
			if ttl <= 0 || ttl > 30*time.Second {
				t.Fatal("unsafe presence TTL")
			}
		}
		next("HOTSPOT_DISSOLVED", 35*time.Second)
		time.Sleep(time.Second)
		for _, key := range keys {
			if redis.Client.Exists(ctx, key).Val() != 0 {
				t.Fatalf("coordinate bucket survived expiry: %s", key)
			}
		}
		_, snapshot := request("GET", "/hotspots", token, nil)
		if len(snapshot["hotspots"].([]interface{})) != 0 {
			t.Fatal("expired hotspot survives snapshot")
		}
	})

	t.Run("websocket remains usable past original sixty second deadline", func(t *testing.T) {
		remaining := 65*time.Second - time.Since(connectedAt)
		if remaining > 0 {
			select {
			case err := <-heartbeatErrors:
				t.Fatal(err)
			case <-time.After(remaining):
			}
		}
		for len(heartbeatEvents) > 0 {
			<-heartbeatEvents
		}
		hub.Broadcast([]byte("{\"type\":\"MEETUP_CREATED\",\"data\":{\"meetupId\":\"heartbeat-check\"}}"))
		select {
		case e := <-heartbeatEvents:
			if e["type"] != "MEETUP_CREATED" {
				t.Fatal(e)
			}
		case err := <-heartbeatErrors:
			t.Fatal(err)
		case <-time.After(3 * time.Second):
			t.Fatal("socket no longer delivers")
		}
	})

	t.Run("cookie session logout and origin protection", func(t *testing.T) {
		req, _ := http.NewRequest("GET", base+"/auth/me", nil)
		req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: token})
		response, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		io.Copy(io.Discard, response.Body)
		response.Body.Close()
		if response.StatusCode != 200 {
			t.Fatal("cookie session failed")
		}
		req, _ = http.NewRequest("POST", base+"/auth/logout", nil)
		req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: token})
		req.Header.Set("Origin", "https://untrusted.example")
		response, err = client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		response.Body.Close()
		if response.StatusCode != 403 {
			t.Fatal("cross-origin session mutation allowed")
		}
		status, _ := request("POST", "/auth/logout", token, nil)
		if status != 200 {
			t.Fatal("logout failed")
		}
	})
	fmt.Fprintln(os.Stdout, "Local MVP integration checks completed against real Redis and MongoDB")
}
