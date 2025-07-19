package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"app/api/handlers/dtos"
	"app/entity"
	"app/mocks"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestNewEnvelopeHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecaseEnvelope := mocks.NewMockIUsecaseEnvelope(ctrl)
	mockUsecaseSignatory := mocks.NewMockIUsecaseSignatory(ctrl)
	logger := logrus.New()

	handler := NewEnvelopeHandler(mockUsecaseEnvelope, mockUsecaseSignatory, logger)

	assert.NotNil(t, handler)
	assert.Equal(t, mockUsecaseEnvelope, handler.UsecaseEnvelope)
	assert.Equal(t, mockUsecaseSignatory, handler.UsecaseSignatory)
	assert.Equal(t, logger, handler.Logger)
}

func TestCreateEnvelopeHandler_WithoutSignatories_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	gin.SetMode(gin.TestMode)

	mockUsecaseEnvelope := mocks.NewMockIUsecaseEnvelope(ctrl)
	mockUsecaseSignatory := mocks.NewMockIUsecaseSignatory(ctrl)
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel) // Suprimir logs durante teste

	handler := NewEnvelopeHandler(mockUsecaseEnvelope, mockUsecaseSignatory, logger)

	// Request DTO sem signatários (compatibilidade retroativa)
	requestDTO := dtos.EnvelopeCreateRequestDTO{
		Name:            "Test Envelope",
		Description:     "Test Description",
		SignatoryEmails: []string{"test@example.com"},
		Documents: []dtos.EnvelopeDocumentRequest{
			{
				Name:              "Test Document",
				FileContentBase64: "JVBERi0xLjQKJeLjz9MKMSAwIG9iago8PAovVHlwZSAvQ2F0YWxvZwovUGFnZXMgMiAwIFIKPj4KZW5kb2JqCjIgMCBvYmoKPDwKL1R5cGUgL1BhZ2VzCi9LaWRzIFszIDAgUl0KL0NvdW50IDEKPD4KZW5kb2JqCjMgMCBvYmoKPDwKL1R5cGUgL1BhZ2UKL1BhcmVudCAyIDAgUgovTWVkaWFCb3ggWzAgMCA2MTIgNzkyXQo+PgplbmRvYmoKeHJlZgowIDQKMDAwMDAwMDAwMCA2NTUzNSBmCjAwMDAwMDAwMDkgMDAwMDAgbgowMDAwMDAwMDU4IDAwMDAwIG4KMDAwMDAwMDExNSAwMDAwMCBuCnRyYWlsZXIKPDwKL1NpemUgNAovUm9vdCAxIDAgUgo+PgpzdGFydHhyZWYKMTc4CiUlRU9G", // PDF mínimo válido em base64
				Description:       "Test Document Description",
			},
		},
	}

	// Mock entity esperada
	expectedEnvelope := &entity.EntityEnvelope{
		ID:              1,
		Name:            "Test Envelope",
		Description:     "Test Description",
		Status:          "draft",
		ClicksignKey:    "test-key",
		SignatoryEmails: []string{"test@example.com"},
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Configurar expectativas do mock
	mockUsecaseEnvelope.EXPECT().
		CreateEnvelopeWithDocuments(gomock.Any(), gomock.Any()).
		Return(expectedEnvelope, nil).
		Times(1)

	// Preparar request
	jsonData, err := json.Marshal(requestDTO)
	assert.NoError(t, err)

	req, err := http.NewRequest("POST", "/api/v1/envelopes", bytes.NewBuffer(jsonData))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Correlation-ID", "test-correlation-id")

	// Preparar response recorder
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Executar handler
	handler.CreateEnvelopeHandler(c)


	// Verificar response
	assert.Equal(t, http.StatusCreated, w.Code)

	var responseDTO dtos.EnvelopeResponseDTO
	err = json.Unmarshal(w.Body.Bytes(), &responseDTO)
	assert.NoError(t, err)

	assert.Equal(t, expectedEnvelope.ID, responseDTO.ID)
	assert.Equal(t, expectedEnvelope.Name, responseDTO.Name)
	assert.Equal(t, expectedEnvelope.Description, responseDTO.Description)
	assert.Equal(t, expectedEnvelope.Status, responseDTO.Status)
	assert.Nil(t, responseDTO.Signatories) // Deve ser nil quando não há signatários
}

