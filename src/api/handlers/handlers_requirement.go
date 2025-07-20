package handlers

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"app/api/handlers/dtos"
	"app/usecase/requirement"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

type RequirementHandlers struct {
	UsecaseRequirement requirement.IUsecaseRequirement
	Logger             *logrus.Logger
}

func NewRequirementHandler(usecaseRequirement requirement.IUsecaseRequirement, logger *logrus.Logger) *RequirementHandlers {
	return &RequirementHandlers{
		UsecaseRequirement: usecaseRequirement,
		Logger:             logger,
	}
}

// @Summary Create requirement
// @Description Create a new requirement for an envelope in Clicksign. Requirements define actions (agree, sign, provide_evidence) that signers must complete, along with authentication methods when needed.
// @Tags requirements
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Envelope ID"
// @Param request body dtos.RequirementCreateRequestDTO true "Requirement data"
// @Success 201 {object} dtos.RequirementResponseDTO "Requirement created successfully"
// @Failure 400 {object} dtos.ValidationErrorResponseDTO "Validation error - invalid request data or business rule violation"
// @Failure 404 {object} dtos.ErrorResponseDTO "Envelope not found"
// @Failure 500 {object} dtos.ErrorResponseDTO "Internal server error - requirement creation failed"
// @Router /api/v1/envelopes/{id}/requirements [post]
func (h *RequirementHandlers) CreateRequirementHandler(c *gin.Context) {
	correlationID := c.GetHeader("X-Correlation-ID")
	if correlationID == "" {
		correlationID = strconv.FormatInt(time.Now().Unix(), 10)
	}
	ctx := context.WithValue(c.Request.Context(), "correlation_id", correlationID)

	h.Logger.WithFields(logrus.Fields{
		"endpoint":       "POST /api/v1/envelopes/{id}/requirements",
		"correlation_id": correlationID,
		"step":           "request_start",
	}).Info("Starting requirement creation request")

	// Parse envelope_id from URL
	envelopeIDParam := c.Param("id")
	envelopeID, err := strconv.Atoi(envelopeIDParam)
	if err != nil {
		h.Logger.WithFields(logrus.Fields{
			"envelope_id_param": envelopeIDParam,
			"error":             err.Error(),
			"correlation_id":    correlationID,
			"step":              "envelope_id_parse",
		}).Error("Failed to parse envelope_id parameter")

		errorResponse := &dtos.ErrorResponseDTO{
			Message: "envelope_id deve ser um número inteiro válido",
			Error:   "invalid_envelope_id",
		}
		c.JSON(http.StatusBadRequest, errorResponse)
		return
	}

	// Parse request body
	var request dtos.RequirementCreateRequestDTO
	if err := c.ShouldBindJSON(&request); err != nil {
		h.Logger.WithFields(logrus.Fields{
			"envelope_id":    envelopeID,
			"error":          err.Error(),
			"correlation_id": correlationID,
			"step":           "request_bind",
		}).Error("Failed to bind request JSON")

		validationErrors := h.extractValidationErrors(err)
		if len(validationErrors) > 0 {
			validationResponse := &dtos.ValidationErrorResponseDTO{
				Message: "Dados de entrada inválidos",
				Error:   "validation_failed",
				Details: validationErrors,
			}
			c.JSON(http.StatusBadRequest, validationResponse)
			return
		}

		errorResponse := &dtos.ErrorResponseDTO{
			Message: "Formato JSON inválido",
			Error:   "invalid_json",
		}
		c.JSON(http.StatusBadRequest, errorResponse)
		return
	}

	// Validate business rules
	if err := request.Validate(); err != nil {
		h.Logger.WithFields(logrus.Fields{
			"envelope_id":    envelopeID,
			"error":          err.Error(),
			"correlation_id": correlationID,
			"step":           "business_validation",
		}).Error("Business rule validation failed")

		errorResponse := &dtos.ErrorResponseDTO{
			Message: err.Error(),
			Error:   "validation_error",
		}
		c.JSON(http.StatusBadRequest, errorResponse)
		return
	}

	h.Logger.WithFields(logrus.Fields{
		"envelope_id":         envelopeID,
		"requirement_action":  request.Action,
		"requirement_role":    request.Role,
		"requirement_auth":    request.Auth,
		"correlation_id":      correlationID,
		"step":                "request_validation_success",
	}).Debug("Request validation completed successfully")

	// Convert DTO to entity
	requirementEntity := request.ToEntity(envelopeID)

	// Create requirement
	createdRequirement, err := h.UsecaseRequirement.CreateRequirement(ctx, requirementEntity)
	if err != nil {
		h.Logger.WithFields(logrus.Fields{
			"envelope_id":    envelopeID,
			"error":          err.Error(),
			"correlation_id": correlationID,
			"step":           "requirement_creation",
		}).Error("Failed to create requirement")

		// Check if it's a validation error (envelope not found, etc.)
		if contains(err.Error(), "envelope not found") {
			errorResponse := &dtos.ErrorResponseDTO{
				Message: "Envelope não encontrado",
				Error:   "envelope_not_found",
			}
			c.JSON(http.StatusNotFound, errorResponse)
			return
		}

		if contains(err.Error(), "envelope must be created in Clicksign") {
			errorResponse := &dtos.ErrorResponseDTO{
				Message: "Envelope deve ser criado no Clicksign antes de adicionar requisitos",
				Error:   "envelope_not_in_clicksign",
			}
			c.JSON(http.StatusBadRequest, errorResponse)
			return
		}

		errorResponse := &dtos.ErrorResponseDTO{
			Message: "Falha ao criar requisito",
			Error:   "creation_failed",
		}
		c.JSON(http.StatusInternalServerError, errorResponse)
		return
	}

	// Convert to response DTO
	responseDTO := &dtos.RequirementResponseDTO{}
	response := responseDTO.FromEntity(createdRequirement)

	h.Logger.WithFields(logrus.Fields{
		"envelope_id":         envelopeID,
		"requirement_id":      createdRequirement.ID,
		"requirement_action":  createdRequirement.Action,
		"clicksign_key":       createdRequirement.ClicksignKey,
		"correlation_id":      correlationID,
		"step":                "requirement_creation_success",
	}).Info("Requirement created successfully")

	c.JSON(http.StatusCreated, response)
}

