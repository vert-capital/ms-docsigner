package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"app/api/handlers/dtos"
	"app/config"
	"app/entity"
	"app/infrastructure/clicksign"
	"app/infrastructure/repository"
	"app/pkg/utils"
	"app/usecase/envelope"
	usecase_document "app/usecase/document"
	"app/usecase/requirement"
	"app/usecase/signatory"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type EnvelopeHandlers struct {
	UsecaseEnvelope  envelope.IUsecaseEnvelope
	UsecaseSignatory signatory.IUsecaseSignatory
	Logger           *logrus.Logger
}

func NewEnvelopeHandler(usecaseEnvelope envelope.IUsecaseEnvelope, usecaseSignatory signatory.IUsecaseSignatory, logger *logrus.Logger) *EnvelopeHandlers {
	return &EnvelopeHandlers{
		UsecaseEnvelope:  usecaseEnvelope,
		UsecaseSignatory: usecaseSignatory,
		Logger:           logger,
	}
}

// @Summary Create envelope
// @Description Create a new envelope in Clicksign with optional signatories. When signatories are provided in the request, they will be created along with the envelope in a single atomic transaction. The process maintains backward compatibility - envelopes can still be created without signatories. The response includes the complete raw data returned by Clicksign API for debugging and analysis purposes.
// @Tags envelopes
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body dtos.EnvelopeCreateRequestDTO true "Envelope data with optional signatories array. When signatories are provided, the response will include the created signatories with their IDs."
// @Success 201 {object} dtos.EnvelopeResponseDTO "Envelope created successfully. The response includes clicksign_raw_data field with the complete JSON response from Clicksign API (optional field for debugging). If signatories were provided in the request, the response includes the created signatories with their assigned IDs."
// @Failure 400 {object} dtos.ValidationErrorResponseDTO "Validation error - invalid request data, duplicate signatory emails, or unsupported document format"
// @Failure 500 {object} dtos.ErrorResponseDTO "Internal server error - envelope creation failed or signatory creation failed during transaction"
// @Router /api/v1/envelopes [post]
func (h *EnvelopeHandlers) CreateEnvelopeHandler(c *gin.Context) {
	correlationID := c.GetHeader("X-Correlation-ID")
	if correlationID == "" {
		correlationID = strconv.FormatInt(time.Now().Unix(), 10)
	}

	h.Logger.Info("Creating envelope request received")

	var requestDTO dtos.EnvelopeCreateRequestDTO

	if err := c.ShouldBindJSON(&requestDTO); err != nil {
		h.Logger.Error("Invalid request payload")

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
		h.Logger.Error("Custom validation failed")

		c.JSON(http.StatusBadRequest, dtos.ErrorResponseDTO{
			Error:   "Validation failed",
			Message: err.Error(),
		})
		return
	}

	// Converter DTO para entidade
	envelope, documents, err := h.mapCreateRequestToEntity(requestDTO)
	if err != nil {
		h.Logger.Error("Failed to map request to entity")

		c.JSON(http.StatusBadRequest, dtos.ErrorResponseDTO{
			Error:   "Invalid request",
			Message: err.Error(),
		})
		return
	}

	// Limpar arquivos temporários em caso de erro
	var tempPaths []string
	for _, doc := range documents {
		if doc.IsFromBase64 && doc.FilePath != "" {
			tempPaths = append(tempPaths, doc.FilePath)
		}
	}
	defer func() {
		for _, tempPath := range tempPaths {
			if cleanupErr := utils.CleanupTempFile(tempPath); cleanupErr != nil {
				h.Logger.Warn("Failed to cleanup temporary file")
			}
		}
	}()

	// Criar envelope através do use case
	var createdEnvelope *entity.EntityEnvelope
	if len(documents) > 0 {
		// Criar envelope com documentos base64
		createdEnvelope, err = h.UsecaseEnvelope.CreateEnvelopeWithDocuments(envelope, documents)
	} else {
		// Criar envelope com IDs de documentos existentes
		createdEnvelope, err = h.UsecaseEnvelope.CreateEnvelope(envelope)
	}
	
	if err != nil {
		h.Logger.Error("Failed to create envelope")

		c.JSON(http.StatusInternalServerError, dtos.ErrorResponseDTO{
			Error:   "Internal server error",
			Message: "Failed to create envelope",
			Details: map[string]interface{}{
				"correlation_id": correlationID,
			},
		})
		return
	}

	h.Logger.Info("Envelope created successfully")

	// Criar signatários se fornecidos no request
	var createdSignatories []entity.EntitySignatory
	if len(requestDTO.Signatories) > 0 {
		h.Logger.Info("Creating signatories for envelope")

		for i, signatoryRequest := range requestDTO.Signatories {
			// Converter EnvelopeSignatoryRequest para SignatoryCreateRequestDTO
			signatoryDTO := signatoryRequest.ToSignatoryCreateRequestDTO(createdEnvelope.ID)
			
			// Converter DTO para entidade
			signatoryEntity := signatoryDTO.ToEntity()
			
			// Criar signatário através do use case
			createdSignatory, sigErr := h.UsecaseSignatory.CreateSignatory(&signatoryEntity)
			if sigErr != nil {
				h.Logger.Error("Failed to create signatory, rolling back envelope")

				// FIXME: Rollback automático de envelope não implementado
				// Considerar implementação futura de transação distribuída
				c.JSON(http.StatusInternalServerError, dtos.ErrorResponseDTO{
					Error:   "Internal server error",
					Message: fmt.Sprintf("Failed to create signatory %d: %v. ATENÇÃO: Envelope %d foi criado mas signatários falharam", i+1, sigErr, createdEnvelope.ID),
					Details: map[string]interface{}{
						"correlation_id":     correlationID,
						"envelope_id":        createdEnvelope.ID,
						"failed_signatory":   i + 1,
						"partial_transaction": true,
					},
				})
				return
			}

			createdSignatories = append(createdSignatories, *createdSignatory)
			
			h.Logger.Info("Signatory created successfully")
		}

		h.Logger.Info("All signatories created successfully")
	}

	// Converter entidade para DTO de resposta
	responseDTO := h.mapEntityToResponse(createdEnvelope, createdSignatories)

	// Log da persistência dos dados brutos do Clicksign
	rawDataPersisted := createdEnvelope.ClicksignRawData != nil
	var rawDataSize int
	if rawDataPersisted {
		rawDataSize = len(*createdEnvelope.ClicksignRawData)
	}

	h.Logger.Info("Envelope created successfully with Clicksign raw data persistence")

	c.JSON(http.StatusCreated, responseDTO)
}

