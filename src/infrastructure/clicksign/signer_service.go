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

	createRequest := s.mapSignerDataToCreateRequest(signerData)

	// Fazer chamada para API do Clicksign usando o endpoint correto para signatários
	endpoint := fmt.Sprintf("/api/v3/envelopes/%s/signers", envelopeID)
	resp, err := s.clicksignClient.Post(ctx, endpoint, createRequest)
	if err != nil {
		return "", fmt.Errorf("failed to create signer in Clicksign envelope: %w", err)
	}
	defer resp.Body.Close()

	// Ler resposta
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response from Clicksign: %w", err)
	}

	// Verificar se houve erro na resposta
	if resp.StatusCode >= 400 {
		var errorResp dto.ClicksignErrorResponse
		if err := json.Unmarshal(body, &errorResp); err != nil {
			return "", fmt.Errorf("Clicksign API error (status %d): %s", resp.StatusCode, string(body))
		}

		return "", fmt.Errorf("Clicksign API error: %s - %s", errorResp.Error.Type, errorResp.Error.Message)
	}

	// Fazer parse da resposta de sucesso usando estrutura JSON API
	var createResponse dto.SignerCreateResponseWrapper
	if err := json.Unmarshal(body, &createResponse); err != nil {
		return "", fmt.Errorf("failed to parse JSON API response from Clicksign: %w", err)
	}

	return createResponse.Data.ID, nil
}

// DeleteSigner deleta um signatário do envelope no Clicksign
func (s *SignerService) DeleteSigner(ctx context.Context, envelopeID string, signerID string) error {
	// Fazer chamada para API do Clicksign usando o endpoint correto para deletar signatário
	endpoint := fmt.Sprintf("/api/v3/envelopes/%s/signers/%s", envelopeID, signerID)
	resp, err := s.clicksignClient.Delete(ctx, endpoint)
	if err != nil {
		return fmt.Errorf("failed to delete signer from Clicksign envelope: %w", err)
	}
	defer resp.Body.Close()

	// Verificar se houve erro na resposta
	if resp.StatusCode >= 400 {
		// Ler resposta para obter detalhes do erro
		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return fmt.Errorf("Clicksign API error (status %d): failed to read response", resp.StatusCode)
		}

		var errorResp dto.ClicksignErrorResponse
		if err := json.Unmarshal(body, &errorResp); err != nil {
			return fmt.Errorf("Clicksign API error (status %d): %s", resp.StatusCode, string(body))
		}

		return fmt.Errorf("Clicksign API error: %s - %s", errorResp.Error.Type, errorResp.Error.Message)
	}

	return nil
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