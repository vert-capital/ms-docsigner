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
	"github.com/stretchr/testify/require"
)

func setupSignatoryHandlerTest(t *testing.T) (*gin.Engine, *mocks.MockIUsecaseSignatory, *mocks.MockIUsecaseEnvelope, *gomock.Controller) {
	ctrl := gomock.NewController(t)
	mockUsecaseSignatory := mocks.NewMockIUsecaseSignatory(ctrl)
	mockUsecaseEnvelope := mocks.NewMockIUsecaseEnvelope(ctrl)

	// Configurar Gin em modo de teste
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Criar logger silencioso para testes
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)

	// Criar handler
	signatoryHandler := NewSignatoryHandler(mockUsecaseSignatory, mockUsecaseEnvelope, logger)

	// Configurar rotas de teste
	router.POST("/api/v1/envelopes/:id/signatories", signatoryHandler.CreateSignatoryHandler)
	router.GET("/api/v1/envelopes/:id/signatories", signatoryHandler.GetSignatoriesHandler)
	router.GET("/api/v1/signatories/:id", signatoryHandler.GetSignatoryHandler)
	router.PUT("/api/v1/signatories/:id", signatoryHandler.UpdateSignatoryHandler)
	router.DELETE("/api/v1/signatories/:id", signatoryHandler.DeleteSignatoryHandler)
	router.POST("/api/v1/envelopes/:id/send", signatoryHandler.SendSignatoriesToClicksignHandler)

	return router, mockUsecaseSignatory, mockUsecaseEnvelope, ctrl
}

