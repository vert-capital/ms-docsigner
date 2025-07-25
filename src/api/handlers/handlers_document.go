package handlers

import (
	"app/api/handlers/dtos"
	"app/config"
	"app/entity"
	"app/infrastructure/clicksign"
	"app/infrastructure/repository"
	"app/pkg/utils"
	usecase_document "app/usecase/document"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type DocumentHandlers struct {
	UsecaseDocument usecase_document.IUsecaseDocument
	Logger          *logrus.Logger
}

func NewDocumentHandler(usecaseDocument usecase_document.IUsecaseDocument, logger *logrus.Logger) *DocumentHandlers {
	return &DocumentHandlers{
		UsecaseDocument: usecaseDocument,
		Logger:          logger,
	}
}

// @Summary Criar documento
// @Description Cria um novo documento usando file_path ou conteúdo base64
// @Description Aceita documentos através de file_path (caminho absoluto) ou file_content_base64 (conteúdo em base64)
// @Description Para file_path: file_size e mime_type são obrigatórios
// @Description Para file_content_base64: file_size e mime_type são opcionais (detectados automaticamente)
// @Description Tipos suportados: PDF, JPEG, PNG, GIF
// @Description Tamanho máximo: 7.5MB após decodificação
// @Tags Documents
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param document body dtos.DocumentCreateRequestDTO true "Dados do documento"
// @Success 201 {object} dtos.DocumentResponseDTO "Documento criado com sucesso"
// @Failure 400 {object} dtos.ErrorResponseDTO "Dados inválidos"
// @Failure 401 {object} dtos.ErrorResponseDTO "Não autorizado"
// @Failure 500 {object} dtos.ErrorResponseDTO "Erro interno"
// @Router /api/v1/documents [post]
func (h DocumentHandlers) CreateDocumentHandler(c *gin.Context) {
	correlationID := c.GetHeader("X-Correlation-ID")
	if correlationID == "" {
		correlationID = strconv.FormatInt(time.Now().Unix(), 10)
	}

	var requestDTO dtos.DocumentCreateRequestDTO

	if err := c.ShouldBindJSON(&requestDTO); err != nil {
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
		c.JSON(http.StatusBadRequest, dtos.ErrorResponseDTO{
			Error:   "Validation failed",
			Message: err.Error(),
		})
		return
	}

	document := &entity.EntityDocument{
		Name:        requestDTO.Name,
		Description: requestDTO.Description,
		Status:      "draft",
	}

	var tempPath string
	var err error

	// Processar base64 ou file_path
	if requestDTO.FileContentBase64 != "" {
		// Processar base64
		fileInfo, base64Err := utils.DecodeBase64File(requestDTO.FileContentBase64)
		if base64Err != nil {

			c.JSON(http.StatusBadRequest, dtos.ErrorResponseDTO{
				Error:   "Invalid base64",
				Message: base64Err.Error(),
			})
			return
		}

		// Validar MIME type
		if validateErr := utils.ValidateMimeType(fileInfo.MimeType); validateErr != nil {
			utils.CleanupTempFile(fileInfo.TempPath)

			c.JSON(http.StatusBadRequest, dtos.ErrorResponseDTO{
				Error:   "Unsupported file type",
				Message: validateErr.Error(),
			})
			return
		}

		document.FilePath = fileInfo.TempPath
		document.FileSize = fileInfo.Size
		document.MimeType = fileInfo.MimeType
		document.IsFromBase64 = true
		tempPath = fileInfo.TempPath

	} else {
		// Usar file_path tradicional
		document.FilePath = requestDTO.FilePath
		document.FileSize = requestDTO.FileSize
		document.MimeType = requestDTO.MimeType
		document.IsFromBase64 = false

	}

	// Limpar arquivo temporário em caso de erro ou sucesso
	if tempPath != "" {
		defer func() {
			if cleanupErr := utils.CleanupTempFile(tempPath); cleanupErr != nil {
				h.Logger.Warn("Failed to cleanup temporary file")
			} else {
				h.Logger.Debug("Temporary file cleaned up successfully")
			}
		}()
	}

	err = h.UsecaseDocument.Create(document)
	if err != nil {

		c.JSON(http.StatusInternalServerError, dtos.ErrorResponseDTO{
			Error:   "Internal server error",
			Message: "Failed to create document",
			Details: map[string]interface{}{
				"correlation_id": correlationID,
			},
		})
		return
	}

	responseDTO := h.mapEntityToResponse(document)

	jsonResponse(c, http.StatusCreated, responseDTO)
}

// @Summary Buscar documento por ID
// @Description Retorna um documento específico pelo ID
// @Tags Documents
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "ID do documento"
// @Success 200 {object} dtos.DocumentResponseDTO "Documento encontrado"
// @Failure 401 {object} dtos.ErrorResponseDTO "Não autorizado"
// @Failure 404 {object} dtos.ErrorResponseDTO "Documento não encontrado"
// @Failure 500 {object} dtos.ErrorResponseDTO "Erro interno"
// @Router /api/v1/documents/{id} [get]
func (h DocumentHandlers) GetDocumentHandler(c *gin.Context) {
	_ = c.GetHeader("X-Correlation-ID")

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {

		c.JSON(http.StatusBadRequest, dtos.ErrorResponseDTO{
			Error:   "Invalid ID",
			Message: "Document ID must be a valid integer",
		})
		return
	}

	document, err := h.UsecaseDocument.GetDocument(id)
	if err != nil {

		c.JSON(http.StatusNotFound, dtos.ErrorResponseDTO{
			Error:   "Document not found",
			Message: "The requested document does not exist",
		})
		return
	}

	responseDTO := h.mapEntityToResponse(document)

	jsonResponse(c, http.StatusOK, responseDTO)
}

// @Summary Listar documentos
// @Description Retorna uma lista de documentos com filtros opcionais
// @Tags Documents
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param search query string false "Buscar por nome"
// @Param status query string false "Filtrar por status"
// @Param clicksign_key query string false "Filtrar por chave Clicksign"
// @Success 200 {object} dtos.DocumentListResponseDTO "Lista de documentos"
// @Failure 401 {object} dtos.ErrorResponseDTO "Não autorizado"
// @Failure 500 {object} dtos.ErrorResponseDTO "Erro interno"
// @Router /api/v1/documents [get]
func (h DocumentHandlers) GetDocumentsHandler(c *gin.Context) {
	_ = c.GetHeader("X-Correlation-ID")

	var filters entity.EntityDocumentFilters
	filters.Search = c.Query("search")
	filters.Status = c.Query("status")
	filters.ClicksignKey = c.Query("clicksign_key")

	documents, err := h.UsecaseDocument.GetDocuments(filters)
	if err != nil {

		c.JSON(http.StatusInternalServerError, dtos.ErrorResponseDTO{
			Error:   "Internal server error",
			Message: "Failed to retrieve documents",
		})
		return
	}

	var responseDTOs []dtos.DocumentResponseDTO
	for _, document := range documents {
		responseDTOs = append(responseDTOs, h.mapEntityToResponse(&document))
	}

	responseDTO := dtos.DocumentListResponseDTO{
		Documents: responseDTOs,
		Total:     len(responseDTOs),
	}

	jsonResponse(c, http.StatusOK, responseDTO)
}

// @Summary Atualizar documento
// @Description Atualiza um documento existente
// @Tags Documents
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "ID do documento"
// @Param document body dtos.DocumentUpdateRequestDTO true "Dados para atualização"
// @Success 200 {object} dtos.DocumentResponseDTO "Documento atualizado"
// @Failure 400 {object} dtos.ErrorResponseDTO "Dados inválidos"
// @Failure 401 {object} dtos.ErrorResponseDTO "Não autorizado"
// @Failure 404 {object} dtos.ErrorResponseDTO "Documento não encontrado"
// @Failure 500 {object} dtos.ErrorResponseDTO "Erro interno"
// @Router /api/v1/documents/{id} [put]
func (h DocumentHandlers) UpdateDocumentHandler(c *gin.Context) {
	correlationID := c.GetHeader("X-Correlation-ID")
	if correlationID == "" {
		correlationID = strconv.FormatInt(time.Now().Unix(), 10)
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponseDTO{
			Error:   "Invalid ID",
			Message: "Document ID must be a valid integer",
		})
		return
	}

	var requestDTO dtos.DocumentUpdateRequestDTO
	if err := c.ShouldBindJSON(&requestDTO); err != nil {
		validationErrors := h.extractValidationErrors(err)
		c.JSON(http.StatusBadRequest, dtos.ValidationErrorResponseDTO{
			Error:   "Validation failed",
			Message: "Invalid request payload",
			Details: validationErrors,
		})
		return
	}

	document, err := h.UsecaseDocument.GetDocument(id)
	if err != nil {
		c.JSON(http.StatusNotFound, dtos.ErrorResponseDTO{
			Error:   "Document not found",
			Message: "The requested document does not exist",
		})
		return
	}

	if requestDTO.Name != nil {
		document.Name = *requestDTO.Name
	}
	if requestDTO.Description != nil {
		document.Description = *requestDTO.Description
	}
	if requestDTO.Status != nil {
		err := document.SetStatus(*requestDTO.Status)
		if err != nil {
			c.JSON(http.StatusBadRequest, dtos.ErrorResponseDTO{
				Error:   "Invalid status",
				Message: "Invalid status transition",
			})
			return
		}
	}

	err = h.UsecaseDocument.Update(document)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponseDTO{
			Error:   "Internal server error",
			Message: "Failed to update document",
			Details: map[string]interface{}{
				"correlation_id": correlationID,
			},
		})
		return
	}

	responseDTO := h.mapEntityToResponse(document)

	jsonResponse(c, http.StatusOK, responseDTO)
}

