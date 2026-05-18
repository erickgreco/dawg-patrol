package websockets

import (
	"context"
	"net/http"
	"time"

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

	if err := ws.service.MarkWSStarted(ctx, reservationID); err != nil {
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

	defer func() {
		releaseCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		ws.service.ReleaseReservation(releaseCtx, reservationID)
	}()

	go ws.service.KeepReservationAlive(extendCtx, reservationID)
	go ws.service.MockTelemetry(extendCtx, ws.hub, reservationID)

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

func (ws *WSHandler) UserTelemetryWS(w http.ResponseWriter, r *http.Request) {
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

	client := &UserClient{
		hub:           ws.hub,
		conn:          conn,
		sendChan:      make(chan []byte, 256),
		userID:        userID,
		reservationID: reservationID,
	}

	ws.hub.subscribe <- client
	go client.writePump()
	client.readPump()
}
