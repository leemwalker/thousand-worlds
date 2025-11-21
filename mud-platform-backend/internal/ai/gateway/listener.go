package gateway

import (
	"encoding/json"
	"strings"

	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
)

// Subscriber defines the interface for subscribing to messages.
type Subscriber interface {
	Subscribe(subj string, cb nats.MsgHandler) (*nats.Subscription, error)
}

func StartListener(nc Subscriber) error {
	_, err := nc.Subscribe("ai.request.>", func(msg *nats.Msg) {
		var req AIRequest
		if err := json.Unmarshal(msg.Data, &req); err != nil {
			log.Error().Err(err).Msg("Failed to unmarshal AI request")
			return
		}

		// Determine response subject
		if msg.Reply != "" {
			req.ResponseSubject = msg.Reply
		} else if req.ID != "" {
			req.ResponseSubject = "ai.response." + req.ID
		} else {
			// Try to extract ID from subject if missing in body
			tokens := strings.Split(msg.Subject, ".")
			if len(tokens) >= 3 {
				req.ID = tokens[2]
				req.ResponseSubject = "ai.response." + req.ID
			}
		}

		// Push to queue without blocking
		select {
		case RequestQueue <- req:
			log.Debug().Str("id", req.ID).Msg("Queued AI request")
		default:
			log.Warn().Str("id", req.ID).Msg("Request queue full, dropping request")
		}
	})

	if err != nil {
		return err
	}

	log.Info().Msg("Listening on ai.request.>")
	return nil
}
