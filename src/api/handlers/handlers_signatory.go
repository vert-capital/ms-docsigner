package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"app/api/handlers/dtos"
	"app/config"
	"app/entity"
	"app/infrastructure/clicksign"
	"app/infrastructure/repository"
	"app/usecase/document"
	"app/usecase/envelope"
	"app/usecase/signatory"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type SignatoryHandlers struct {
	UsecaseSignatory signatory.IUsecaseSignatory
	UsecaseEnvelope  envelope.IUsecaseEnvelope
	Logger           *logrus.Logger
}

func NewSignatoryHandler(
	usecaseSignatory signatory.IUsecaseSignatory,
	usecaseEnvelope envelope.IUsecaseEnvelope,
	logger *logrus.Logger,
) *SignatoryHandlers {
	return &SignatoryHandlers{
		UsecaseSignatory: usecaseSignatory,
		UsecaseEnvelope:  usecaseEnvelope,
		Logger:           logger,
	}
}

// @Summary Create signatory
// @Description Create a new signatory for an envelope
// @Tags signatories
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Envelope ID"
// @Param request body dtos.SignatoryCreateRequestDTO true "Signatory data"
// @Success 201 {object} dtos.SignatoryResponseDTO
// @Failure 400 {object} dtos.ValidationErrorResponseDTO
// @Failure 404 {object} dtos.ErrorResponseDTO
// @Failure 500 {object} dtos.ErrorResponseDTO
// @Router /api/v1/envelopes/{id}/signatories [post]
func (h *SignatoryHandlers) CreateSignatoryHandler(c *gin.Context) {
	correlationID := c.GetHeader("X-Correlation-ID")
	if correlationID == "" {
		correlationID = strconv.FormatInt(time.Now().Unix(), 10)
	}

	envelopeIDStr := c.Param("id")
	envelopeID, err := strconv.Atoi(envelopeIDStr)
	if err != nil {
		h.Logger.WithFields(logrus.Fields{
			"error":          err.Error(),
			"envelope_id":    envelopeIDStr,
			"correlation_id": correlationID,
		}).Error("Invalid envelope ID")

		c.JSON(http.StatusBadRequest, dtos.ErrorResponseDTO{
			Error:   "Invalid ID",
			Message: "Envelope ID must be a valid integer",
		})
		return
	}

	h.Logger.WithFields(logrus.Fields{
		"correlation_id": correlationID,
		"endpoint":       "POST /api/v1/envelopes/{id}/signatories",
		"envelope_id":    envelopeID,
	}).Info("Creating signatory request received")

	var requestDTO dtos.SignatoryCreateRequestDTO

	if err := c.ShouldBindJSON(&requestDTO); err != nil {
		h.Logger.WithFields(logrus.Fields{
			"error":          err.Error(),
			"correlation_id": correlationID,
			"envelope_id":    envelopeID,
		}).Error("Invalid request payload")

		validationErrors := h.extractValidationErrors(err)
		c.JSON(http.StatusBadRequest, dtos.ValidationErrorResponseDTO{
			Error:   "Validation failed",
			Message: "Invalid request payload",
			Details: validationErrors,
		})
		return
	}

	// Definir envelope_id do path parameter
	requestDTO.EnvelopeID = envelopeID

	// Validação customizada do DTO
	if err := requestDTO.Validate(); err != nil {
		h.Logger.WithFields(logrus.Fields{
			"error":          err.Error(),
			"correlation_id": correlationID,
			"envelope_id":    envelopeID,
		}).Error("Custom validation failed")

		c.JSON(http.StatusBadRequest, dtos.ErrorResponseDTO{
			Error:   "Validation failed",
			Message: err.Error(),
		})
		return
	}

	// Verificar se o envelope existe
	_, err = h.UsecaseEnvelope.GetEnvelope(envelopeID)
	if err != nil {
		h.Logger.WithFields(logrus.Fields{
			"error":          err.Error(),
			"envelope_id":    envelopeID,
			"correlation_id": correlationID,
		}).Error("Envelope not found")

		c.JSON(http.StatusNotFound, dtos.ErrorResponseDTO{
			Error:   "Envelope not found",
			Message: "The specified envelope does not exist",
		})
		return
	}

	// Converter DTO para entidade
	signatoryEntity := requestDTO.ToEntity()

	// Criar signatário através do use case
	createdSignatory, err := h.UsecaseSignatory.CreateSignatory(&signatoryEntity)
	if err != nil {
		h.Logger.WithFields(logrus.Fields{
			"error":           err.Error(),
			"envelope_id":     envelopeID,
			"signatory_email": requestDTO.Email,
			"correlation_id":  correlationID,
		}).Error("Failed to create signatory")

		c.JSON(http.StatusInternalServerError, dtos.ErrorResponseDTO{
			Error:   "Internal server error",
			Message: "Failed to create signatory",
			Details: map[string]interface{}{
				"correlation_id": correlationID,
			},
		})
		return
	}

	// Converter entidade para DTO de resposta
	responseDTO := h.mapEntityToResponse(createdSignatory)

	h.Logger.WithFields(logrus.Fields{
		"signatory_id":    createdSignatory.ID,
		"envelope_id":     envelopeID,
		"signatory_email": createdSignatory.Email,
		"correlation_id":  correlationID,
	}).Info("Signatory created successfully")

	c.JSON(http.StatusCreated, responseDTO)
}

