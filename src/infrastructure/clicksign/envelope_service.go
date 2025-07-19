package clicksign

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"app/entity"
	"app/infrastructure/clicksign/dto"
	"app/usecase/clicksign"
	"github.com/sirupsen/logrus"
)

type EnvelopeService struct {
	clicksignClient clicksign.ClicksignClientInterface
	logger          *logrus.Logger
}

func NewEnvelopeService(clicksignClient clicksign.ClicksignClientInterface, logger *logrus.Logger) *EnvelopeService {
	return &EnvelopeService{
		clicksignClient: clicksignClient,
		logger:          logger,
	}
}

func (s *EnvelopeService) CreateEnvelope(ctx context.Context, envelope *entity.EntityEnvelope) (string, string, error) {
	correlationID := ctx.Value("correlation_id")

	s.logger.WithFields(logrus.Fields{
		"envelope_name":   envelope.Name,
		"documents_count": len(envelope.DocumentsIDs),
		"signers_count":   len(envelope.SignatoryEmails),
		"api_format":      "JSON API",
		"content_type":    "application/vnd.api+json",
		"correlation_id":  correlationID,
	}).Info("Creating envelope in Clicksign using JSON API format")

	// Mapear entidade para DTO do Clicksign
	createRequest := s.mapEntityToCreateRequest(envelope)

	s.logger.WithFields(logrus.Fields{
		"data_type":        createRequest.Data.Type,
		"attributes_count": 6, // name, locale, auto_close, remind_interval, block_after_refusal, deadline_at
		"envelope_name":    createRequest.Data.Attributes.Name,
		"correlation_id":   correlationID,
	}).Debug("JSON API request structure prepared")

	// Fazer chamada para API do Clicksign
	resp, err := s.clicksignClient.Post(ctx, "/api/v3/envelopes", createRequest)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"error":          err.Error(),
			"envelope_name":  envelope.Name,
			"correlation_id": correlationID,
		}).Error("Failed to create envelope in Clicksign")
		return "", "", fmt.Errorf("failed to create envelope in Clicksign: %w", err)
	}
	defer resp.Body.Close()

	// Ler resposta
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"error":          err.Error(),
			"envelope_name":  envelope.Name,
			"correlation_id": correlationID,
		}).Error("Failed to read response from Clicksign")
		return "", "", fmt.Errorf("failed to read response from Clicksign: %w", err)
	}

	// Verificar se houve erro na resposta
	if resp.StatusCode >= 400 {
		var errorResp dto.ClicksignErrorResponse
		if err := json.Unmarshal(body, &errorResp); err != nil {
			s.logger.WithFields(logrus.Fields{
				"error":          err.Error(),
				"status_code":    resp.StatusCode,
				"response_body":  string(body),
				"correlation_id": correlationID,
			}).Error("Failed to parse error response from Clicksign")
			return "", "", fmt.Errorf("Clicksign API error (status %d): %s", resp.StatusCode, string(body))
		}

		s.logger.WithFields(logrus.Fields{
			"error_type":     errorResp.Error.Type,
			"error_message":  errorResp.Error.Message,
			"error_code":     errorResp.Error.Code,
			"status_code":    resp.StatusCode,
			"correlation_id": correlationID,
		}).Error("Clicksign API returned error")

		return "", "", fmt.Errorf("Clicksign API error: %s - %s", errorResp.Error.Type, errorResp.Error.Message)
	}

	// Preservar dados brutos antes do parse
	rawData := string(body)
	
	// Fazer parse da resposta de sucesso usando estrutura JSON API
	var createResponse dto.EnvelopeCreateResponseWrapper
	if err := json.Unmarshal(body, &createResponse); err != nil {
		s.logger.WithFields(logrus.Fields{
			"error":          err.Error(),
			"response_body":  string(body),
			"expected_format": "JSON API (data.type.attributes)",
			"correlation_id": correlationID,
		}).Error("Failed to parse JSON API response from Clicksign")
		return "", "", fmt.Errorf("failed to parse JSON API response from Clicksign: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"response_type":    createResponse.Data.Type,
		"response_id":      createResponse.Data.ID,
		"attributes_parsed": true,
		"correlation_id":   correlationID,
	}).Debug("JSON API response parsed successfully")

	s.logger.WithFields(logrus.Fields{
		"envelope_id":    createResponse.Data.ID,
		"envelope_name":  createResponse.Data.Attributes.Name,
		"status":         createResponse.Data.Attributes.Status,
		"type":           createResponse.Data.Type,
		"raw_data_size":  len(rawData),
		"correlation_id": correlationID,
	}).Info("Envelope created successfully in Clicksign using JSON API format")

	return createResponse.Data.ID, rawData, nil
}

