package ws

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	core_auth "github.com/PopovMarko/tailverse-petnet/internal/core/auth"
	core_domain "github.com/PopovMarko/tailverse-petnet/internal/core/domain"
	core_logger "github.com/PopovMarko/tailverse-petnet/internal/core/logger"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// The mobile app applies these messages as they are, so their JSON shape is a contract.
func TestEncodeEventContract(t *testing.T) {
	spotId := "11111111-1111-4111-8111-111111111111"
	startsAt := time.Date(2026, 10, 3, 9, 0, 0, 0, time.UTC)
	createdAt := time.Date(2026, 10, 3, 8, 30, 0, 0, time.UTC)

	cases := map[string]struct {
		event any
		want  string
	}{
		"spot_update": {
			event: core_domain.SpotUpdatedEvent{SpotId: spotId, PresentCount: 3},
			want:  `{"type":"spot_update","spot_id":"11111111-1111-4111-8111-111111111111","present_count":3}`,
		},
		"announcement_created at a spot": {
			event: core_domain.AnnouncementCreatedEvent{Announcement: core_domain.Announcement{
				Id: "a1", PetId: "p1", SpotId: &spotId, StartsAt: startsAt, DurationMin: 60, Status: "active", CreatedAt: createdAt,
			}},
			want: `{"type":"announcement_created","announcement":{"id":"a1","pet_id":"p1","spot_id":"11111111-1111-4111-8111-111111111111",` +
				`"custom_point":null,"starts_at":"2026-10-03T09:00:00Z","duration_min":60,"status":"active","created_at":"2026-10-03T08:30:00Z"}}`,
		},
		"announcement_created at a custom point": {
			event: core_domain.AnnouncementCreatedEvent{Announcement: core_domain.Announcement{
				Id: "a2", PetId: "p1", CustomPoint: &core_domain.GeoPoint{Lat: 47.91, Lng: 33.34}, StartsAt: startsAt, DurationMin: 30,
				Status: "active", CreatedAt: createdAt,
			}},
			want: `{"type":"announcement_created","announcement":{"id":"a2","pet_id":"p1","spot_id":null,` +
				`"custom_point":{"lat":47.91,"lng":33.34},"starts_at":"2026-10-03T09:00:00Z","duration_min":30,"status":"active","created_at":"2026-10-03T08:30:00Z"}}`,
		},
	}
	for name, c := range cases {
		got, err := encodeEvent(c.event)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if string(got) != c.want {
			t.Errorf("%s:\n got %s\nwant %s", name, got, c.want)
		}
	}

	if _, err := encodeEvent(struct{}{}); err == nil {
		t.Error("an unknown event must not be encoded")
	}
}

func testLogger() *core_logger.Logger {
	return &core_logger.Logger{Logger: zap.NewNop()}
}

// newTestServer serves Presence the way main.go does, with the auth middleware replaced by a fixed owner.
func newTestServer(t *testing.T, hub *Hub) string {
	t.Helper()
	handler := NewWsHttpHandler(hub)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := core_logger.ToContext(r.Context(), testLogger())
		ctx = core_auth.ToContext(ctx, "owner-1")
		handler.Presence(w, r.WithContext(ctx))
	}))
	t.Cleanup(server.Close)
	return "ws" + strings.TrimPrefix(server.URL, "http")
}

func dial(t *testing.T, url string) *websocket.Conn {
	t.Helper()
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { conn.Close() })
	return conn
}

func waitForClients(t *testing.T, hub *Hub, want int) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for hub.ClientsCount() != want {
		if time.Now().After(deadline) {
			t.Fatalf("clients = %d, want %d", hub.ClientsCount(), want)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func TestHubBroadcastsToEveryClient(t *testing.T) {
	hub := NewHub(testLogger())
	url := newTestServer(t, hub)
	first, second := dial(t, url), dial(t, url)
	waitForClients(t, hub, 2)

	// A client message (location_update) is accepted and does not break the connection.
	if err := first.WriteJSON(map[string]any{"type": "location_update", "lat": 47.9, "lng": 33.3}); err != nil {
		t.Fatal(err)
	}

	hub.Publish(core_domain.SpotUpdatedEvent{SpotId: "s1", PresentCount: 2})

	for name, conn := range map[string]*websocket.Conn{"first": first, "second": second} {
		conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		var message map[string]any
		if err := conn.ReadJSON(&message); err != nil {
			t.Fatalf("%s client: %v", name, err)
		}
		if message["type"] != "spot_update" || message["spot_id"] != "s1" || message["present_count"] != float64(2) {
			t.Errorf("%s client got %v", name, message)
		}
	}
}

func TestHubForgetsClosedClientsAndClosesOnShutdown(t *testing.T) {
	hub := NewHub(testLogger())
	url := newTestServer(t, hub)
	leaving, staying := dial(t, url), dial(t, url)
	waitForClients(t, hub, 2)

	leaving.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "bye"))
	leaving.Close()
	waitForClients(t, hub, 1)

	hub.Close()
	waitForClients(t, hub, 0)
	staying.SetReadDeadline(time.Now().Add(2 * time.Second))
	if _, _, err := staying.ReadMessage(); err == nil {
		t.Error("the connection must be closed on shutdown")
	}

	// After shutdown new connections are refused by the hub.
	late := dial(t, url)
	late.SetReadDeadline(time.Now().Add(2 * time.Second))
	if _, _, err := late.ReadMessage(); err == nil {
		t.Error("a connection after shutdown must be closed")
	}
	if hub.ClientsCount() != 0 {
		t.Errorf("clients = %d after shutdown", hub.ClientsCount())
	}
}

func TestEncodedMessagesAreValidJson(t *testing.T) {
	message, err := encodeEvent(core_domain.SpotUpdatedEvent{SpotId: "s1"})
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(message, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded["present_count"] != float64(0) {
		t.Errorf("a zero count must still be sent: %s", message)
	}
}