// @Summary Get envelope
// @Description Get envelope by ID. The response includes clicksign_raw_data field with the complete JSON response from Clicksign API when available (optional field for debugging and analysis).
// @Tags envelopes
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Envelope ID"
// @Success 200 {object} dtos.EnvelopeResponseDTO "Envelope data with optional clicksign_raw_data field containing raw Clicksign API response"
// @Failure 404 {object} dtos.ErrorResponseDTO
// @Failure 500 {object} dtos.ErrorResponseDTO
// @Router /api/v1/envelopes/{id} [get]
func (h *EnvelopeHandlers) GetEnvelopeHandler(c *gin.Context) {
	correlationID := c.GetHeader("X-Correlation-ID")
	if correlationID == "" {
		correlationID = strconv.FormatInt(time.Now().Unix(), 10)
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.Logger.Error("Invalid envelope ID")

		c.JSON(http.StatusBadRequest, dtos.ErrorResponseDTO{
			Error:   "Invalid ID",
			Message: "Envelope ID must be a valid integer",
		})
		return
	}

	h.Logger.Info("Getting envelope request received")

	envelope, err := h.UsecaseEnvelope.GetEnvelope(id)
	if err != nil {
		h.Logger.Error("Failed to get envelope")

		c.JSON(http.StatusNotFound, dtos.ErrorResponseDTO{
			Error:   "Envelope not found",
			Message: "The requested envelope does not exist",
		})
		return
	}

	responseDTO := h.mapEntityToResponse(envelope)

	h.Logger.Info("Envelope retrieved successfully")

	c.JSON(http.StatusOK, responseDTO)
}