func TestCreateSignatoryHandler_Success(t *testing.T) {
	router, mockUsecaseSignatory, mockUsecaseEnvelope, ctrl := setupSignatoryHandlerTest(t)
	defer ctrl.Finish()

	envelopeID := 1
	envelope := &entity.EntityEnvelope{
		ID:           envelopeID,
		Name:         "Test Envelope",
		ClicksignKey: "test-key",
		Status:       "draft",
	}

	requestDTO := dtos.SignatoryCreateRequestDTO{
		Name:       "João Silva",
		Email:      "joao@example.com",
		EnvelopeID: envelopeID,
	}

	expectedSignatory := &entity.EntitySignatory{
		ID:         1,
		Name:       "João Silva",
		Email:      "joao@example.com",
		EnvelopeID: envelopeID,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Mock expectations
	mockUsecaseEnvelope.EXPECT().
		GetEnvelope(envelopeID).
		Return(envelope, nil)

	mockUsecaseSignatory.EXPECT().
		CreateSignatory(gomock.Any()).
		Return(expectedSignatory, nil)

	// Preparar request
	jsonBody, _ := json.Marshal(requestDTO)
	req, _ := http.NewRequest("POST", "/api/v1/envelopes/1/signatories", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Correlation-ID", "test-correlation-123")

	// Executar request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verificar resposta
	assert.Equal(t, http.StatusCreated, w.Code)

	var response dtos.SignatoryResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, expectedSignatory.ID, response.ID)
	assert.Equal(t, expectedSignatory.Name, response.Name)
	assert.Equal(t, expectedSignatory.Email, response.Email)
	assert.Equal(t, expectedSignatory.EnvelopeID, response.EnvelopeID)
}

func TestCreateSignatoryHandler_EnvelopeNotFound(t *testing.T) {
	router, mockUsecaseSignatory, mockUsecaseEnvelope, ctrl := setupSignatoryHandlerTest(t)
	defer ctrl.Finish()

	// Suprimir warning de variável não utilizada
	_ = mockUsecaseSignatory

	envelopeID := 999
	requestDTO := dtos.SignatoryCreateRequestDTO{
		Name:       "João Silva",
		Email:      "joao@example.com",
		EnvelopeID: envelopeID,
	}

	// Mock expectations
	mockUsecaseEnvelope.EXPECT().
		GetEnvelope(envelopeID).
		Return(nil, fmt.Errorf("envelope not found"))

	// Preparar request
	jsonBody, _ := json.Marshal(requestDTO)
	req, _ := http.NewRequest("POST", "/api/v1/envelopes/999/signatories", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// Executar request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verificar resposta
	assert.Equal(t, http.StatusNotFound, w.Code)

	var response dtos.ErrorResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Envelope not found", response.Error)
}

func TestCreateSignatoryHandler_InvalidEnvelopeID(t *testing.T) {
	router, _, _, ctrl := setupSignatoryHandlerTest(t)
	defer ctrl.Finish()

	requestDTO := dtos.SignatoryCreateRequestDTO{
		Name:  "João Silva",
		Email: "joao@example.com",
	}

	// Preparar request com ID inválido
	jsonBody, _ := json.Marshal(requestDTO)
	req, _ := http.NewRequest("POST", "/api/v1/envelopes/invalid/signatories", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// Executar request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verificar resposta
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response dtos.ErrorResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Invalid ID", response.Error)
}

func TestCreateSignatoryHandler_ValidationError(t *testing.T) {
	router, mockUsecaseSignatory, mockUsecaseEnvelope, ctrl := setupSignatoryHandlerTest(t)
	defer ctrl.Finish()

	// Suprimir warnings de variáveis não utilizadas
	_ = mockUsecaseSignatory
	_ = mockUsecaseEnvelope

	// Request DTO com campos obrigatórios faltando
	requestDTO := dtos.SignatoryCreateRequestDTO{
		Name: "", // Campo obrigatório vazio
		// Email ausente
	}

	// Preparar request
	jsonBody, _ := json.Marshal(requestDTO)
	req, _ := http.NewRequest("POST", "/api/v1/envelopes/1/signatories", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// Executar request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verificar resposta
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response dtos.ValidationErrorResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Validation failed", response.Error)
	assert.NotEmpty(t, response.Details)
}

func TestGetSignatoriesHandler_Success(t *testing.T) {
	router, mockUsecaseSignatory, mockUsecaseEnvelope, ctrl := setupSignatoryHandlerTest(t)
	defer ctrl.Finish()

	envelopeID := 1
	envelope := &entity.EntityEnvelope{
		ID:           envelopeID,
		Name:         "Test Envelope",
		ClicksignKey: "test-key",
		Status:       "draft",
	}

	expectedSignatories := []entity.EntitySignatory{
		{
			ID:         1,
			Name:       "João Silva",
			Email:      "joao@example.com",
			EnvelopeID: envelopeID,
		},
		{
			ID:         2,
			Name:       "Maria Santos",
			Email:      "maria@example.com",
			EnvelopeID: envelopeID,
		},
	}

	// Mock expectations
	mockUsecaseEnvelope.EXPECT().
		GetEnvelope(envelopeID).
		Return(envelope, nil)

	mockUsecaseSignatory.EXPECT().
		GetSignatoriesByEnvelope(envelopeID).
		Return(expectedSignatories, nil)

	// Preparar request
	req, _ := http.NewRequest("GET", "/api/v1/envelopes/1/signatories", nil)
	req.Header.Set("X-Correlation-ID", "test-correlation-123")

	// Executar request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verificar resposta
	assert.Equal(t, http.StatusOK, w.Code)

	var response dtos.SignatoryListResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, len(expectedSignatories), response.Total)
	assert.Len(t, response.Signatories, len(expectedSignatories))
	assert.Equal(t, expectedSignatories[0].Name, response.Signatories[0].Name)
	assert.Equal(t, expectedSignatories[1].Name, response.Signatories[1].Name)
}

func TestGetSignatoryHandler_Success(t *testing.T) {
	router, mockUsecaseSignatory, _, ctrl := setupSignatoryHandlerTest(t)
	defer ctrl.Finish()

	signatoryID := 1
	expectedSignatory := &entity.EntitySignatory{
		ID:         signatoryID,
		Name:       "João Silva",
		Email:      "joao@example.com",
		EnvelopeID: 1,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Mock expectations
	mockUsecaseSignatory.EXPECT().
		GetSignatory(signatoryID).
		Return(expectedSignatory, nil)

	// Preparar request
	req, _ := http.NewRequest("GET", "/api/v1/signatories/1", nil)
	req.Header.Set("X-Correlation-ID", "test-correlation-123")

	// Executar request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verificar resposta
	assert.Equal(t, http.StatusOK, w.Code)

	var response dtos.SignatoryResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, expectedSignatory.ID, response.ID)
	assert.Equal(t, expectedSignatory.Name, response.Name)
	assert.Equal(t, expectedSignatory.Email, response.Email)
	assert.Equal(t, expectedSignatory.EnvelopeID, response.EnvelopeID)
}

func TestGetSignatoryHandler_NotFound(t *testing.T) {
	router, mockUsecaseSignatory, _, ctrl := setupSignatoryHandlerTest(t)
	defer ctrl.Finish()

	signatoryID := 999

	// Mock expectations
	mockUsecaseSignatory.EXPECT().
		GetSignatory(signatoryID).
		Return(nil, fmt.Errorf("signatory not found"))

	// Preparar request
	req, _ := http.NewRequest("GET", "/api/v1/signatories/999", nil)

	// Executar request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verificar resposta
	assert.Equal(t, http.StatusNotFound, w.Code)

	var response dtos.ErrorResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Signatory not found", response.Error)
}

func TestUpdateSignatoryHandler_Success(t *testing.T) {
	router, mockUsecaseSignatory, _, ctrl := setupSignatoryHandlerTest(t)
	defer ctrl.Finish()

	signatoryID := 1
	existingSignatory := &entity.EntitySignatory{
		ID:         signatoryID,
		Name:       "João Silva",
		Email:      "joao@example.com",
		EnvelopeID: 1,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	updateDTO := dtos.SignatoryUpdateRequestDTO{
		Name: stringPtr("João Santos"),
	}

	// Mock expectations
	mockUsecaseSignatory.EXPECT().
		GetSignatory(signatoryID).
		Return(existingSignatory, nil)

	mockUsecaseSignatory.EXPECT().
		UpdateSignatory(gomock.Any()).
		Return(nil)

	// Preparar request
	jsonBody, _ := json.Marshal(updateDTO)
	req, _ := http.NewRequest("PUT", "/api/v1/signatories/1", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Correlation-ID", "test-correlation-123")

	// Executar request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verificar resposta
	assert.Equal(t, http.StatusOK, w.Code)

	var response dtos.SignatoryResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "João Santos", response.Name)            // Nome atualizado
	assert.Equal(t, existingSignatory.Email, response.Email) // Email inalterado
}

func TestDeleteSignatoryHandler_Success(t *testing.T) {
	router, mockUsecaseSignatory, _, ctrl := setupSignatoryHandlerTest(t)
	defer ctrl.Finish()

	signatoryID := 1
	existingSignatory := &entity.EntitySignatory{
		ID:         signatoryID,
		Name:       "João Silva",
		Email:      "joao@example.com",
		EnvelopeID: 1,
	}

	// Mock expectations
	mockUsecaseSignatory.EXPECT().
		GetSignatory(signatoryID).
		Return(existingSignatory, nil)

	mockUsecaseSignatory.EXPECT().
		DeleteSignatory(signatoryID).
		Return(nil)

	// Preparar request
	req, _ := http.NewRequest("DELETE", "/api/v1/signatories/1", nil)
	req.Header.Set("X-Correlation-ID", "test-correlation-123")

	// Executar request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verificar resposta
	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Empty(t, w.Body.String())
}

func TestDeleteSignatoryHandler_NotFound(t *testing.T) {
	router, mockUsecaseSignatory, _, ctrl := setupSignatoryHandlerTest(t)
	defer ctrl.Finish()

	signatoryID := 999

	// Mock expectations
	mockUsecaseSignatory.EXPECT().
		GetSignatory(signatoryID).
		Return(nil, fmt.Errorf("signatory not found"))

	// Preparar request
	req, _ := http.NewRequest("DELETE", "/api/v1/signatories/999", nil)

	// Executar request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verificar resposta
	assert.Equal(t, http.StatusNotFound, w.Code)

	var response dtos.ErrorResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Signatory not found", response.Error)
}

func TestSendSignatoriesToClicksignHandler_Success(t *testing.T) {
	router, mockUsecaseSignatory, mockUsecaseEnvelope, ctrl := setupSignatoryHandlerTest(t)
	defer ctrl.Finish()

	envelopeID := 1
	envelope := &entity.EntityEnvelope{
		ID:           envelopeID,
		Name:         "Test Envelope",
		ClicksignKey: "test-clicksign-key", // Importante ter a chave
		Status:       "draft",
	}

	signatories := []entity.EntitySignatory{
		{
			ID:         1,
			Name:       "João Silva",
			Email:      "joao@example.com",
			EnvelopeID: envelopeID,
		},
	}

	// Mock expectations
	mockUsecaseEnvelope.EXPECT().
		GetEnvelope(envelopeID).
		Return(envelope, nil)

	mockUsecaseSignatory.EXPECT().
		GetSignatoriesByEnvelope(envelopeID).
		Return(signatories, nil)

	// Preparar request
	req, _ := http.NewRequest("POST", "/api/v1/envelopes/1/send", nil)
	req.Header.Set("X-Correlation-ID", "test-correlation-123")

	// Executar request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verificar resposta
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Como não temos o cliente Clicksign mockado, a chamada falhará,
	// mas o teste verifica se chegou até o ponto de tentar enviar
	assert.Contains(t, response, "signatories")
	assert.Contains(t, response, "total")
}

func TestSendSignatoriesToClicksignHandler_EnvelopeNotReady(t *testing.T) {
	router, mockUsecaseSignatory, mockUsecaseEnvelope, ctrl := setupSignatoryHandlerTest(t)
	defer ctrl.Finish()

	// Suprimir warning de variável não utilizada
	_ = mockUsecaseSignatory

	envelopeID := 1
	envelope := &entity.EntityEnvelope{
		ID:           envelopeID,
		Name:         "Test Envelope",
		ClicksignKey: "", // Sem chave Clicksign
		Status:       "draft",
	}

	// Mock expectations
	mockUsecaseEnvelope.EXPECT().
		GetEnvelope(envelopeID).
		Return(envelope, nil)

	// Preparar request
	req, _ := http.NewRequest("POST", "/api/v1/envelopes/1/send", nil)

	// Executar request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verificar resposta
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response dtos.ErrorResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Envelope not ready", response.Error)
	assert.Contains(t, response.Message, "Clicksign")
}

func TestSendSignatoriesToClicksignHandler_NoSignatories(t *testing.T) {
	router, mockUsecaseSignatory, mockUsecaseEnvelope, ctrl := setupSignatoryHandlerTest(t)
	defer ctrl.Finish()

	envelopeID := 1
	envelope := &entity.EntityEnvelope{
		ID:           envelopeID,
		Name:         "Test Envelope",
		ClicksignKey: "test-clicksign-key",
		Status:       "draft",
	}

	emptySignatories := []entity.EntitySignatory{}

	// Mock expectations
	mockUsecaseEnvelope.EXPECT().
		GetEnvelope(envelopeID).
		Return(envelope, nil)

	mockUsecaseSignatory.EXPECT().
		GetSignatoriesByEnvelope(envelopeID).
		Return(emptySignatories, nil)

	// Preparar request
	req, _ := http.NewRequest("POST", "/api/v1/envelopes/1/send", nil)

	// Executar request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verificar resposta
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response dtos.ErrorResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "No signatories", response.Error)
}

// Helper functions para testes
func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}

func boolPtr(b bool) *bool {
	return &b
}
