package pulse

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"

	"pulse2mqtt/include/settings"
)

var ErrWebSocketUnavailable = errors.New("Pulse WebSocket unavailable")

func StreamModernData(ctx context.Context, onFrame func([]byte)) error {
	url := "ws://" + settings.Load.Service.Pulse.IP + "/ws"
	header := http.Header{}
	credentials := settings.Load.Service.Pulse.User + ":" + settings.Load.Service.Pulse.Password
	header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(credentials)))

	conn, response, err := (&websocket.Dialer{}).DialContext(ctx, url, header)
	if err != nil {
		if response != nil && (response.StatusCode == http.StatusNotFound || response.StatusCode == http.StatusUnauthorized || response.StatusCode == http.StatusForbidden) {
			return fmt.Errorf("%w: HTTP %s", ErrWebSocketUnavailable, response.Status)
		}
		return fmt.Errorf("Pulse WebSocket connection failed: %w", err)
	}
	defer conn.Close()

	for {
		messageType, payload, err := conn.ReadMessage()
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return fmt.Errorf("Pulse WebSocket read failed: %w", err)
		}
		if messageType != websocket.TextMessage && messageType != websocket.BinaryMessage {
			continue
		}
		body := payload
		if len(body) > 1 && body[0] == '<' {
			if end := strings.IndexByte(string(body), '>'); end >= 0 {
				body = body[end+1:]
			}
		}
		if len(body) > 0 {
			onFrame(body)
		}
	}
}