// @Summary Deletar documento
// @Description Remove um documento do sistema
// @Tags Documents
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "ID do documento"
// @Success 200 {object} map[string]string "Documento deletado com sucesso"
// @Failure 401 {object} dtos.ErrorResponseDTO "Não autorizado"
// @Failure 404 {object} dtos.ErrorResponseDTO "Documento não encontrado"
// @Failure 500 {object} dtos.ErrorResponseDTO "Erro interno"
// @Router /api/v1/documents/{id} [delete]
func (h DocumentHandlers) DeleteDocumentHandler(c *gin.Context) {
	correlationID := c.GetHeader("X-Correlation-ID")
	if correlationID == "" {
		correlationID = strconv.FormatInt(time.Now().Unix(), 10)
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponseDTO{
			Error:   "Invalid ID",
			Message: "Document ID must be a valid integer",
		})
		return
	}

	document, err := h.UsecaseDocument.GetDocument(id)
	if err != nil {
		c.JSON(http.StatusNotFound, dtos.ErrorResponseDTO{
			Error:   "Document not found",
			Message: "The requested document does not exist",
		})
		return
	}

	err = h.UsecaseDocument.Delete(document)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponseDTO{
			Error:   "Internal server error",
			Message: "Failed to delete document",
			Details: map[string]interface{}{
				"correlation_id": correlationID,
			},
		})
		return
	}

	jsonResponse(c, http.StatusOK, gin.H{"message": "Documento deletado com sucesso"})
}

