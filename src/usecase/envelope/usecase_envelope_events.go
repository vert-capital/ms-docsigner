package envelope

import (
	"context"
	"fmt"

	"app/infrastructure/clicksign"

	"github.com/sirupsen/logrus"
)


// CheckEventsFromClicksignAPI verifica eventos da API da Clicksign e retorna os eventos encontrados
func (u *UsecaseEnvelopeService) CheckEventsFromClicksignAPI(ctx context.Context, envelopeID int) (*CheckEventsResult, error) {
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

		return &CheckEventsResult{
			Events:        []SignatureEventData{},
			ProcessedCount: 0,
			EnvelopeKey:   envelope.ClicksignKey,
		}, nil
	}

	// Buscar eventos via API da Clicksign
	eventsService := clicksign.NewEventsService(u.clicksignClient, u.logger)
	signatureStatuses, err := eventsService.GetSignaturesStatus(ctx, envelope.ClicksignKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get events from Clicksign API: %w", err)
	}

	var events []SignatureEventData

	u.logger.WithFields(logrus.Fields{
		"envelope_id":      envelopeID,
		"api_events_found": len(signatureStatuses),
	}).Info("Retrieved events from Clicksign API")

	// Processar eventos de assinatura encontrados
	for signerKey, status := range signatureStatuses {
		if status.Signed && status.SignedAt != nil {
			events = append(events, SignatureEventData{
				SignerKey: signerKey,
				Email:     status.Email,
				Name:      status.Name,
				SignedAt:  *status.SignedAt,
			})

			u.logger.WithFields(logrus.Fields{
				"signer_key":  signerKey,
				"envelope_id": envelopeID,
				"signed_at":   status.SignedAt,
				"email":       status.Email,
			}).Info("Found signature event via API")
		}
	}

	u.logger.WithFields(logrus.Fields{
		"envelope_id":    envelopeID,
		"events_found":   len(events),
	}).Info("Event check completed, returning events for processing")

	return &CheckEventsResult{
		Events:        events,
		ProcessedCount: len(events),
		EnvelopeKey:   envelope.ClicksignKey,
	}, nil
}