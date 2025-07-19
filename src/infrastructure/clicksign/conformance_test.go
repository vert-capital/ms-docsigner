package clicksign

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"app/entity"
	"app/infrastructure/clicksign/dto"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockClicksignClient é um mock do client Clicksign para testes
type MockClicksignClient struct {
	mock.Mock
}

func (m *MockClicksignClient) Get(ctx context.Context, endpoint string) (*http.Response, error) {
	args := m.Called(ctx, endpoint)
	return args.Get(0).(*http.Response), args.Error(1)
}

func (m *MockClicksignClient) Post(ctx context.Context, endpoint string, body interface{}) (*http.Response, error) {
	args := m.Called(ctx, endpoint, body)
	return args.Get(0).(*http.Response), args.Error(1)
}

func (m *MockClicksignClient) Put(ctx context.Context, endpoint string, body interface{}) (*http.Response, error) {
	args := m.Called(ctx, endpoint, body)
	return args.Get(0).(*http.Response), args.Error(1)
}

func (m *MockClicksignClient) Delete(ctx context.Context, endpoint string) (*http.Response, error) {
	args := m.Called(ctx, endpoint)
	return args.Get(0).(*http.Response), args.Error(1)
}

func (m *MockClicksignClient) Patch(ctx context.Context, endpoint string, body interface{}) (*http.Response, error) {
	args := m.Called(ctx, endpoint, body)
	return args.Get(0).(*http.Response), args.Error(1)
}

func TestEnvelopeCreateRequestConformance(t *testing.T) {
	t.Run("should create envelope request with correct JSON API structure", func(t *testing.T) {
		// Arrange
		mockClient := &MockClicksignClient{}
		logger := logrus.New()
		service := NewEnvelopeService(mockClient, logger)

		expectedResponse := dto.EnvelopeCreateResponseWrapper{
			Data: dto.EnvelopeCreateResponseData{
				Type: "envelopes",
				ID:   "test-envelope-id",
				Attributes: dto.EnvelopeCreateResponseAttributes{
					Name:   "Test Envelope",
					Status: "draft",
					Locale: "pt-BR",
				},
			},
		}

		responseBody, _ := json.Marshal(expectedResponse)
		mockResponse := &http.Response{
			StatusCode: 201,
			Body:       io.NopCloser(bytes.NewReader(responseBody)),
		}

		// Configurar mock para capturar a estrutura da requisição
		var capturedBody interface{}
		mockClient.On("Post", mock.Anything, "/api/v3/envelopes", mock.AnythingOfType("*dto.EnvelopeCreateRequestWrapper")).
			Run(func(args mock.Arguments) {
				capturedBody = args.Get(2)
			}).
			Return(mockResponse, nil)

		// Act
		envelope := &MockEntityEnvelope{
			Name:           "Test Envelope",
			AutoClose:      true,
			RemindInterval: 3,
		}
		
		_, _, err := service.CreateEnvelope(context.Background(), (*entity.EntityEnvelope)(envelope))

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, capturedBody)

		// Verificar estrutura JSON API
		requestWrapper, ok := capturedBody.(*dto.EnvelopeCreateRequestWrapper)
		assert.True(t, ok, "Request body should be EnvelopeCreateRequestWrapper")
		assert.Equal(t, "envelopes", requestWrapper.Data.Type)
		assert.Equal(t, "Test Envelope", requestWrapper.Data.Attributes.Name)
		assert.Equal(t, "pt-BR", requestWrapper.Data.Attributes.Locale)
		assert.True(t, requestWrapper.Data.Attributes.AutoClose)
		assert.Equal(t, 3, requestWrapper.Data.Attributes.RemindInterval)

		mockClient.AssertExpectations(t)
	})
}

