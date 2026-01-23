package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"app/api/handlers/dtos"
	"app/config"
	"app/entity"
	"app/infrastructure/clicksign"
	"app/infrastructure/provider"
	"app/infrastructure/provider_factory"
	"app/infrastructure/repository"
	"app/pkg/utils"
	"app/usecase/document"
	usecase_envelope "app/usecase/envelope"
	"app/usecase/requirement"
	"app/usecase/signatory"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// EnvelopeV2Handlers gerencia handlers para a rota v2 de envelopes
type EnvelopeV2Handlers struct {
	ProviderFactory       *provider_factory.ProviderFactory
	UsecaseDocuments      document.IUsecaseDocument
	UsecaseRequirement    requirement.IUsecaseRequirement
	UsecaseSignatory      signatory.IUsecaseSignatory
	RepositoryEnvelope    usecase_envelope.IRepositoryEnvelope
	RepositorySignatory   signatory.IRepositorySignatory
	RepositoryRequirement requirement.IRepositoryRequirement
	Logger                *logrus.Logger
}

// NewEnvelopeV2Handler cria uma nova instância do EnvelopeV2Handlers
func NewEnvelopeV2Handler(
	providerFactory *provider_factory.ProviderFactory,
	usecaseDocuments document.IUsecaseDocument,
	usecaseRequirement requirement.IUsecaseRequirement,
	usecaseSignatory signatory.IUsecaseSignatory,
	repositoryEnvelope usecase_envelope.IRepositoryEnvelope,
	repositorySignatory signatory.IRepositorySignatory,
	repositoryRequirement requirement.IRepositoryRequirement,
	logger *logrus.Logger,
) *EnvelopeV2Handlers {
	return &EnvelopeV2Handlers{
		ProviderFactory:       providerFactory,
		UsecaseDocuments:      usecaseDocuments,
		UsecaseRequirement:    usecaseRequirement,
		UsecaseSignatory:      usecaseSignatory,
		RepositoryEnvelope:    repositoryEnvelope,
		RepositorySignatory:   repositorySignatory,
		RepositoryRequirement: repositoryRequirement,
		Logger:                logger,
	}
}

