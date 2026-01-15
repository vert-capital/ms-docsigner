package clicksign

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"app/entity"
	"app/infrastructure/clicksign/dto"
	"app/pkg/utils"

	"github.com/sirupsen/logrus"
)

type DocumentService struct {
	clicksignClient ClicksignClientInterface
	logger          *logrus.Logger
}

func NewDocumentService(clicksignClient ClicksignClientInterface, logger *logrus.Logger) *DocumentService {
	return &DocumentService{
		clicksignClient: clicksignClient,
		logger:          logger,
	}
}

// UploadDocument faz upload de um documento para o Clicksign
func (s *DocumentService) UploadDocument(ctx context.Context, document *entity.EntityDocument) (string, error) {

	var uploadRequest *dto.DocumentUploadRequestWrapper
	var err error

	if document.IsFromBase64 {
		// Documento veio de base64, usar conteúdo base64 diretamente
		uploadRequest, err = s.prepareBase64Upload(document)
		if err != nil {
			return "", fmt.Errorf("failed to prepare base64 upload: %w", err)
		}
	} else {
		// Documento veio de file_path, ler arquivo e converter para base64
		uploadRequest, err = s.prepareFilePathUpload(document)
		if err != nil {
			return "", fmt.Errorf("failed to prepare file path upload: %w", err)
		}
	}

	// Fazer upload para Clicksign
	resp, err := s.clicksignClient.Post(ctx, "/api/v3/documents", uploadRequest)
	if err != nil {
		return "", fmt.Errorf("failed to upload document to Clicksign: %w", err)
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

		if errorResp.Error.Type == "" && errorResp.Error.Message == "" {
			return "", fmt.Errorf("Clicksign API error (status %d): %s", resp.StatusCode, string(body))
		}
		return "", fmt.Errorf("Clicksign API error: %s - %s", errorResp.Error.Type, errorResp.Error.Message)
	}

	// Fazer parse da resposta de sucesso
	var uploadResponse dto.DocumentUploadResponseWrapper
	if err := json.Unmarshal(body, &uploadResponse); err != nil {
		return "", fmt.Errorf("failed to parse upload response from Clicksign: %w", err)
	}

	return uploadResponse.Data.ID, nil
}

// CreateDocument cria um documento no envelope usando a estrutura JSON API correta
func (s *DocumentService) CreateDocument(ctx context.Context, envelopeID string, document *entity.EntityDocument, internalEnvelopeID int) (string, error) {

	var createRequest *dto.DocumentCreateRequestWrapper
	var err error

	if document.IsFromBase64 {
		createRequest, err = s.prepareBase64CreateRequest(document, internalEnvelopeID)
		if err != nil {
			return "", fmt.Errorf("failed to prepare base64 create request: %w", err)
		}
	} else {
		createRequest, err = s.prepareFilePathCreateRequest(document, internalEnvelopeID)
		if err != nil {
			return "", fmt.Errorf("failed to prepare file path create request: %w", err)
		}
	}

	// Fazer chamada para API do Clicksign usando o endpoint correto para documentos
	endpoint := fmt.Sprintf("/api/v3/envelopes/%s/documents", envelopeID)
	resp, err := s.clicksignClient.Post(ctx, endpoint, createRequest)
	if err != nil {
		return "", fmt.Errorf("failed to create document in Clicksign envelope: %w", err)
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

		if errorResp.Error.Type == "" && errorResp.Error.Message == "" {
			return "", fmt.Errorf("Clicksign API error (status %d): %s", resp.StatusCode, string(body))
		}
		return "", fmt.Errorf("Clicksign API error: %s - %s", errorResp.Error.Type, errorResp.Error.Message)
	}

	// Fazer parse da resposta de sucesso usando estrutura JSON API
	var createResponse dto.DocumentCreateResponseWrapper
	if err := json.Unmarshal(body, &createResponse); err != nil {
		return "", fmt.Errorf("failed to parse JSON API response from Clicksign: %w", err)
	}

	return createResponse.Data.ID, nil
}

