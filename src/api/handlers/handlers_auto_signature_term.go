package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"app/api/handlers/dtos"
	"app/config"
	"app/entity"
	"app/infrastructure/clicksign"
	"app/infrastructure/repository"
	"app/usecase/auto_signature_term"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type AutoSignatureTermHandlers struct {
	UsecaseAutoSignatureTerm auto_signature_term.IUsecaseAutoSignatureTerm
	Logger                   *logrus.Logger
}

func NewAutoSignatureTermHandler(usecaseAutoSignatureTerm auto_signature_term.IUsecaseAutoSignatureTerm, logger *logrus.Logger) *AutoSignatureTermHandlers {
	return &AutoSignatureTermHandlers{
		UsecaseAutoSignatureTerm: usecaseAutoSignatureTerm,
		Logger:                   logger,
	}
}

// @Summary Create auto signature term
// @Description Create a new auto signature term in Clicksign. This endpoint creates a term that allows automatic signature for a specific signer.
// @Tags auto-signature-terms
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body dtos.AutoSignatureTermCreateRequestDTO true "Auto signature term data"
// @Success 201 {object} dtos.AutoSignatureTermResponseDTO "Auto signature term created successfully"
// @Failure 400 {object} dtos.ValidationErrorResponseDTO "Validation error - invalid request data"
// @Failure 409 {object} dtos.ErrorResponseDTO "Conflict - term already exists for this signer and operator"
// @Failure 500 {object} dtos.ErrorResponseDTO "Internal server error - term creation failed"
// @Router /api/v1/auto-signature/terms [post]
func (h *AutoSignatureTermHandlers) CreateAutoSignatureTermHandler(c *gin.Context) {
	correlationID := c.GetHeader("X-Correlation-ID")
	if correlationID == "" {
		correlationID = strconv.FormatInt(time.Now().Unix(), 10)
	}

	var requestDTO dtos.AutoSignatureTermCreateRequestDTO

	if err := c.ShouldBindJSON(&requestDTO); err != nil {
		validationErrors := h.extractValidationErrors(err)
		c.JSON(http.StatusBadRequest, dtos.ValidationErrorResponseDTO{
			Error:   "Validation failed",
			Message: "Invalid request payload",
			Details: validationErrors,
		})
		return
	}

	// Converter DTO para entidade
	term, err := h.mapCreateRequestToEntity(requestDTO)
	if err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponseDTO{
			Error:   "Invalid request",
			Message: err.Error(),
		})
		return
	}

	// Criar o termo
	createdTerm, err := h.UsecaseAutoSignatureTerm.CreateAutoSignatureTerm(c.Request.Context(), term)
	if err != nil {
		h.Logger.WithError(err).Error("Failed to create auto signature term")

		// Verificar se é o erro específico de termo já existente
		if h.isTermAlreadyExistsError(err) {
			c.JSON(http.StatusConflict, dtos.ErrorResponseDTO{
				Error:   "Term already exists",
				Message: "Já existe um termo de assinatura automática para este signatário e operador.",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, dtos.ErrorResponseDTO{
			Error:   "Internal server error",
			Message: "Failed to create auto signature term",
		})
		return
	}

	// Converter entidade para DTO de resposta
	responseDTO := h.mapEntityToResponse(createdTerm)

	c.JSON(http.StatusCreated, responseDTO)
}

// @Summary Get auto signature term by ID
// @Description Get an auto signature term by its ID
// @Tags auto-signature-terms
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Auto signature term ID"
// @Success 200 {object} dtos.AutoSignatureTermResponseDTO "Auto signature term found"
// @Failure 404 {object} dtos.ErrorResponseDTO "Auto signature term not found"
// @Failure 500 {object} dtos.ErrorResponseDTO "Internal server error"
// @Router /api/v1/auto-signature/terms/{id} [get]
func (h *AutoSignatureTermHandlers) GetAutoSignatureTermHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponseDTO{
			Error:   "Invalid ID",
			Message: "ID must be a valid integer",
		})
		return
	}

	term, err := h.UsecaseAutoSignatureTerm.GetAutoSignatureTerm(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, dtos.ErrorResponseDTO{
				Error:   "Not found",
				Message: "Auto signature term not found",
			})
			return
		}
		h.Logger.WithError(err).Error("Failed to get auto signature term")
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponseDTO{
			Error:   "Internal server error",
			Message: "Failed to get auto signature term",
		})
		return
	}

	responseDTO := h.mapEntityToResponse(term)
	c.JSON(http.StatusOK, responseDTO)
}

// @Summary Get all auto signature terms
// @Description Get all auto signature terms
// @Tags auto-signature-terms
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} dtos.AutoSignatureTermListResponseDTO "List of auto signature terms"
// @Failure 500 {object} dtos.ErrorResponseDTO "Internal server error"
// @Router /api/v1/auto-signature/terms [get]
func (h *AutoSignatureTermHandlers) GetAllAutoSignatureTermsHandler(c *gin.Context) {
	terms, err := h.UsecaseAutoSignatureTerm.GetAllAutoSignatureTerms()
	if err != nil {
		h.Logger.WithError(err).Error("Failed to get all auto signature terms")
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponseDTO{
			Error:   "Internal server error",
			Message: "Failed to get auto signature terms",
		})
		return
	}

	responseDTO := h.mapEntitiesToResponseList(terms)
	c.JSON(http.StatusOK, responseDTO)
}