// @Summary Get requirements by envelope
// @Description Get all requirements for a specific envelope
// @Tags requirements
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Envelope ID"
// @Success 200 {object} dtos.RequirementListResponseDTO "Requirements retrieved successfully"
// @Failure 404 {object} dtos.ErrorResponseDTO "Envelope not found"
// @Failure 500 {object} dtos.ErrorResponseDTO "Internal server error"
// @Router /api/v1/envelopes/{id}/requirements [get]
func (h *RequirementHandlers) GetRequirementsByEnvelopeHandler(c *gin.Context) {
	correlationID := c.GetHeader("X-Correlation-ID")
	if correlationID == "" {
		correlationID = strconv.FormatInt(time.Now().Unix(), 10)
	}
	ctx := context.WithValue(c.Request.Context(), "correlation_id", correlationID)

	h.Logger.WithFields(logrus.Fields{
		"endpoint":       "GET /api/v1/envelopes/{id}/requirements",
		"correlation_id": correlationID,
		"step":           "request_start",
	}).Info("Starting get requirements by envelope request")

	// Parse envelope_id from URL
	envelopeIDParam := c.Param("id")
	envelopeID, err := strconv.Atoi(envelopeIDParam)
	if err != nil {
		h.Logger.WithFields(logrus.Fields{
			"envelope_id_param": envelopeIDParam,
			"error":             err.Error(),
			"correlation_id":    correlationID,
			"step":              "envelope_id_parse",
		}).Error("Failed to parse envelope_id parameter")

		errorResponse := &dtos.ErrorResponseDTO{
			Message: "envelope_id deve ser um número inteiro válido",
			Error:   "invalid_envelope_id",
		}
		c.JSON(http.StatusBadRequest, errorResponse)
		return
	}

	// Get requirements
	requirements, err := h.UsecaseRequirement.GetRequirementsByEnvelopeID(ctx, envelopeID)
	if err != nil {
		h.Logger.WithFields(logrus.Fields{
			"envelope_id":    envelopeID,
			"error":          err.Error(),
			"correlation_id": correlationID,
			"step":           "requirements_fetch",
		}).Error("Failed to fetch requirements")

		if contains(err.Error(), "envelope not found") {
			errorResponse := &dtos.ErrorResponseDTO{
				Message: "Envelope não encontrado",
				Error:   "envelope_not_found",
			}
			c.JSON(http.StatusNotFound, errorResponse)
			return
		}

		errorResponse := &dtos.ErrorResponseDTO{
			Message: "Falha ao buscar requisitos",
			Error:   "fetch_failed",
		}
		c.JSON(http.StatusInternalServerError, errorResponse)
		return
	}

	// Convert to response DTO
	responseDTO := &dtos.RequirementListResponseDTO{}
	response := responseDTO.FromEntityList(requirements)

	h.Logger.WithFields(logrus.Fields{
		"envelope_id":        envelopeID,
		"requirements_count": len(requirements),
		"correlation_id":     correlationID,
		"step":               "requirements_fetch_success",
	}).Info("Requirements fetched successfully")

	c.JSON(http.StatusOK, response)
}

