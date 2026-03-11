package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"app/api/handlers/dtos"
	"app/config"
	"app/entity"
	"app/infrastructure/clicksign"
	"app/infrastructure/repository"
	"app/infrastructure/vertc_assinaturas"
	usecase_auto_signature_term "app/usecase/auto_signature_term"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type AutoSignatureTermHandlers struct {
	UsecaseAutoSignatureTerm usecase_auto_signature_term.IUsecaseAutoSignatureTerm
	VertcAutomaticSignature  *vertc_assinaturas.AutomaticSignatureService
	Logger                   *logrus.Logger
}

func NewAutoSignatureTermHandler(
	usecaseAutoSignatureTerm usecase_auto_signature_term.IUsecaseAutoSignatureTerm,
	vertcAutomaticSignature *vertc_assinaturas.AutomaticSignatureService,
	logger *logrus.Logger,
) *AutoSignatureTermHandlers {
	return &AutoSignatureTermHandlers{
		UsecaseAutoSignatureTerm: usecaseAutoSignatureTerm,
		VertcAutomaticSignature:  vertcAutomaticSignature,
		Logger:                   logger,
	}
}

// @Summary Create auto signature term
// @Description Create a new auto signature term in Clicksign. This endpoint creates a term that allows automatic signature for a specific signer.
// @Tags auto-signature-terms
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param provider query string false "Provider name (clicksign|vert-sign). Defaults to clicksign"
// @Param request body dtos.AutoSignatureTermCreateRequestDTO true "Auto signature term data"
// @Success 201 {object} dtos.AutoSignatureTermResponseDTO "Auto signature term created successfully"
// @Failure 400 {object} dtos.ValidationErrorResponseDTO "Validation error - invalid request data"
// @Failure 409 {object} dtos.ErrorResponseDTO "Conflict - term already exists for this signer and operator"
// @Failure 500 {object} dtos.ErrorResponseDTO "Internal server error - term creation failed"
// @Router /api/v1/auto-signature/terms [post]
func (h *AutoSignatureTermHandlers) CreateAutoSignatureTermHandler(c *gin.Context) {
	providerName, isValidProvider := normalizeAutoSignatureProviderWithDefault(c.Query("provider"))
	if !isValidProvider {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponseDTO{
			Error:   "Invalid provider",
			Message: fmt.Sprintf("Unsupported provider: %s. Supported providers: clicksign, vert-sign", c.Query("provider")),
		})
		return
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

	if providerName == "vert-sign" {
		h.createAutoSignatureTermForVertSign(c, requestDTO)
		return
	}

	validationDetails := validateClicksignCreateTermRequest(requestDTO)
	if len(validationDetails) > 0 {
		c.JSON(http.StatusBadRequest, dtos.ValidationErrorResponseDTO{
			Error:   "Validation failed",
			Message: "Missing required fields for clicksign provider",
			Details: validationDetails,
		})
		return
	}

	h.createAutoSignatureTermForClicksign(c, requestDTO)
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

// @Summary Check auto signature term status by provider and signer email
// @Description Check if the signer already has an active and signed auto signature permission in the selected provider.
// @Tags auto-signature-terms
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param provider query string true "Provider name (clicksign|vert-sign)"
// @Param email query string true "Signer e-mail"
// @Success 200 {object} dtos.AutoSignatureTermStatusResponseDTO "Auto signature status for the signer and provider"
// @Failure 400 {object} dtos.ValidationErrorResponseDTO "Validation error - invalid query params"
// @Failure 501 {object} dtos.ErrorResponseDTO "Provider does not support e-mail based check"
// @Failure 502 {object} dtos.ErrorResponseDTO "Provider integration error"
// @Router /api/v1/auto-signature/terms/status [get]
func (h *AutoSignatureTermHandlers) CheckAutoSignatureTermStatusHandler(c *gin.Context) {
	var queryDTO dtos.AutoSignatureTermStatusQueryDTO
	if err := c.ShouldBindQuery(&queryDTO); err != nil {
		validationErrors := h.extractValidationErrors(err)
		c.JSON(http.StatusBadRequest, dtos.ValidationErrorResponseDTO{
			Error:   "Validation failed",
			Message: "Invalid query params",
			Details: validationErrors,
		})
		return
	}

	providerName, isValidProvider := normalizeAutoSignatureProvider(queryDTO.Provider)
	if !isValidProvider {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponseDTO{
			Error:   "Invalid provider",
			Message: fmt.Sprintf("Unsupported provider: %s. Supported providers: clicksign, vert-sign", queryDTO.Provider),
			Details: map[string]interface{}{
				"provider": queryDTO.Provider,
			},
		})
		return
	}

	if providerName == "clicksign" {
		c.JSON(http.StatusNotImplemented, dtos.ErrorResponseDTO{
			Error:   "Provider not supported for e-mail validation",
			Message: "Clicksign não disponibiliza rota para validação de termo de assinatura automática por e-mail.",
			Details: map[string]interface{}{
				"provider": providerName,
				"email":    queryDTO.Email,
			},
		})
		return
	}

	if h.VertcAutomaticSignature == nil {
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponseDTO{
			Error:   "Internal server error",
			Message: "VertSign auto signature service is not configured",
		})
		return
	}

	result, err := h.VertcAutomaticSignature.CheckSignedTermByEmail(c.Request.Context(), queryDTO.Email)
	if err != nil {
		h.Logger.WithError(err).WithFields(logrus.Fields{
			"provider": providerName,
			"email":    queryDTO.Email,
		}).Error("Failed to check auto signature status")

		c.JSON(http.StatusBadGateway, dtos.ErrorResponseDTO{
			Error:   "Provider integration error",
			Message: "Failed to check auto signature status in provider",
			Details: map[string]interface{}{
				"provider": providerName,
				"email":    queryDTO.Email,
			},
		})
		return
	}

	response := dtos.AutoSignatureTermStatusResponseDTO{
		Provider:        providerName,
		Email:           queryDTO.Email,
		HasSignedTerm:   result.HasSignedTerm,
		PermissionFound: result.PermissionFound,
		PermissionID:    result.PermissionID,
		ContractStatus:  result.ContractStatus,
		IsActive:        result.IsActive,
	}

	c.JSON(http.StatusOK, response)
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