// @Summary Create envelope (v2)
// @Description Create a new envelope with provider selection. Supports multiple providers (clicksign, vertc-assinaturas). The provider field is required.
// @Tags envelopes-v2
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body dtos.EnvelopeV2CreateRequestDTO true "Envelope data with provider field"
// @Success 201 {object} dtos.EnvelopeResponseDTO "Envelope created successfully"
// @Failure 400 {object} dtos.ValidationErrorResponseDTO "Validation error or invalid provider"
// @Failure 501 {object} dtos.ErrorResponseDTO "Provider not implemented"
// @Failure 500 {object} dtos.ErrorResponseDTO "Internal server error"
// @Router /api/v2/envelopes [post]
func (h *EnvelopeV2Handlers) CreateEnvelopeV2Handler(c *gin.Context) {
	correlationID := c.GetHeader("X-Correlation-ID")
	if correlationID == "" {
		correlationID = strconv.FormatInt(time.Now().Unix(), 10)
	}

	var requestDTO dtos.EnvelopeV2CreateRequestDTO

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

	// Validar e obter provider
	envelopeProvider, err := h.ProviderFactory.GetProvider(requestDTO.Provider)
	if err != nil {
		// Verificar se é provider não implementado ou inválido
		if !h.ProviderFactory.IsProviderSupported(requestDTO.Provider) {
			c.JSON(http.StatusBadRequest, dtos.ErrorResponseDTO{
				Error:   "Invalid provider",
				Message: err.Error(),
				Details: map[string]interface{}{
					"correlation_id": correlationID,
					"provider":       requestDTO.Provider,
				},
			})
			return
		}

		// Provider suportado mas não implementado
		c.JSON(http.StatusNotImplemented, dtos.ErrorResponseDTO{
			Error:   "Provider not implemented",
			Message: err.Error(),
			Details: map[string]interface{}{
				"correlation_id": correlationID,
				"provider":       requestDTO.Provider,
			},
		})
		return
	}

	// Converter DTO para entidade (reutilizar lógica do handler v1)
	envelope, documents, err := h.mapCreateRequestToEntityV2(requestDTO)
	if err != nil {
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

	// Criar use case com provider
	envelopeProviderService := usecase_envelope.NewUsecaseEnvelopeProviderService(
		h.RepositoryEnvelope,
		envelopeProvider,
		h.UsecaseDocuments,
		h.UsecaseRequirement,
		h.Logger,
	)

	// Criar envelope através do use case
	createdEnvelope, err := envelopeProviderService.CreateEnvelope(envelope)
	if err != nil {
		status := http.StatusInternalServerError
		var ce *clicksign.ClicksignError
		if errors.As(err, &ce) && ce.StatusCode > 0 {
			status = ce.StatusCode
		}
		c.JSON(status, dtos.ErrorResponseDTO{
			Error:   http.StatusText(status),
			Message: "Failed to create envelope: " + err.Error(),
			Details: map[string]interface{}{
				"correlation_id": correlationID,
				"provider":       requestDTO.Provider,
			},
		})
		return
	}

	// Criar documentos base64 se fornecidos
	if len(documents) > 0 {
		for _, doc := range documents {
			err := h.UsecaseDocuments.Create(doc)
			if err != nil {
				c.JSON(http.StatusInternalServerError, dtos.ErrorResponseDTO{
					Error:   "Internal server error",
					Message: fmt.Sprintf("Failed to create document '%s': %v", doc.Name, err),
					Details: map[string]interface{}{
						"correlation_id": correlationID,
						"provider":       requestDTO.Provider,
					},
				})
				return
			}
			// Adicionar documento ao envelope criado
			createdEnvelope.DocumentsIDs = append(createdEnvelope.DocumentsIDs, doc.ID)

			// Enviar documento para o provider e obter a chave
			doc.ClicksignKey, err = envelopeProviderService.CreateDocument(
				c.Request.Context(),
				createdEnvelope.ClicksignKey,
				doc,
				createdEnvelope.ID,
			)

			if err != nil {
				c.JSON(http.StatusInternalServerError, dtos.ErrorResponseDTO{
					Error:   "Internal server error",
					Message: fmt.Sprintf("Failed to upload document '%s' to provider: %v", doc.Name, err),
					Details: map[string]interface{}{
						"correlation_id": correlationID,
						"provider":       requestDTO.Provider,
					},
				})
				return
			}

			// Atualizar documento no banco com a chave do provider
			err = h.UsecaseDocuments.Update(doc)
			if err != nil {
				c.JSON(http.StatusInternalServerError, dtos.ErrorResponseDTO{
					Error:   "Internal server error",
					Message: fmt.Sprintf("Failed to update document '%s' with provider key: %v", doc.Name, err),
					Details: map[string]interface{}{
						"correlation_id": correlationID,
						"provider":       requestDTO.Provider,
					},
				})
				return
			}
		}

		// Atualizar envelope no banco com os IDs dos documentos
		err = envelopeProviderService.UpdateEnvelope(createdEnvelope)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dtos.ErrorResponseDTO{
				Error:   "Internal server error",
				Message: fmt.Sprintf("Failed to update envelope with document IDs: %v", err),
				Details: map[string]interface{}{
					"correlation_id": correlationID,
					"provider":       requestDTO.Provider,
				},
			})
			return
		}
	}

	// Criar signatários se fornecidos no request
	var createdSignatories []entity.EntitySignatory
	if len(requestDTO.Signatories) > 0 {
		for i, signatoryRequest := range requestDTO.Signatories {
			// Converter EnvelopeSignatoryRequest para SignatoryCreateRequestDTO
			signatoryDTO := signatoryRequest.ToSignatoryCreateRequestDTO(createdEnvelope.ID)

			// Converter DTO para entidade
			signatoryEntity := signatoryDTO.ToEntity()

			// Validar entidade antes de criar
			if err := signatoryEntity.Validate(); err != nil {
				c.JSON(http.StatusBadRequest, dtos.ErrorResponseDTO{
					Error:   "Validation failed",
					Message: fmt.Sprintf("Signatory %d validation failed: %v", i+1, err),
					Details: map[string]interface{}{
						"correlation_id": correlationID,
						"provider":       requestDTO.Provider,
					},
				})
				return
			}

			// Criar signatário localmente primeiro (sem chamar provider via usecase)
			// Usar repository diretamente para evitar acoplamento com Clicksign
			if err := h.RepositorySignatory.Create(&signatoryEntity); err != nil {
				c.JSON(http.StatusInternalServerError, dtos.ErrorResponseDTO{
					Error:   "Internal server error",
					Message: fmt.Sprintf("Failed to create signatory %d locally: %v", i+1, err),
					Details: map[string]interface{}{
						"correlation_id":      correlationID,
						"envelope_id":         createdEnvelope.ID,
						"failed_signatory":    i + 1,
						"partial_transaction": true,
						"provider":            requestDTO.Provider,
					},
				})
				return
			}

			// Obter envelope para pegar a chave do provider
			envelope, err := h.RepositoryEnvelope.GetByID(createdEnvelope.ID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, dtos.ErrorResponseDTO{
					Error:   "Internal server error",
					Message: fmt.Sprintf("Failed to get envelope: %v", err),
					Details: map[string]interface{}{
						"correlation_id": correlationID,
						"provider":       requestDTO.Provider,
					},
				})
				return
			}

			// Mapear para SignerData do provider
			// Valores padrão conforme EntitySignatory.NewSignatory
			defaultGroup := 1
			defaultHasDoc := false
			defaultRefusable := true

			signerData := provider.SignerData{
				Name:              signatoryEntity.Name,
				Email:             signatoryEntity.Email,
				Birthday:          "",
				HasDocumentation:  defaultHasDoc,
				Refusable:         defaultRefusable,
				Group:             defaultGroup,
			}

			if signatoryEntity.Birthday != nil {
				signerData.Birthday = *signatoryEntity.Birthday
			}
			if signatoryEntity.Documentation != nil {
				signerData.Documentation = signatoryEntity.Documentation
			}
			if signatoryEntity.PhoneNumber != nil {
				signerData.PhoneNumber = signatoryEntity.PhoneNumber
			}
			if signatoryEntity.HasDocumentation != nil {
				signerData.HasDocumentation = *signatoryEntity.HasDocumentation
			}
			if signatoryEntity.Refusable != nil {
				signerData.Refusable = *signatoryEntity.Refusable
			}
			// Group deve ser maior que 0 (Clicksign requirement)
			// Se não fornecido ou 0, usar padrão 1
			if signatoryEntity.Group != nil && *signatoryEntity.Group > 0 {
				signerData.Group = *signatoryEntity.Group
			} else {
				signerData.Group = defaultGroup
			}

			// Criar signatário no provider
			providerSignerKey, err := envelopeProvider.CreateSigner(c.Request.Context(), envelope.ClicksignKey, signerData)
			if err != nil {
				// Tentar reverter criação local (best effort)
				_ = h.RepositorySignatory.Delete(&signatoryEntity)

				c.JSON(http.StatusInternalServerError, dtos.ErrorResponseDTO{
					Error:   "Internal server error",
					Message: fmt.Sprintf("Failed to create signatory in provider: %v. ATENÇÃO: Signatário foi criado localmente mas falhou no provider", err),
					Details: map[string]interface{}{
						"correlation_id":      correlationID,
						"envelope_id":         createdEnvelope.ID,
						"failed_signatory":    i + 1,
						"partial_transaction": true,
						"provider":            requestDTO.Provider,
					},
				})
				return
			}

			// Atualizar signatário com chave do provider
			signatoryEntity.SetClicksignKey(providerSignerKey)
			if err := h.RepositorySignatory.Update(&signatoryEntity); err != nil {
				c.JSON(http.StatusInternalServerError, dtos.ErrorResponseDTO{
					Error:   "Internal server error",
					Message: fmt.Sprintf("Failed to update signatory with provider key: %v", err),
					Details: map[string]interface{}{
						"correlation_id": correlationID,
						"provider":       requestDTO.Provider,
					},
				})
				return
			}

			createdSignatories = append(createdSignatories, signatoryEntity)
		}
	}

	// Criar requirements se fornecidos no request
	if len(requestDTO.Requirements) > 0 {
		for i, requirementRequest := range requestDTO.Requirements {
			// Verificar se há signatários suficientes
			if i >= len(createdSignatories) {
				c.JSON(http.StatusBadRequest, dtos.ErrorResponseDTO{
					Error:   "Bad Request",
					Message: fmt.Sprintf("Não há signatários suficientes para o requirement %d. Enviados: %d, Necessários: %d", i+1, len(createdSignatories), len(requestDTO.Requirements)),
					Details: map[string]interface{}{
						"correlation_id": correlationID,
						"envelope_id":    createdEnvelope.ID,
						"provider":       requestDTO.Provider,
					},
				})
				return
			}

			signatory := createdSignatories[i]

			for _, document := range documents {
				// Converter RequirementCreateRequest para EntityRequirement
				reqData := provider.RequirementData{
					Action:     requirementRequest.Action,
					DocumentID: document.ClicksignKey,
					SignerID:   signatory.ClicksignKey,
				}

				if requirementRequest.Auth != nil {
					reqData.Auth = *requirementRequest.Auth
				}

				// Criar requirement no provider
				providerReqKey, err := envelopeProvider.CreateRequirement(
					c.Request.Context(),
					createdEnvelope.ClicksignKey,
					reqData,
				)

				if err != nil {
					status := http.StatusInternalServerError
					var ce *clicksign.ClicksignError
					if errors.As(err, &ce) && ce.StatusCode > 0 {
						status = ce.StatusCode
					}
					c.JSON(status, dtos.ErrorResponseDTO{
						Error:   http.StatusText(status),
						Message: fmt.Sprintf("Failed to create requirement for envelope %d: %v", createdEnvelope.ID, err),
						Details: map[string]interface{}{
							"correlation_id": correlationID,
							"envelope_id":    createdEnvelope.ID,
							"provider":       requestDTO.Provider,
						},
					})
					return
				}

				// Criar requirement localmente
				// Nota: Já criamos no provider acima, então apenas salvamos localmente com a chave retornada
				requirementEntity := &entity.EntityRequirement{
					EnvelopeID:   createdEnvelope.ID,
					ClicksignKey: providerReqKey,
					DocumentID:   &document.ClicksignKey,
					SignerID:     &signatory.ClicksignKey,
					Action:       requirementRequest.Action,
				}

				if requirementRequest.Auth != nil {
					requirementEntity.Auth = requirementRequest.Auth
				}

				_, err = h.RepositoryRequirement.Create(c.Request.Context(), requirementEntity)
				if err != nil {
					c.JSON(http.StatusInternalServerError, dtos.ErrorResponseDTO{
						Error:   "Internal server error",
						Message: fmt.Sprintf("Failed to create requirement locally: %v", err),
						Details: map[string]interface{}{
							"correlation_id": correlationID,
							"envelope_id":    createdEnvelope.ID,
							"provider":       requestDTO.Provider,
						},
					})
					return
				}
			}
		}
	}

	// Criar qualificadores se fornecidos no request
	if len(requestDTO.Qualifiers) > 0 {
		for i, qualifierRequest := range requestDTO.Qualifiers {
			// Verificar se há signatários suficientes
			if i >= len(createdSignatories) {
				c.JSON(http.StatusBadRequest, dtos.ErrorResponseDTO{
					Error:   "Bad Request",
					Message: fmt.Sprintf("Não há signatários suficientes para o qualifier %d. Enviados: %d, Necessários: %d", i+1, len(createdSignatories), len(requestDTO.Qualifiers)),
					Details: map[string]interface{}{
						"correlation_id": correlationID,
						"envelope_id":    createdEnvelope.ID,
						"provider":       requestDTO.Provider,
					},
				})
				return
			}

			signatory := createdSignatories[i]

			for _, document := range documents {
			// Converter RequirementCreateRequest para EntityRequirement
			reqData := provider.RequirementData{
				Action:     qualifierRequest.Action,
				DocumentID: document.ClicksignKey,
				SignerID:   signatory.ClicksignKey,
			}
			if qualifierRequest.Role != "" {
				reqData.Role = qualifierRequest.Role
			}

				// Criar requirement no provider
				providerReqKey, err := envelopeProvider.CreateRequirement(
					c.Request.Context(),
					createdEnvelope.ClicksignKey,
					reqData,
				)

				if err != nil {
					status := http.StatusInternalServerError
					var ce *clicksign.ClicksignError
					if errors.As(err, &ce) && ce.StatusCode > 0 {
						status = ce.StatusCode
					}
					c.JSON(status, dtos.ErrorResponseDTO{
						Error:   http.StatusText(status),
						Message: fmt.Sprintf("Failed to create qualifier for envelope %d: %v", createdEnvelope.ID, err),
						Details: map[string]interface{}{
							"correlation_id": correlationID,
							"envelope_id":    createdEnvelope.ID,
							"provider":       requestDTO.Provider,
						},
					})
					return
				}

				// Criar requirement localmente (qualifier)
				// Nota: Já criamos no provider acima, então apenas salvamos localmente com a chave retornada
				requirementEntity := &entity.EntityRequirement{
					EnvelopeID:   createdEnvelope.ID,
					ClicksignKey: providerReqKey,
					DocumentID:   &document.ClicksignKey,
					SignerID:     &signatory.ClicksignKey,
					Action:       qualifierRequest.Action,
					Role:         qualifierRequest.Role,
				}

				_, err = h.RepositoryRequirement.Create(c.Request.Context(), requirementEntity)
				if err != nil {
					c.JSON(http.StatusInternalServerError, dtos.ErrorResponseDTO{
						Error:   "Internal server error",
						Message: fmt.Sprintf("Failed to create qualifier locally: %v", err),
						Details: map[string]interface{}{
							"correlation_id": correlationID,
							"envelope_id":    createdEnvelope.ID,
							"provider":       requestDTO.Provider,
						},
					})
					return
				}
			}
		}
	}

	if requestDTO.Approved {
		// Ativar envelope se aprovado
		createdEnvelope, err = envelopeProviderService.ActivateEnvelope(createdEnvelope.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dtos.ErrorResponseDTO{
				Error:   "Internal server error",
				Message: fmt.Sprintf("Failed to activate envelope %v: %v", createdEnvelope.ClicksignKey, err),
				Details: map[string]interface{}{
					"correlation_id": correlationID,
					"provider":       requestDTO.Provider,
				},
			})
			return
		}
	}

	// Converter entidade para DTO de resposta
	responseDTO := h.mapEntityToResponseV2(createdEnvelope, createdSignatories)

	c.JSON(http.StatusCreated, responseDTO)
}