// @Summary Get signatories by envelope
// @Description Get list of signatories for a specific envelope
// @Tags signatories
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Envelope ID"
// @Success 200 {object} dtos.SignatoryListResponseDTO
// @Failure 400 {object} dtos.ErrorResponseDTO
// @Failure 404 {object} dtos.ErrorResponseDTO
// @Failure 500 {object} dtos.ErrorResponseDTO
// @Router /api/v1/envelopes/{id}/signatories [get]
func (h *SignatoryHandlers) GetSignatoriesHandler(c *gin.Context) {
	correlationID := c.GetHeader("X-Correlation-ID")
	if correlationID == "" {
		correlationID = strconv.FormatInt(time.Now().Unix(), 10)
	}

	envelopeIDStr := c.Param("id")
	envelopeID, err := strconv.Atoi(envelopeIDStr)
	if err != nil {
		h.Logger.WithFields(logrus.Fields{
			"error":          err.Error(),
			"envelope_id":    envelopeIDStr,
			"correlation_id": correlationID,
		}).Error("Invalid envelope ID")

		c.JSON(http.StatusBadRequest, dtos.ErrorResponseDTO{
			Error:   "Invalid ID",
			Message: "Envelope ID must be a valid integer",
		})
		return
	}

	h.Logger.WithFields(logrus.Fields{
		"correlation_id": correlationID,
		"endpoint":       "GET /api/v1/envelopes/{id}/signatories",
		"envelope_id":    envelopeID,
	}).Info("Getting signatories request received")

	// Verificar se o envelope existe
	_, err = h.UsecaseEnvelope.GetEnvelope(envelopeID)
	if err != nil {
		h.Logger.WithFields(logrus.Fields{
			"error":          err.Error(),
			"envelope_id":    envelopeID,
			"correlation_id": correlationID,
		}).Error("Envelope not found")

		c.JSON(http.StatusNotFound, dtos.ErrorResponseDTO{
			Error:   "Envelope not found",
			Message: "The specified envelope does not exist",
		})
		return
	}

	// Buscar signatários do envelope
	signatories, err := h.UsecaseSignatory.GetSignatoriesByEnvelope(envelopeID)
	if err != nil {
		h.Logger.WithFields(logrus.Fields{
			"error":          err.Error(),
			"envelope_id":    envelopeID,
			"correlation_id": correlationID,
		}).Error("Failed to get signatories")

		c.JSON(http.StatusInternalServerError, dtos.ErrorResponseDTO{
			Error:   "Internal server error",
			Message: "Failed to retrieve signatories",
		})
		return
	}

	responseDTO := h.mapSignatoryListToResponse(signatories)

	h.Logger.WithFields(logrus.Fields{
		"envelope_id":        envelopeID,
		"signatories_count":  len(signatories),
		"correlation_id":     correlationID,
	}).Info("Signatories retrieved successfully")

	c.JSON(http.StatusOK, responseDTO)
}

