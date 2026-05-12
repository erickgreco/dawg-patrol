package websockets

import "github.com/google/uuid"

func NewHub() *Hub {
	return &Hub{
		sessions:   make(map[uuid.UUID]*Session),
		register:   make(chan *RobotClient),
		subscribe:  make(chan *UserClient),
		unregister: make(chan uuid.UUID),
		broadcast:  make(chan *TelemetryMessage),
	}
}

/*
This method creates sessions according to an existing reservationID
previously created when a robot is reserved.
Session starts with a robot and allows a user to join session
*/
func (h *Hub) Run() {
	for {
		select {
		// Creates the session when the robot connects based on the reservationID
		case robot := <-h.register:
			h.sessions[robot.reservationID] = &Session{robot: robot}

			// Adds user to an existing connection according to reservationID
		case user := <-h.subscribe:
			if session, ok := h.sessions[user.reservationID]; ok {
				session.user = user
			}

			// Closes both channels and eliminates session
		case reservationID := <-h.unregister:
			if session, ok := h.sessions[reservationID]; ok {
				if session.robot != nil {
					close(session.robot.sendChan)
				}
				if session.user != nil {
					close(session.user.sendChan)
				}
				delete(h.sessions, reservationID)
			}
			// Sends telemetry to user and if users channel is blocked it disconnects
		case msg := <-h.broadcast:
			if session, ok := h.sessions[msg.reservationID]; ok {
				if session.user != nil {
					select {
					case session.user.sendChan <- msg.data:
					default:
						close(session.user.sendChan)
						session.user = nil
					}
				}
			}
		}
	}
}
