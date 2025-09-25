package envelope

import (
	"context"
	"fmt"

	"app/api/handlers/dtos"
	"app/usecase/interfaces"

	"github.com/sirupsen/logrus"
)

// CheckEventsFromClicksignAPI verifica eventos da API da Clicksign e dispara webhooks internos
func (u *UsecaseEnvelopeService) CheckEventsFromClicksignAPI(ctx context.Context, envelopeID int, webhookUsecase interfaces.UsecaseWebhookMinimalInterface) (*dtos.WebhookProcessResponseDTO, error) {
	u.logger.WithField("envelope_id", envelopeID).Info("Checking events from Clicksign API as webhook fallback")

	// Buscar envelope
	envelope, err := u.repositoryEnvelope.GetByID(envelopeID)
	if err != nil {
		return nil, fmt.Errorf("envelope not found: %w", err)
	}

	if envelope.ClicksignKey == "" {
		return nil, fmt.Errorf("envelope does not have clicksign_key")
	}

	// Verificar se envelope já foi processado completamente
	if envelope.Status == "completed" || envelope.Status == "cancelled" {
		u.logger.WithFields(logrus.Fields{
			"envelope_id": envelopeID,
			"status":      envelope.Status,
		}).Info("Envelope already in final state, skipping event check")

		return &dtos.WebhookProcessResponseDTO{
			Success: true,
			Message: fmt.Sprintf("Envelope is already in '%s' state, no events processed", envelope.Status),
		}, nil
	}

	// Verificar se já existem webhooks de assinatura processados para este envelope
	existingWebhooks, err := webhookUsecase.GetWebhooksByDocumentKey(envelope.ClicksignKey)
	if err != nil {
		u.logger.WithError(err).Warn("Failed to check existing webhooks, continuing with event check")
	}

	processedSignEvents := 0
	for _, webhook := range existingWebhooks {
		if webhook.EventName == "sign" && webhook.Status == "processed" {
			processedSignEvents++
		}
	}

	u.logger.WithFields(logrus.Fields{
		"envelope_id":           envelopeID,
		"existing_sign_events":  processedSignEvents,
		"total_webhooks":        len(existingWebhooks),
	}).Info("Found existing webhooks for envelope")

	// Buscar eventos via API da Clicksign
	// TODO: Implementar serviço de eventos quando disponível
	u.logger.Warn("EventsService not implemented yet, skipping event check")
	
	// Retornar resposta indicando que não foi possível verificar eventos
	return &dtos.WebhookProcessResponseDTO{
		Success: false,
		Message: "EventsService not implemented yet",
	}, nil
}