// @Summary Get envelope (v2)
// @Description Get envelope by ID. The response includes clicksign_raw_data field with the complete JSON response from provider API when available.
// @Tags envelopes-v2
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Envelope ID"
// @Success 200 {object} dtos.EnvelopeResponseDTO "Envelope data with optional clicksign_raw_data field containing raw provider API response"
// @Failure 404 {object} dtos.ErrorResponseDTO
// @Failure 500 {object} dtos.ErrorResponseDTO
// @Router /api/v2/envelopes/{id} [get]
func (h *EnvelopeV2Handlers) GetEnvelopeV2Handler(c *gin.Context) {
	correlationID := c.GetHeader("X-Correlation-ID")
	if correlationID == "" {
		correlationID = strconv.FormatInt(time.Now().Unix(), 10)
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponseDTO{
			Error:   "Invalid ID",
			Message: "Envelope ID must be a valid integer",
			Details: map[string]interface{}{
				"correlation_id": correlationID,
			},
		})
		return
	}

	// Obter envelope do repository (não precisa de provider para leitura)
	envelope, err := h.RepositoryEnvelope.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, dtos.ErrorResponseDTO{
			Error:   "Envelope not found",
			Message: "The requested envelope does not exist",
			Details: map[string]interface{}{
				"correlation_id": correlationID,
			},
		})
		return
	}

	responseDTO := h.mapEntityToResponseV2(envelope)

	c.JSON(http.StatusOK, responseDTO)
}

