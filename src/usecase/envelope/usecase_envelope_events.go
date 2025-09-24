package envelope

import (
	"context"
	"fmt"
	"time"

	"app/api/handlers/dtos"
	"app/infrastructure/clicksign"
	"app/usecase/webhook"

	"github.com/sirupsen/logrus"
)

// CheckEventsFromClicksignAPI verifica eventos da API da Clicksign e dispara webhooks internos
func (u *UsecaseEnvelopeService) CheckEventsFromClicksignAPI(ctx context.Context, envelopeID int, webhookUsecase webhook.UsecaseWebhookInterface) (*dtos.WebhookProcessResponseDTO, error) {
	u.logger.WithField("envelope_id", envelopeID).Info("Checking events from Clicksign API as webhook fallback")

	// Buscar envelope
	envelope, err := u.repositoryEnvelope.GetByID(envelopeID)
	if err != nil {
		return nil, fmt.Errorf("envelope not found: %w", err)
	}

	if envelope.ClicksignKey == "" {
		return nil, fmt.Errorf("envelope does not have clicksign_key")
	}

	// Buscar eventos via API da Clicksign
	eventsService := clicksign.NewEventsService(u.clicksignClient, u.logger)
	signatureStatuses, err := eventsService.GetSignaturesStatus(ctx, envelope.ClicksignKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get events from Clicksign API: %w", err)
	}

	processedEvents := 0

	// Processar eventos de assinatura encontrados
	for signerKey, status := range signatureStatuses {
		if status.Signed && status.SignedAt != nil {
			// Criar webhook DTO simulando evento de assinatura
			webhookDTO := &dtos.WebhookRequestDTO{
				Event: dtos.WebhookEventDTO{
					Name:       "sign",
					OccurredAt: status.SignedAt.Format(time.RFC3339),
					Data: map[string]interface{}{
						"signer": map[string]interface{}{
							"key":   signerKey,
							"email": status.Email,
							"name":  status.Name,
						},
					},
				},
				Document: dtos.WebhookDocumentDTO{
					Key:        envelope.ClicksignKey,
					AccountKey: "api-fallback", // Identificar como fallback manual
					Status:     "running",
				},
			}

			// Disparar o processamento de webhook existente
			rawPayload := fmt.Sprintf(`{"source":"api_fallback","signer_key":"%s","envelope_id":%d,"signed_at":"%s"}`,
				signerKey, envelopeID, status.SignedAt.Format(time.RFC3339))

			_, err := webhookUsecase.ProcessWebhook(webhookDTO, rawPayload)
			if err != nil {
				u.logger.WithError(err).WithFields(logrus.Fields{
					"signer_key":  signerKey,
					"envelope_id": envelopeID,
				}).Error("Failed to process sign event via internal webhook")
				continue
			}

			processedEvents++

			u.logger.WithFields(logrus.Fields{
				"signer_key":  signerKey,
				"envelope_id": envelopeID,
				"signed_at":   status.SignedAt,
				"email":       status.Email,
				"name":        status.Name,
			}).Info("Processed sign event via API fallback - webhook triggered internally")
		}
	}

	return &dtos.WebhookProcessResponseDTO{
		Success: true,
		Message: fmt.Sprintf("Checked Clicksign events API and processed %d sign events. Internal webhooks were triggered for each signature found.", processedEvents),
	}, nil
}