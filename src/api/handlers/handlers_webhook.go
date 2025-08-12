package handlers

import (
	"app/api/handlers/dtos"
	"app/usecase/webhook"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// WebhookHandler representa o handler de webhooks
type WebhookHandler struct {
	webhookUsecase webhook.UsecaseWebhookInterface
	logger         *logrus.Logger
}

// NewWebhookHandler cria uma nova instância do handler de webhooks
func NewWebhookHandler(webhookUsecase webhook.UsecaseWebhookInterface, logger *logrus.Logger) *WebhookHandler {
	return &WebhookHandler{
		webhookUsecase: webhookUsecase,
		logger:         logger,
	}
}

// ReceiveWebhook recebe um webhook do Clicksign
// @Summary Recebe webhook do Clicksign
// @Description Recebe e processa webhooks enviados pelo Clicksign
// @Tags webhooks
// @Accept json
// @Produce json
// @Param webhook body dtos.WebhookRequestDTO true "Dados do webhook"
// @Success 200 {object} dtos.WebhookResponseDTO
// @Failure 400 {object} dtos.ErrorResponseDTO
// @Failure 500 {object} dtos.ErrorResponseDTO
// @Router /api/v1/webhooks [post]
func (h *WebhookHandler) ReceiveWebhook(c *gin.Context) {
	// Ler o body da requisição
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.logger.Error("Failed to read request body", map[string]interface{}{
			"error": err.Error(),
		})
		c.JSON(http.StatusBadRequest, dtos.ErrorResponseDTO{
			Error:   "BAD_REQUEST",
			Message: "Falha ao ler o corpo da requisição",
		})
		return
	}

	// Parse do JSON
	var webhookDTO dtos.WebhookRequestDTO
	if err := json.Unmarshal(body, &webhookDTO); err != nil {
		h.logger.Error("Failed to parse webhook JSON", map[string]interface{}{
			"error": err.Error(),
			"body":  string(body),
		})
		c.JSON(http.StatusBadRequest, dtos.ErrorResponseDTO{
			Error:   "INVALID_JSON",
			Message: "JSON inválido",
		})
		return
	}

	// Validar dados obrigatórios
	if webhookDTO.Event.Name == "" {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponseDTO{
			Error:   "MISSING_EVENT_NAME",
			Message: "Nome do evento é obrigatório",
		})
		return
	}

	if webhookDTO.Document.Key == "" {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponseDTO{
			Error:   "MISSING_DOCUMENT_KEY",
			Message: "Chave do documento é obrigatória",
		})
		return
	}

	// Processar webhook
	webhook, err := h.webhookUsecase.ProcessWebhook(&webhookDTO, string(body))
	if err != nil {
		h.logger.Error("Failed to process webhook", map[string]interface{}{
			"error":        err.Error(),
			"event_name":   webhookDTO.Event.Name,
			"document_key": webhookDTO.Document.Key,
		})
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponseDTO{
			Error:   "PROCESSING_ERROR",
			Message: "Erro ao processar webhook",
		})
		return
	}

	// Retornar resposta
	response := dtos.WebhookResponseDTO{
		ID:          webhook.ID,
		EventName:   webhook.EventName,
		DocumentKey: webhook.DocumentKey,
		AccountKey:  webhook.AccountKey,
		Status:      webhook.Status,
		ProcessedAt: webhook.ProcessedAt,
		Error:       webhook.Error,
		CreatedAt:   webhook.CreatedAt,
		UpdatedAt:   webhook.UpdatedAt,
	}

	c.JSON(http.StatusOK, response)
}

// GetWebhookByID busca um webhook por ID
// @Summary Busca webhook por ID
// @Description Retorna um webhook específico pelo ID
// @Tags webhooks
// @Accept json
// @Produce json
// @Param id path int true "ID do webhook"
// @Success 200 {object} dtos.WebhookResponseDTO
// @Failure 404 {object} dtos.ErrorResponseDTO
// @Failure 500 {object} dtos.ErrorResponseDTO
// @Router /api/v1/webhooks/{id} [get]
func (h *WebhookHandler) GetWebhookByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponseDTO{
			Error:   "INVALID_ID",
			Message: "ID inválido",
		})
		return
	}

	webhook, err := h.webhookUsecase.GetWebhookByID(id)
	if err != nil {
		h.logger.Error("Failed to get webhook by ID", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		})
		c.JSON(http.StatusNotFound, dtos.ErrorResponseDTO{
			Error:   "WEBHOOK_NOT_FOUND",
			Message: "Webhook não encontrado",
		})
		return
	}

	response := dtos.WebhookResponseDTO{
		ID:          webhook.ID,
		EventName:   webhook.EventName,
		DocumentKey: webhook.DocumentKey,
		AccountKey:  webhook.AccountKey,
		Status:      webhook.Status,
		ProcessedAt: webhook.ProcessedAt,
		Error:       webhook.Error,
		CreatedAt:   webhook.CreatedAt,
		UpdatedAt:   webhook.UpdatedAt,
	}

	c.JSON(http.StatusOK, response)
}

