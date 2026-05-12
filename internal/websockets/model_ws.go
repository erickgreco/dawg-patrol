package websockets

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

/*
Represents the active ws conn of a robot
Links conn with its active reservation so the
Hub can route telemetry to the right user
*/
type RobotClient struct {
	hub           *Hub
	conn          *websocket.Conn
	sendChan      chan []byte
	robotID       uuid.UUID
	reservationID uuid.UUID
}

/*
Represents the active ws conn of a user
Links conn with its active reservation so
user can receive robots telemetry
*/
type UserClient struct {
	hub           *Hub
	conn          *websocket.Conn
	sendChan      chan []byte
	userID        uuid.UUID
	reservationID uuid.UUID
}

/*
Session groups robot and user conn with the
same reservationID.
The Hub uses this structure to route robots
telemetry to correspondant user
*/
type Session struct {
	robot *RobotClient
	user  *UserClient
}

/*
Represents a message sent by a robot including reservationID
so the Hub can identify to which session it belongs and resends
to right user
*/
type TelemetryMessage struct {
	reservationID uuid.UUID
	data          []byte
}

/*
Hub has the register of all active ws sessions,
runs on its own goroutine and coordinates robots registration
with users subscriptions and telemetry resend
*/
type Hub struct {
	sessions   map[uuid.UUID]*Session
	register   chan *RobotClient
	subscribe  chan *UserClient
	unregister chan uuid.UUID
	broadcast  chan *TelemetryMessage
}