// @Summary Get signatory
// @Description Get signatory by ID
// @Tags signatories
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Signatory ID"
// @Success 200 {object} dtos.SignatoryResponseDTO
// @Failure 400 {object} dtos.ErrorResponseDTO
// @Failure 404 {object} dtos.ErrorResponseDTO
// @Failure 500 {object} dtos.ErrorResponseDTO
// @Router /api/v1/signatories/{id} [get]
func (h *SignatoryHandlers) GetSignatoryHandler(c *gin.Context) {
	correlationID := c.GetHeader("X-Correlation-ID")
	if correlationID == "" {
		correlationID = strconv.FormatInt(time.Now().Unix(), 10)
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.Logger.WithFields(logrus.Fields{
			"error":          err.Error(),
			"id":             idStr,
			"correlation_id": correlationID,
		}).Error("Invalid signatory ID")

		c.JSON(http.StatusBadRequest, dtos.ErrorResponseDTO{
			Error:   "Invalid ID",
			Message: "Signatory ID must be a valid integer",
		})
		return
	}

	h.Logger.WithFields(logrus.Fields{
		"signatory_id":   id,
		"correlation_id": correlationID,
		"endpoint":       "GET /api/v1/signatories/{id}",
	}).Info("Getting signatory request received")

	signatory, err := h.UsecaseSignatory.GetSignatory(id)
	if err != nil {
		h.Logger.WithFields(logrus.Fields{
			"error":          err.Error(),
			"signatory_id":   id,
			"correlation_id": correlationID,
		}).Error("Failed to get signatory")

		c.JSON(http.StatusNotFound, dtos.ErrorResponseDTO{
			Error:   "Signatory not found",
			Message: "The requested signatory does not exist",
		})
		return
	}

	responseDTO := h.mapEntityToResponse(signatory)

	h.Logger.WithFields(logrus.Fields{
		"signatory_id":    signatory.ID,
		"signatory_email": signatory.Email,
		"envelope_id":     signatory.EnvelopeID,
		"correlation_id":  correlationID,
	}).Info("Signatory retrieved successfully")

	c.JSON(http.StatusOK, responseDTO)
}