// @Summary List envelopes (v2)
// @Description Get list of envelopes with optional filters
// @Tags envelopes-v2
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param search query string false "Search term"
// @Param status query string false "Status filter"
// @Param clicksign_key query string false "Clicksign key filter"
// @Success 200 {object} dtos.EnvelopeListResponseDTO
// @Failure 500 {object} dtos.ErrorResponseDTO
// @Router /api/v2/envelopes [get]
func (h *EnvelopeV2Handlers) GetEnvelopesV2Handler(c *gin.Context) {
	correlationID := c.GetHeader("X-Correlation-ID")
	if correlationID == "" {
		correlationID = strconv.FormatInt(time.Now().Unix(), 10)
	}

	var filters entity.EntityEnvelopeFilters
	filters.Search = c.Query("search")
	filters.Status = c.Query("status")
	filters.ClicksignKey = c.Query("clicksign_key")

	// Obter envelopes do repository (não precisa de provider para leitura)
	envelopes, err := h.RepositoryEnvelope.GetEnvelopes(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponseDTO{
			Error:   "Internal server error",
			Message: "Failed to retrieve envelopes",
			Details: map[string]interface{}{
				"correlation_id": correlationID,
			},
		})
		return
	}

	responseDTO := h.mapEnvelopeListToResponseV2(envelopes)

	c.JSON(http.StatusOK, responseDTO)
}