// Helper methods

func (h DocumentHandlers) mapEntityToResponse(document *entity.EntityDocument) dtos.DocumentResponseDTO {
	return dtos.DocumentResponseDTO{
		ID:           document.ID,
		Name:         document.Name,
		FilePath:     document.FilePath,
		FileSize:     document.FileSize,
		MimeType:     document.MimeType,
		Status:       document.Status,
		ClicksignKey: document.ClicksignKey,
		Description:  document.Description,
		CreatedAt:    document.CreatedAt,
		UpdatedAt:    document.UpdatedAt,
	}
}

func (h DocumentHandlers) extractValidationErrors(err error) []dtos.ValidationErrorDetail {
	var validationErrors []dtos.ValidationErrorDetail

	if validationErr, ok := err.(validator.ValidationErrors); ok {
		for _, fieldError := range validationErr {
			validationErrors = append(validationErrors, dtos.ValidationErrorDetail{
				Field:   fieldError.Field(),
				Message: h.getValidationErrorMessage(fieldError),
				Value:   fieldError.Value().(string),
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

func (h DocumentHandlers) getValidationErrorMessage(fieldError validator.FieldError) string {
	switch fieldError.Tag() {
	case "required":
		return "This field is required"
	case "min":
		return "This field must have at least " + fieldError.Param() + " characters/items"
	case "max":
		return "This field must have at most " + fieldError.Param() + " characters/items"
	case "gt":
		return "This field must be greater than " + fieldError.Param()
	default:
		return "This field is invalid"
	}
}

func MountDocumentHandlers(gin *gin.Engine, conn *gorm.DB, logger *logrus.Logger) {
	// Inicializar cliente Clicksign
	clicksignClient := clicksign.NewClicksignClient(config.EnvironmentVariables, logger)

	documentHandlers := NewDocumentHandler(
		usecase_document.NewUsecaseDocumentServiceWithClicksign(
			repository.NewRepositoryDocument(conn),
			clicksignClient,
			logger,
		),
		logger,
	)

	group := gin.Group("/api/v1/documents")
	SetAuthMiddleware(conn, group)

	group.POST("/", documentHandlers.CreateDocumentHandler)
	group.GET("/:id", documentHandlers.GetDocumentHandler)
	group.GET("/", documentHandlers.GetDocumentsHandler)
	group.PUT("/:id", documentHandlers.UpdateDocumentHandler)
	group.DELETE("/:id", documentHandlers.DeleteDocumentHandler)
}