func TestCreateEnvelopeHandler_WithSignatories_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	gin.SetMode(gin.TestMode)

	mockUsecaseEnvelope := mocks.NewMockIUsecaseEnvelope(ctrl)
	mockUsecaseSignatory := mocks.NewMockIUsecaseSignatory(ctrl)
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel) // Suprimir logs durante teste

	handler := NewEnvelopeHandler(mockUsecaseEnvelope, mockUsecaseSignatory, logger)

	// Request DTO com signatários
	requestDTO := dtos.EnvelopeCreateRequestDTO{
		Name:            "Test Envelope",
		Description:     "Test Description",
		SignatoryEmails: []string{"test@example.com"},
		Signatories: []dtos.EnvelopeSignatoryRequest{
			{
				Name:  "Test Signatory",
				Email: "signatory@example.com",
			},
		},
		Documents: []dtos.EnvelopeDocumentRequest{
			{
				Name:              "Test Document",
				FileContentBase64: "JVBERi0xLjQKJeLjz9MKMSAwIG9iago8PAovVHlwZSAvQ2F0YWxvZwovUGFnZXMgMiAwIFIKPj4KZW5kb2JqCjIgMCBvYmoKPDwKL1R5cGUgL1BhZ2VzCi9LaWRzIFszIDAgUl0KL0NvdW50IDEKPD4KZW5kb2JqCjMgMCBvYmoKPDwKL1R5cGUgL1BhZ2UKL1BhcmVudCAyIDAgUgovTWVkaWFCb3ggWzAgMCA2MTIgNzkyXQo+PgplbmRvYmoKeHJlZgowIDQKMDAwMDAwMDAwMCA2NTUzNSBmCjAwMDAwMDAwMDkgMDAwMDAgbgowMDAwMDAwMDU4IDAwMDAwIG4KMDAwMDAwMDExNSAwMDAwMCBuCnRyYWlsZXIKPDwKL1NpemUgNAovUm9vdCAxIDAgUgo+PgpzdGFydHhyZWYKMTc4CiUlRU9G", // PDF mínimo válido em base64
				Description:       "Test Document Description",
			},
		},
	}

	// Mock entities esperadas
	expectedEnvelope := &entity.EntityEnvelope{
		ID:              1,
		Name:            "Test Envelope",
		Description:     "Test Description",
		Status:          "draft",
		ClicksignKey:    "test-key",
		SignatoryEmails: []string{"test@example.com"},
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	expectedSignatory := &entity.EntitySignatory{
		ID:         1,
		Name:       "Test Signatory",
		Email:      "signatory@example.com",
		EnvelopeID: 1,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Configurar expectativas dos mocks
	mockUsecaseEnvelope.EXPECT().
		CreateEnvelopeWithDocuments(gomock.Any(), gomock.Any()).
		Return(expectedEnvelope, nil).
		Times(1)

	mockUsecaseSignatory.EXPECT().
		CreateSignatory(gomock.Any()).
		Return(expectedSignatory, nil).
		Times(1)

	// Preparar request
	jsonData, err := json.Marshal(requestDTO)
	assert.NoError(t, err)

	req, err := http.NewRequest("POST", "/api/v1/envelopes", bytes.NewBuffer(jsonData))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Correlation-ID", "test-correlation-id")

	// Preparar response recorder
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Executar handler
	handler.CreateEnvelopeHandler(c)

	// Verificar response
	assert.Equal(t, http.StatusCreated, w.Code)

	var responseDTO dtos.EnvelopeResponseDTO
	err = json.Unmarshal(w.Body.Bytes(), &responseDTO)
	assert.NoError(t, err)

	assert.Equal(t, expectedEnvelope.ID, responseDTO.ID)
	assert.Equal(t, expectedEnvelope.Name, responseDTO.Name)
	assert.Equal(t, expectedEnvelope.Description, responseDTO.Description)
	assert.Equal(t, expectedEnvelope.Status, responseDTO.Status)
	assert.NotNil(t, responseDTO.Signatories)
	assert.Len(t, responseDTO.Signatories, 1)
	assert.Equal(t, expectedSignatory.ID, responseDTO.Signatories[0].ID)
	assert.Equal(t, expectedSignatory.Name, responseDTO.Signatories[0].Name)
	assert.Equal(t, expectedSignatory.Email, responseDTO.Signatories[0].Email)
}

func TestCreateEnvelopeHandler_EnvelopeCreationFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	gin.SetMode(gin.TestMode)

	mockUsecaseEnvelope := mocks.NewMockIUsecaseEnvelope(ctrl)
	mockUsecaseSignatory := mocks.NewMockIUsecaseSignatory(ctrl)
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel) // Suprimir logs durante teste

	handler := NewEnvelopeHandler(mockUsecaseEnvelope, mockUsecaseSignatory, logger)

	requestDTO := dtos.EnvelopeCreateRequestDTO{
		Name:            "Test Envelope",
		Description:     "Test Description",
		SignatoryEmails: []string{"test@example.com"},
		Documents: []dtos.EnvelopeDocumentRequest{
			{
				Name:              "Test Document",
				FileContentBase64: "JVBERi0xLjQKJeLjz9MKMSAwIG9iago8PAovVHlwZSAvQ2F0YWxvZwovUGFnZXMgMiAwIFIKPj4KZW5kb2JqCjIgMCBvYmoKPDwKL1R5cGUgL1BhZ2VzCi9LaWRzIFszIDAgUl0KL0NvdW50IDEKPD4KZW5kb2JqCjMgMCBvYmoKPDwKL1R5cGUgL1BhZ2UKL1BhcmVudCAyIDAgUgovTWVkaWFCb3ggWzAgMCA2MTIgNzkyXQo+PgplbmRvYmoKeHJlZgowIDQKMDAwMDAwMDAwMCA2NTUzNSBmCjAwMDAwMDAwMDkgMDAwMDAgbgowMDAwMDAwMDU4IDAwMDAwIG4KMDAwMDAwMDExNSAwMDAwMCBuCnRyYWlsZXIKPDwKL1NpemUgNAovUm9vdCAxIDAgUgo+PgpzdGFydHhyZWYKMTc4CiUlRU9G", // PDF mínimo válido em base64
				Description:       "Test Document Description",
			},
		},
	}

	// Mock falha na criação do envelope
	mockUsecaseEnvelope.EXPECT().
		CreateEnvelopeWithDocuments(gomock.Any(), gomock.Any()).
		Return(nil, fmt.Errorf("database error")).
		Times(1)

	// Preparar request
	jsonData, err := json.Marshal(requestDTO)
	assert.NoError(t, err)

	req, err := http.NewRequest("POST", "/api/v1/envelopes", bytes.NewBuffer(jsonData))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Correlation-ID", "test-correlation-id")

	// Preparar response recorder
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Executar handler
	handler.CreateEnvelopeHandler(c)

	// Verificar response
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var errorResponse dtos.ErrorResponseDTO
	err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)

	assert.Equal(t, "Internal server error", errorResponse.Error)
	assert.Equal(t, "Failed to create envelope", errorResponse.Message)
}