func TestDocumentCreateRequestConformance(t *testing.T) {
	t.Run("should create document request with correct JSON API structure", func(t *testing.T) {
		// Arrange
		mockClient := &MockClicksignClient{}
		logger := logrus.New()
		service := NewDocumentService(mockClient, logger)

		expectedResponse := dto.DocumentCreateResponseWrapper{
			Data: dto.DocumentCreateResponseData{
				Type: "documents",
				ID:   "test-document-id",
				Attributes: dto.DocumentCreateResponseAttributes{
					Filename:    "test.pdf",
					ContentType: "application/pdf",
					Filesize:    1024,
				},
			},
		}

		responseBody, _ := json.Marshal(expectedResponse)
		mockResponse := &http.Response{
			StatusCode: 201,
			Body:       io.NopCloser(bytes.NewReader(responseBody)),
		}

		// Configurar mock para capturar a estrutura da requisição
		var capturedBody interface{}
		mockClient.On("Post", mock.Anything, "/api/v3/envelopes/test-envelope/documents", mock.AnythingOfType("*dto.DocumentCreateRequestWrapper")).
			Run(func(args mock.Arguments) {
				capturedBody = args.Get(2)
			}).
			Return(mockResponse, nil)

		// Act
		document := &MockEntityDocument{
			ID:            1,
			Name:          "test",
			MimeType:      "application/pdf",
			IsFromBase64:  false,
			FilePath:      "/tmp/test.pdf",
		}

		// Criar arquivo temporário para o teste
		testContent := []byte("test pdf content")
		CreateTempFile(t, document.FilePath, testContent)

		_, err := service.CreateDocument(context.Background(), "test-envelope", (*entity.EntityDocument)(document))

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, capturedBody)

		// Verificar estrutura JSON API
		requestWrapper, ok := capturedBody.(*dto.DocumentCreateRequestWrapper)
		assert.True(t, ok, "Request body should be DocumentCreateRequestWrapper")
		assert.Equal(t, "documents", requestWrapper.Data.Type)
		assert.Equal(t, "test.pdf", requestWrapper.Data.Attributes.Filename)
		assert.NotEmpty(t, requestWrapper.Data.Attributes.ContentBase64)
		assert.NotNil(t, requestWrapper.Data.Attributes.Metadata)
		assert.Equal(t, "private", requestWrapper.Data.Attributes.Metadata.Type)
		assert.Equal(t, 1, requestWrapper.Data.Attributes.Metadata.ID)

		mockClient.AssertExpectations(t)
	})
}

func TestSignerCreateRequestConformance(t *testing.T) {
	t.Run("should create signer request with correct JSON API structure", func(t *testing.T) {
		// Arrange
		mockClient := &MockClicksignClient{}
		logger := logrus.New()
		service := NewSignerService(mockClient, logger)

		expectedResponse := dto.SignerCreateResponseWrapper{
			Data: dto.SignerCreateResponseData{
				Type: "signers",
				ID:   "test-signer-id",
				Attributes: dto.SignerCreateResponseAttributes{
					Name:  "João Silva",
					Email: "joao.silva@example.com",
				},
			},
		}

		responseBody, _ := json.Marshal(expectedResponse)
		mockResponse := &http.Response{
			StatusCode: 201,
			Body:       io.NopCloser(bytes.NewReader(responseBody)),
		}

		// Configurar mock para capturar a estrutura da requisição
		var capturedBody interface{}
		mockClient.On("Post", mock.Anything, "/api/v3/envelopes/test-envelope/signers", mock.AnythingOfType("*dto.SignerCreateRequestWrapper")).
			Run(func(args mock.Arguments) {
				capturedBody = args.Get(2)
			}).
			Return(mockResponse, nil)

		// Act
		signerData := SignerData{
			Name:             "João Silva",
			Email:            "joao.silva@example.com",
			Birthday:         "1990-01-01",
			HasDocumentation: true,
			Refusable:        false,
			Group:            1,
			CommunicateEvents: &SignerCommunicateEventsData{
				DocumentSigned:    "email",
				SignatureRequest:  "email",
				SignatureReminder: "email",
			},
		}

		_, err := service.CreateSigner(context.Background(), "test-envelope", signerData)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, capturedBody)

		// Verificar estrutura JSON API
		requestWrapper, ok := capturedBody.(*dto.SignerCreateRequestWrapper)
		assert.True(t, ok, "Request body should be SignerCreateRequestWrapper")
		assert.Equal(t, "signers", requestWrapper.Data.Type)
		assert.Equal(t, "João Silva", requestWrapper.Data.Attributes.Name)
		assert.Equal(t, "joao.silva@example.com", requestWrapper.Data.Attributes.Email)
		assert.Equal(t, "1990-01-01", requestWrapper.Data.Attributes.Birthday)
		assert.True(t, requestWrapper.Data.Attributes.HasDocumentation)
		assert.False(t, requestWrapper.Data.Attributes.Refusable)
		assert.Equal(t, 1, requestWrapper.Data.Attributes.Group)
		assert.NotNil(t, requestWrapper.Data.Attributes.CommunicateEvents)
		assert.Equal(t, "email", requestWrapper.Data.Attributes.CommunicateEvents.DocumentSigned)

		mockClient.AssertExpectations(t)
	})
}

