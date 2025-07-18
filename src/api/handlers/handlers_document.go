package handlers

import (
	"app/api/handlers/dtos"
	"app/entity"
	"app/infrastructure/repository"
	usecase_document "app/usecase/document"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type DocumentHandlers struct {
	UsecaseDocument usecase_document.IUsecaseDocument
}

func NewDocumentHandler(usecaseDocument usecase_document.IUsecaseDocument) *DocumentHandlers {
	return &DocumentHandlers{UsecaseDocument: usecaseDocument}
}

// @Summary Criar documento
// @Description Cria um novo documento
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
	var requestDTO dtos.DocumentCreateRequestDTO

	if err := c.ShouldBindJSON(&requestDTO); err != nil {
		handleError(c, err)
		return
	}

	document := &entity.EntityDocument{
		Name:        requestDTO.Name,
		FilePath:    requestDTO.FilePath,
		FileSize:    requestDTO.FileSize,
		MimeType:    requestDTO.MimeType,
		Description: requestDTO.Description,
		Status:      "draft",
	}

	err := h.UsecaseDocument.Create(document)
	if exception := handleError(c, err); exception {
		return
	}

	responseDTO := dtos.DocumentResponseDTO{
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
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		handleError(c, err)
		return
	}

	document, err := h.UsecaseDocument.GetDocument(id)
	if exception := handleError(c, err); exception {
		return
	}

	responseDTO := dtos.DocumentResponseDTO{
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
	var filters entity.EntityDocumentFilters

	filters.Search = c.Query("search")
	filters.Status = c.Query("status")
	filters.ClicksignKey = c.Query("clicksign_key")

	documents, err := h.UsecaseDocument.GetDocuments(filters)
	if exception := handleError(c, err); exception {
		return
	}

	var responseDTOs []dtos.DocumentResponseDTO
	for _, document := range documents {
		responseDTOs = append(responseDTOs, dtos.DocumentResponseDTO{
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
		})
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
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		handleError(c, err)
		return
	}

	var requestDTO dtos.DocumentUpdateRequestDTO
	if err := c.ShouldBindJSON(&requestDTO); err != nil {
		handleError(c, err)
		return
	}

	document, err := h.UsecaseDocument.GetDocument(id)
	if exception := handleError(c, err); exception {
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
			handleError(c, err)
			return
		}
	}

	err = h.UsecaseDocument.Update(document)
	if exception := handleError(c, err); exception {
		return
	}

	responseDTO := dtos.DocumentResponseDTO{
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
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		handleError(c, err)
		return
	}

	document, err := h.UsecaseDocument.GetDocument(id)
	if exception := handleError(c, err); exception {
		return
	}

	err = h.UsecaseDocument.Delete(document)
	if exception := handleError(c, err); exception {
		return
	}

	jsonResponse(c, http.StatusOK, gin.H{"message": "Documento deletado com sucesso"})
}

func MountDocumentHandlers(gin *gin.Engine, conn *gorm.DB) {
	documentHandlers := NewDocumentHandler(
		usecase_document.NewUsecaseDocumentService(
			repository.NewRepositoryDocument(conn),
		),
	)

	group := gin.Group("/api/v1/documents")
	SetAuthMiddleware(conn, group)

	group.POST("/", documentHandlers.CreateDocumentHandler)
	group.GET("/:id", documentHandlers.GetDocumentHandler)
	group.GET("/", documentHandlers.GetDocumentsHandler)
	group.PUT("/:id", documentHandlers.UpdateDocumentHandler)
	group.DELETE("/:id", documentHandlers.DeleteDocumentHandler)
}