func TestCreateEnvelopeHandler_SignatoryCreationFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	gin.SetMode(gin.TestMode)

	mockUsecaseEnvelope := mocks.NewMockIUsecaseEnvelope(ctrl)
	mockUsecaseSignatory := mocks.NewMockIUsecaseSignatory(ctrl)
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel) // Suprimir logs durante teste

	handler := NewEnvelopeHandler(mockUsecaseEnvelope, mockUsecaseSignatory, logger)

	requestDTO := dtos.EnvelopeCreateRequestDTO{
		Name:            "Test Envelope",
		Description:     "Test Description",
		SignatoryEmails: []string{"test@example.com"},
		Signatories: []dtos.EnvelopeSignatoryRequest{
			{
				Name:  "Test Signatory",
				Email: "signatory@example.com",
			},
		},
		Documents: []dtos.EnvelopeDocumentRequest{
			{
				Name:              "Test Document",
				FileContentBase64: "JVBERi0xLjQKJeLjz9MKMSAwIG9iago8PAovVHlwZSAvQ2F0YWxvZwovUGFnZXMgMiAwIFIKPj4KZW5kb2JqCjIgMCBvYmoKPDwKL1R5cGUgL1BhZ2VzCi9LaWRzIFszIDAgUl0KL0NvdW50IDEKPD4KZW5kb2JqCjMgMCBvYmoKPDwKL1R5cGUgL1BhZ2UKL1BhcmVudCAyIDAgUgovTWVkaWFCb3ggWzAgMCA2MTIgNzkyXQo+PgplbmRvYmoKeHJlZgowIDQKMDAwMDAwMDAwMCA2NTUzNSBmCjAwMDAwMDAwMDkgMDAwMDAgbgowMDAwMDAwMDU4IDAwMDAwIG4KMDAwMDAwMDExNSAwMDAwMCBuCnRyYWlsZXIKPDwKL1NpemUgNAovUm9vdCAxIDAgUgo+PgpzdGFydHhyZWYKMTc4CiUlRU9G", // PDF mínimo válido em base64
				Description:       "Test Document Description",
			},
		},
	}

	expectedEnvelope := &entity.EntityEnvelope{
		ID:              1,
		Name:            "Test Envelope",
		Description:     "Test Description",
		Status:          "draft",
		ClicksignKey:    "test-key",
		SignatoryEmails: []string{"test@example.com"},
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Mock envelope criado com sucesso, mas signatory falha
	mockUsecaseEnvelope.EXPECT().
		CreateEnvelopeWithDocuments(gomock.Any(), gomock.Any()).
		Return(expectedEnvelope, nil).
		Times(1)

	mockUsecaseSignatory.EXPECT().
		CreateSignatory(gomock.Any()).
		Return(nil, fmt.Errorf("duplicate email")).
		Times(1)

	// Preparar request
	jsonData, err := json.Marshal(requestDTO)
	assert.NoError(t, err)

	req, err := http.NewRequest("POST", "/api/v1/envelopes", bytes.NewBuffer(jsonData))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Correlation-ID", "test-correlation-id")

	// Preparar response recorder
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Executar handler
	handler.CreateEnvelopeHandler(c)

	// Verificar response
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var errorResponse dtos.ErrorResponseDTO
	err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)

	assert.Equal(t, "Internal server error", errorResponse.Error)
	assert.Contains(t, errorResponse.Message, "Failed to create signatory 1")
}

