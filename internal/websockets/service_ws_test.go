package websockets

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

/*
newTestHub creates and starts a Hub goroutine for test use.
Each test gets its own hub so sessions don't leak between cases.
*/
func newTestHub(t *testing.T) *Hub {
	t.Helper()
	hub := NewHub()
	go hub.Run()
	return hub
}

/*
dialWS converts an httptest server URL (http://...) to WS (ws://...)
and dials it, failing the test immediately on error.
*/
func dialWS(t *testing.T, serverURL string) *websocket.Conn {
	t.Helper()
	wsURL := "ws" + strings.TrimPrefix(serverURL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to dial WS: %v", err)
	}
	return conn
}

/*
newRobotServer creates a test HTTP server that upgrades to WS, registers
a RobotClient with the hub, and runs its read/write pumps.
Simulates the robot side of the connection without going through the full handler.
*/
func newRobotServer(t *testing.T, hub *Hub, reservationID uuid.UUID) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		client := &RobotClient{
			hub:           hub,
			conn:          conn,
			sendChan:      make(chan []byte, 256),
			robotID:       uuid.New(),
			reservationID: reservationID,
		}
		hub.register <- client
		go client.writePump()
		client.readPump()
	}))
}

/*
newUserServer creates a test HTTP server that upgrades to WS, subscribes
a UserClient to the hub, and runs its read/write pumps.
Simulates the user side of the connection without going through the full handler.
*/
func newUserServer(t *testing.T, hub *Hub, reservationID uuid.UUID) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		client := &UserClient{
			hub:           hub,
			conn:          conn,
			sendChan:      make(chan []byte, 256),
			userID:        uuid.New(),
			reservationID: reservationID,
		}
		hub.subscribe <- client
		go client.writePump()
		client.readPump()
	}))
}

/*
TestRobotDisconnectedDuringControl verifies that when the robot WS connection
drops mid-session, the hub unregisters the session and propagates the close
to the user connection.
*/
func TestRobotDisconnectedDuringControl(t *testing.T) {
	hub := newTestHub(t)
	reservationID := uuid.New()

	robotServer := newRobotServer(t, hub, reservationID)
	defer robotServer.Close()
	userServer := newUserServer(t, hub, reservationID)
	defer userServer.Close()

	robotConn := dialWS(t, robotServer.URL)
	userConn := dialWS(t, userServer.URL)
	defer userConn.Close()

	time.Sleep(50 * time.Millisecond)

	// Confirm session is active: robot sends, user receives
	robotConn.WriteMessage(websocket.TextMessage, []byte(`{"speed":1.5}`))

	userConn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	_, got, err := userConn.ReadMessage()
	if err != nil {
		t.Fatalf("expected telemetry before disconnect: %v", err)
	}
	if string(got) != `{"speed":1.5}` {
		t.Errorf("unexpected message: %s", got)
	}

	// Robot drops connection
	robotConn.Close()

	// Hub should close user connection after unregistering session
	userConn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	_, _, err = userConn.ReadMessage()
	if err == nil {
		t.Error("expected user connection to close after robot disconnect")
	}
}

/*
TestUserDisconnected verifies that when the user WS connection drops,
the hub unregisters the session and closes the robot connection via its sendChan.
*/
func TestUserDisconnected(t *testing.T) {
	hub := newTestHub(t)
	reservationID := uuid.New()

	robotServer := newRobotServer(t, hub, reservationID)
	defer robotServer.Close()
	userServer := newUserServer(t, hub, reservationID)
	defer userServer.Close()

	robotConn := dialWS(t, robotServer.URL)
	defer robotConn.Close()
	userConn := dialWS(t, userServer.URL)

	time.Sleep(50 * time.Millisecond)

	// User drops connection
	userConn.Close()

	// Hub closes robot sendChan → writePump sends CloseMessage → robot conn closes
	robotConn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	_, _, err := robotConn.ReadMessage()
	if err == nil {
		t.Error("expected robot connection to close after user disconnect")
	}
}

