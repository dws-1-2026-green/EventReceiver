package models

import "encoding/json"

type APIResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type EventCreateMessage struct {
	EventType string          `json:"eventType"`
	Event     json.RawMessage `json:"event"`
}