// @Summary Activate envelope (v2)
// @Description Activate envelope to start signing process using the provider that created it
// @Tags envelopes-v2
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Envelope ID"
// @Success 200 {object} dtos.EnvelopeResponseDTO
// @Failure 400 {object} dtos.ErrorResponseDTO
// @Failure 404 {object} dtos.ErrorResponseDTO
// @Failure 500 {object} dtos.ErrorResponseDTO
// @Router /api/v2/envelopes/{id}/activate [post]
func (h *EnvelopeV2Handlers) ActivateEnvelopeV2Handler(c *gin.Context) {
	correlationID := c.GetHeader("X-Correlation-ID")
	if correlationID == "" {
		correlationID = strconv.FormatInt(time.Now().Unix(), 10)
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponseDTO{
			Error:   "Invalid ID",
			Message: "Envelope ID must be a valid integer",
			Details: map[string]interface{}{
				"correlation_id": correlationID,
			},
		})
		return
	}

	// Obter envelope para determinar qual provider foi usado
	// Por enquanto, assumimos Clicksign se tiver ClicksignKey
	// TODO: Adicionar campo provider na entidade quando suportar múltiplos providers
	envelope, err := h.RepositoryEnvelope.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, dtos.ErrorResponseDTO{
			Error:   "Envelope not found",
			Message: "The requested envelope does not exist",
			Details: map[string]interface{}{
				"correlation_id": correlationID,
			},
		})
		return
	}

	// Determinar provider baseado no envelope
	// Por enquanto, se tem ClicksignKey, assume Clicksign
	// No futuro, podemos ter um campo provider na entidade
	providerName := "clicksign" // Default para envelopes existentes
	if envelope.ClicksignKey == "" {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponseDTO{
			Error:   "Bad Request",
			Message: "Envelope does not have a provider key",
			Details: map[string]interface{}{
				"correlation_id": correlationID,
			},
		})
		return
	}

	// Obter provider
	envelopeProvider, err := h.ProviderFactory.GetProvider(providerName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponseDTO{
			Error:   "Internal server error",
			Message: fmt.Sprintf("Failed to get provider: %v", err),
			Details: map[string]interface{}{
				"correlation_id": correlationID,
				"provider":       providerName,
			},
		})
		return
	}

	// Criar use case com provider
	usecaseDocument := h.UsecaseDocuments
	usecaseRequirement := h.UsecaseRequirement
	envelopeProviderService := usecase_envelope.NewUsecaseEnvelopeProviderService(
		h.RepositoryEnvelope,
		envelopeProvider,
		usecaseDocument,
		usecaseRequirement,
		h.Logger,
	)

	// Ativar envelope
	activatedEnvelope, err := envelopeProviderService.ActivateEnvelope(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponseDTO{
			Error:   "Internal server error",
			Message: "Failed to activate envelope: " + err.Error(),
			Details: map[string]interface{}{
				"correlation_id": correlationID,
				"provider":       providerName,
			},
		})
		return
	}

	responseDTO := h.mapEntityToResponseV2(activatedEnvelope)

	c.JSON(http.StatusOK, responseDTO)
}