// @Summary Get requirement by ID
// @Description Get a specific requirement by its ID
// @Tags requirements
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param requirement_id path int true "Requirement ID"
// @Success 200 {object} dtos.RequirementResponseDTO "Requirement retrieved successfully"
// @Failure 404 {object} dtos.ErrorResponseDTO "Requirement not found"
// @Failure 500 {object} dtos.ErrorResponseDTO "Internal server error"
// @Router /api/v1/requirements/{requirement_id} [get]
func (h *RequirementHandlers) GetRequirementHandler(c *gin.Context) {
	correlationID := c.GetHeader("X-Correlation-ID")
	if correlationID == "" {
		correlationID = strconv.FormatInt(time.Now().Unix(), 10)
	}
	ctx := context.WithValue(c.Request.Context(), "correlation_id", correlationID)

	h.Logger.WithFields(logrus.Fields{
		"endpoint":       "GET /api/v1/requirements/{requirement_id}",
		"correlation_id": correlationID,
		"step":           "request_start",
	}).Debug("Starting get requirement request")

	// Parse requirement_id from URL
	requirementIDParam := c.Param("requirement_id")
	requirementID, err := strconv.Atoi(requirementIDParam)
	if err != nil {
		h.Logger.WithFields(logrus.Fields{
			"requirement_id_param": requirementIDParam,
			"error":                err.Error(),
			"correlation_id":       correlationID,
			"step":                 "requirement_id_parse",
		}).Error("Failed to parse requirement_id parameter")

		errorResponse := &dtos.ErrorResponseDTO{
			Message: "requirement_id deve ser um número inteiro válido",
			Error:   "invalid_requirement_id",
		}
		c.JSON(http.StatusBadRequest, errorResponse)
		return
	}

	// Get requirement
	requirement, err := h.UsecaseRequirement.GetRequirement(ctx, requirementID)
	if err != nil {
		h.Logger.WithFields(logrus.Fields{
			"requirement_id": requirementID,
			"error":          err.Error(),
			"correlation_id": correlationID,
			"step":           "requirement_fetch",
		}).Error("Failed to fetch requirement")

		if contains(err.Error(), "failed to fetch requirement") {
			errorResponse := &dtos.ErrorResponseDTO{
				Message: "Requisito não encontrado",
				Error:   "requirement_not_found",
			}
			c.JSON(http.StatusNotFound, errorResponse)
			return
		}

		errorResponse := &dtos.ErrorResponseDTO{
			Message: "Falha ao buscar requisito",
			Error:   "fetch_failed",
		}
		c.JSON(http.StatusInternalServerError, errorResponse)
		return
	}

	// Convert to response DTO
	responseDTO := &dtos.RequirementResponseDTO{}
	response := responseDTO.FromEntity(requirement)

	h.Logger.WithFields(logrus.Fields{
		"requirement_id":     requirement.ID,
		"envelope_id":        requirement.EnvelopeID,
		"requirement_action": requirement.Action,
		"correlation_id":     correlationID,
		"step":               "requirement_fetch_success",
	}).Debug("Requirement fetched successfully")

	c.JSON(http.StatusOK, response)
}

