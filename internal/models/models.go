package models

import (
	"encoding/json"
	"time"
)

type APIResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type APIRequest struct {
	Id        string          `json:"id"`
	Type      string          `json:"type"`
	CreatedAt time.Time       `json:"created_at"`
	Data      json.RawMessage `json:"data,omitempty"`
}

type TargetEvent struct {
	Id     string          `json:"id"`
	Source string          `json:"source"`
	Type   string          `json:"type"`
	Data   json.RawMessage `json:"data"`
}

type EventCreateMessage struct {
	Event      TargetEvent `json:"event"`
	IngestedAt time.Time   `json:"ingested_at"`
	TraceId    string      `json:"trace_id"`
}