func (s *EnvelopeService) GetEnvelope(ctx context.Context, clicksignKey string) (*dto.EnvelopeGetResponse, error) {
	correlationID := ctx.Value("correlation_id")

	s.logger.WithFields(logrus.Fields{
		"clicksign_key":  clicksignKey,
		"correlation_id": correlationID,
	}).Info("Getting envelope from Clicksign")

	endpoint := fmt.Sprintf("/api/v3/envelopes/%s", clicksignKey)
	resp, err := s.clicksignClient.Get(ctx, endpoint)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"error":          err.Error(),
			"clicksign_key":  clicksignKey,
			"correlation_id": correlationID,
		}).Error("Failed to get envelope from Clicksign")
		return nil, fmt.Errorf("failed to get envelope from Clicksign: %w", err)
	}
	defer resp.Body.Close()

	// Ler resposta
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"error":          err.Error(),
			"clicksign_key":  clicksignKey,
			"correlation_id": correlationID,
		}).Error("Failed to read response from Clicksign")
		return nil, fmt.Errorf("failed to read response from Clicksign: %w", err)
	}

	// Verificar se houve erro na resposta
	if resp.StatusCode >= 400 {
		var errorResp dto.ClicksignErrorResponse
		if err := json.Unmarshal(body, &errorResp); err != nil {
			s.logger.WithFields(logrus.Fields{
				"error":          err.Error(),
				"status_code":    resp.StatusCode,
				"response_body":  string(body),
				"correlation_id": correlationID,
			}).Error("Failed to parse error response from Clicksign")
			return nil, fmt.Errorf("Clicksign API error (status %d): %s", resp.StatusCode, string(body))
		}

		s.logger.WithFields(logrus.Fields{
			"error_type":     errorResp.Error.Type,
			"error_message":  errorResp.Error.Message,
			"error_code":     errorResp.Error.Code,
			"status_code":    resp.StatusCode,
			"correlation_id": correlationID,
		}).Error("Clicksign API returned error")

		return nil, fmt.Errorf("Clicksign API error: %s - %s", errorResp.Error.Type, errorResp.Error.Message)
	}

	// Fazer parse da resposta de sucesso
	var getResponse dto.EnvelopeGetResponse
	if err := json.Unmarshal(body, &getResponse); err != nil {
		s.logger.WithFields(logrus.Fields{
			"error":          err.Error(),
			"response_body":  string(body),
			"correlation_id": correlationID,
		}).Error("Failed to parse success response from Clicksign")
		return nil, fmt.Errorf("failed to parse response from Clicksign: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"envelope_id":    getResponse.ID,
		"envelope_name":  getResponse.Name,
		"status":         getResponse.Status,
		"correlation_id": correlationID,
	}).Info("Envelope retrieved successfully from Clicksign")

	return &getResponse, nil
}

func (s *EnvelopeService) ActivateEnvelope(ctx context.Context, clicksignKey string) error {
	correlationID := ctx.Value("correlation_id")

	s.logger.WithFields(logrus.Fields{
		"clicksign_key":  clicksignKey,
		"correlation_id": correlationID,
	}).Info("Activating envelope in Clicksign")

	updateRequest := dto.EnvelopeUpdateRequest{
		Status: stringPtr("running"),
	}

	endpoint := fmt.Sprintf("/api/v3/envelopes/%s", clicksignKey)
	resp, err := s.clicksignClient.Patch(ctx, endpoint, updateRequest)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"error":          err.Error(),
			"clicksign_key":  clicksignKey,
			"correlation_id": correlationID,
		}).Error("Failed to activate envelope in Clicksign")
		return fmt.Errorf("failed to activate envelope in Clicksign: %w", err)
	}
	defer resp.Body.Close()

	// Ler resposta
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"error":          err.Error(),
			"clicksign_key":  clicksignKey,
			"correlation_id": correlationID,
		}).Error("Failed to read response from Clicksign")
		return fmt.Errorf("failed to read response from Clicksign: %w", err)
	}

	// Verificar se houve erro na resposta
	if resp.StatusCode >= 400 {
		var errorResp dto.ClicksignErrorResponse
		if err := json.Unmarshal(body, &errorResp); err != nil {
			s.logger.WithFields(logrus.Fields{
				"error":          err.Error(),
				"status_code":    resp.StatusCode,
				"response_body":  string(body),
				"correlation_id": correlationID,
			}).Error("Failed to parse error response from Clicksign")
			return fmt.Errorf("Clicksign API error (status %d): %s", resp.StatusCode, string(body))
		}

		s.logger.WithFields(logrus.Fields{
			"error_type":     errorResp.Error.Type,
			"error_message":  errorResp.Error.Message,
			"error_code":     errorResp.Error.Code,
			"status_code":    resp.StatusCode,
			"correlation_id": correlationID,
		}).Error("Clicksign API returned error")

		return fmt.Errorf("Clicksign API error: %s - %s", errorResp.Error.Type, errorResp.Error.Message)
	}

	s.logger.WithFields(logrus.Fields{
		"clicksign_key":  clicksignKey,
		"correlation_id": correlationID,
	}).Info("Envelope activated successfully in Clicksign")

	return nil
}

func (s *EnvelopeService) mapEntityToCreateRequest(envelope *entity.EntityEnvelope) *dto.EnvelopeCreateRequestWrapper {
	req := &dto.EnvelopeCreateRequestWrapper{
		Data: dto.EnvelopeCreateData{
			Type: "envelopes",
			Attributes: dto.EnvelopeCreateAttributes{
				Name:              envelope.Name,
				Locale:            "pt-BR",
				AutoClose:         envelope.AutoClose,
				RemindInterval:    envelope.RemindInterval,
				BlockAfterRefusal: true,
				DeadlineAt:        envelope.DeadlineAt,
			},
		},
	}

	if envelope.Message != "" {
		req.Data.Attributes.DefaultSubject = envelope.Message
	}

	return req
}

func stringPtr(s string) *string {
	return &s
}
