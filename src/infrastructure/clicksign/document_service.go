package clicksign

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"app/entity"
	"app/infrastructure/clicksign/dto"
	"app/pkg/utils"
	"app/usecase/clicksign"
	"github.com/sirupsen/logrus"
)

type DocumentService struct {
	clicksignClient clicksign.ClicksignClientInterface
	logger          *logrus.Logger
}

func NewDocumentService(clicksignClient clicksign.ClicksignClientInterface, logger *logrus.Logger) *DocumentService {
	return &DocumentService{
		clicksignClient: clicksignClient,
		logger:          logger,
	}
}

// UploadDocument faz upload de um documento para o Clicksign
func (s *DocumentService) UploadDocument(ctx context.Context, document *entity.EntityDocument) (string, error) {
	correlationID := ctx.Value("correlation_id")

	s.logger.WithFields(logrus.Fields{
		"document_id":    document.ID,
		"document_name":  document.Name,
		"is_from_base64": document.IsFromBase64,
		"correlation_id": correlationID,
	}).Info("Starting document upload to Clicksign")

	var uploadRequest *dto.DocumentUploadRequestWrapper
	var err error

	if document.IsFromBase64 {
		// Documento veio de base64, usar conteúdo base64 diretamente
		uploadRequest, err = s.prepareBase64Upload(document)
		if err != nil {
			s.logger.WithFields(logrus.Fields{
				"error":          err.Error(),
				"document_id":    document.ID,
				"correlation_id": correlationID,
			}).Error("Failed to prepare base64 upload")
			return "", fmt.Errorf("failed to prepare base64 upload: %w", err)
		}
	} else {
		// Documento veio de file_path, ler arquivo e converter para base64
		uploadRequest, err = s.prepareFilePathUpload(document)
		if err != nil {
			s.logger.WithFields(logrus.Fields{
				"error":          err.Error(),
				"document_id":    document.ID,
				"file_path":      document.FilePath,
				"correlation_id": correlationID,
			}).Error("Failed to prepare file path upload")
			return "", fmt.Errorf("failed to prepare file path upload: %w", err)
		}
	}

	// Fazer upload para Clicksign
	resp, err := s.clicksignClient.Post(ctx, "/api/v3/documents", uploadRequest)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"error":          err.Error(),
			"document_id":    document.ID,
			"correlation_id": correlationID,
		}).Error("Failed to upload document to Clicksign")
		return "", fmt.Errorf("failed to upload document to Clicksign: %w", err)
	}
	defer resp.Body.Close()

	// Ler resposta
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"error":          err.Error(),
			"document_id":    document.ID,
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

	// Fazer parse da resposta de sucesso
	var uploadResponse dto.DocumentUploadResponseWrapper
	if err := json.Unmarshal(body, &uploadResponse); err != nil {
		s.logger.WithFields(logrus.Fields{
			"error":          err.Error(),
			"response_body":  string(body),
			"correlation_id": correlationID,
		}).Error("Failed to parse upload response from Clicksign")
		return "", fmt.Errorf("failed to parse upload response from Clicksign: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"document_id":        document.ID,
		"clicksign_doc_id":   uploadResponse.Data.ID,
		"uploaded_filename":  uploadResponse.Data.Attributes.Filename,
		"uploaded_filesize":  uploadResponse.Data.Attributes.Filesize,
		"correlation_id":     correlationID,
	}).Info("Document uploaded successfully to Clicksign")

	return uploadResponse.Data.ID, nil
}

// CreateDocument cria um documento no envelope usando a estrutura JSON API correta
func (s *DocumentService) CreateDocument(ctx context.Context, envelopeID string, document *entity.EntityDocument) (string, error) {
	correlationID := ctx.Value("correlation_id")

	s.logger.WithFields(logrus.Fields{
		"envelope_id":    envelopeID,
		"document_id":    document.ID,
		"document_name":  document.Name,
		"is_from_base64": document.IsFromBase64,
		"correlation_id": correlationID,
	}).Info("Creating document in envelope using JSON API format")

	var createRequest *dto.DocumentCreateRequestWrapper
	var err error

	if document.IsFromBase64 {
		createRequest, err = s.prepareBase64CreateRequest(document)
		if err != nil {
			s.logger.WithFields(logrus.Fields{
				"error":          err.Error(),
				"document_id":    document.ID,
				"correlation_id": correlationID,
			}).Error("Failed to prepare base64 create request")
			return "", fmt.Errorf("failed to prepare base64 create request: %w", err)
		}
	} else {
		createRequest, err = s.prepareFilePathCreateRequest(document)
		if err != nil {
			s.logger.WithFields(logrus.Fields{
				"error":          err.Error(),
				"document_id":    document.ID,
				"file_path":      document.FilePath,
				"correlation_id": correlationID,
			}).Error("Failed to prepare file path create request")
			return "", fmt.Errorf("failed to prepare file path create request: %w", err)
		}
	}

	// Fazer chamada para API do Clicksign usando o endpoint correto para documentos
	endpoint := fmt.Sprintf("/api/v3/envelopes/%s/documents", envelopeID)
	resp, err := s.clicksignClient.Post(ctx, endpoint, createRequest)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"error":          err.Error(),
			"envelope_id":    envelopeID,
			"document_id":    document.ID,
			"correlation_id": correlationID,
		}).Error("Failed to create document in Clicksign envelope")
		return "", fmt.Errorf("failed to create document in Clicksign envelope: %w", err)
	}
	defer resp.Body.Close()

	// Ler resposta
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"error":          err.Error(),
			"envelope_id":    envelopeID,
			"document_id":    document.ID,
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
	var createResponse dto.DocumentCreateResponseWrapper
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
		"document_id":       document.ID,
		"clicksign_doc_id":  createResponse.Data.ID,
		"created_filename":  createResponse.Data.Attributes.Filename,
		"created_filesize":  createResponse.Data.Attributes.Filesize,
		"response_type":     createResponse.Data.Type,
		"correlation_id":    correlationID,
	}).Info("Document created successfully in Clicksign envelope using JSON API format")

	return createResponse.Data.ID, nil
}

