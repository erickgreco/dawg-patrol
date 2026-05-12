package websockets

import (
	"context"
	"net/http"

	"github.com/erickgreco/dawg-patrol/internal/apimiddleware"
	"github.com/erickgreco/dawg-patrol/pkg/myerrors"
)

type WSHandler struct {
	service *WSService
	hub     *Hub
}

func NewWebSocketHandler(service *WSService, hub *Hub) *WSHandler {
	return &WSHandler{
		service: service,
		hub:     hub,
	}
}

func (ws *WSHandler) RobotTelemetryWS(w http.ResponseWriter, r *http.Request) {
	userID, err := apimiddleware.GetUserIDFromClaimsCtx(r)
	if err != nil {
		myerrors.BadRequestResponse(w, r, err)
		return
	}

	robotID, err := apimiddleware.GetRobotIDFromCtx(r)
	if err != nil {
		myerrors.BadRequestResponse(w, r, err)
		return
	}

	reservationID, err := apimiddleware.GetReservationIDFromCtx(r)
	if err != nil {
		myerrors.BadRequestResponse(w, r, err)
		return
	}

	ctx := r.Context()

	if err := ws.service.SessionServiceValidator(ctx, reservationID, userID, robotID); err != nil {
		myerrors.BadRequestResponse(w, r, err)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		myerrors.BadRequestResponse(w, r, err)
		return
	}

	extendCtx, cancelExtend := context.WithCancel(ctx)
	defer cancelExtend()

	go ws.service.KeepReservationAlive(extendCtx, reservationID)

	client := &RobotClient{
		hub:           ws.hub,
		conn:          conn,
		sendChan:      make(chan []byte, 256),
		robotID:       robotID,
		reservationID: reservationID,
	}

	ws.hub.register <- client
	go client.writePump()
	client.readPump()
}
