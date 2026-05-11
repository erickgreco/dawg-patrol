package robots

import (
	"context"
	"time"

	"github.com/erickgreco/dawg-patrol/pkg/myerrors"
)

/*
This goroutine executes a first cleanUp as soon as it starts running
and then it keeps periodically cleaning every minute, this way if
server is restarted expired reservations will inmediatelly be cleaned
avoiding waiting periods if server restarts
*/
func ReservationCleanUpWorker(store RobotsRepo) {
	ticker := time.NewTicker(1 * time.Minute)

	go func() {
		defer ticker.Stop()

		runCleanUp := func() {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			if err := store.CleanExpiredReservations(ctx); err != nil {
				myerrors.CleanUpWorkerError(err)
			}
		}
		runCleanUp()

		for range ticker.C {
			runCleanUp()
		}
	}()
}
