package webhook

import (
	"app/api/handlers/dtos"
	"app/entity"
)

type UsecaseWebhookInterface interface {
	// ProcessWebhook processa um webhook recebido
	ProcessWebhook(webhookDTO *dtos.WebhookRequestDTO, rawPayload string) (*entity.EntityWebhook, error)

	// ProcessAutoCloseEvent processa especificamente eventos de fechamento automático
	ProcessAutoCloseEvent(webhookDTO *dtos.WebhookRequestDTO, webhook *entity.EntityWebhook) error

	// ProcessSignEvent processa eventos de assinatura
	ProcessSignEvent(webhookDTO *dtos.WebhookRequestDTO, webhook *entity.EntityWebhook) error

	// ProcessSignatureStartedEvent processa eventos de início de assinatura
	ProcessSignatureStartedEvent(webhookDTO *dtos.WebhookRequestDTO, webhook *entity.EntityWebhook) error

	// ProcessAddSignerEvent processa eventos de adição de signatário
	ProcessAddSignerEvent(webhookDTO *dtos.WebhookRequestDTO, webhook *entity.EntityWebhook) error

	// ProcessUploadEvent processa eventos de upload
	ProcessUploadEvent(webhookDTO *dtos.WebhookRequestDTO, webhook *entity.EntityWebhook) error

	// GetWebhookByID busca um webhook por ID
	GetWebhookByID(id int) (*entity.EntityWebhook, error)

	// GetWebhooksByDocumentKey busca webhooks por document key
	GetWebhooksByDocumentKey(documentKey string) ([]entity.EntityWebhook, error)

	// GetWebhooksByAccountKey busca webhooks por account key
	GetWebhooksByAccountKey(accountKey string) ([]entity.EntityWebhook, error)

	// GetWebhooksByEventName busca webhooks por nome do evento
	GetWebhooksByEventName(eventName string) ([]entity.EntityWebhook, error)

	// GetWebhooksByStatus busca webhooks por status
	GetWebhooksByStatus(status string) ([]entity.EntityWebhook, error)

	// GetPendingWebhooks busca webhooks pendentes
	GetPendingWebhooks() ([]entity.EntityWebhook, error)

	// GetFailedWebhooks busca webhooks que falharam
	GetFailedWebhooks() ([]entity.EntityWebhook, error)

	// GetAllWebhooks busca todos os webhooks com paginação
	GetAllWebhooks(page, limit int) ([]entity.EntityWebhook, int64, error)

	// GetWebhooksByFilters busca webhooks com filtros
	GetWebhooksByFilters(filters *dtos.WebhookFiltersDTO) ([]entity.EntityWebhook, int64, error)

	// RetryWebhook tenta reprocessar um webhook que falhou
	RetryWebhook(id int) error

	// DeleteWebhook deleta um webhook
	DeleteWebhook(id int) error

	// MarkWebhookAsProcessed marca um webhook como processado
	MarkWebhookAsProcessed(id int) error

	// MarkWebhookAsFailed marca um webhook como falhou
	MarkWebhookAsFailed(id int, errorMsg string) error
}