// prepareBase64CreateRequest prepara a requisição de criação de documento que veio de base64
func (s *DocumentService) prepareBase64CreateRequest(document *entity.EntityDocument) (*dto.DocumentCreateRequestWrapper, error) {
	// Ler arquivo temporário e converter para base64
	fileData, err := os.ReadFile(document.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read temporary file: %w", err)
	}

	base64Content := base64.StdEncoding.EncodeToString(fileData)
	filename := s.generateFilename(document)

	createRequest := &dto.DocumentCreateRequestWrapper{
		Data: dto.DocumentCreateData{
			Type: "documents",
			Attributes: dto.DocumentCreateAttributes{
				Filename:      filename,
				ContentBase64: base64Content,
				Metadata: &dto.DocumentMetadata{
					Type: "private",
					ID:   int(document.ID),
					User: 1, // TODO: Mapear user correto quando tivermos contexto de usuário
				},
			},
		},
	}

	return createRequest, nil
}

// prepareFilePathCreateRequest prepara a requisição de criação de documento que veio de file_path
func (s *DocumentService) prepareFilePathCreateRequest(document *entity.EntityDocument) (*dto.DocumentCreateRequestWrapper, error) {
	// Ler arquivo do sistema e converter para base64
	fileData, err := os.ReadFile(document.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file from path: %w", err)
	}

	base64Content := base64.StdEncoding.EncodeToString(fileData)
	filename := filepath.Base(document.FilePath)

	createRequest := &dto.DocumentCreateRequestWrapper{
		Data: dto.DocumentCreateData{
			Type: "documents",
			Attributes: dto.DocumentCreateAttributes{
				Filename:      filename,
				ContentBase64: base64Content,
				Metadata: &dto.DocumentMetadata{
					Type: "private",
					ID:   int(document.ID),
					User: 1, // TODO: Mapear user correto quando tivermos contexto de usuário
				},
			},
		},
	}

	return createRequest, nil
}

// prepareBase64Upload prepara o upload de documento que veio de base64
func (s *DocumentService) prepareBase64Upload(document *entity.EntityDocument) (*dto.DocumentUploadRequestWrapper, error) {
	// Ler arquivo temporário e converter para base64
	fileData, err := os.ReadFile(document.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read temporary file: %w", err)
	}

	base64Content := base64.StdEncoding.EncodeToString(fileData)
	filename := s.generateFilename(document)

	uploadRequest := &dto.DocumentUploadRequestWrapper{
		Data: dto.DocumentUploadData{
			Type: "documents",
			Attributes: dto.DocumentUploadAttributes{
				ContentBase64: base64Content,
				Filename:      filename,
			},
		},
	}

	return uploadRequest, nil
}

// prepareFilePathUpload prepara o upload de documento que veio de file_path
func (s *DocumentService) prepareFilePathUpload(document *entity.EntityDocument) (*dto.DocumentUploadRequestWrapper, error) {
	// Ler arquivo do sistema e converter para base64
	fileData, err := os.ReadFile(document.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file from path: %w", err)
	}

	base64Content := base64.StdEncoding.EncodeToString(fileData)
	filename := filepath.Base(document.FilePath)

	uploadRequest := &dto.DocumentUploadRequestWrapper{
		Data: dto.DocumentUploadData{
			Type: "documents",
			Attributes: dto.DocumentUploadAttributes{
				ContentBase64: base64Content,
				Filename:      filename,
			},
		},
	}

	return uploadRequest, nil
}

// generateFilename gera um nome de arquivo baseado no documento e extensão do MIME type
func (s *DocumentService) generateFilename(document *entity.EntityDocument) string {
	extension := utils.GetFileExtensionFromMimeType(document.MimeType)
	return fmt.Sprintf("%s_%d%s", document.Name, document.ID, extension)
}