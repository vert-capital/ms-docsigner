package vertc_assinaturas

import (
	"app/entity"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// QuickSendRequest representa a requisição para o endpoint quick-send
type QuickSendRequest struct {
	Envelope  EnvelopeData  `json:"envelope"`
	Signers   []SignerDto   `json:"signers"`
	Documents []DocumentDto `json:"documents"`
}

// EnvelopeData representa os dados do envelope para quick-send
type EnvelopeData struct {
	Name     string `json:"name"`
	Subject  string `json:"subject"`
	Message  string `json:"message"`
	ExpireIn string `json:"expireIn"` // Formato: "YYYY-MM-DD HH:mm"
}

// SignerDto representa um signatário para quick-send
type SignerDto struct {
	Email           string   `json:"email"`
	Name            string   `json:"name"`
	IsRequired      bool     `json:"isRequired"`
	RequiredMethods []string `json:"requiredMethods"`
}

// DocumentDto representa um documento para quick-send
type DocumentDto struct {
	URL        string `json:"url"`
	Name       string `json:"name"`
	MimeType   string `json:"mimeType"`
	SplitPages bool   `json:"splitPages"`
}

// QuickSendResponse representa a resposta do endpoint quick-send
type QuickSendResponse struct {
	Status            string                      `json:"status"`
	Message           string                      `json:"message"`
	EnvelopeID        string                      `json:"envelopeId"`
	Documents         []interface{}               `json:"documents"`
	Signers           []interface{}               `json:"signers"`
	ProviderDocuments []QuickSendProviderDocument `json:"provider_documents,omitempty"`
}

type QuickSendProviderDocument struct {
	BackendDocumentID int    `json:"backend_document_id"`
	ProviderDocumentID string `json:"provider_document_id"`
	Name              string `json:"name,omitempty"`
}

// QuickSendService gerencia operações relacionadas ao quick-send
type QuickSendService struct {
	client *VertcAssinaturasClient
	logger *logrus.Logger
}

// NewQuickSendService cria uma nova instância do QuickSendService
func NewQuickSendService(client *VertcAssinaturasClient, logger *logrus.Logger) *QuickSendService {
	return &QuickSendService{
		client: client,
		logger: logger,
	}
}

// QuickSend executa o quick-send com os dados fornecidos
func (s *QuickSendService) QuickSend(ctx context.Context, data QuickSendData) (*QuickSendResponse, error) {
	// Mapear dados para formato quick-send
	request, err := s.mapToQuickSendRequest(data)
	if err != nil {
		return nil, fmt.Errorf("failed to map data to quick-send request: %w", err)
	}

	// Gerar idempotency key
	idempotencyKey := s.generateIdempotencyKey()

	// Fazer requisição
	resp, err := s.client.Post(ctx, "/api/v1/envelopes/quick-send", request, idempotencyKey)
	if err != nil {
		return nil, fmt.Errorf("failed to call quick-send endpoint: %w", err)
	}
	defer resp.Body.Close()

	// Ler resposta
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read quick-send response: %w", err)
	}
	s.logger.Infof(
		"[VERT_SIGN_CREATE_ENVELOPE][QUICK_SEND_RESPONSE] status=%d body=%s",
		resp.StatusCode,
		string(body),
	)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, &VertcAssinaturasError{
			Type:       ErrorTypeClient,
			Message:    fmt.Sprintf("quick-send failed with status %d: %s", resp.StatusCode, string(body)),
			StatusCode: resp.StatusCode,
		}
	}

	var quickSendResp QuickSendResponse
	if err := json.Unmarshal(body, &quickSendResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal quick-send response: %w", err)
	}
	quickSendResp.ProviderDocuments = s.buildProviderDocumentMappings(quickSendResp.Documents, data.Documents)

	s.logger.Debugf("Quick-send realizado com sucesso. Envelope ID: %s", quickSendResp.EnvelopeID)
	return &quickSendResp, nil
}

func (s *QuickSendService) buildProviderDocumentMappings(
	providerDocuments []interface{},
	requestDocuments []*entity.EntityDocument,
) []QuickSendProviderDocument {
	mappings := make([]QuickSendProviderDocument, 0)

	for index, providerDoc := range providerDocuments {
		if index >= len(requestDocuments) {
			break
		}

		providerDocMap, ok := providerDoc.(map[string]interface{})
		if !ok {
			continue
		}

		providerDocumentID, _ := providerDocMap["id"].(string)
		if providerDocumentID == "" {
			continue
		}

		backendDocumentID := s.extractBackendDocumentIDFromMetadata(requestDocuments[index].Metadata)
		if backendDocumentID <= 0 {
			continue
		}

		name, _ := providerDocMap["name"].(string)
		mappings = append(mappings, QuickSendProviderDocument{
			BackendDocumentID: backendDocumentID,
			ProviderDocumentID: providerDocumentID,
			Name:              name,
		})
	}

	return mappings
}

