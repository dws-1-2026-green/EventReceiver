package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"event-receiver/kafka"
	"event-receiver/models"
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
	eventType := r.Header.Get("Event-Type")
	if len(eventType) == 0 {
		sendJSONResponse(w, 400, models.APIResponse{
			Status:  "Fail",
			Message: "EventType header required",
		})
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		sendJSONResponse(w, 400, models.APIResponse{
			Status:  "Wrong request",
			Message: "Body required",
		})
		return
	}
	request := models.EventCreateMessage{
		EventType: eventType,
		Event:     json.RawMessage{},
	}
	err = request.Event.UnmarshalJSON(body)
	if err != nil {
		sendJSONResponse(w, 400, models.APIResponse{
			Status:  "Wrong request",
			Message: "Can't parse JSON from body",
		})
		return
	}

	err = h.kafkaProducer.SendMessage("event", request)
	if err != nil {
		log.Printf("Fail to send message %s", err)
		sendJSONResponse(w, 500, models.APIResponse{
			Status:  "Internal error",
			Message: "Fail to handle event",
		})
		return
	}

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
	json.NewEncoder(w).Encode(data)
}