func TestCreateEnvelopeHandler_ValidationErrors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	gin.SetMode(gin.TestMode)

	mockUsecaseEnvelope := mocks.NewMockIUsecaseEnvelope(ctrl)
	mockUsecaseSignatory := mocks.NewMockIUsecaseSignatory(ctrl)
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel) // Suprimir logs durante teste

	handler := NewEnvelopeHandler(mockUsecaseEnvelope, mockUsecaseSignatory, logger)

	// Request DTO com dados inválidos
	requestDTO := dtos.EnvelopeCreateRequestDTO{
		Name:            "AB", // Muito curto (min=3)
		Description:     "Test Description",
		SignatoryEmails: []string{}, // Vazio (required,min=1)
	}

	// Preparar request
	jsonData, err := json.Marshal(requestDTO)
	assert.NoError(t, err)

	req, err := http.NewRequest("POST", "/api/v1/envelopes", bytes.NewBuffer(jsonData))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Correlation-ID", "test-correlation-id")

	// Preparar response recorder
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Executar handler
	handler.CreateEnvelopeHandler(c)

	// Verificar response
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var validationErrorResponse dtos.ValidationErrorResponseDTO
	err = json.Unmarshal(w.Body.Bytes(), &validationErrorResponse)
	assert.NoError(t, err)

	assert.Equal(t, "Validation failed", validationErrorResponse.Error)
	assert.NotEmpty(t, validationErrorResponse.Details)
}