func (s *QuickSendService) extractBackendDocumentIDFromMetadata(metadata []byte) int {
	if len(metadata) == 0 {
		return 0
	}

	var metadataMap map[string]interface{}
	if err := json.Unmarshal(metadata, &metadataMap); err != nil {
		return 0
	}

	value, exists := metadataMap["backend_document_id"]
	if !exists {
		return 0
	}

	switch v := value.(type) {
	case float64:
		return int(v)
	case int:
		return v
	default:
		return 0
	}
}

// mapToQuickSendRequest mapeia os dados internos para o formato quick-send
func (s *QuickSendService) mapToQuickSendRequest(data QuickSendData) (*QuickSendRequest, error) {
	s.logger.Debugf("Mapeando dados para quick-send. Envelope: %s, Documentos: %d, Signatários: %d",
		data.Envelope.Name, len(data.Documents), len(data.Signers))

	// Mapear envelope
	envelopeData := EnvelopeData{
		Name:    data.Envelope.Name,
		Subject: data.Envelope.Name,
		Message: data.Envelope.Description,
	}

	// Se message estiver vazio, usar description
	if envelopeData.Message == "" {
		envelopeData.Message = data.Envelope.Message
	}

	// Converter deadline_at para expireIn
	if data.Envelope.DeadlineAt != nil {
		envelopeData.ExpireIn = data.Envelope.DeadlineAt.Format("2006-01-02 15:04")
	} else {
		// Se não houver deadline, usar 30 dias a partir de agora
		futureDate := time.Now().Add(30 * 24 * time.Hour)
		envelopeData.ExpireIn = futureDate.Format("2006-01-02 15:04")
	}

	// Mapear signatários
	signers := make([]SignerDto, 0, len(data.Signers))
	for _, signer := range data.Signers {
		isRequired := !signer.Refusable // Inverter lógica: refusable=false significa required=true

		// Mapear auth_method para requiredMethods
		// "email" -> "code_email" (padrão do vertc-assinaturas)
		requiredMethods := []string{"code_email"} // Valor padrão
		if signer.AuthMethod != "" {
			requiredMethods = s.mapAuthMethodToRequiredMethods(signer.AuthMethod)
		}

		signerDto := SignerDto{
			Email:           signer.Email,
			Name:            signer.Name,
			IsRequired:      isRequired,
			RequiredMethods: requiredMethods,
		}
		signers = append(signers, signerDto)
	}

	// Mapear documentos
	documents := make([]DocumentDto, 0, len(data.Documents))
	for _, doc := range data.Documents {
		// Para vertc-assinaturas, precisamos da URL do documento
		// Se o documento tem FilePath que é uma URL, usar diretamente
		// Caso contrário, precisamos de uma URL pública
		// Por enquanto, vamos assumir que FilePath contém a URL
		docURL := doc.FilePath

		// Validar que é uma URL válida
		if !isValidURL(docURL) {
			return nil, fmt.Errorf("document '%s' does not have a valid URL. FilePath: %s", doc.Name, docURL)
		}

		docDto := DocumentDto{
			URL:        docURL,
			Name:       doc.Name,
			MimeType:   doc.MimeType,
			SplitPages: false, // Valor padrão
		}
		documents = append(documents, docDto)
	}

	return &QuickSendRequest{
		Envelope:  envelopeData,
		Signers:   signers,
		Documents: documents,
	}, nil
}

// generateIdempotencyKey gera uma chave única de idempotência
func (s *QuickSendService) generateIdempotencyKey() string {
	return uuid.New().String()
}

// isValidURL verifica se uma string é uma URL válida
func isValidURL(url string) bool {
	return len(url) >= 4 && (url[:4] == "http" || (len(url) >= 5 && url[:5] == "https"))
}

// mapAuthMethodToRequiredMethods mapeia o método de autenticação do gerador-documentos
// para o formato requiredMethods do vertc-assinaturas
// "email" -> ["code_email"]
func (s *QuickSendService) mapAuthMethodToRequiredMethods(authMethod string) []string {
	switch authMethod {
	case "email":
		return []string{"code_email"}
	case "auto_signature":
		return []string{"automatic_signature"}
	// Futuros métodos podem ser adicionados aqui:
	// case "icp_brasil":
	//     return []string{"certificate"}
	default:
		// Se não reconhecido, usar padrão
		s.logger.Warnf("Auth method '%s' não reconhecido, usando padrão 'code_email'", authMethod)
		return []string{"code_email"}
	}
}