// GetWebhooks lista webhooks com filtros
// @Summary Lista webhooks
// @Description Retorna uma lista de webhooks com filtros opcionais
// @Tags webhooks
// @Accept json
// @Produce json
// @Param event_name query string false "Nome do evento"
// @Param document_key query string false "Chave do documento"
// @Param account_key query string false "Chave da conta"
// @Param status query string false "Status do webhook"
// @Param page query int false "Página (padrão: 1)"
// @Param limit query int false "Limite por página (padrão: 10)"
// @Success 200 {object} dtos.WebhookListResponseDTO
// @Failure 500 {object} dtos.ErrorResponseDTO
// @Router /api/v1/webhooks [get]
func (h *WebhookHandler) GetWebhooks(c *gin.Context) {
	// Parse dos parâmetros de query
	filters := &dtos.WebhookFiltersDTO{
		EventName:   c.Query("event_name"),
		DocumentKey: c.Query("document_key"),
		AccountKey:  c.Query("account_key"),
		Status:      c.Query("status"),
	}

	// Parse de page e limit
	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil {
			filters.Page = page
		}
	}
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			filters.Limit = limit
		}
	}

	// Valores padrão
	if filters.Page <= 0 {
		filters.Page = 1
	}
	if filters.Limit <= 0 {
		filters.Limit = 10
	}

	webhooks, total, err := h.webhookUsecase.GetWebhooksByFilters(filters)
	if err != nil {
		h.logger.Error("Failed to get webhooks", map[string]interface{}{
			"error": err.Error(),
		})
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponseDTO{
			Error:   "DATABASE_ERROR",
			Message: "Erro ao buscar webhooks",
		})
		return
	}

	// Converter para DTOs de resposta
	webhookResponses := make([]dtos.WebhookResponseDTO, len(webhooks))
	for i, webhook := range webhooks {
		webhookResponses[i] = dtos.WebhookResponseDTO{
			ID:          webhook.ID,
			EventName:   webhook.EventName,
			DocumentKey: webhook.DocumentKey,
			AccountKey:  webhook.AccountKey,
			Status:      webhook.Status,
			ProcessedAt: webhook.ProcessedAt,
			Error:       webhook.Error,
			CreatedAt:   webhook.CreatedAt,
			UpdatedAt:   webhook.UpdatedAt,
		}
	}

	response := dtos.WebhookListResponseDTO{
		Webhooks: webhookResponses,
		Total:    int(total),
	}

	c.JSON(http.StatusOK, response)
}

// GetWebhooksByDocumentKey busca webhooks por document key
// @Summary Busca webhooks por document key
// @Description Retorna webhooks de um documento específico
// @Tags webhooks
// @Accept json
// @Produce json
// @Param document_key path string true "Chave do documento"
// @Success 200 {object} dtos.WebhookListResponseDTO
// @Failure 500 {object} dtos.ErrorResponseDTO
// @Router /api/v1/webhooks/document/{document_key} [get]
func (h *WebhookHandler) GetWebhooksByDocumentKey(c *gin.Context) {
	documentKey := c.Param("document_key")
	if documentKey == "" {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponseDTO{
			Error:   "MISSING_DOCUMENT_KEY",
			Message: "Chave do documento é obrigatória",
		})
		return
	}

	webhooks, err := h.webhookUsecase.GetWebhooksByDocumentKey(documentKey)
	if err != nil {
		h.logger.Error("Failed to get webhooks by document key", map[string]interface{}{
			"error":        err.Error(),
			"document_key": documentKey,
		})
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponseDTO{
			Error:   "DATABASE_ERROR",
			Message: "Erro ao buscar webhooks",
		})
		return
	}

	// Converter para DTOs de resposta
	webhookResponses := make([]dtos.WebhookResponseDTO, len(webhooks))
	for i, webhook := range webhooks {
		webhookResponses[i] = dtos.WebhookResponseDTO{
			ID:          webhook.ID,
			EventName:   webhook.EventName,
			DocumentKey: webhook.DocumentKey,
			AccountKey:  webhook.AccountKey,
			Status:      webhook.Status,
			ProcessedAt: webhook.ProcessedAt,
			Error:       webhook.Error,
			CreatedAt:   webhook.CreatedAt,
			UpdatedAt:   webhook.UpdatedAt,
		}
	}

	response := dtos.WebhookListResponseDTO{
		Webhooks: webhookResponses,
		Total:    len(webhookResponses),
	}

	c.JSON(http.StatusOK, response)
}

// RetryWebhook tenta reprocessar um webhook que falhou
// @Summary Reprocessa webhook
// @Description Tenta reprocessar um webhook que falhou
// @Tags webhooks
// @Accept json
// @Produce json
// @Param id path int true "ID do webhook"
// @Success 200 {object} dtos.WebhookProcessResponseDTO
// @Failure 400 {object} dtos.ErrorResponseDTO
// @Failure 404 {object} dtos.ErrorResponseDTO
// @Failure 500 {object} dtos.ErrorResponseDTO
// @Router /api/v1/webhooks/{id}/retry [post]
func (h *WebhookHandler) RetryWebhook(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponseDTO{
			Error:   "INVALID_ID",
			Message: "ID inválido",
		})
		return
	}

	err = h.webhookUsecase.RetryWebhook(id)
	if err != nil {
		h.logger.Error("Failed to retry webhook", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		})
		c.JSON(http.StatusBadRequest, dtos.ErrorResponseDTO{
			Error:   "RETRY_ERROR",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dtos.WebhookProcessResponseDTO{
		Success: true,
		Message: "Webhook marcado para reprocessamento",
	})
}

