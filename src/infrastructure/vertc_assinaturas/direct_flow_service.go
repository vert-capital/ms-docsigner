package vertc_assinaturas

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"app/entity"
	"app/infrastructure/provider"
	"app/pkg/utils"

	"github.com/sirupsen/logrus"
)

type directEnvelopeCreateRequest struct {
	Name     string `json:"name"`
	ExpireIn string `json:"expireIn,omitempty"`
}

type directEnvelopeResponse struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

type directSignerRequest struct {
	EnvelopeID      string   `json:"envelopeId"`
	Email           string   `json:"email"`
	Name            string   `json:"name"`
	IsRequired      bool     `json:"isRequired"`
	RequiredMethods []string `json:"requiredMethods,omitempty"`
}

type directSendRequest struct {
	Subject string `json:"subject,omitempty"`
	Message string `json:"message,omitempty"`
}

// DirectFlowExecutionResult representa o resultado do fluxo direto de criação/envio.
type DirectFlowExecutionResult struct {
	Mode                  string `json:"mode"`
	EnvelopeID            string `json:"envelope_id"`
	EnvelopeStatus        string `json:"envelope_status,omitempty"`
	DocumentsUploaded     int    `json:"documents_uploaded"`
	SignersCreated        int    `json:"signers_created"`
	NotificationTriggered bool   `json:"notification_triggered"`
}

// DirectFlowService executa o fluxo "create envelope -> upload docs -> create signers -> send".
type DirectFlowService struct {
	client *VertcAssinaturasClient
	logger *logrus.Logger
}

// NewDirectFlowService cria uma nova instância do serviço de fluxo direto.
func NewDirectFlowService(client *VertcAssinaturasClient, logger *logrus.Logger) *DirectFlowService {
	return &DirectFlowService{
		client: client,
		logger: logger,
	}
}

// CreateEnvelopeWithDocumentsAndSigners cria e envia um envelope sem usar quick-send.
func (s *DirectFlowService) CreateEnvelopeWithDocumentsAndSigners(
	ctx context.Context,
	data QuickSendData,
) (*DirectFlowExecutionResult, error) {
	if data.Envelope == nil {
		return nil, fmt.Errorf("envelope data is required for direct flow")
	}

	if len(data.Documents) == 0 {
		return nil, fmt.Errorf("at least one document is required for direct flow")
	}

	if len(data.Signers) == 0 {
		return nil, fmt.Errorf("at least one signer is required for direct flow")
	}

	envelopeResp, err := s.createEnvelope(ctx, data.Envelope)
	if err != nil {
		return nil, err
	}

	for _, document := range data.Documents {
		if err := s.uploadDocument(ctx, envelopeResp.ID, document); err != nil {
			return nil, err
		}
	}

	signersCreated, err := s.createSigners(ctx, envelopeResp.ID, data.Signers)
	if err != nil {
		return nil, err
	}

	if err := s.sendEnvelope(ctx, envelopeResp.ID, data.Envelope); err != nil {
		return nil, err
	}

	return &DirectFlowExecutionResult{
		Mode:                  "direct",
		EnvelopeID:            envelopeResp.ID,
		EnvelopeStatus:        "sent",
		DocumentsUploaded:     len(data.Documents),
		SignersCreated:        signersCreated,
		NotificationTriggered: true,
	}, nil
}

func (s *DirectFlowService) createEnvelope(ctx context.Context, envelope *entity.EntityEnvelope) (*directEnvelopeResponse, error) {
	request := directEnvelopeCreateRequest{
		Name: envelope.Name,
	}

	if envelope.DeadlineAt != nil {
		request.ExpireIn = envelope.DeadlineAt.Format(time.RFC3339)
	}

	resp, err := s.client.Post(ctx, "/api/v1/envelopes", request, "")
	if err != nil {
		return nil, fmt.Errorf("failed to create vert-sign envelope: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read vert-sign envelope creation response: %w", err)
	}

	var envelopeResp directEnvelopeResponse
	if err := json.Unmarshal(body, &envelopeResp); err != nil {
		return nil, fmt.Errorf("failed to parse vert-sign envelope creation response: %w", err)
	}

	if envelopeResp.ID == "" {
		return nil, fmt.Errorf("vert-sign envelope creation response does not contain id")
	}

	return &envelopeResp, nil
}

func (s *DirectFlowService) uploadDocument(ctx context.Context, envelopeID string, document *entity.EntityDocument) error {
	fileContent, err := s.readDocumentContent(document)
	if err != nil {
		return fmt.Errorf("failed to read document '%s' for direct flow upload: %w", document.Name, err)
	}

	endpoint := fmt.Sprintf("/api/v1/documents/%s", envelopeID)
	resp, err := s.client.PostMultipartFile(ctx, endpoint, "files", document.Name, fileContent)
	if err != nil {
		return fmt.Errorf("failed to upload document '%s' to vert-sign: %w", document.Name, err)
	}
	defer resp.Body.Close()

	return nil
}

func (s *DirectFlowService) createSigners(ctx context.Context, envelopeID string, signers []provider.SignerData) (int, error) {
	requests := make([]directSignerRequest, 0, len(signers))

	for _, signer := range signers {
		authMethod := strings.TrimSpace(signer.AuthMethod)
		if authMethod == "" {
			authMethod = "email"
		}

		req := directSignerRequest{
			EnvelopeID: envelopeID,
			Email:      signer.Email,
			Name:       signer.Name,
			IsRequired: !signer.Refusable,
		}

		switch authMethod {
		case "email":
			req.RequiredMethods = []string{"code_email"}
		case "auto_signature":
			req.RequiredMethods = []string{"automatic_signature"}
		default:
			return 0, fmt.Errorf("unsupported auth method '%s' for vert-sign direct flow", authMethod)
		}

		requests = append(requests, req)
	}

	resp, err := s.client.Post(ctx, "/api/v1/signers", requests, "")
	if err != nil {
		return 0, fmt.Errorf("failed to create signers in vert-sign: %w", err)
	}
	defer resp.Body.Close()

	return len(requests), nil
}

func (s *DirectFlowService) sendEnvelope(ctx context.Context, envelopeID string, envelope *entity.EntityEnvelope) error {
	message := strings.TrimSpace(envelope.Description)
	if message == "" {
		message = strings.TrimSpace(envelope.Message)
	}

	request := directSendRequest{
		Subject: strings.TrimSpace(envelope.Name),
		Message: message,
	}

	endpoint := fmt.Sprintf("/api/v1/envelopes/%s/send", envelopeID)
	resp, err := s.client.Post(ctx, endpoint, request, "")
	if err != nil {
		return fmt.Errorf("failed to send vert-sign envelope: %w", err)
	}
	defer resp.Body.Close()

	return nil
}

func (s *DirectFlowService) readDocumentContent(document *entity.EntityDocument) ([]byte, error) {
	if strings.HasPrefix(document.FilePath, "http://") || strings.HasPrefix(document.FilePath, "https://") {
		fileInfo, err := utils.DownloadFileFromURL(document.FilePath)
		if err != nil {
			return nil, err
		}

		defer func() {
			if cleanupErr := utils.CleanupTempFile(fileInfo.TempPath); cleanupErr != nil {
				s.logger.WithError(cleanupErr).WithField("document_name", document.Name).Warn("Failed to cleanup downloaded URL temp file")
			}
		}()

		return fileInfo.DecodedData, nil
	}

	fileContent, err := os.ReadFile(document.FilePath)
	if err != nil {
		return nil, err
	}

	return fileContent, nil
}