// @Summary Notify envelope (v2)
// @Description Send notification to envelope signatories using the provider that created it
// @Tags envelopes-v2
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Envelope ID"
// @Param request body dtos.EnvelopeNotificationRequestDTO true "Notification data"
// @Success 200 {object} dtos.EnvelopeNotificationResponseDTO
// @Failure 400 {object} dtos.ErrorResponseDTO
// @Failure 404 {object} dtos.ErrorResponseDTO
// @Failure 500 {object} dtos.ErrorResponseDTO
// @Router /api/v2/envelopes/{id}/notify [post]
func (h *EnvelopeV2Handlers) NotifyEnvelopeV2Handler(c *gin.Context) {
	correlationID := c.GetHeader("X-Correlation-ID")
	if correlationID == "" {
		correlationID = strconv.FormatInt(time.Now().Unix(), 10)
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponseDTO{
			Error:   "Invalid ID",
			Message: "Envelope ID must be a valid integer",
			Details: map[string]interface{}{
				"correlation_id": correlationID,
			},
		})
		return
	}

	var requestDTO dtos.EnvelopeNotificationRequestDTO

	if err := c.ShouldBindJSON(&requestDTO); err != nil {
		validationErrors := h.extractValidationErrors(err)
		c.JSON(http.StatusBadRequest, dtos.ValidationErrorResponseDTO{
			Error:   "Validation failed",
			Message: "Invalid request payload",
			Details: validationErrors,
		})
		return
	}

	// Obter envelope para determinar qual provider foi usado
	envelope, err := h.RepositoryEnvelope.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, dtos.ErrorResponseDTO{
			Error:   "Envelope not found",
			Message: "The requested envelope does not exist",
			Details: map[string]interface{}{
				"correlation_id": correlationID,
			},
		})
		return
	}

	// Determinar provider baseado no envelope
	// Por enquanto, se tem ClicksignKey, assume Clicksign
	providerName := "clicksign" // Default para envelopes existentes
	if envelope.ClicksignKey == "" {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponseDTO{
			Error:   "Bad Request",
			Message: "Envelope does not have a provider key",
			Details: map[string]interface{}{
				"correlation_id": correlationID,
			},
		})
		return
	}

	// Obter provider
	envelopeProvider, err := h.ProviderFactory.GetProvider(providerName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponseDTO{
			Error:   "Internal server error",
			Message: fmt.Sprintf("Failed to get provider: %v", err),
			Details: map[string]interface{}{
				"correlation_id": correlationID,
				"provider":       providerName,
			},
		})
		return
	}

	// Criar use case com provider
	usecaseDocument := h.UsecaseDocuments
	usecaseRequirement := h.UsecaseRequirement
	envelopeProviderService := usecase_envelope.NewUsecaseEnvelopeProviderService(
		h.RepositoryEnvelope,
		envelopeProvider,
		usecaseDocument,
		usecaseRequirement,
		h.Logger,
	)

	// Enviar notificação através do use case
	err = envelopeProviderService.NotifyEnvelope(c.Request.Context(), id, requestDTO.Message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponseDTO{
			Error:   "Internal server error",
			Message: "Failed to send notification: " + err.Error(),
			Details: map[string]interface{}{
				"correlation_id": correlationID,
				"provider":       providerName,
			},
		})
		return
	}

	responseDTO := dtos.EnvelopeNotificationResponseDTO{
		Success: true,
		Message: "Notification sent successfully",
	}

	c.JSON(http.StatusOK, responseDTO)
}