// @Summary Update requirement
// @Description Update a specific requirement (currently only status can be updated)
// @Tags requirements
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param requirement_id path int true "Requirement ID"
// @Param request body dtos.RequirementUpdateRequestDTO true "Requirement update data"
// @Success 200 {object} dtos.RequirementResponseDTO "Requirement updated successfully"
// @Failure 400 {object} dtos.ValidationErrorResponseDTO "Validation error - invalid request data"
// @Failure 404 {object} dtos.ErrorResponseDTO "Requirement not found"
// @Failure 500 {object} dtos.ErrorResponseDTO "Internal server error"
// @Router /api/v1/requirements/{requirement_id} [put]
func (h *RequirementHandlers) UpdateRequirementHandler(c *gin.Context) {
	correlationID := c.GetHeader("X-Correlation-ID")
	if correlationID == "" {
		correlationID = strconv.FormatInt(time.Now().Unix(), 10)
	}
	ctx := context.WithValue(c.Request.Context(), "correlation_id", correlationID)

	h.Logger.WithFields(logrus.Fields{
		"endpoint":       "PUT /api/v1/requirements/{requirement_id}",
		"correlation_id": correlationID,
		"step":           "request_start",
	}).Info("Starting requirement update request")

	// Parse requirement_id from URL
	requirementIDParam := c.Param("requirement_id")
	requirementID, err := strconv.Atoi(requirementIDParam)
	if err != nil {
		h.Logger.WithFields(logrus.Fields{
			"requirement_id_param": requirementIDParam,
			"error":                err.Error(),
			"correlation_id":       correlationID,
			"step":                 "requirement_id_parse",
		}).Error("Failed to parse requirement_id parameter")

		errorResponse := &dtos.ErrorResponseDTO{
			Message: "requirement_id deve ser um número inteiro válido",
			Error:   "invalid_requirement_id",
		}
		c.JSON(http.StatusBadRequest, errorResponse)
		return
	}

	// Parse request body
	var request dtos.RequirementUpdateRequestDTO
	if err := c.ShouldBindJSON(&request); err != nil {
		h.Logger.WithFields(logrus.Fields{
			"requirement_id": requirementID,
			"error":          err.Error(),
			"correlation_id": correlationID,
			"step":           "request_bind",
		}).Error("Failed to bind request JSON")

		validationErrors := h.extractValidationErrors(err)
		if len(validationErrors) > 0 {
			validationResponse := &dtos.ValidationErrorResponseDTO{
				Message: "Dados de entrada inválidos",
				Error:   "validation_failed",
				Details: validationErrors,
			}
			c.JSON(http.StatusBadRequest, validationResponse)
			return
		}

		errorResponse := &dtos.ErrorResponseDTO{
			Message: "Formato JSON inválido",
			Error:   "invalid_json",
		}
		c.JSON(http.StatusBadRequest, errorResponse)
		return
	}

	// Get existing requirement
	requirement, err := h.UsecaseRequirement.GetRequirement(ctx, requirementID)
	if err != nil {
		h.Logger.WithFields(logrus.Fields{
			"requirement_id": requirementID,
			"error":          err.Error(),
			"correlation_id": correlationID,
			"step":           "requirement_fetch",
		}).Error("Failed to fetch requirement for update")

		errorResponse := &dtos.ErrorResponseDTO{
			Message: "Requisito não encontrado",
			Error:   "requirement_not_found",
		}
		c.JSON(http.StatusNotFound, errorResponse)
		return
	}

	// Apply updates
	if request.Status != "" {
		requirement.SetStatus(request.Status)
	}

	// Update requirement
	updatedRequirement, err := h.UsecaseRequirement.UpdateRequirement(ctx, requirement)
	if err != nil {
		h.Logger.WithFields(logrus.Fields{
			"requirement_id": requirementID,
			"error":          err.Error(),
			"correlation_id": correlationID,
			"step":           "requirement_update",
		}).Error("Failed to update requirement")

		errorResponse := &dtos.ErrorResponseDTO{
			Message: "Falha ao atualizar requisito",
			Error:   "update_failed",
		}
		c.JSON(http.StatusInternalServerError, errorResponse)
		return
	}

	// Convert to response DTO
	responseDTO := &dtos.RequirementResponseDTO{}
	response := responseDTO.FromEntity(updatedRequirement)

	h.Logger.WithFields(logrus.Fields{
		"requirement_id": updatedRequirement.ID,
		"envelope_id":    updatedRequirement.EnvelopeID,
		"new_status":     updatedRequirement.Status,
		"correlation_id": correlationID,
		"step":           "requirement_update_success",
	}).Info("Requirement updated successfully")

	c.JSON(http.StatusOK, response)
}