func TestCreateEnvelopeHandler_DuplicateSignatoryEmails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	gin.SetMode(gin.TestMode)

	mockUsecaseEnvelope := mocks.NewMockIUsecaseEnvelope(ctrl)
	mockUsecaseSignatory := mocks.NewMockIUsecaseSignatory(ctrl)
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel) // Suprimir logs durante teste

	handler := NewEnvelopeHandler(mockUsecaseEnvelope, mockUsecaseSignatory, logger)

	// Request DTO com emails duplicados de signatários
	requestDTO := dtos.EnvelopeCreateRequestDTO{
		Name:            "Test Envelope",
		Description:     "Test Description",
		SignatoryEmails: []string{"test@example.com"},
		Signatories: []dtos.EnvelopeSignatoryRequest{
			{
				Name:  "Test Signatory 1",
				Email: "duplicate@example.com",
			},
			{
				Name:  "Test Signatory 2",
				Email: "duplicate@example.com", // Email duplicado
			},
		},
		Documents: []dtos.EnvelopeDocumentRequest{
			{
				Name:              "Test Document",
				FileContentBase64: "JVBERi0xLjQKJeLjz9MKMSAwIG9iago8PAovVHlwZSAvQ2F0YWxvZwovUGFnZXMgMiAwIFIKPj4KZW5kb2JqCjIgMCBvYmoKPDwKL1R5cGUgL1BhZ2VzCi9LaWRzIFszIDAgUl0KL0NvdW50IDEKPD4KZW5kb2JqCjMgMCBvYmoKPDwKL1R5cGUgL1BhZ2UKL1BhcmVudCAyIDAgUgovTWVkaWFCb3ggWzAgMCA2MTIgNzkyXQo+PgplbmRvYmoKeHJlZgowIDQKMDAwMDAwMDAwMCA2NTUzNSBmCjAwMDAwMDAwMDkgMDAwMDAgbgowMDAwMDAwMDU4IDAwMDAwIG4KMDAwMDAwMDExNSAwMDAwMCBuCnRyYWlsZXIKPDwKL1NpemUgNAovUm9vdCAxIDAgUgo+PgpzdGFydHhyZWYKMTc4CiUlRU9G", // PDF mínimo válido em base64
				Description:       "Test Document Description",
			},
		},
	}

	// Preparar request
	jsonData, err := json.Marshal(requestDTO)
	assert.NoError(t, err)

	req, err := http.NewRequest("POST", "/api/v1/envelopes", bytes.NewBuffer(jsonData))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Correlation-ID", "test-correlation-id")

	// Preparar response recorder
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Executar handler
	handler.CreateEnvelopeHandler(c)

	// Verificar response
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var errorResponse dtos.ErrorResponseDTO
	err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)

	assert.Equal(t, "Validation failed", errorResponse.Error)
	assert.Contains(t, errorResponse.Message, "email duplicado encontrado")
}