// mapEntityToResponseV2 converte EntityEnvelope para DTO de resposta (reutiliza lógica do v1)
func (h *EnvelopeV2Handlers) mapEntityToResponseV2(envelope *entity.EntityEnvelope, signatories ...[]entity.EntitySignatory) *dtos.EnvelopeResponseDTO {
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

// mapEnvelopeListToResponseV2 converte lista de envelopes para DTO de resposta
func (h *EnvelopeV2Handlers) mapEnvelopeListToResponseV2(envelopes []entity.EntityEnvelope) *dtos.EnvelopeListResponseDTO {
	envelopeList := make([]dtos.EnvelopeResponseDTO, len(envelopes))
	for i, envelope := range envelopes {
		envelopeList[i] = *h.mapEntityToResponseV2(&envelope)
	}

	return &dtos.EnvelopeListResponseDTO{
		Envelopes: envelopeList,
		Total:     len(envelopes),
	}
}

// mapCreateRequestToEntityV2 converte EnvelopeV2CreateRequestDTO para EntityEnvelope e documentos
// Reutiliza a mesma lógica do handler v1
func (h *EnvelopeV2Handlers) mapCreateRequestToEntityV2(dto dtos.EnvelopeV2CreateRequestDTO) (*entity.EntityEnvelope, []*entity.EntityDocument, error) {
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

		// Converter metadata do DTO (map[string]interface{}) para datatypes.JSON
		var metadataJSON datatypes.JSON
		if docRequest.Metadata != nil {
			metadataBytes, err := json.Marshal(docRequest.Metadata)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to marshal metadata for document '%s': %w", docRequest.Name, err)
			}
			metadataJSON = datatypes.JSON(metadataBytes)
		}

		document := &entity.EntityDocument{
			Name:         docRequest.Name,
			Description:  docRequest.Description,
			FilePath:     fileInfo.TempPath,
			FileSize:     fileInfo.Size,
			MimeType:     fileInfo.MimeType,
			IsFromBase64: true,
			Status:       "draft",
			Metadata:     metadataJSON,
		}

		documents = append(documents, document)
	}

	return envelope, documents, nil
}

