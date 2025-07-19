package clicksign

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"app/infrastructure/clicksign/dto"
	"app/usecase/clicksign"
	"github.com/sirupsen/logrus"
)

type SignerService struct {
	clicksignClient clicksign.ClicksignClientInterface
	logger          *logrus.Logger
}

func NewSignerService(clicksignClient clicksign.ClicksignClientInterface, logger *logrus.Logger) *SignerService {
	return &SignerService{
		clicksignClient: clicksignClient,
		logger:          logger,
	}
}

// CreateSigner cria um signatário no envelope usando a estrutura JSON API conforme Postman Collection
func (s *SignerService) CreateSigner(ctx context.Context, envelopeID string, signerData SignerData) (string, error) {
	correlationID := ctx.Value("correlation_id")

	s.logger.WithFields(logrus.Fields{
		"envelope_id":    envelopeID,
		"signer_name":    signerData.Name,
		"signer_email":   signerData.Email,
		"correlation_id": correlationID,
	}).Info("Creating signer in envelope using JSON API format")

	createRequest := s.mapSignerDataToCreateRequest(signerData)

	s.logger.WithFields(logrus.Fields{
		"data_type":           createRequest.Data.Type,
		"signer_name":         createRequest.Data.Attributes.Name,
		"signer_email":        createRequest.Data.Attributes.Email,
		"has_documentation":   createRequest.Data.Attributes.HasDocumentation,
		"refusable":          createRequest.Data.Attributes.Refusable,
		"group":              createRequest.Data.Attributes.Group,
		"correlation_id":     correlationID,
	}).Debug("JSON API request structure prepared for signer")

	// Fazer chamada para API do Clicksign usando o endpoint correto para signatários
	endpoint := fmt.Sprintf("/api/v3/envelopes/%s/signers", envelopeID)
	resp, err := s.clicksignClient.Post(ctx, endpoint, createRequest)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"error":          err.Error(),
			"envelope_id":    envelopeID,
			"signer_email":   signerData.Email,
			"correlation_id": correlationID,
		}).Error("Failed to create signer in Clicksign envelope")
		return "", fmt.Errorf("failed to create signer in Clicksign envelope: %w", err)
	}
	defer resp.Body.Close()

	// Ler resposta
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"error":          err.Error(),
			"envelope_id":    envelopeID,
			"signer_email":   signerData.Email,
			"correlation_id": correlationID,
		}).Error("Failed to read response from Clicksign")
		return "", fmt.Errorf("failed to read response from Clicksign: %w", err)
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
			return "", fmt.Errorf("Clicksign API error (status %d): %s", resp.StatusCode, string(body))
		}

		s.logger.WithFields(logrus.Fields{
			"error_type":     errorResp.Error.Type,
			"error_message":  errorResp.Error.Message,
			"error_code":     errorResp.Error.Code,
			"status_code":    resp.StatusCode,
			"correlation_id": correlationID,
		}).Error("Clicksign API returned error")

		return "", fmt.Errorf("Clicksign API error: %s - %s", errorResp.Error.Type, errorResp.Error.Message)
	}

	// Fazer parse da resposta de sucesso usando estrutura JSON API
	var createResponse dto.SignerCreateResponseWrapper
	if err := json.Unmarshal(body, &createResponse); err != nil {
		s.logger.WithFields(logrus.Fields{
			"error":          err.Error(),
			"response_body":  string(body),
			"expected_format": "JSON API (data.type.attributes)",
			"correlation_id": correlationID,
		}).Error("Failed to parse JSON API response from Clicksign")
		return "", fmt.Errorf("failed to parse JSON API response from Clicksign: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"envelope_id":       envelopeID,
		"signer_id":        createResponse.Data.ID,
		"signer_name":      createResponse.Data.Attributes.Name,
		"signer_email":     createResponse.Data.Attributes.Email,
		"response_type":    createResponse.Data.Type,
		"correlation_id":   correlationID,
	}).Info("Signer created successfully in Clicksign envelope using JSON API format")

	return createResponse.Data.ID, nil
}

// mapSignerDataToCreateRequest mapeia os dados do signatário para a estrutura JSON API
func (s *SignerService) mapSignerDataToCreateRequest(signerData SignerData) *dto.SignerCreateRequestWrapper {
	req := &dto.SignerCreateRequestWrapper{
		Data: dto.SignerCreateData{
			Type: "signers",
			Attributes: dto.SignerCreateAttributes{
				Name:              signerData.Name,
				Email:             signerData.Email,
				Birthday:          signerData.Birthday,
				PhoneNumber:       signerData.PhoneNumber,
				HasDocumentation:  signerData.HasDocumentation,
				Refusable:         signerData.Refusable,
				Group:             signerData.Group,
			},
		},
	}

	// Configurar eventos de comunicação se fornecidos
	if signerData.CommunicateEvents != nil {
		req.Data.Attributes.CommunicateEvents = &dto.SignerCommunicateEvents{
			DocumentSigned:    signerData.CommunicateEvents.DocumentSigned,
			SignatureRequest:  signerData.CommunicateEvents.SignatureRequest,
			SignatureReminder: signerData.CommunicateEvents.SignatureReminder,
		}
	}

	return req
}

// SignerData representa os dados necessários para criar um signatário
type SignerData struct {
	Name              string
	Email             string
	Birthday          string
	PhoneNumber       *string
	HasDocumentation  bool
	Refusable         bool
	Group             int
	CommunicateEvents *SignerCommunicateEventsData
}

// SignerCommunicateEventsData representa as configurações de comunicação
type SignerCommunicateEventsData struct {
	DocumentSigned    string
	SignatureRequest  string
	SignatureReminder string
}