package interfaces

import (
	"app/api/handlers/dtos"
	"app/entity"
)

// UsecaseEnvelopeMinimalInterface interface mínima para evitar ciclo de dependência
type UsecaseEnvelopeMinimalInterface interface {
	GetEnvelope(id int) (*entity.EntityEnvelope, error)
	GetEnvelopeByClicksignKey(key string) (*entity.EntityEnvelope, error)
	UpdateEnvelopeForWebhook(envelope *entity.EntityEnvelope) error
}

// UsecaseWebhookMinimalInterface interface mínima para evitar ciclo de dependência
type UsecaseWebhookMinimalInterface interface {
	ProcessWebhook(webhookDTO *dtos.WebhookRequestDTO, rawPayload string) (*entity.EntityWebhook, error)
	GetWebhooksByDocumentKey(documentKey string) ([]entity.EntityWebhook, error)
}