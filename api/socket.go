package api

import (
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

// NewSocketConnection creates a new websocket connection and keeps it open by responding
// with a pong messages until the connection is closed through the returned `done` channel.
func (a *Api) NewSocketConnection(ctx context.Context, endpoint string) (chan struct{}, *websocket.Conn, error) {
	host := a.cfg.GetWorkbenchUrl()[8:]
	url := fmt.Sprintf("wss://%s/%s", host, endpoint)
	header := http.Header{"Authorization": []string{"Bearer " + a.cfg.Root.User.Token}}

	c, _, err := websocket.DefaultDialer.DialContext(ctx, url, header)
	if err != nil {
		log.Fatal("dial:", err)
	}

	done := make(chan struct{})
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	ticker := time.NewTicker(time.Second)

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return

			case <-ticker.C:
				err = c.WriteMessage(websocket.PongMessage, []byte{})
				if err != nil {
					log.Println("websocket write error:", err)
					return
				}

			case <-interrupt:
				fmt.Println(" interrupt received, stopping zepl")
				closeMsg := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")
				if err = c.WriteMessage(websocket.CloseMessage, closeMsg); err != nil {
					return
				}
				select {
				case <-done:
				case <-time.After(time.Second):
				}
				return
			}
		}
	}()

	return done, c, nil
}