// DeleteWebhook deleta um webhook
// @Summary Deleta webhook
// @Description Remove um webhook do sistema
// @Tags webhooks
// @Accept json
// @Produce json
// @Param id path int true "ID do webhook"
// @Success 200 {object} dtos.WebhookProcessResponseDTO
// @Failure 400 {object} dtos.ErrorResponseDTO
// @Failure 404 {object} dtos.ErrorResponseDTO
// @Failure 500 {object} dtos.ErrorResponseDTO
// @Router /api/v1/webhooks/{id} [delete]
func (h *WebhookHandler) DeleteWebhook(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponseDTO{
			Error:   "INVALID_ID",
			Message: "ID inválido",
		})
		return
	}

	err = h.webhookUsecase.DeleteWebhook(id)
	if err != nil {
		h.logger.Error("Failed to delete webhook", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		})
		c.JSON(http.StatusNotFound, dtos.ErrorResponseDTO{
			Error:   "DELETE_ERROR",
			Message: "Erro ao deletar webhook",
		})
		return
	}

	c.JSON(http.StatusOK, dtos.WebhookProcessResponseDTO{
		Success: true,
		Message: "Webhook deletado com sucesso",
	})
}

// GetPendingWebhooks busca webhooks pendentes
// @Summary Lista webhooks pendentes
// @Description Retorna webhooks que estão pendentes de processamento
// @Tags webhooks
// @Accept json
// @Produce json
// @Success 200 {object} dtos.WebhookListResponseDTO
// @Failure 500 {object} dtos.ErrorResponseDTO
// @Router /api/v1/webhooks/pending [get]
func (h *WebhookHandler) GetPendingWebhooks(c *gin.Context) {
	webhooks, err := h.webhookUsecase.GetPendingWebhooks()
	if err != nil {
		h.logger.Error("Failed to get pending webhooks", map[string]interface{}{
			"error": err.Error(),
		})
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponseDTO{
			Error:   "DATABASE_ERROR",
			Message: "Erro ao buscar webhooks pendentes",
		})
		return
	}

	// Converter para DTOs de resposta
	webhookResponses := make([]dtos.WebhookResponseDTO, len(webhooks))
	for i, webhook := range webhooks {
		webhookResponses[i] = dtos.WebhookResponseDTO{
			ID:          webhook.ID,
			EventName:   webhook.EventName,
			DocumentKey: webhook.DocumentKey,
			AccountKey:  webhook.AccountKey,
			Status:      webhook.Status,
			ProcessedAt: webhook.ProcessedAt,
			Error:       webhook.Error,
			CreatedAt:   webhook.CreatedAt,
			UpdatedAt:   webhook.UpdatedAt,
		}
	}

	response := dtos.WebhookListResponseDTO{
		Webhooks: webhookResponses,
		Total:    len(webhookResponses),
	}

	c.JSON(http.StatusOK, response)
}

// GetFailedWebhooks busca webhooks que falharam
// @Summary Lista webhooks que falharam
// @Description Retorna webhooks que falharam no processamento
// @Tags webhooks
// @Accept json
// @Produce json
// @Success 200 {object} dtos.WebhookListResponseDTO
// @Failure 500 {object} dtos.ErrorResponseDTO
// @Router /api/v1/webhooks/failed [get]
func (h *WebhookHandler) GetFailedWebhooks(c *gin.Context) {
	webhooks, err := h.webhookUsecase.GetFailedWebhooks()
	if err != nil {
		h.logger.Error("Failed to get failed webhooks", map[string]interface{}{
			"error": err.Error(),
		})
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponseDTO{
			Error:   "DATABASE_ERROR",
			Message: "Erro ao buscar webhooks que falharam",
		})
		return
	}

	// Converter para DTOs de resposta
	webhookResponses := make([]dtos.WebhookResponseDTO, len(webhooks))
	for i, webhook := range webhooks {
		webhookResponses[i] = dtos.WebhookResponseDTO{
			ID:          webhook.ID,
			EventName:   webhook.EventName,
			DocumentKey: webhook.DocumentKey,
			AccountKey:  webhook.AccountKey,
			Status:      webhook.Status,
			ProcessedAt: webhook.ProcessedAt,
			Error:       webhook.Error,
			CreatedAt:   webhook.CreatedAt,
			UpdatedAt:   webhook.UpdatedAt,
		}
	}

	response := dtos.WebhookListResponseDTO{
		Webhooks: webhookResponses,
		Total:    len(webhookResponses),
	}

	c.JSON(http.StatusOK, response)
}
