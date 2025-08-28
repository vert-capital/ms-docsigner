package webhook

import (
	"app/api/handlers/dtos"
	"app/entity"
	"app/infrastructure/repository"
	"app/usecase/document"
	"app/usecase/envelope"
	"encoding/json"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

type UsecaseWebhookService struct {
	webhookRepository *repository.RepositoryWebhook
	envelopeUsecase   envelope.IUsecaseEnvelope
	documentUsecase   document.IUsecaseDocument
	logger            *logrus.Logger
}

func NewUsecaseWebhookService(
	webhookRepository *repository.RepositoryWebhook,
	envelopeUsecase envelope.IUsecaseEnvelope,
	documentUsecase document.IUsecaseDocument,
	logger *logrus.Logger,
) *UsecaseWebhookService {
	return &UsecaseWebhookService{
		webhookRepository: webhookRepository,
		envelopeUsecase:   envelopeUsecase,
		documentUsecase:   documentUsecase,
		logger:            logger,
	}
}

// ProcessWebhook processa um webhook recebido
func (u *UsecaseWebhookService) ProcessWebhook(webhookDTO *dtos.WebhookRequestDTO, rawPayload string) (*entity.EntityWebhook, error) {
	u.logger.Info("Processing webhook", map[string]interface{}{
		"event_name":   webhookDTO.Event.Name,
		"document_key": webhookDTO.Document.Key,
		"account_key":  webhookDTO.Document.AccountKey,
	})

	// Criar entidade webhook
	webhook, err := entity.NewWebhook(
		webhookDTO.Event.Name,
		webhookDTO.Document.Key,
		webhookDTO.Document.AccountKey,
		rawPayload,
	)
	if err != nil {
		u.logger.Error("Failed to create webhook entity", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to create webhook entity: %w", err)
	}

	// Salvar webhook no banco
	err = u.webhookRepository.Create(webhook)
	if err != nil {
		u.logger.Error("Failed to save webhook", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to save webhook: %w", err)
	}

	// Processar evento específico
	err = u.processSpecificEvent(webhookDTO, webhook)
	if err != nil {
		u.logger.Error("Failed to process specific event", map[string]interface{}{
			"error":      err.Error(),
			"webhook_id": webhook.ID,
		})

		// Marcar webhook como falhou
		webhook.MarkAsFailed(err.Error())
		u.webhookRepository.Update(webhook)

		return webhook, fmt.Errorf("failed to process specific event: %w", err)
	}

	// Marcar webhook como processado
	webhook.MarkAsProcessed()
	err = u.webhookRepository.Update(webhook)
	if err != nil {
		u.logger.Error("Failed to mark webhook as processed", map[string]interface{}{
			"error":      err.Error(),
			"webhook_id": webhook.ID,
		})
		return webhook, fmt.Errorf("failed to mark webhook as processed: %w", err)
	}

	u.logger.Info("Webhook processed successfully", map[string]interface{}{
		"webhook_id": webhook.ID,
		"event_name": webhookDTO.Event.Name,
	})

	return webhook, nil
}

// processSpecificEvent processa o evento específico baseado no tipo
func (u *UsecaseWebhookService) processSpecificEvent(webhookDTO *dtos.WebhookRequestDTO, webhook *entity.EntityWebhook) error {
	switch webhookDTO.Event.Name {
	case "auto_close":
		return u.ProcessAutoCloseEvent(webhookDTO, webhook)
	case "sign":
		return u.ProcessSignEvent(webhookDTO, webhook)
	case "signature_started":
		return u.ProcessSignatureStartedEvent(webhookDTO, webhook)
	case "add_signer":
		return u.ProcessAddSignerEvent(webhookDTO, webhook)
	case "upload":
		return u.ProcessUploadEvent(webhookDTO, webhook)
	default:
		u.logger.Warn("Unknown event type", map[string]interface{}{
			"event_name": webhookDTO.Event.Name,
		})
		return nil // Evento desconhecido não é erro, apenas ignorado
	}
}

// ProcessAutoCloseEvent processa especificamente eventos de fechamento automático
func (u *UsecaseWebhookService) ProcessAutoCloseEvent(webhookDTO *dtos.WebhookRequestDTO, webhook *entity.EntityWebhook) error {
	u.logger.Info("Processing auto close event", map[string]interface{}{
		"document_key": webhookDTO.Document.Key,
		"status":       webhookDTO.Document.Status,
	})

	// Buscar envelope pelo ID que está no metadata do documento
	var envelopeID int
	if webhookDTO.Document.Metadata != nil {
		if id, ok := webhookDTO.Document.Metadata["envelope_id"].(float64); ok {
			envelopeID = int(id)
		}
	}

	var envelope *entity.EntityEnvelope
	var err error

	if envelopeID > 0 {
		// Tentar buscar pelo ID primeiro
		envelope, err = u.envelopeUsecase.GetEnvelope(envelopeID)
		if err != nil {
			u.logger.Warn("Failed to find envelope by ID, trying by document key", map[string]interface{}{
				"envelope_id":  envelopeID,
				"document_key": webhookDTO.Document.Key,
				"error":        err.Error(),
			})
		}
	}

	// Se não encontrou pelo ID, tentar pelo document key
	if envelope == nil {
		envelope, err = u.envelopeUsecase.GetEnvelopeByClicksignKey(webhookDTO.Document.Key)
		if err != nil {
			return fmt.Errorf("failed to find envelope by document key: %w", err)
		}
	}

	if envelope == nil {
		u.logger.Warn("No envelope found for document key", map[string]interface{}{
			"document_key": webhookDTO.Document.Key,
		})
		return nil // Não é erro se não encontrar envelope
	}

	// Verificar se o envelope já está finalizado
	if envelope.Status == "completed" {
		u.logger.Warn("Envelope already completed, ignoring auto close event", map[string]interface{}{
			"envelope_id":  envelope.ID,
			"document_key": webhookDTO.Document.Key,
			"status":       envelope.Status,
		})
		return fmt.Errorf("envelope is already completed and cannot be processed again. Envelope ID: %d, Document Key: %s", envelope.ID, webhookDTO.Document.Key)
	}

	// Log do status atual do envelope antes da atualização
	u.logger.Info("Envelope found, updating status to completed", map[string]interface{}{
		"envelope_id":    envelope.ID,
		"document_key":   webhookDTO.Document.Key,
		"current_status": envelope.Status,
		"new_status":     "completed",
	})

	// Atualizar status do envelope para completed
	err = envelope.SetStatus("completed")
	if err != nil {
		return fmt.Errorf("failed to set envelope status to completed: %w", err)
	}

	// Salvar dados raw do Clicksign
	rawData, _ := json.Marshal(webhookDTO)
	envelope.SetClicksignRawData(string(rawData))

	// Atualizar envelope no banco usando método específico para webhooks
	err = u.envelopeUsecase.UpdateEnvelopeForWebhook(envelope)
	if err != nil {
		return fmt.Errorf("failed to update envelope: %w", err)
	}

	u.logger.Info("Envelope updated successfully for auto close event", map[string]interface{}{
		"envelope_id":  envelope.ID,
		"document_key": webhookDTO.Document.Key,
		"new_status":   envelope.Status,
	})

	// Atualizar status do documento para "sent" (finalizado)
	// Buscar documento pelo clicksign_key
	document, err := u.documentUsecase.GetDocumentByClicksignKey(webhookDTO.Document.Key)
	if err != nil {
		u.logger.Warn("Failed to find document by clicksign key", map[string]interface{}{
			"document_key": webhookDTO.Document.Key,
			"error":        err.Error(),
		})
		// Não é erro crítico se não encontrar o documento
	} else {
		// Atualizar status do documento para "sent"
		err = document.SetStatus("sent")
		if err != nil {
			u.logger.Warn("Failed to set document status to sent", map[string]interface{}{
				"document_id": document.ID,
				"error":       err.Error(),
			})
		} else {
			// Atualizar documento no banco
			err = u.documentUsecase.Update(document)
			if err != nil {
				u.logger.Warn("Failed to update document", map[string]interface{}{
					"document_id": document.ID,
					"error":       err.Error(),
				})
			} else {
				u.logger.Info("Document updated successfully for auto close event", map[string]interface{}{
					"document_id":  document.ID,
					"document_key": webhookDTO.Document.Key,
					"new_status":   document.Status,
				})
			}
		}
	}

	// Salvar dados do evento no webhook
	err = webhook.SetEventData(webhookDTO)
	if err != nil {
		return fmt.Errorf("failed to set event data: %w", err)
	}

	return nil
}

// ProcessSignEvent processa eventos de assinatura
func (u *UsecaseWebhookService) ProcessSignEvent(webhookDTO *dtos.WebhookRequestDTO, webhook *entity.EntityWebhook) error {
	u.logger.Info("Processing sign event", map[string]interface{}{
		"document_key": webhookDTO.Document.Key,
	})

	// Salvar dados do evento no webhook
	err := webhook.SetEventData(webhookDTO)
	if err != nil {
		return fmt.Errorf("failed to set event data: %w", err)
	}

	// Aqui você pode adicionar lógica específica para eventos de assinatura
	// Por exemplo, notificar outros sistemas, atualizar métricas, etc.

	return nil
}

// ProcessSignatureStartedEvent processa eventos de início de assinatura
func (u *UsecaseWebhookService) ProcessSignatureStartedEvent(webhookDTO *dtos.WebhookRequestDTO, webhook *entity.EntityWebhook) error {
	u.logger.Info("Processing signature started event", map[string]interface{}{
		"document_key": webhookDTO.Document.Key,
	})

	// Salvar dados do evento no webhook
	err := webhook.SetEventData(webhookDTO)
	if err != nil {
		return fmt.Errorf("failed to set event data: %w", err)
	}

	// Aqui você pode adicionar lógica específica para início de assinatura
	// Por exemplo, enviar notificações, atualizar status, etc.

	return nil
}

// ProcessAddSignerEvent processa eventos de adição de signatário
func (u *UsecaseWebhookService) ProcessAddSignerEvent(webhookDTO *dtos.WebhookRequestDTO, webhook *entity.EntityWebhook) error {
	u.logger.Info("Processing add signer event", map[string]interface{}{
		"document_key": webhookDTO.Document.Key,
	})

	// Salvar dados do evento no webhook
	err := webhook.SetEventData(webhookDTO)
	if err != nil {
		return fmt.Errorf("failed to set event data: %w", err)
	}

	// Aqui você pode adicionar lógica específica para adição de signatário
	// Por exemplo, sincronizar signatários, enviar notificações, etc.

	return nil
}

// ProcessUploadEvent processa eventos de upload
func (u *UsecaseWebhookService) ProcessUploadEvent(webhookDTO *dtos.WebhookRequestDTO, webhook *entity.EntityWebhook) error {
	u.logger.Info("Processing upload event", map[string]interface{}{
		"document_key": webhookDTO.Document.Key,
	})

	// Salvar dados do evento no webhook
	err := webhook.SetEventData(webhookDTO)
	if err != nil {
		return fmt.Errorf("failed to set event data: %w", err)
	}

	// Aqui você pode adicionar lógica específica para upload
	// Por exemplo, sincronizar documentos, atualizar metadados, etc.

	return nil
}

// GetWebhookByID busca um webhook por ID
func (u *UsecaseWebhookService) GetWebhookByID(id int) (*entity.EntityWebhook, error) {
	return u.webhookRepository.GetByID(id)
}

// GetWebhooksByDocumentKey busca webhooks por document key
func (u *UsecaseWebhookService) GetWebhooksByDocumentKey(documentKey string) ([]entity.EntityWebhook, error) {
	return u.webhookRepository.GetByDocumentKey(documentKey)
}

// GetWebhooksByAccountKey busca webhooks por account key
func (u *UsecaseWebhookService) GetWebhooksByAccountKey(accountKey string) ([]entity.EntityWebhook, error) {
	return u.webhookRepository.GetByAccountKey(accountKey)
}

// GetWebhooksByEventName busca webhooks por nome do evento
func (u *UsecaseWebhookService) GetWebhooksByEventName(eventName string) ([]entity.EntityWebhook, error) {
	return u.webhookRepository.GetByEventName(eventName)
}

// GetWebhooksByStatus busca webhooks por status
func (u *UsecaseWebhookService) GetWebhooksByStatus(status string) ([]entity.EntityWebhook, error) {
	return u.webhookRepository.GetByStatus(status)
}

// GetPendingWebhooks busca webhooks pendentes
func (u *UsecaseWebhookService) GetPendingWebhooks() ([]entity.EntityWebhook, error) {
	return u.webhookRepository.GetPending()
}

// GetFailedWebhooks busca webhooks que falharam
func (u *UsecaseWebhookService) GetFailedWebhooks() ([]entity.EntityWebhook, error) {
	return u.webhookRepository.GetFailed()
}

// GetAllWebhooks busca todos os webhooks com paginação
func (u *UsecaseWebhookService) GetAllWebhooks(page, limit int) ([]entity.EntityWebhook, int64, error) {
	return u.webhookRepository.GetAll(page, limit)
}

// GetWebhooksByFilters busca webhooks com filtros
func (u *UsecaseWebhookService) GetWebhooksByFilters(filters *dtos.WebhookFiltersDTO) ([]entity.EntityWebhook, int64, error) {
	if filters.Page <= 0 {
		filters.Page = 1
	}
	if filters.Limit <= 0 {
		filters.Limit = 10
	}

	return u.webhookRepository.GetByFilters(
		filters.EventName,
		filters.DocumentKey,
		filters.AccountKey,
		filters.Status,
		filters.Page,
		filters.Limit,
	)
}

// RetryWebhook tenta reprocessar um webhook que falhou
func (u *UsecaseWebhookService) RetryWebhook(id int) error {
	webhook, err := u.webhookRepository.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get webhook: %w", err)
	}

	if !webhook.IsFailed() {
		return fmt.Errorf("webhook is not in failed status, current status: %s", webhook.Status)
	}

	// Reset status para pending
	webhook.Status = "pending"
	webhook.Error = nil
	webhook.UpdatedAt = time.Now()

	err = u.webhookRepository.Update(webhook)
	if err != nil {
		return fmt.Errorf("failed to update webhook status: %w", err)
	}

	u.logger.Info("Webhook marked for retry", map[string]interface{}{
		"webhook_id": webhook.ID,
	})

	return nil
}

// DeleteWebhook deleta um webhook
func (u *UsecaseWebhookService) DeleteWebhook(id int) error {
	return u.webhookRepository.Delete(id)
}

// MarkWebhookAsProcessed marca um webhook como processado
func (u *UsecaseWebhookService) MarkWebhookAsProcessed(id int) error {
	return u.webhookRepository.MarkAsProcessed(id)
}

// MarkWebhookAsFailed marca um webhook como falhou
func (u *UsecaseWebhookService) MarkWebhookAsFailed(id int, errorMsg string) error {
	return u.webhookRepository.MarkAsFailed(id, errorMsg)
}