func TestRequirementCreateRequestConformance(t *testing.T) {
	t.Run("should create requirement request with correct JSON API structure and relationships", func(t *testing.T) {
		// Arrange
		mockClient := &MockClicksignClient{}
		logger := logrus.New()
		service := NewRequirementService(mockClient, logger)

		expectedResponse := dto.RequirementCreateResponseWrapper{
			Data: dto.RequirementCreateResponseData{
				Type: "requirements",
				ID:   "test-requirement-id",
				Attributes: dto.RequirementCreateResponseAttributes{
					Action: "agree",
					Role:   "sign",
				},
			},
		}

		responseBody, _ := json.Marshal(expectedResponse)
		mockResponse := &http.Response{
			StatusCode: 201,
			Body:       io.NopCloser(bytes.NewReader(responseBody)),
		}

		// Configurar mock para capturar a estrutura da requisição
		var capturedBody interface{}
		mockClient.On("Post", mock.Anything, "/api/v3/envelopes/test-envelope/requirements", mock.AnythingOfType("*dto.RequirementCreateRequestWrapper")).
			Run(func(args mock.Arguments) {
				capturedBody = args.Get(2)
			}).
			Return(mockResponse, nil)

		// Act
		reqData := RequirementData{
			Action:     "agree",
			Role:       "sign",
			DocumentID: "test-document-id",
			SignerID:   "test-signer-id",
		}

		_, err := service.CreateRequirement(context.Background(), "test-envelope", reqData)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, capturedBody)

		// Verificar estrutura JSON API
		requestWrapper, ok := capturedBody.(*dto.RequirementCreateRequestWrapper)
		assert.True(t, ok, "Request body should be RequirementCreateRequestWrapper")
		assert.Equal(t, "requirements", requestWrapper.Data.Type)
		assert.Equal(t, "agree", requestWrapper.Data.Attributes.Action)
		assert.Equal(t, "sign", requestWrapper.Data.Attributes.Role)

		// Verificar relacionamentos conforme JSON API spec
		assert.NotNil(t, requestWrapper.Data.Relationships)
		assert.NotNil(t, requestWrapper.Data.Relationships.Document)
		assert.Equal(t, "documents", requestWrapper.Data.Relationships.Document.Data.Type)
		assert.Equal(t, "test-document-id", requestWrapper.Data.Relationships.Document.Data.ID)
		assert.NotNil(t, requestWrapper.Data.Relationships.Signer)
		assert.Equal(t, "signers", requestWrapper.Data.Relationships.Signer.Data.Type)
		assert.Equal(t, "test-signer-id", requestWrapper.Data.Relationships.Signer.Data.ID)

		mockClient.AssertExpectations(t)
	})
}

