package handlers

import (
	"app/infrastructure/repository"
	"app/usecase/document"
	"app/usecase/envelope"
	"app/usecase/webhook"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// MountWebhookHandlers monta as rotas de webhook
func MountWebhookHandlers(r *gin.Engine, db *gorm.DB, logger *logrus.Logger) {
	// Criar repositórios
	webhookRepository := repository.NewRepositoryWebhook(db)
	envelopeRepository := repository.NewRepositoryEnvelope(db)
	documentRepository := repository.NewRepositoryDocument(db)

	// Criar usecases
	documentUsecase := document.NewUsecaseDocumentService(documentRepository)
	envelopeUsecase := envelope.NewUsecaseEnvelopeService(
		envelopeRepository,
		nil,             // clicksignClient - será configurado se necessário
		documentUsecase, // usecaseDocument
		nil,             // usecaseRequirement - será configurado se necessário
		logger,
	)
	webhookUsecase := webhook.NewUsecaseWebhookService(webhookRepository, envelopeUsecase, documentUsecase, logger)

	// Criar handler
	webhookHandler := NewWebhookHandler(webhookUsecase, logger)

	// Grupo de rotas para webhooks
	webhookGroup := r.Group("/api/v1/webhooks")
	{
		// POST /api/v1/webhooks - Receber webhook do Clicksign
		webhookGroup.POST("/", webhookHandler.ReceiveWebhook)

		// GET /api/v1/webhooks - Listar webhooks com filtros
		webhookGroup.GET("/", webhookHandler.GetWebhooks)

		// GET /api/v1/webhooks/pending - Listar webhooks pendentes
		webhookGroup.GET("/pending", webhookHandler.GetPendingWebhooks)

		// GET /api/v1/webhooks/failed - Listar webhooks que falharam
		webhookGroup.GET("/failed", webhookHandler.GetFailedWebhooks)

		// GET /api/v1/webhooks/document/:document_key - Buscar webhooks por document key
		webhookGroup.GET("/document/:document_key", webhookHandler.GetWebhooksByDocumentKey)

		// GET /api/v1/webhooks/:id - Buscar webhook por ID
		webhookGroup.GET("/:id", webhookHandler.GetWebhookByID)

		// POST /api/v1/webhooks/:id/retry - Reprocessar webhook
		webhookGroup.POST("/:id/retry", webhookHandler.RetryWebhook)

		// DELETE /api/v1/webhooks/:id - Deletar webhook
		webhookGroup.DELETE("/:id", webhookHandler.DeleteWebhook)
	}
}
