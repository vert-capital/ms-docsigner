package usecase_envelope

// CheckEventsFromClicksignAPI verifica eventos da API da Clicksign e dispara webhooks internos
// func (u *UsecaseEnvelopeService) CheckEventsFromClicksignAPI(ctx context.Context, envelopeID int, webhookUsecase webhook.UsecaseWebhookInterface) (*dtos.WebhookProcessResponseDTO, error) {
// 	u.logger.WithField("envelope_id", envelopeID).Info("Checking events from Clicksign API as webhook fallback")

// 	// Buscar envelope
// 	envelope, err := u.repositoryEnvelope.GetByID(envelopeID)
// 	if err != nil {
// 		return nil, fmt.Errorf("envelope not found: %w", err)
// 	}

// 	if envelope.ClicksignKey == "" {
// 		return nil, fmt.Errorf("envelope does not have clicksign_key")
// 	}

// 	// Verificar se envelope já foi processado completamente
// 	if envelope.Status == "completed" || envelope.Status == "cancelled" {
// 		u.logger.WithFields(logrus.Fields{
// 			"envelope_id": envelopeID,
// 			"status":      envelope.Status,
// 		}).Info("Envelope already in final state, skipping event check")

// 		return &dtos.WebhookProcessResponseDTO{
// 			Success: true,
// 			Message: fmt.Sprintf("Envelope is already in '%s' state, no events processed", envelope.Status),
// 		}, nil
// 	}

// 	// Verificar se já existem webhooks de assinatura processados para este envelope
// 	existingWebhooks, err := webhookUsecase.GetWebhooksByDocumentKey(envelope.ClicksignKey)
// 	if err != nil {
// 		u.logger.WithError(err).Warn("Failed to check existing webhooks, continuing with event check")
// 	}

// 	processedSignEvents := 0
// 	for _, webhook := range existingWebhooks {
// 		if webhook.EventName == "sign" && webhook.Status == "processed" {
// 			processedSignEvents++
// 		}
// 	}

// 	u.logger.WithFields(logrus.Fields{
// 		"envelope_id":          envelopeID,
// 		"existing_sign_events": processedSignEvents,
// 		"total_webhooks":       len(existingWebhooks),
// 	}).Info("Found existing webhooks for envelope")

// 	// Buscar eventos via API da Clicksign
// 	eventsService := clicksign.NewEventsService(u.clicksignClient, u.logger)
// 	signatureStatuses, err := eventsService.GetSignaturesStatus(ctx, envelope.ClicksignKey)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get events from Clicksign API: %w", err)
// 	}

// 	processedEvents := 0
// 	skippedEvents := 0

// 	// Criar mapa de signers já processados via webhook
// 	processedSigners := make(map[string]bool)
// 	for _, webhook := range existingWebhooks {
// 		if webhook.EventName == "sign" && webhook.Status == "processed" {
// 			// Tentar extrair signer_key do raw payload
// 			var payloadData map[string]interface{}
// 			if err := json.Unmarshal([]byte(webhook.RawPayload), &payloadData); err == nil {
// 				if eventData, ok := payloadData["event"].(map[string]interface{}); ok {
// 					if data, ok := eventData["data"].(map[string]interface{}); ok {
// 						if signer, ok := data["signer"].(map[string]interface{}); ok {
// 							if signerKey, ok := signer["key"].(string); ok {
// 								processedSigners[signerKey] = true
// 							}
// 						}
// 					}
// 				}
// 			}
// 		}
// 	}

// 	u.logger.WithField("processed_signers_count", len(processedSigners)).Info("Mapped already processed signers")

// 	// Processar eventos de assinatura encontrados
// 	for signerKey, status := range signatureStatuses {
// 		if status.Signed && status.SignedAt != nil {
// 			// Verificar se esta assinatura já foi processada via webhook
// 			if processedSigners[signerKey] {
// 				skippedEvents++
// 				u.logger.WithFields(logrus.Fields{
// 					"signer_key": signerKey,
// 					"email":      status.Email,
// 					"signed_at":  status.SignedAt,
// 				}).Info("Skipping already processed signature event")
// 				continue
// 			}
// 			// Criar webhook DTO simulando evento de assinatura
// 			webhookDTO := &dtos.WebhookRequestDTO{
// 				Event: dtos.WebhookEventDTO{
// 					Name:       "sign",
// 					OccurredAt: status.SignedAt.Format(time.RFC3339),
// 					Data: map[string]interface{}{
// 						"signer": map[string]interface{}{
// 							"key":   signerKey,
// 							"email": status.Email,
// 							"name":  status.Name,
// 						},
// 					},
// 				},
// 				Document: dtos.WebhookDocumentDTO{
// 					Key:        envelope.ClicksignKey,
// 					AccountKey: "api-fallback", // Identificar como fallback manual
// 					Status:     "running",
// 				},
// 			}

// 			// Disparar o processamento de webhook existente
// 			rawPayload := fmt.Sprintf(`{"source":"api_fallback","signer_key":"%s","envelope_id":%d,"signed_at":"%s"}`,
// 				signerKey, envelopeID, status.SignedAt.Format(time.RFC3339))

// 			_, err := webhookUsecase.ProcessWebhook(webhookDTO, rawPayload)
// 			if err != nil {
// 				u.logger.WithError(err).WithFields(logrus.Fields{
// 					"signer_key":  signerKey,
// 					"envelope_id": envelopeID,
// 				}).Error("Failed to process sign event via internal webhook")
// 				continue
// 			}

// 			processedEvents++

// 			u.logger.WithFields(logrus.Fields{
// 				"signer_key":  signerKey,
// 				"envelope_id": envelopeID,
// 				"signed_at":   status.SignedAt,
// 				"email":       status.Email,
// 				"name":        status.Name,
// 			}).Info("Processed sign event via API fallback - webhook triggered internally")
// 		}
// 	}

// 	// Montar mensagem de resposta detalhada
// 	message := fmt.Sprintf("Checked Clicksign events API: processed %d new sign events", processedEvents)

// 	if skippedEvents > 0 {
// 		message += fmt.Sprintf(", skipped %d already processed events", skippedEvents)
// 	}

// 	if processedEvents > 0 {
// 		message += ". Internal webhooks were triggered for new signatures"
// 	}

// 	if processedEvents == 0 && skippedEvents == 0 {
// 		message += ". No signature events found in API"
// 	}

// 	u.logger.WithFields(logrus.Fields{
// 		"envelope_id":      envelopeID,
// 		"processed_events": processedEvents,
// 		"skipped_events":   skippedEvents,
// 		"total_api_events": len(signatureStatuses),
// 	}).Info("Event check completed")

// 	return &dtos.WebhookProcessResponseDTO{
// 		Success: true,
// 		Message: message,
// 	}, nil
// }