// @Summary Update signatory
// @Description Update signatory information
// @Tags signatories
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Signatory ID"
// @Param request body dtos.SignatoryUpdateRequestDTO true "Signatory update data"
// @Success 200 {object} dtos.SignatoryResponseDTO
// @Failure 400 {object} dtos.ValidationErrorResponseDTO
// @Failure 404 {object} dtos.ErrorResponseDTO
// @Failure 500 {object} dtos.ErrorResponseDTO
// @Router /api/v1/signatories/{id} [put]
func (h *SignatoryHandlers) UpdateSignatoryHandler(c *gin.Context) {
	correlationID := c.GetHeader("X-Correlation-ID")
	if correlationID == "" {
		correlationID = strconv.FormatInt(time.Now().Unix(), 10)
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.Logger.WithFields(logrus.Fields{
			"error":          err.Error(),
			"id":             idStr,
			"correlation_id": correlationID,
		}).Error("Invalid signatory ID")

		c.JSON(http.StatusBadRequest, dtos.ErrorResponseDTO{
			Error:   "Invalid ID",
			Message: "Signatory ID must be a valid integer",
		})
		return
	}

	h.Logger.WithFields(logrus.Fields{
		"signatory_id":   id,
		"correlation_id": correlationID,
		"endpoint":       "PUT /api/v1/signatories/{id}",
	}).Info("Updating signatory request received")

	var requestDTO dtos.SignatoryUpdateRequestDTO

	if err := c.ShouldBindJSON(&requestDTO); err != nil {
		h.Logger.WithFields(logrus.Fields{
			"error":          err.Error(),
			"signatory_id":   id,
			"correlation_id": correlationID,
		}).Error("Invalid request payload")

		validationErrors := h.extractValidationErrors(err)
		c.JSON(http.StatusBadRequest, dtos.ValidationErrorResponseDTO{
			Error:   "Validation failed",
			Message: "Invalid request payload",
			Details: validationErrors,
		})
		return
	}

	// Validação customizada do DTO
	if err := requestDTO.Validate(); err != nil {
		h.Logger.WithFields(logrus.Fields{
			"error":          err.Error(),
			"signatory_id":   id,
			"correlation_id": correlationID,
		}).Error("Custom validation failed")

		c.JSON(http.StatusBadRequest, dtos.ErrorResponseDTO{
			Error:   "Validation failed",
			Message: err.Error(),
		})
		return
	}

	// Buscar signatário existente
	existingSignatory, err := h.UsecaseSignatory.GetSignatory(id)
	if err != nil {
		h.Logger.WithFields(logrus.Fields{
			"error":          err.Error(),
			"signatory_id":   id,
			"correlation_id": correlationID,
		}).Error("Signatory not found")

		c.JSON(http.StatusNotFound, dtos.ErrorResponseDTO{
			Error:   "Signatory not found",
			Message: "The requested signatory does not exist",
		})
		return
	}

	// Aplicar mudanças do DTO na entidade
	requestDTO.ApplyToEntity(existingSignatory)

	// Atualizar signatário através do use case
	err = h.UsecaseSignatory.UpdateSignatory(existingSignatory)
	if err != nil {
		h.Logger.WithFields(logrus.Fields{
			"error":          err.Error(),
			"signatory_id":   id,
			"correlation_id": correlationID,
		}).Error("Failed to update signatory")

		c.JSON(http.StatusInternalServerError, dtos.ErrorResponseDTO{
			Error:   "Internal server error",
			Message: "Failed to update signatory",
			Details: map[string]interface{}{
				"correlation_id": correlationID,
			},
		})
		return
	}

	// Converter entidade para DTO de resposta
	responseDTO := h.mapEntityToResponse(existingSignatory)

	h.Logger.WithFields(logrus.Fields{
		"signatory_id":    existingSignatory.ID,
		"signatory_email": existingSignatory.Email,
		"envelope_id":     existingSignatory.EnvelopeID,
		"correlation_id":  correlationID,
	}).Info("Signatory updated successfully")

	c.JSON(http.StatusOK, responseDTO)
}

// @Summary Delete signatory
// @Description Delete signatory by ID
// @Tags signatories
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Signatory ID"
// @Success 204 "No Content"
// @Failure 400 {object} dtos.ErrorResponseDTO
// @Failure 404 {object} dtos.ErrorResponseDTO
// @Failure 500 {object} dtos.ErrorResponseDTO
// @Router /api/v1/signatories/{id} [delete]
func (h *SignatoryHandlers) DeleteSignatoryHandler(c *gin.Context) {
	correlationID := c.GetHeader("X-Correlation-ID")
	if correlationID == "" {
		correlationID = strconv.FormatInt(time.Now().Unix(), 10)
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.Logger.WithFields(logrus.Fields{
			"error":          err.Error(),
			"id":             idStr,
			"correlation_id": correlationID,
		}).Error("Invalid signatory ID")

		c.JSON(http.StatusBadRequest, dtos.ErrorResponseDTO{
			Error:   "Invalid ID",
			Message: "Signatory ID must be a valid integer",
		})
		return
	}

	h.Logger.WithFields(logrus.Fields{
		"signatory_id":   id,
		"correlation_id": correlationID,
		"endpoint":       "DELETE /api/v1/signatories/{id}",
	}).Info("Deleting signatory request received")

	// Verificar se o signatário existe antes de tentar deletar
	_, err = h.UsecaseSignatory.GetSignatory(id)
	if err != nil {
		h.Logger.WithFields(logrus.Fields{
			"error":          err.Error(),
			"signatory_id":   id,
			"correlation_id": correlationID,
		}).Error("Signatory not found")

		c.JSON(http.StatusNotFound, dtos.ErrorResponseDTO{
			Error:   "Signatory not found",
			Message: "The requested signatory does not exist",
		})
		return
	}

	// Deletar signatário através do use case
	err = h.UsecaseSignatory.DeleteSignatory(id)
	if err != nil {
		h.Logger.WithFields(logrus.Fields{
			"error":          err.Error(),
			"signatory_id":   id,
			"correlation_id": correlationID,
		}).Error("Failed to delete signatory")

		c.JSON(http.StatusInternalServerError, dtos.ErrorResponseDTO{
			Error:   "Internal server error",
			Message: "Failed to delete signatory",
			Details: map[string]interface{}{
				"correlation_id": correlationID,
			},
		})
		return
	}

	h.Logger.WithFields(logrus.Fields{
		"signatory_id":   id,
		"correlation_id": correlationID,
	}).Info("Signatory deleted successfully")

	c.Status(http.StatusNoContent)
}

