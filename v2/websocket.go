package binance

import (
	"context"
	"nhooyr.io/websocket"
	"time"
)

// WsHandler handle raw websocket message
type WsHandler func(message []byte)

// ErrHandler handles errors
type ErrHandler func(err error)

// WsConfig webservice configuration
type WsConfig struct {
	Endpoint string
}

func newWsConfig(endpoint string) *WsConfig {
	return &WsConfig{
		Endpoint: endpoint,
	}
}

var wsServe = func(cfg *WsConfig, handler WsHandler, errHandler ErrHandler) (doneC, stopC chan struct{}, err error) {
	ctx, cancel := context.WithCancel(context.Background())
	//defer cancel()
	c, _, err := websocket.Dial(ctx, cfg.Endpoint, nil)
	if err != nil {
		cancel()
		return nil, nil, err
	}
	doneC = make(chan struct{})
	stopC = make(chan struct{})
	go func() {
		defer close(doneC)
		defer cancel()
		if WebsocketKeepalive {
			go keepAlive(ctx, c, WebsocketTimeout)
		}
		silent := false
		go func() {
			select {
			case <-stopC:
				silent = true
			case <-doneC:
			}
			c.Close(websocket.StatusNormalClosure, "normal closure")
		}()
		for {
			_, message, readErr := c.Read(ctx)
			if readErr != nil {
				if !silent {
					errHandler(readErr)
				}
				return
			}
			handler(message)
		}
	}()
	return
}

func keepAlive(ctx context.Context, c *websocket.Conn, d time.Duration) {
	t := time.NewTimer(d)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}

		err := c.Ping(ctx)
		if err != nil {
			return
		}

		t.Reset(time.Minute)
	}
}