func (h *AutoSignatureTermHandlers) createAutoSignatureTermForClicksign(c *gin.Context, requestDTO dtos.AutoSignatureTermCreateRequestDTO) {
	term, err := h.mapCreateRequestToEntity(requestDTO)
	if err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponseDTO{
			Error:   "Invalid request",
			Message: err.Error(),
		})
		return
	}

	createdTerm, err := h.UsecaseAutoSignatureTerm.CreateAutoSignatureTerm(c.Request.Context(), term)
	if err != nil {
		h.Logger.WithError(err).Error("Failed to create auto signature term in clicksign")

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

	responseDTO := h.mapEntityToResponse(createdTerm)
	c.JSON(http.StatusCreated, responseDTO)
}

func (h *AutoSignatureTermHandlers) createAutoSignatureTermForVertSign(c *gin.Context, requestDTO dtos.AutoSignatureTermCreateRequestDTO) {
	if h.VertcAutomaticSignature == nil {
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponseDTO{
			Error:   "Internal server error",
			Message: "VertSign auto signature service is not configured",
		})
		return
	}

	result, err := h.VertcAutomaticSignature.CreateTermEnsuringUser(c.Request.Context(), requestDTO.Signer.Email, requestDTO.Signer.Name)
	if err != nil {
		h.Logger.WithError(err).WithFields(logrus.Fields{
			"provider": "vert-sign",
			"email":    requestDTO.Signer.Email,
		}).Error("Failed to create auto signature term in vert-sign")

		if vertc_assinaturas.IsAutomaticSignaturePermissionAlreadyExistsError(err) {
			c.JSON(http.StatusConflict, dtos.ErrorResponseDTO{
				Error:   "Term already exists",
				Message: "Já existe uma permissão ativa de assinatura automática para este signatário no VertSign.",
			})
			return
		}

		c.JSON(http.StatusBadGateway, dtos.ErrorResponseDTO{
			Error:   "Provider integration error",
			Message: "Failed to create auto signature term in VertSign",
			Details: map[string]interface{}{
				"provider": "vert-sign",
				"email":    requestDTO.Signer.Email,
			},
		})
		return
	}

	response := dtos.AutoSignatureTermProviderCreateResponseDTO{
		Provider:         "vert-sign",
		Email:            requestDTO.Signer.Email,
		PermissionID:     result.PermissionID,
		EnvelopeID:       result.EnvelopeID,
		ContractStatus:   result.ContractStatus,
		IsActive:         result.IsActive,
		NotificationSent: result.NotificationSent,
		UserCreated:      result.UserCreated,
		UserExisted:      result.UserExisted,
	}

	if result.NotificationError != nil {
		response.NotificationError = *result.NotificationError
	}

	c.JSON(http.StatusCreated, response)
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