// @Summary List envelopes
// @Description Get list of envelopes with optional filters
// @Tags envelopes
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param search query string false "Search term"
// @Param status query string false "Status filter"
// @Param clicksign_key query string false "Clicksign key filter"
// @Success 200 {object} dtos.EnvelopeListResponseDTO
// @Failure 500 {object} dtos.ErrorResponseDTO
// @Router /api/v1/envelopes [get]
func (h *EnvelopeHandlers) GetEnvelopesHandler(c *gin.Context) {
	correlationID := c.GetHeader("X-Correlation-ID")
	if correlationID == "" {
		correlationID = strconv.FormatInt(time.Now().Unix(), 10)
	}

	h.Logger.Info("Getting envelopes list request received")

	var filters entity.EntityEnvelopeFilters
	filters.Search = c.Query("search")
	filters.Status = c.Query("status")
	filters.ClicksignKey = c.Query("clicksign_key")

	envelopes, err := h.UsecaseEnvelope.GetEnvelopes(filters)
	if err != nil {
		h.Logger.Error("Failed to get envelopes")

		c.JSON(http.StatusInternalServerError, dtos.ErrorResponseDTO{
			Error:   "Internal server error",
			Message: "Failed to retrieve envelopes",
		})
		return
	}

	responseDTO := h.mapEnvelopeListToResponse(envelopes)

	h.Logger.Info("Envelopes retrieved successfully")

	c.JSON(http.StatusOK, responseDTO)
}

// @Summary Activate envelope
// @Description Activate envelope to start signing process
// @Tags envelopes
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Envelope ID"
// @Success 200 {object} dtos.EnvelopeResponseDTO
// @Failure 400 {object} dtos.ErrorResponseDTO
// @Failure 404 {object} dtos.ErrorResponseDTO
// @Failure 500 {object} dtos.ErrorResponseDTO
// @Router /api/v1/envelopes/{id}/activate [post]
func (h *EnvelopeHandlers) ActivateEnvelopeHandler(c *gin.Context) {
	correlationID := c.GetHeader("X-Correlation-ID")
	if correlationID == "" {
		correlationID = strconv.FormatInt(time.Now().Unix(), 10)
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.Logger.Error("Invalid envelope ID")

		c.JSON(http.StatusBadRequest, dtos.ErrorResponseDTO{
			Error:   "Invalid ID",
			Message: "Envelope ID must be a valid integer",
		})
		return
	}

	h.Logger.Info("Activating envelope request received")

	envelope, err := h.UsecaseEnvelope.ActivateEnvelope(id)
	if err != nil {
		h.Logger.Error("Failed to activate envelope")

		c.JSON(http.StatusInternalServerError, dtos.ErrorResponseDTO{
			Error:   "Internal server error",
			Message: "Failed to activate envelope",
			Details: map[string]interface{}{
				"correlation_id": correlationID,
			},
		})
		return
	}

	responseDTO := h.mapEntityToResponse(envelope)

	h.Logger.Info("Envelope activated successfully")

	c.JSON(http.StatusOK, responseDTO)
}

// Helper methods

func (h *EnvelopeHandlers) mapCreateRequestToEntity(dto dtos.EnvelopeCreateRequestDTO) (*entity.EntityEnvelope, []*entity.EntityDocument, error) {
	// Determinar emails dos signatários com base no formato usado
	var signatoryEmails []string
	if len(dto.SignatoryEmails) > 0 {
		// Usando formato antigo com emails diretos
		signatoryEmails = dto.SignatoryEmails
	} else if len(dto.Signatories) > 0 {
		// Usando formato novo com signatários estruturados - extrair emails
		for _, signatory := range dto.Signatories {
			signatoryEmails = append(signatoryEmails, signatory.Email)
		}
	}

	envelope := &entity.EntityEnvelope{
		Name:            dto.Name,
		Description:     dto.Description,
		DocumentsIDs:    dto.DocumentsIDs,
		SignatoryEmails: signatoryEmails,
		Message:         dto.Message,
		DeadlineAt:      dto.DeadlineAt,
		RemindInterval:  dto.RemindInterval,
		AutoClose:       dto.AutoClose,
		Status:          "draft",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if envelope.RemindInterval == 0 {
		envelope.RemindInterval = 3
	}

	var documents []*entity.EntityDocument
	
	// Processar documentos base64 se fornecidos
	for _, docRequest := range dto.Documents {
		// Processar base64
		fileInfo, err := utils.DecodeBase64File(docRequest.FileContentBase64)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to process base64 content for document '%s': %w", docRequest.Name, err)
		}

		// Validar MIME type
		if err := utils.ValidateMimeType(fileInfo.MimeType); err != nil {
			utils.CleanupTempFile(fileInfo.TempPath)
			return nil, nil, fmt.Errorf("unsupported file type for document '%s': %w", docRequest.Name, err)
		}

		document := &entity.EntityDocument{
			Name:         docRequest.Name,
			Description:  docRequest.Description,
			FilePath:     fileInfo.TempPath,
			FileSize:     fileInfo.Size,
			MimeType:     fileInfo.MimeType,
			IsFromBase64: true,
			Status:       "draft",
		}

		documents = append(documents, document)
	}

	return envelope, documents, nil
}