// extractValidationErrors extrai erros de validação (reutiliza do handler v1)
func (h *EnvelopeV2Handlers) extractValidationErrors(err error) []dtos.ValidationErrorDetail {
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

// getValidationErrorMessage retorna mensagem de erro de validação (reutiliza do handler v1)
func (h *EnvelopeV2Handlers) getValidationErrorMessage(fieldError validator.FieldError) string {
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

// MountEnvelopeV2Handlers monta as rotas v2 de envelopes
func MountEnvelopeV2Handlers(gin *gin.Engine, conn *gorm.DB, logger *logrus.Logger) {
	// Criar factory de providers
	providerFactory := provider_factory.NewProviderFactory(config.EnvironmentVariables, logger)

	// Criar usecase de documento
	clicksignClient := clicksign.NewClicksignClient(config.EnvironmentVariables, logger)
	usecaseDocument := document.NewUsecaseDocumentServiceWithClicksign(
		repository.NewRepositoryDocument(conn),
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

	// Criar usecase de signatory
	usecaseSignatory := signatory.NewUsecaseSignatoryService(
		repository.NewRepositorySignatory(conn),
		repository.NewRepositoryEnvelope(conn),
		clicksignClient,
		logger,
	)

	// Criar repositories
	repositoryEnvelope := repository.NewRepositoryEnvelope(conn)
	repositorySignatory := repository.NewRepositorySignatory(conn)
	repositoryRequirement := repository.NewRepositoryRequirement(conn)

	envelopeV2Handlers := NewEnvelopeV2Handler(
		providerFactory,
		usecaseDocument,
		usecaseRequirement,
		usecaseSignatory,
		repositoryEnvelope,
		repositorySignatory,
		repositoryRequirement,
		logger,
	)

	group := gin.Group("/api/v2/envelopes")
	SetAuthMiddleware(conn, group)

	group.POST("/", envelopeV2Handlers.CreateEnvelopeV2Handler)
	group.GET("/:id", envelopeV2Handlers.GetEnvelopeV2Handler)
	group.GET("/", envelopeV2Handlers.GetEnvelopesV2Handler)
	group.POST("/:id/activate", envelopeV2Handlers.ActivateEnvelopeV2Handler)
	group.POST("/:id/notify", envelopeV2Handlers.NotifyEnvelopeV2Handler)
}