/*
TestReconnectRobotSameID verifies that a robot can reconnect with the same
reservationID after a disconnect. The hub must create a fresh session so a
new user can subscribe and receive telemetry again.
*/
func TestReconnectRobotSameID(t *testing.T) {
	hub := newTestHub(t)
	reservationID := uuid.New()

	robotServer := newRobotServer(t, hub, reservationID)
	defer robotServer.Close()

	// First connection: connect then disconnect
	robotConn1 := dialWS(t, robotServer.URL)
	time.Sleep(50 * time.Millisecond)
	robotConn1.Close()
	time.Sleep(100 * time.Millisecond) // let hub process the unregister

	// Robot reconnects with the same reservationID
	robotConn2 := dialWS(t, robotServer.URL)
	defer robotConn2.Close()
	time.Sleep(50 * time.Millisecond)

	// User subscribes to the fresh session
	userServer := newUserServer(t, hub, reservationID)
	defer userServer.Close()
	userConn := dialWS(t, userServer.URL)
	defer userConn.Close()
	time.Sleep(50 * time.Millisecond)

	// Telemetry should flow normally after reconnect
	expected := `{"reconnected":true}`
	robotConn2.WriteMessage(websocket.TextMessage, []byte(expected))

	userConn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	_, received, err := userConn.ReadMessage()
	if err != nil {
		t.Fatalf("expected telemetry after reconnect: %v", err)
	}
	if string(received) != expected {
		t.Errorf("expected %q, got %q", expected, received)
	}
}

/*
TestMalformedTelemetry covers two failure modes:
  - Non-JSON bytes are forwarded as-is to the user without the hub crashing.
  - A message exceeding maxMessageSize causes the robot connection to close.
*/
func TestMalformedTelemetry(t *testing.T) {
	t.Run("non-json forwarded as-is", func(t *testing.T) {
		hub := newTestHub(t)
		reservationID := uuid.New()

		robotServer := newRobotServer(t, hub, reservationID)
		defer robotServer.Close()
		userServer := newUserServer(t, hub, reservationID)
		defer userServer.Close()

		robotConn := dialWS(t, robotServer.URL)
		defer robotConn.Close()
		userConn := dialWS(t, userServer.URL)
		defer userConn.Close()
		time.Sleep(50 * time.Millisecond)

		malformed := []byte("not-json-{{{")
		robotConn.WriteMessage(websocket.TextMessage, malformed)

		userConn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
		_, received, err := userConn.ReadMessage()
		if err != nil {
			t.Fatalf("expected malformed bytes to be forwarded: %v", err)
		}
		if string(received) != string(malformed) {
			t.Errorf("expected raw bytes forwarded as-is, got %q", received)
		}
	})

	t.Run("oversized message closes connection", func(t *testing.T) {
		hub := newTestHub(t)
		reservationID := uuid.New()

		robotServer := newRobotServer(t, hub, reservationID)
		defer robotServer.Close()

		robotConn := dialWS(t, robotServer.URL)
		time.Sleep(50 * time.Millisecond)

		// Send a message that exceeds the maxMessageSize limit (512 bytes)
		robotConn.WriteMessage(websocket.TextMessage, make([]byte, maxMessageSize+1))

		robotConn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
		_, _, err := robotConn.ReadMessage()
		if err == nil {
			t.Error("expected connection to close after oversized message")
		}
	})
}

/*
TestDelayedPackets verifies that telemetry frames sent with deliberate delays
between them are all received by the user in order without loss.
*/
func TestDelayedPackets(t *testing.T) {
	hub := newTestHub(t)
	reservationID := uuid.New()

	robotServer := newRobotServer(t, hub, reservationID)
	defer robotServer.Close()
	userServer := newUserServer(t, hub, reservationID)
	defer userServer.Close()

	robotConn := dialWS(t, robotServer.URL)
	defer robotConn.Close()
	userConn := dialWS(t, userServer.URL)
	defer userConn.Close()
	time.Sleep(50 * time.Millisecond)

	frames := []string{
		`{"seq":1,"speed":0.5}`,
		`{"seq":2,"speed":1.0}`,
		`{"seq":3,"speed":1.5}`,
	}

	for _, frame := range frames {
		robotConn.WriteMessage(websocket.TextMessage, []byte(frame))
		time.Sleep(100 * time.Millisecond)
	}

	for _, expected := range frames {
		userConn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
		_, received, err := userConn.ReadMessage()
		if err != nil {
			t.Fatalf("failed to receive delayed packet: %v", err)
		}
		if string(received) != expected {
			t.Errorf("expected %q, got %q", expected, received)
		}
	}
}