// @Summary Send signatories to Clicksign
// @Description Send envelope signatories to Clicksign for processing
// @Tags signatories
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Envelope ID"
// @Success 200 {object} dtos.SignatoryListResponseDTO
// @Failure 400 {object} dtos.ErrorResponseDTO
// @Failure 404 {object} dtos.ErrorResponseDTO
// @Failure 500 {object} dtos.ErrorResponseDTO
// @Router /api/v1/envelopes/{id}/send [post]
func (h *SignatoryHandlers) SendSignatoriesToClicksignHandler(c *gin.Context) {
	correlationID := c.GetHeader("X-Correlation-ID")
	if correlationID == "" {
		correlationID = strconv.FormatInt(time.Now().Unix(), 10)
	}

	envelopeIDStr := c.Param("id")
	envelopeID, err := strconv.Atoi(envelopeIDStr)
	if err != nil {
		h.Logger.WithFields(logrus.Fields{
			"error":          err.Error(),
			"envelope_id":    envelopeIDStr,
			"correlation_id": correlationID,
		}).Error("Invalid envelope ID")

		c.JSON(http.StatusBadRequest, dtos.ErrorResponseDTO{
			Error:   "Invalid ID",
			Message: "Envelope ID must be a valid integer",
		})
		return
	}

	h.Logger.WithFields(logrus.Fields{
		"correlation_id": correlationID,
		"endpoint":       "POST /api/v1/envelopes/{id}/send",
		"envelope_id":    envelopeID,
	}).Info("Sending signatories to Clicksign request received")

	// Verificar se o envelope existe e obter suas informações
	envelope, err := h.UsecaseEnvelope.GetEnvelope(envelopeID)
	if err != nil {
		h.Logger.WithFields(logrus.Fields{
			"error":          err.Error(),
			"envelope_id":    envelopeID,
			"correlation_id": correlationID,
		}).Error("Envelope not found")

		c.JSON(http.StatusNotFound, dtos.ErrorResponseDTO{
			Error:   "Envelope not found",
			Message: "The specified envelope does not exist",
		})
		return
	}

	// Validar se o envelope pode ser enviado (deve ter ClicksignKey)
	if envelope.ClicksignKey == "" {
		h.Logger.WithFields(logrus.Fields{
			"envelope_id":    envelopeID,
			"envelope_status": envelope.Status,
			"correlation_id": correlationID,
		}).Error("Envelope not ready for sending - missing Clicksign key")

		c.JSON(http.StatusBadRequest, dtos.ErrorResponseDTO{
			Error:   "Envelope not ready",
			Message: "Envelope must be created in Clicksign before sending signatories",
		})
		return
	}

	// Buscar signatários do envelope
	signatories, err := h.UsecaseSignatory.GetSignatoriesByEnvelope(envelopeID)
	if err != nil {
		h.Logger.WithFields(logrus.Fields{
			"error":          err.Error(),
			"envelope_id":    envelopeID,
			"correlation_id": correlationID,
		}).Error("Failed to get signatories for envelope")

		c.JSON(http.StatusInternalServerError, dtos.ErrorResponseDTO{
			Error:   "Internal server error",
			Message: "Failed to retrieve signatories for envelope",
		})
		return
	}

	if len(signatories) == 0 {
		h.Logger.WithFields(logrus.Fields{
			"envelope_id":    envelopeID,
			"correlation_id": correlationID,
		}).Warn("No signatories found for envelope")

		c.JSON(http.StatusBadRequest, dtos.ErrorResponseDTO{
			Error:   "No signatories",
			Message: "Envelope must have at least one signatory before sending",
		})
		return
	}

	// Criar o cliente Clicksign e SignerService
	clicksignClient := clicksign.NewClicksignClient(config.EnvironmentVariables, h.Logger)
	signerService := clicksign.NewSignerService(clicksignClient, h.Logger)
	signatoryMapper := clicksign.NewSignatoryMapper()

	// Enviar cada signatário para o Clicksign
	var successCount int
	var errors []string
	
	for _, signatory := range signatories {
		h.Logger.WithFields(logrus.Fields{
			"signatory_id":    signatory.ID,
			"signatory_email": signatory.Email,
			"envelope_id":     envelopeID,
			"clicksign_key":   envelope.ClicksignKey,
			"correlation_id":  correlationID,
		}).Info("Sending signatory to Clicksign")

		// Mapear entidade para estrutura Clicksign
		createRequest := signatoryMapper.ToClicksignCreateRequest(&signatory)
		
		// Converter para SignerData (estrutura esperada pelo SignerService)
		signerData := clicksign.SignerData{
			Name:             createRequest.Data.Attributes.Name,
			Email:            createRequest.Data.Attributes.Email,
			Birthday:         createRequest.Data.Attributes.Birthday,
			PhoneNumber:      createRequest.Data.Attributes.PhoneNumber,
			HasDocumentation: createRequest.Data.Attributes.HasDocumentation,
			Refusable:        createRequest.Data.Attributes.Refusable,
			Group:            createRequest.Data.Attributes.Group,
		}

		if createRequest.Data.Attributes.CommunicateEvents != nil {
			signerData.CommunicateEvents = &clicksign.SignerCommunicateEventsData{
				DocumentSigned:    createRequest.Data.Attributes.CommunicateEvents.DocumentSigned,
				SignatureRequest:  createRequest.Data.Attributes.CommunicateEvents.SignatureRequest,
				SignatureReminder: createRequest.Data.Attributes.CommunicateEvents.SignatureReminder,
			}
		}

		// Criar contexto com correlation ID
		ctx := context.WithValue(context.Background(), "correlation_id", correlationID)
		
		// Enviar para Clicksign
		clicksignSignerID, err := signerService.CreateSigner(ctx, envelope.ClicksignKey, signerData)
		if err != nil {
			errorMsg := fmt.Sprintf("Failed to send signatory %s (%s) to Clicksign: %v", signatory.Name, signatory.Email, err)
			errors = append(errors, errorMsg)
			
			h.Logger.WithFields(logrus.Fields{
				"error":           err.Error(),
				"signatory_id":    signatory.ID,
				"signatory_email": signatory.Email,
				"envelope_id":     envelopeID,
				"clicksign_key":   envelope.ClicksignKey,
				"correlation_id":  correlationID,
			}).Error("Failed to send signatory to Clicksign")
			continue
		}

		successCount++
		h.Logger.WithFields(logrus.Fields{
			"signatory_id":       signatory.ID,
			"signatory_email":    signatory.Email,
			"clicksign_signer_id": clicksignSignerID,
			"envelope_id":        envelopeID,
			"clicksign_key":      envelope.ClicksignKey,
			"correlation_id":     correlationID,
		}).Info("Signatory sent to Clicksign successfully")
	}

	// Preparar resposta
	responseDTO := h.mapSignatoryListToResponse(signatories)
	
	if len(errors) > 0 {
		// Houveram erros parciais
		h.Logger.WithFields(logrus.Fields{
			"envelope_id":        envelopeID,
			"total_signatories":  len(signatories),
			"successful_sends":   successCount,
			"failed_sends":       len(errors),
			"correlation_id":     correlationID,
		}).Warn("Signatories sent to Clicksign with partial failures")

		c.JSON(http.StatusOK, map[string]interface{}{
			"signatories":     responseDTO.Signatories,
			"total":           responseDTO.Total,
			"successful_sends": successCount,
			"failed_sends":    len(errors),
			"errors":          errors,
		})
		return
	}

	// Todos os signatários foram enviados com sucesso
	h.Logger.WithFields(logrus.Fields{
		"envelope_id":        envelopeID,
		"total_signatories":  len(signatories),
		"successful_sends":   successCount,
		"correlation_id":     correlationID,
	}).Info("All signatories sent to Clicksign successfully")

	c.JSON(http.StatusOK, map[string]interface{}{
		"signatories":     responseDTO.Signatories,
		"total":           responseDTO.Total,
		"successful_sends": successCount,
		"message":         "All signatories sent to Clicksign successfully",
	})
}