// @Summary Delete auto signature term
// @Description Delete an auto signature term by its ID
// @Tags auto-signature-terms
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Auto signature term ID"
// @Success 204 "Auto signature term deleted successfully"
// @Failure 404 {object} dtos.ErrorResponseDTO "Auto signature term not found"
// @Failure 500 {object} dtos.ErrorResponseDTO "Internal server error"
// @Router /api/v1/auto-signature/terms/{id} [delete]
func (h *AutoSignatureTermHandlers) DeleteAutoSignatureTermHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponseDTO{
			Error:   "Invalid ID",
			Message: "ID must be a valid integer",
		})
		return
	}

	err = h.UsecaseAutoSignatureTerm.DeleteAutoSignatureTerm(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, dtos.ErrorResponseDTO{
				Error:   "Not found",
				Message: "Auto signature term not found",
			})
			return
		}
		h.Logger.WithError(err).Error("Failed to delete auto signature term")
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponseDTO{
			Error:   "Internal server error",
			Message: "Failed to delete auto signature term",
		})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *AutoSignatureTermHandlers) mapCreateRequestToEntity(requestDTO dtos.AutoSignatureTermCreateRequestDTO) (*entity.EntityAutoSignatureTerm, error) {
	term := entity.EntityAutoSignatureTerm{
		SignerDocumentation: requestDTO.Signer.Documentation,
		SignerBirthday:      requestDTO.Signer.Birthday,
		SignerEmail:         requestDTO.Signer.Email,
		SignerName:          requestDTO.Signer.Name,
		AdminEmail:          requestDTO.AdminEmail,
		APIEmail:            requestDTO.APIEmail,
	}

	return entity.NewAutoSignatureTerm(term)
}

func (h *AutoSignatureTermHandlers) mapEntityToResponse(term *entity.EntityAutoSignatureTerm) dtos.AutoSignatureTermResponseDTO {
	return dtos.AutoSignatureTermResponseDTO{
		ID: term.ID,
		Signer: dtos.SignerInfoDTO{
			Documentation: term.SignerDocumentation,
			Birthday:      term.SignerBirthday,
			Email:         term.SignerEmail,
			Name:          term.SignerName,
		},
		AdminEmail:       term.AdminEmail,
		APIEmail:         term.APIEmail,
		ClicksignKey:     term.ClicksignKey,
		ClicksignRawData: term.ClicksignRawData,
		CreatedAt:        term.CreatedAt,
		UpdatedAt:        term.UpdatedAt,
	}
}

func (h *AutoSignatureTermHandlers) mapEntitiesToResponseList(terms []entity.EntityAutoSignatureTerm) dtos.AutoSignatureTermListResponseDTO {
	var responseList []dtos.AutoSignatureTermResponseDTO
	for _, term := range terms {
		responseList = append(responseList, h.mapEntityToResponse(&term))
	}

	return dtos.AutoSignatureTermListResponseDTO{
		Terms: responseList,
		Total: len(terms),
	}
}

func (h *AutoSignatureTermHandlers) extractValidationErrors(err error) []dtos.ValidationErrorDetail {
	var validationErrors []dtos.ValidationErrorDetail

	if validationErr, ok := err.(validator.ValidationErrors); ok {
		for _, fieldError := range validationErr {
			validationErrors = append(validationErrors, dtos.ValidationErrorDetail{
				Field:   fieldError.Field(),
				Message: h.getValidationErrorMessage(fieldError),
				Value:   fmt.Sprintf("%v", fieldError.Value()),
			})
		}
	} else {
		validationErrors = append(validationErrors, dtos.ValidationErrorDetail{
			Field:   "general",
			Message: err.Error(),
		})
	}

	return validationErrors
}

func (h *AutoSignatureTermHandlers) getValidationErrorMessage(fieldError validator.FieldError) string {
	switch fieldError.Tag() {
	case "required":
		return "This field is required"
	case "min":
		return "This field must have at least " + fieldError.Param() + " characters/items"
	case "max":
		return "This field must have at most " + fieldError.Param() + " characters/items"
	case "email":
		return "This field must be a valid email address"
	default:
		return "This field is invalid"
	}
}

// isTermAlreadyExistsError verifica se o erro é relacionado a um termo já existente
func (h *AutoSignatureTermHandlers) isTermAlreadyExistsError(err error) bool {
	errorStr := err.Error()
	return strings.Contains(errorStr, "Já existe um termo de assinatura automática para este signatário e operador")
}

func MountAutoSignatureTermHandlers(gin *gin.Engine, conn *gorm.DB, logger *logrus.Logger) {
	clicksignClient := clicksign.NewClicksignClient(config.EnvironmentVariables, logger).(*clicksign.ClicksignClient)

	autoSignatureTermHandlers := NewAutoSignatureTermHandler(
		auto_signature_term.NewUsecaseAutoSignatureTermService(
			repository.NewRepositoryAutoSignatureTerm(conn),
			clicksignClient,
			logger,
		),
		logger,
	)

	group := gin.Group("/api/v1/auto-signature/terms")
	SetAuthMiddleware(conn, group)

	group.POST("/", autoSignatureTermHandlers.CreateAutoSignatureTermHandler)
	group.GET("/:id", autoSignatureTermHandlers.GetAutoSignatureTermHandler)
	group.GET("/", autoSignatureTermHandlers.GetAllAutoSignatureTermsHandler)
	group.DELETE("/:id", autoSignatureTermHandlers.DeleteAutoSignatureTermHandler)
}
