package payload_handler

import "context"

type PayloadHandler interface {
	InsertPayload(payload Payload)
	GetHandlerType() string
	CountUnProcessedPayload() uint
	Start(ctx context.Context)
}

type Payload struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type MockPayloadHandler struct {
	Payload []Payload
}

func (m *MockPayloadHandler) InsertPayload(payload Payload) {
	m.Payload = append(m.Payload, payload)
}

func (m *MockPayloadHandler) GetHandlerType() string {
	return "mock_data"
}

func (m *MockPayloadHandler) CountUnProcessedPayload() uint {
	return uint(len(m.Payload))
}

func (m *MockPayloadHandler) Start(ctx context.Context) {
}
