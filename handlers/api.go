package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"event-receiver/kafka"
	"event-receiver/models"

	"github.com/google/uuid"
)

type APIHandler struct {
	kafkaProducer *kafka.Producer
}

func NewAPIHandler(kp *kafka.Producer) *APIHandler {
	return &APIHandler{
		kafkaProducer: kp,
	}
}

func (h *APIHandler) PostEvent(w http.ResponseWriter, r *http.Request) {
	sourceName := r.PathValue("source_name")
	if len(sourceName) == 0 {
		sendJSONResponse(w, 400, models.APIResponse{
			Status:  "Wrong request",
			Message: "url must match pattern /sources/{source_name}/events",
		})
		return
	}

	request := models.APIRequest{}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		sendJSONResponse(w, 400, models.APIResponse{
			Status:  "Wrong request",
			Message: "Can't parse JSON from body",
		})
		slog.Warn(
			"Fail to parse request json",
			slog.Any("error", err),
		)
		return
	}

	slog.Info(
		"Receive event",
		slog.String("source_name", sourceName),
		slog.String("type", request.Type),
		slog.String("id", request.Id),
	)

	message := models.EventCreateMessage{
		Event: models.TargetEvent{
			Id:     request.Id,
			Source: sourceName,
			Type:   request.Type,
			Data:   request.Data,
		},
		IngestedAt: time.Now(),
		TraceId:    uuid.NewString(),
	}

	if err := h.kafkaProducer.SendMessage("event", message); err != nil {
		slog.Error(
			"Fail to send message",
			slog.Any("error", err),
		)
		sendJSONResponse(w, 500, models.APIResponse{
			Status:  "Internal error",
			Message: "Fail to handle event",
		})
		return
	}

	slog.Info(
		"Event created",
		slog.String("source_name", message.Event.Source),
		slog.String("type", message.Event.Type),
		slog.String("id", message.Event.Id),
		slog.String("trace_id", message.TraceId),
	)

	sendJSONResponse(w, 200, models.APIResponse{
		Status:  "OK",
		Message: "Event sended",
	})
}

func (h *APIHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	response := models.APIResponse{
		Status:  "OK",
		Message: "Service is running",
	}
	sendJSONResponse(w, http.StatusOK, response)
}

func sendJSONResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		slog.Error(
			"Fail to encode repsonse json",
			slog.Int("status", status),
			slog.Any("data", data),
			slog.Any("error", err),
		)
	}
}