func TestMapEntityToResponse_WithSignatories(t *testing.T) {
	logger := logrus.New()
	handler := &EnvelopeHandlers{Logger: logger}

	envelope := &entity.EntityEnvelope{
		ID:              1,
		Name:            "Test Envelope",
		Description:     "Test Description",
		Status:          "draft",
		ClicksignKey:    "test-key",
		SignatoryEmails: []string{"test@example.com"},
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	signatories := []entity.EntitySignatory{
		{
			ID:         1,
			Name:       "Test Signatory",
			Email:      "signatory@example.com",
			EnvelopeID: 1,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
	}

	result := handler.mapEntityToResponse(envelope, signatories)

	assert.NotNil(t, result)
	assert.Equal(t, envelope.ID, result.ID)
	assert.Equal(t, envelope.Name, result.Name)
	assert.NotNil(t, result.Signatories)
	assert.Len(t, result.Signatories, 1)
	assert.Equal(t, signatories[0].ID, result.Signatories[0].ID)
	assert.Equal(t, signatories[0].Name, result.Signatories[0].Name)
	assert.Equal(t, signatories[0].Email, result.Signatories[0].Email)
}

func TestMapEntityToResponse_WithoutSignatories(t *testing.T) {
	logger := logrus.New()
	handler := &EnvelopeHandlers{Logger: logger}

	envelope := &entity.EntityEnvelope{
		ID:              1,
		Name:            "Test Envelope",
		Description:     "Test Description",
		Status:          "draft",
		ClicksignKey:    "test-key",
		SignatoryEmails: []string{"test@example.com"},
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	result := handler.mapEntityToResponse(envelope)

	assert.NotNil(t, result)
	assert.Equal(t, envelope.ID, result.ID)
	assert.Equal(t, envelope.Name, result.Name)
	assert.Nil(t, result.Signatories) // Deve ser nil quando não há signatários
}

func TestCreateEnvelopeHandler_WithClicksignRawData_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	gin.SetMode(gin.TestMode)

	mockUsecaseEnvelope := mocks.NewMockIUsecaseEnvelope(ctrl)
	mockUsecaseSignatory := mocks.NewMockIUsecaseSignatory(ctrl)
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)

	handler := NewEnvelopeHandler(mockUsecaseEnvelope, mockUsecaseSignatory, logger)

	requestDTO := dtos.EnvelopeCreateRequestDTO{
		Name:            "Test Envelope",
		Description:     "Test Description",
		SignatoryEmails: []string{"test@example.com"},
		Documents: []dtos.EnvelopeDocumentRequest{
			{
				Name:              "Test Document",
				FileContentBase64: "JVBERi0xLjQKJeLjz9MKMSAwIG9iago8PAovVHlwZSAvQ2F0YWxvZwovUGFnZXMgMiAwIFIKPj4KZW5kb2JqCjIgMCBvYmoKPDwKL1R5cGUgL1BhZ2VzCi9LaWRzIFszIDAgUl0KL0NvdW50IDEKPD4KZW5kb2JqCjMgMCBvYmoKPDwKL1R5cGUgL1BhZ2UKL1BhcmVudCAyIDAgUgovTWVkaWFCb3ggWzAgMCA2MTIgNzkyXQo+PgplbmRvYmoKeHJlZgowIDQKMDAwMDAwMDAwMCA2NTUzNSBmCjAwMDAwMDAwMDkgMDAwMDAgbgowMDAwMDAwMDU4IDAwMDAwIG4KMDAwMDAwMDExNSAwMDAwMCBuCnRyYWlsZXIKPDwKL1NpemUgNAovUm9vdCAxIDAgUgo+PgpzdGFydHhyZWYKMTc4CiUlRU9G",
				Description:       "Test Document Description",
			},
		},
	}

	rawData := `{"data":{"id":"test-key","type":"envelopes","attributes":{"name":"Test Envelope","status":"draft"}}}`
	expectedEnvelope := &entity.EntityEnvelope{
		ID:               1,
		Name:             "Test Envelope",
		Description:      "Test Description",
		Status:           "draft",
		ClicksignKey:     "test-key",
		ClicksignRawData: &rawData,
		SignatoryEmails:  []string{"test@example.com"},
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	mockUsecaseEnvelope.EXPECT().
		CreateEnvelopeWithDocuments(gomock.Any(), gomock.Any()).
		Return(expectedEnvelope, nil).
		Times(1)

	jsonData, err := json.Marshal(requestDTO)
	assert.NoError(t, err)

	req, err := http.NewRequest("POST", "/api/v1/envelopes", bytes.NewBuffer(jsonData))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Correlation-ID", "test-correlation-id")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.CreateEnvelopeHandler(c)

	assert.Equal(t, http.StatusCreated, w.Code)

	var responseDTO dtos.EnvelopeResponseDTO
	err = json.Unmarshal(w.Body.Bytes(), &responseDTO)
	assert.NoError(t, err)

	assert.Equal(t, expectedEnvelope.ID, responseDTO.ID)
	assert.Equal(t, expectedEnvelope.Name, responseDTO.Name)
	assert.Equal(t, expectedEnvelope.ClicksignKey, responseDTO.ClicksignKey)
	assert.NotNil(t, responseDTO.ClicksignRawData)
	assert.Equal(t, rawData, *responseDTO.ClicksignRawData)
}

func TestCreateEnvelopeHandler_WithoutClicksignRawData_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	gin.SetMode(gin.TestMode)

	mockUsecaseEnvelope := mocks.NewMockIUsecaseEnvelope(ctrl)
	mockUsecaseSignatory := mocks.NewMockIUsecaseSignatory(ctrl)
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)

	handler := NewEnvelopeHandler(mockUsecaseEnvelope, mockUsecaseSignatory, logger)

	requestDTO := dtos.EnvelopeCreateRequestDTO{
		Name:            "Test Envelope",
		Description:     "Test Description",
		SignatoryEmails: []string{"test@example.com"},
		Documents: []dtos.EnvelopeDocumentRequest{
			{
				Name:              "Test Document",
				FileContentBase64: "JVBERi0xLjQKJeLjz9MKMSAwIG9iago8PAovVHlwZSAvQ2F0YWxvZwovUGFnZXMgMiAwIFIKPj4KZW5kb2JqCjIgMCBvYmoKPDwKL1R5cGUgL1BhZ2VzCi9LaWRzIFszIDAgUl0KL0NvdW50IDEKPD4KZW5kb2JqCjMgMCBvYmoKPDwKL1R5cGUgL1BhZ2UKL1BhcmVudCAyIDAgUgovTWVkaWFCb3ggWzAgMCA2MTIgNzkyXQo+PgplbmRvYmoKeHJlZgowIDQKMDAwMDAwMDAwMCA2NTUzNSBmCjAwMDAwMDAwMDkgMDAwMDAgbgowMDAwMDAwMDU4IDAwMDAwIG4KMDAwMDAwMDExNSAwMDAwMCBuCnRyYWlsZXIKPDwKL1NpemUgNAovUm9vdCAxIDAgUgo+PgpzdGFydHhyZWYKMTc4CiUlRU9G",
				Description:       "Test Document Description",
			},
		},
	}

	expectedEnvelope := &entity.EntityEnvelope{
		ID:               1,
		Name:             "Test Envelope",
		Description:      "Test Description",
		Status:           "draft",
		ClicksignKey:     "test-key",
		ClicksignRawData: nil,
		SignatoryEmails:  []string{"test@example.com"},
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	mockUsecaseEnvelope.EXPECT().
		CreateEnvelopeWithDocuments(gomock.Any(), gomock.Any()).
		Return(expectedEnvelope, nil).
		Times(1)

	jsonData, err := json.Marshal(requestDTO)
	assert.NoError(t, err)

	req, err := http.NewRequest("POST", "/api/v1/envelopes", bytes.NewBuffer(jsonData))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Correlation-ID", "test-correlation-id")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.CreateEnvelopeHandler(c)

	assert.Equal(t, http.StatusCreated, w.Code)

	var responseDTO dtos.EnvelopeResponseDTO
	err = json.Unmarshal(w.Body.Bytes(), &responseDTO)
	assert.NoError(t, err)

	assert.Equal(t, expectedEnvelope.ID, responseDTO.ID)
	assert.Equal(t, expectedEnvelope.Name, responseDTO.Name)
	assert.Equal(t, expectedEnvelope.ClicksignKey, responseDTO.ClicksignKey)
	assert.Nil(t, responseDTO.ClicksignRawData)
}

func TestMapEntityToResponse_WithClicksignRawData(t *testing.T) {
	logger := logrus.New()
	handler := &EnvelopeHandlers{Logger: logger}

	rawData := `{"data":{"id":"test-key","type":"envelopes","attributes":{"name":"Test Envelope"}}}`
	envelope := &entity.EntityEnvelope{
		ID:               1,
		Name:             "Test Envelope",
		Description:      "Test Description",
		Status:           "draft",
		ClicksignKey:     "test-key",
		ClicksignRawData: &rawData,
		SignatoryEmails:  []string{"test@example.com"},
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	result := handler.mapEntityToResponse(envelope)

	assert.NotNil(t, result)
	assert.Equal(t, envelope.ID, result.ID)
	assert.Equal(t, envelope.Name, result.Name)
	assert.Equal(t, envelope.ClicksignKey, result.ClicksignKey)
	assert.NotNil(t, result.ClicksignRawData)
	assert.Equal(t, rawData, *result.ClicksignRawData)
}

func TestMapEntityToResponse_WithoutClicksignRawData(t *testing.T) {
	logger := logrus.New()
	handler := &EnvelopeHandlers{Logger: logger}

	envelope := &entity.EntityEnvelope{
		ID:               1,
		Name:             "Test Envelope",
		Description:      "Test Description",
		Status:           "draft",
		ClicksignKey:     "test-key",
		ClicksignRawData: nil,
		SignatoryEmails:  []string{"test@example.com"},
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	result := handler.mapEntityToResponse(envelope)

	assert.NotNil(t, result)
	assert.Equal(t, envelope.ID, result.ID)
	assert.Equal(t, envelope.Name, result.Name)
	assert.Equal(t, envelope.ClicksignKey, result.ClicksignKey)
	assert.Nil(t, result.ClicksignRawData)
}