func TestBulkRequirementsConformance(t *testing.T) {
	t.Run("should create bulk requirements with atomic operations structure", func(t *testing.T) {
		// Arrange
		mockClient := &MockClicksignClient{}
		logger := logrus.New()
		service := NewRequirementService(mockClient, logger)

		expectedResponse := dto.BulkRequirementsResponseWrapper{
			AtomicResults: []dto.AtomicResult{
				{
					Data: &dto.RequirementCreateResponseData{
						Type: "requirements",
						ID:   "req-1",
					},
				},
			},
		}

		responseBody, _ := json.Marshal(expectedResponse)
		mockResponse := &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewReader(responseBody)),
		}

		// Configurar mock para capturar a estrutura da requisição
		var capturedBody interface{}
		mockClient.On("Post", mock.Anything, "/api/v3/envelopes/test-envelope/bulk_requirements", mock.AnythingOfType("*dto.BulkRequirementsRequestWrapper")).
			Run(func(args mock.Arguments) {
				capturedBody = args.Get(2)
			}).
			Return(mockResponse, nil)

		// Act
		operations := []BulkOperation{
			{
				Operation:     "remove",
				RequirementID: "old-req-id",
			},
			{
				Operation: "add",
				RequirementData: &RequirementData{
					Action:     "agree",
					Role:       "sign",
					DocumentID: "doc-id",
					SignerID:   "signer-id",
				},
			},
		}

		_, err := service.CreateBulkRequirements(context.Background(), "test-envelope", operations)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, capturedBody)

		// Verificar estrutura atomic operations
		requestWrapper, ok := capturedBody.(*dto.BulkRequirementsRequestWrapper)
		assert.True(t, ok, "Request body should be BulkRequirementsRequestWrapper")
		assert.Len(t, requestWrapper.AtomicOperations, 2)

		// Verificar operação remove
		removeOp := requestWrapper.AtomicOperations[0]
		assert.Equal(t, "remove", removeOp.Op)
		assert.NotNil(t, removeOp.Ref)
		assert.Equal(t, "requirements", removeOp.Ref.Type)
		assert.Equal(t, "old-req-id", removeOp.Ref.ID)
		assert.Nil(t, removeOp.Data)

		// Verificar operação add
		addOp := requestWrapper.AtomicOperations[1]
		assert.Equal(t, "add", addOp.Op)
		assert.Nil(t, addOp.Ref)
		assert.NotNil(t, addOp.Data)
		assert.Equal(t, "requirements", addOp.Data.Type)
		assert.Equal(t, "agree", addOp.Data.Attributes.Action)
		assert.Equal(t, "sign", addOp.Data.Attributes.Role)

		mockClient.AssertExpectations(t)
	})
}

func TestJSONAPIResponseParsing(t *testing.T) {
	t.Run("should parse JSON API error responses correctly", func(t *testing.T) {
		// Arrange
		errorResponse := dto.ClicksignErrorResponse{
			Error: struct {
				Type       string         `json:"type"`
				Message    string         `json:"message"`
				Details    map[string]any `json:"details,omitempty"`
				Code       string         `json:"code,omitempty"`
				StatusCode int            `json:"status_code,omitempty"`
			}{
				Type:    "validation_error",
				Message: "Nome é obrigatório",
				Code:    "required_field",
			},
		}

		responseBody, _ := json.Marshal(errorResponse)

		// Act & Assert
		var parsedError dto.ClicksignErrorResponse
		err := json.Unmarshal(responseBody, &parsedError)

		assert.NoError(t, err)
		assert.Equal(t, "validation_error", parsedError.Error.Type)
		assert.Equal(t, "Nome é obrigatório", parsedError.Error.Message)
		assert.Equal(t, "required_field", parsedError.Error.Code)
	})
}

// Mocks das entidades para testes que são compatíveis com as entidades reais
type MockEntityEnvelope entity.EntityEnvelope

type MockEntityDocument entity.EntityDocument

// Helper para criar arquivos temporários nos testes
func CreateTempFile(t *testing.T, filepath string, content []byte) {
	// Criar diretórios se necessário
	if idx := strings.LastIndex(filepath, "/"); idx != -1 {
		dir := filepath[:idx]
		err := os.MkdirAll(dir, 0755)
		assert.NoError(t, err)
	}

	err := os.WriteFile(filepath, content, 0644)
	assert.NoError(t, err)

	// Limpar arquivo após o teste
	t.Cleanup(func() {
		os.Remove(filepath)
	})
}