// Helper methods

func (h *SignatoryHandlers) mapEntityToResponse(signatory *entity.EntitySignatory) *dtos.SignatoryResponseDTO {
	var responseDTO dtos.SignatoryResponseDTO
	responseDTO.FromEntity(signatory)
	return &responseDTO
}

func (h *SignatoryHandlers) mapSignatoryListToResponse(signatories []entity.EntitySignatory) *dtos.SignatoryListResponseDTO {
	signatoryList := make([]dtos.SignatoryResponseDTO, len(signatories))
	for i, signatory := range signatories {
		signatoryList[i] = *h.mapEntityToResponse(&signatory)
	}

	return &dtos.SignatoryListResponseDTO{
		Signatories: signatoryList,
		Total:       len(signatories),
	}
}

func (h *SignatoryHandlers) extractValidationErrors(err error) []dtos.ValidationErrorDetail {
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

func (h *SignatoryHandlers) getValidationErrorMessage(fieldError validator.FieldError) string {
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

func MountSignatoryHandlers(gin *gin.Engine, conn *gorm.DB, logger *logrus.Logger) {
	// Criar clientes e repositórios
	clicksignClient := clicksign.NewClicksignClient(config.EnvironmentVariables, logger)
	
	// Criar usecase de documento para envelopes
	usecaseDocument := document.NewUsecaseDocumentServiceWithClicksign(
		repository.NewRepositoryDocument(conn),
		clicksignClient,
		logger,
	)

	// Importar as dependências necessárias
	signatoryHandlers := NewSignatoryHandler(
		signatory.NewUsecaseSignatoryService(
			repository.NewRepositorySignatory(conn),
			repository.NewRepositoryEnvelope(conn),
			logger,
		),
		envelope.NewUsecaseEnvelopeService(
			repository.NewRepositoryEnvelope(conn),
			clicksignClient,
			usecaseDocument,
			logger,
		),
		logger,
	)

	// Rotas para signatários por envelope (usando :id para consistência com envelope handlers)
	envelopeGroup := gin.Group("/api/v1/envelopes")
	SetAuthMiddleware(conn, envelopeGroup)
	envelopeGroup.POST("/:id/signatories", signatoryHandlers.CreateSignatoryHandler)
	envelopeGroup.GET("/:id/signatories", signatoryHandlers.GetSignatoriesHandler)
	envelopeGroup.POST("/:id/send", signatoryHandlers.SendSignatoriesToClicksignHandler)

	// Rotas para signatários individuais
	signatoryGroup := gin.Group("/api/v1/signatories")
	SetAuthMiddleware(conn, signatoryGroup)
	signatoryGroup.GET("/:id", signatoryHandlers.GetSignatoryHandler)
	signatoryGroup.PUT("/:id", signatoryHandlers.UpdateSignatoryHandler)
	signatoryGroup.DELETE("/:id", signatoryHandlers.DeleteSignatoryHandler)
}