// prepareBase64CreateRequest prepara a requisição de criação de documento que veio de base64
func (s *DocumentService) prepareBase64CreateRequest(document *entity.EntityDocument, internalEnvelopeID int) (*dto.DocumentCreateRequestWrapper, error) {
	// Ler arquivo temporário e converter para base64
	fileData, err := os.ReadFile(document.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read temporary file: %w", err)
	}

	base64Content := s.generateDataURI(fileData, document.MimeType)
	filename := s.generateFilename(document)

	// Usar metadata customizado do backend se disponível, senão usar padrão
	var metadata *dto.DocumentMetadata
	envelopeID := internalEnvelopeID // padrão

	if len(document.Metadata) > 0 {
		// Fazer unmarshal do datatypes.JSON para extrair envelope_id
		var metadataMap map[string]interface{}
		if err := json.Unmarshal(document.Metadata, &metadataMap); err == nil {
			// Extrair envelope_id do metadata recebido do backend
			// Pode vir como float64 (JSON number), int, ou string
			if envIDFloat, ok := metadataMap["envelope_id"].(float64); ok {
				envelopeID = int(envIDFloat)
			} else if envIDInt, ok := metadataMap["envelope_id"].(int); ok {
				envelopeID = envIDInt
			} else if envIDStr, ok := metadataMap["envelope_id"].(string); ok {
				// Se vier como string, tentar converter para int
				if parsedID, err := strconv.Atoi(envIDStr); err == nil {
					envelopeID = parsedID
				}
			}
		}
	}

	metadata = &dto.DocumentMetadata{
		Type:       "private",
		ID:         int(document.ID),
		User:       1,          // TODO: Mapear user correto quando tivermos contexto de usuário
		EnvelopeID: envelopeID, // Usar envelope_id do backend se fornecido, senão usar padrão
	}

	createRequest := &dto.DocumentCreateRequestWrapper{
		Data: dto.DocumentCreateData{
			Type: "documents",
			Attributes: dto.DocumentCreateAttributes{
				Filename:      filename,
				ContentBase64: base64Content,
				Metadata:      metadata,
			},
		},
	}

	return createRequest, nil
}

// prepareFilePathCreateRequest prepara a requisição de criação de documento que veio de file_path
func (s *DocumentService) prepareFilePathCreateRequest(document *entity.EntityDocument, internalEnvelopeID int) (*dto.DocumentCreateRequestWrapper, error) {
	// Ler arquivo do sistema e converter para base64
	fileData, err := os.ReadFile(document.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file from path: %w", err)
	}

	base64Content := s.generateDataURI(fileData, document.MimeType)
	filename := filepath.Base(document.FilePath)

	// Usar metadata customizado do backend se disponível, senão usar padrão
	var metadata *dto.DocumentMetadata
	envelopeID := internalEnvelopeID // padrão

	if len(document.Metadata) > 0 {
		// Fazer unmarshal do datatypes.JSON para extrair envelope_id
		var metadataMap map[string]interface{}
		if err := json.Unmarshal(document.Metadata, &metadataMap); err == nil {
			// Extrair envelope_id do metadata recebido do backend
			// Pode vir como float64 (JSON number), int, ou string
			if envIDFloat, ok := metadataMap["envelope_id"].(float64); ok {
				envelopeID = int(envIDFloat)
			} else if envIDInt, ok := metadataMap["envelope_id"].(int); ok {
				envelopeID = envIDInt
			} else if envIDStr, ok := metadataMap["envelope_id"].(string); ok {
				// Se vier como string, tentar converter para int
				if parsedID, err := strconv.Atoi(envIDStr); err == nil {
					envelopeID = parsedID
				}
			}
		}
	}

	metadata = &dto.DocumentMetadata{
		Type:       "private",
		ID:         int(document.ID),
		User:       1,          // TODO: Mapear user correto quando tivermos contexto de usuário
		EnvelopeID: envelopeID, // Usar envelope_id do backend se fornecido, senão usar padrão
	}

	createRequest := &dto.DocumentCreateRequestWrapper{
		Data: dto.DocumentCreateData{
			Type: "documents",
			Attributes: dto.DocumentCreateAttributes{
				Filename:      filename,
				ContentBase64: base64Content,
				Metadata:      metadata,
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

	base64Content := s.generateDataURI(fileData, document.MimeType)
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

	base64Content := s.generateDataURI(fileData, document.MimeType)
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
	sanitizedName := s.sanitizeFilename(document.Name)
	return fmt.Sprintf("%s_%d%s", sanitizedName, document.ID, extension)
}

// sanitizeFilename remove caracteres especiais que não são válidos para nomes de arquivo no Clicksign
func (s *DocumentService) sanitizeFilename(filename string) string {
	// Substituir caracteres especiais problemáticos por underscores ou remover
	result := filename

	// Caracteres que devem ser substituídos por underscores
	invalidChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, char := range invalidChars {
		result = strings.ReplaceAll(result, char, "_")
	}

	// Remover múltiplos underscores consecutivos
	result = regexp.MustCompile("_+").ReplaceAllString(result, "_")

	// Remover underscores no início e fim
	result = strings.Trim(result, "_")

	// Se ficou vazio após sanitização, usar um nome padrão
	if result == "" {
		result = "documento"
	}

	return result
}

// generateDataURI gera um data URI com o prefixo correto baseado no MIME type
func (s *DocumentService) generateDataURI(fileData []byte, mimeType string) string {
	base64Data := base64.StdEncoding.EncodeToString(fileData)
	return fmt.Sprintf("data:%s;base64,%s", mimeType, base64Data)
}