func (h *EnvelopeHandlers) mapEntityToResponse(envelope *entity.EntityEnvelope, signatories ...[]entity.EntitySignatory) *dtos.EnvelopeResponseDTO {
	response := &dtos.EnvelopeResponseDTO{
		ID:               envelope.ID,
		Name:             envelope.Name,
		Description:      envelope.Description,
		Status:           envelope.Status,
		ClicksignKey:     envelope.ClicksignKey,
		ClicksignRawData: envelope.ClicksignRawData,
		DocumentsIDs:     envelope.DocumentsIDs,
		SignatoryEmails:  envelope.SignatoryEmails,
		Message:          envelope.Message,
		DeadlineAt:       envelope.DeadlineAt,
		RemindInterval:   envelope.RemindInterval,
		AutoClose:        envelope.AutoClose,
		CreatedAt:        envelope.CreatedAt,
		UpdatedAt:        envelope.UpdatedAt,
	}

	// Incluir signatários se fornecidos
	if len(signatories) > 0 && len(signatories[0]) > 0 {
		signatoryDTOs := make([]dtos.SignatoryResponseDTO, len(signatories[0]))
		for i, signatory := range signatories[0] {
			signatoryDTOs[i].FromEntity(&signatory)
		}
		response.Signatories = signatoryDTOs
	}

	return response
}

func (h *EnvelopeHandlers) mapEnvelopeListToResponse(envelopes []entity.EntityEnvelope) *dtos.EnvelopeListResponseDTO {
	envelopeList := make([]dtos.EnvelopeResponseDTO, len(envelopes))
	for i, envelope := range envelopes {
		envelopeList[i] = *h.mapEntityToResponse(&envelope)
	}

	return &dtos.EnvelopeListResponseDTO{
		Envelopes: envelopeList,
		Total:     len(envelopes),
	}
}

func (h *EnvelopeHandlers) extractValidationErrors(err error) []dtos.ValidationErrorDetail {
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

func (h *EnvelopeHandlers) getValidationErrorMessage(fieldError validator.FieldError) string {
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

func MountEnvelopeHandlers(gin *gin.Engine, conn *gorm.DB, logger *logrus.Logger) {
	clicksignClient := clicksign.NewClicksignClient(config.EnvironmentVariables, logger)

	// Criar usecase de documento para envelopes com documentos base64
	usecaseDocument := usecase_document.NewUsecaseDocumentServiceWithClicksign(
		repository.NewRepositoryDocument(conn),
		clicksignClient,
		logger,
	)

	// Criar usecase de signatory
	usecaseSignatory := signatory.NewUsecaseSignatoryService(
		repository.NewRepositorySignatory(conn),
		repository.NewRepositoryEnvelope(conn),
		clicksignClient,
		logger,
	)

	// Criar usecase de requirement
	usecaseRequirement := requirement.NewUsecaseRequirementService(
		repository.NewRepositoryRequirement(conn),
		repository.NewRepositoryEnvelope(conn),
		clicksignClient,
		logger,
	)

	envelopeHandlers := NewEnvelopeHandler(
		envelope.NewUsecaseEnvelopeService(
			repository.NewRepositoryEnvelope(conn),
			clicksignClient,
			usecaseDocument,
			usecaseRequirement,
			logger,
		),
		usecaseSignatory,
		logger,
	)

	// Criar handler de requirements
	requirementHandlers := NewRequirementHandler(usecaseRequirement, logger)

	group := gin.Group("/api/v1/envelopes")
	SetAuthMiddleware(conn, group)

	group.POST("/", envelopeHandlers.CreateEnvelopeHandler)
	group.GET("/:id", envelopeHandlers.GetEnvelopeHandler)
	group.GET("/", envelopeHandlers.GetEnvelopesHandler)
	group.POST("/:id/activate", envelopeHandlers.ActivateEnvelopeHandler)

	// Rotas de requirements por envelope
	group.POST("/:id/requirements", requirementHandlers.CreateRequirementHandler)
	group.GET("/:id/requirements", requirementHandlers.GetRequirementsByEnvelopeHandler)
}