// @Summary Delete requirement
// @Description Delete a specific requirement
// @Tags requirements
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param requirement_id path int true "Requirement ID"
// @Success 204 "Requirement deleted successfully"
// @Failure 404 {object} dtos.ErrorResponseDTO "Requirement not found"
// @Failure 500 {object} dtos.ErrorResponseDTO "Internal server error"
// @Router /api/v1/requirements/{requirement_id} [delete]
func (h *RequirementHandlers) DeleteRequirementHandler(c *gin.Context) {
	correlationID := c.GetHeader("X-Correlation-ID")
	if correlationID == "" {
		correlationID = strconv.FormatInt(time.Now().Unix(), 10)
	}
	ctx := context.WithValue(c.Request.Context(), "correlation_id", correlationID)

	h.Logger.WithFields(logrus.Fields{
		"endpoint":       "DELETE /api/v1/requirements/{requirement_id}",
		"correlation_id": correlationID,
		"step":           "request_start",
	}).Info("Starting requirement deletion request")

	// Parse requirement_id from URL
	requirementIDParam := c.Param("requirement_id")
	requirementID, err := strconv.Atoi(requirementIDParam)
	if err != nil {
		h.Logger.WithFields(logrus.Fields{
			"requirement_id_param": requirementIDParam,
			"error":                err.Error(),
			"correlation_id":       correlationID,
			"step":                 "requirement_id_parse",
		}).Error("Failed to parse requirement_id parameter")

		errorResponse := &dtos.ErrorResponseDTO{
			Message: "requirement_id deve ser um número inteiro válido",
			Error:   "invalid_requirement_id",
		}
		c.JSON(http.StatusBadRequest, errorResponse)
		return
	}

	// Delete requirement
	err = h.UsecaseRequirement.DeleteRequirement(ctx, requirementID)
	if err != nil {
		h.Logger.WithFields(logrus.Fields{
			"requirement_id": requirementID,
			"error":          err.Error(),
			"correlation_id": correlationID,
			"step":           "requirement_deletion",
		}).Error("Failed to delete requirement")

		if contains(err.Error(), "failed to fetch requirement for deletion") {
			errorResponse := &dtos.ErrorResponseDTO{
				Message: "Requisito não encontrado",
				Error:   "requirement_not_found",
			}
			c.JSON(http.StatusNotFound, errorResponse)
			return
		}

		errorResponse := &dtos.ErrorResponseDTO{
			Message: "Falha ao deletar requisito",
			Error:   "deletion_failed",
		}
		c.JSON(http.StatusInternalServerError, errorResponse)
		return
	}

	h.Logger.WithFields(logrus.Fields{
		"requirement_id": requirementID,
		"correlation_id": correlationID,
		"step":           "requirement_deletion_success",
	}).Info("Requirement deleted successfully")

	c.Status(http.StatusNoContent)
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// extractValidationErrors extrai erros de validação do erro retornado pelo binding
func (h *RequirementHandlers) extractValidationErrors(err error) []dtos.ValidationErrorDetail {
	var validationErrors []dtos.ValidationErrorDetail

	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		for _, validationError := range validationErrs {
			detail := dtos.ValidationErrorDetail{
				Field:   validationError.Field(),
				Value:   validationError.Param(),
				Message: h.getValidationErrorMessage(validationError),
			}
			validationErrors = append(validationErrors, detail)
		}
	}

	return validationErrors
}

// getValidationErrorMessage retorna uma mensagem de erro personalizada baseada na tag de validação
func (h *RequirementHandlers) getValidationErrorMessage(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return "Campo obrigatório"
	case "oneof":
		return "Valor deve ser um dos seguintes: " + err.Param()
	case "min":
		return "Valor mínimo é " + err.Param()
	case "max":
		return "Valor máximo é " + err.Param()
	default:
		return "Valor inválido"
	}
}