func normalizeAutoSignatureProvider(provider string) (string, bool) {
	normalizedProvider := strings.ToLower(strings.TrimSpace(provider))

	switch normalizedProvider {
	case "clicksign":
		return "clicksign", true
	case "vert-sign", "vert_sign", "vertsign", "vertc-assinaturas", "vertc_assinaturas":
		return "vert-sign", true
	default:
		return "", false
	}
}

func normalizeAutoSignatureProviderWithDefault(provider string) (string, bool) {
	normalizedProvider := strings.TrimSpace(provider)
	if normalizedProvider == "" {
		return "clicksign", true
	}

	return normalizeAutoSignatureProvider(normalizedProvider)
}

func validateClicksignCreateTermRequest(requestDTO dtos.AutoSignatureTermCreateRequestDTO) []dtos.ValidationErrorDetail {
	var details []dtos.ValidationErrorDetail

	if strings.TrimSpace(requestDTO.AdminEmail) == "" {
		details = append(details, dtos.ValidationErrorDetail{
			Field:   "admin_email",
			Message: "This field is required for clicksign",
		})
	}

	if strings.TrimSpace(requestDTO.APIEmail) == "" {
		details = append(details, dtos.ValidationErrorDetail{
			Field:   "api_email",
			Message: "This field is required for clicksign",
		})
	}

	if strings.TrimSpace(requestDTO.Signer.Documentation) == "" {
		details = append(details, dtos.ValidationErrorDetail{
			Field:   "signer.documentation",
			Message: "This field is required for clicksign",
		})
	}

	if strings.TrimSpace(requestDTO.Signer.Birthday) == "" {
		details = append(details, dtos.ValidationErrorDetail{
			Field:   "signer.birthday",
			Message: "This field is required for clicksign",
		})
	}

	if strings.TrimSpace(requestDTO.Signer.Name) == "" {
		details = append(details, dtos.ValidationErrorDetail{
			Field:   "signer.name",
			Message: "This field is required for clicksign",
		})
	}

	return details
}

func MountAutoSignatureTermHandlers(gin *gin.Engine, conn *gorm.DB, logger *logrus.Logger) {
	clicksignClient := clicksign.NewClicksignClient(config.EnvironmentVariables, logger).(*clicksign.ClicksignClient)
	vertcAssinaturasClient := vertc_assinaturas.NewVertcAssinaturasClient(config.EnvironmentVariables, logger)

	autoSignatureTermHandlers := NewAutoSignatureTermHandler(
		usecase_auto_signature_term.NewUsecaseAutoSignatureTermService(
			repository.NewRepositoryAutoSignatureTerm(conn),
			clicksignClient,
			logger,
		),
		vertc_assinaturas.NewAutomaticSignatureService(vertcAssinaturasClient, logger),
		logger,
	)

	group := gin.Group("/api/v1/auto-signature/terms")
	SetAuthMiddleware(conn, group)

	group.POST("/", autoSignatureTermHandlers.CreateAutoSignatureTermHandler)
	group.GET("/status", autoSignatureTermHandlers.CheckAutoSignatureTermStatusHandler)
	group.GET("/:id", autoSignatureTermHandlers.GetAutoSignatureTermHandler)
	group.GET("/", autoSignatureTermHandlers.GetAllAutoSignatureTermsHandler)
	group.DELETE("/:id", autoSignatureTermHandlers.DeleteAutoSignatureTermHandler)
}
