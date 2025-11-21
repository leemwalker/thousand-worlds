package gateway

import (
	"encoding/json"
	"time"

	"github.com/rs/zerolog/log"
)

var RequestQueue = make(chan AIRequest, 100)

// AIClient defines the interface for AI generation.
type AIClient interface {
	Generate(prompt string, model string) (string, error)
}

// Publisher defines the interface for publishing messages.
type Publisher interface {
	Publish(subj string, data []byte) error
}

func StartWorker(nc Publisher, client AIClient) {
	go func() {
		log.Info().Msg("AI Worker started")
		for req := range RequestQueue {
			processRequest(nc, client, req)
		}
	}()
}

func processRequest(nc Publisher, client AIClient, req AIRequest) {
	log.Info().Str("id", req.ID).Str("model", req.Model).Msg("Processing AI request")

	start := time.Now()
	response, err := client.Generate(req.Prompt, req.Model)
	duration := time.Since(start)

	aiResp := AIResponse{
		ID:       req.ID,
		Response: response,
	}

	if err != nil {
		log.Error().Err(err).Str("id", req.ID).Msg("Failed to generate AI response")
		aiResp.Error = err.Error()
	} else {
		log.Info().Str("id", req.ID).Dur("duration", duration).Msg("AI response generated")
	}

	respData, err := json.Marshal(aiResp)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal AI response")
		return
	}

	if req.ResponseSubject != "" {
		if err := nc.Publish(req.ResponseSubject, respData); err != nil {
			log.Error().Err(err).Str("subject", req.ResponseSubject).Msg("Failed to publish AI response")
		}
	} else {
		log.Warn().Str("id", req.ID).Msg("No response subject provided, skipping publish")
	}
}
