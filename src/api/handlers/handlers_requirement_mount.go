package handlers

import (
	"app/config"
	"app/infrastructure/clicksign"
	"app/infrastructure/repository"
	"app/usecase/requirement"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func MountRequirementHandlers(gin *gin.Engine, conn *gorm.DB, logger *logrus.Logger) {
	clicksignClient := clicksign.NewClicksignClient(config.EnvironmentVariables, logger)

	// Criar usecase de requirement
	usecaseRequirement := requirement.NewUsecaseRequirementService(
		repository.NewRepositoryRequirement(conn),
		repository.NewRepositoryEnvelope(conn),
		clicksignClient,
		logger,
	)

	// Criar handler de requirements
	requirementHandlers := NewRequirementHandler(usecaseRequirement, logger)

	// Grupo de rotas individuais de requirements
	requirementGroup := gin.Group("/api/v1/requirements")
	SetAuthMiddleware(conn, requirementGroup)

	requirementGroup.GET("/:requirement_id", requirementHandlers.GetRequirementHandler)
	requirementGroup.PUT("/:requirement_id", requirementHandlers.UpdateRequirementHandler)
	requirementGroup.DELETE("/:requirement_id", requirementHandlers.DeleteRequirementHandler)
}