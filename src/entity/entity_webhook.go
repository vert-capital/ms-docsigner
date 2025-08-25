package entity

import (
	"encoding/json"
	"fmt"
	"time"
)

// EntityWebhook representa um webhook recebido
type EntityWebhook struct {
	ID          int        `json:"id" gorm:"primaryKey"`
	EventName   string     `json:"event_name" gorm:"not null" validate:"required"`
	EventData   string     `json:"event_data" gorm:"type:text"`
	DocumentKey string     `json:"document_key" gorm:"index"`
	AccountKey  string     `json:"account_key" gorm:"index"`
	Status      string     `json:"status" gorm:"not null;default:'pending'" validate:"required,oneof=pending processed failed"`
	ProcessedAt *time.Time `json:"processed_at"`
	Error       *string    `json:"error" gorm:"type:text"`
	RawPayload  string     `json:"raw_payload" gorm:"type:text;not null"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// TableName sets the table name for GORM
func (EntityWebhook) TableName() string {
	return "webhooks"
}

// NewWebhook cria uma nova instância de webhook
func NewWebhook(eventName, documentKey, accountKey, rawPayload string) (*EntityWebhook, error) {
	now := time.Now()

	w := &EntityWebhook{
		EventName:   eventName,
		DocumentKey: documentKey,
		AccountKey:  accountKey,
		Status:      "pending",
		RawPayload:  rawPayload,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	err := w.Validate()
	if err != nil {
		return nil, err
	}

	return w, nil
}

// Validate valida a entidade webhook
func (w *EntityWebhook) Validate() error {
	err := validate.Struct(w)
	if err != nil {
		return err
	}

	if w.EventName == "" {
		return fmt.Errorf("event_name is required")
	}

	if w.RawPayload == "" {
		return fmt.Errorf("raw_payload is required")
	}

	return nil
}

// SetEventData define os dados do evento como JSON
func (w *EntityWebhook) SetEventData(data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}
	w.EventData = string(jsonData)
	w.UpdatedAt = time.Now()
	return nil
}

// GetEventData retorna os dados do evento como interface{}
func (w *EntityWebhook) GetEventData() (map[string]interface{}, error) {
	if w.EventData == "" {
		return nil, nil
	}

	var data map[string]interface{}
	err := json.Unmarshal([]byte(w.EventData), &data)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal event data: %w", err)
	}

	return data, nil
}

// MarkAsProcessed marca o webhook como processado
func (w *EntityWebhook) MarkAsProcessed() {
	now := time.Now()
	w.Status = "processed"
	w.ProcessedAt = &now
	w.UpdatedAt = now
}

// MarkAsFailed marca o webhook como falhou
func (w *EntityWebhook) MarkAsFailed(errorMsg string) {
	w.Status = "failed"
	w.Error = &errorMsg
	w.UpdatedAt = time.Now()
}

// IsProcessed verifica se o webhook foi processado
func (w *EntityWebhook) IsProcessed() bool {
	return w.Status == "processed"
}

// IsFailed verifica se o webhook falhou
func (w *EntityWebhook) IsFailed() bool {
	return w.Status == "failed"
}

// IsPending verifica se o webhook está pendente
func (w *EntityWebhook) IsPending() bool {
	return w.Status == "pending"
}
