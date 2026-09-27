package data

import (
	"context"

	log "github.com/DjSni/go-log"

	"pulse2mqtt/include/pulse"
)

func StreamModernData(ctx context.Context, onData func(Data)) error {
	return pulse.StreamModernData(ctx, func(body []byte) {
		messages, err := parseSML(body)
		if err != nil {
			log.Error("Modern WebSocket SML parse error:", err)
			return
		}
		onData(dataFromMessages(messages))
	})
}
