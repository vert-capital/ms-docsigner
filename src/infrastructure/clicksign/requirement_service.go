package clicksign

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"app/infrastructure/clicksign/dto"
	"app/usecase/clicksign"
	"github.com/sirupsen/logrus"
)

type RequirementService struct {
	clicksignClient clicksign.ClicksignClientInterface
	logger          *logrus.Logger
}

func NewRequirementService(clicksignClient clicksign.ClicksignClientInterface, logger *logrus.Logger) *RequirementService {
	return &RequirementService{
		clicksignClient: clicksignClient,
		logger:          logger,
	}
}

// CreateRequirement cria um requisito no envelope usando a estrutura JSON API com relacionamentos
func (s *RequirementService) CreateRequirement(ctx context.Context, envelopeID string, reqData RequirementData) (string, error) {
	correlationID := ctx.Value("correlation_id")

	s.logger.WithFields(logrus.Fields{
		"envelope_id":    envelopeID,
		"action":         reqData.Action,
		"role":           reqData.Role,
		"auth":           reqData.Auth,
		"document_id":    reqData.DocumentID,
		"signer_id":      reqData.SignerID,
		"correlation_id": correlationID,
	}).Info("Creating requirement in envelope using JSON API format with relationships")

	createRequest := s.mapRequirementDataToCreateRequest(reqData)

	s.logger.WithFields(logrus.Fields{
		"data_type":           createRequest.Data.Type,
		"action":             createRequest.Data.Attributes.Action,
		"role":               createRequest.Data.Attributes.Role,
		"auth":               createRequest.Data.Attributes.Auth,
		"has_relationships":  createRequest.Data.Relationships != nil,
		"correlation_id":     correlationID,
	}).Debug("JSON API request structure prepared for requirement")

	// Fazer chamada para API do Clicksign usando o endpoint correto para requisitos
	endpoint := fmt.Sprintf("/api/v3/envelopes/%s/requirements", envelopeID)
	resp, err := s.clicksignClient.Post(ctx, endpoint, createRequest)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"error":          err.Error(),
			"envelope_id":    envelopeID,
			"action":         reqData.Action,
			"correlation_id": correlationID,
		}).Error("Failed to create requirement in Clicksign envelope")
		return "", fmt.Errorf("failed to create requirement in Clicksign envelope: %w", err)
	}
	defer resp.Body.Close()

	// Ler resposta
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"error":          err.Error(),
			"envelope_id":    envelopeID,
			"action":         reqData.Action,
			"correlation_id": correlationID,
		}).Error("Failed to read response from Clicksign")
		return "", fmt.Errorf("failed to read response from Clicksign: %w", err)
	}

	// Verificar se houve erro na resposta
	if resp.StatusCode >= 400 {
		var errorResp dto.ClicksignErrorResponse
		if err := json.Unmarshal(body, &errorResp); err != nil {
			s.logger.WithFields(logrus.Fields{
				"error":          err.Error(),
				"status_code":    resp.StatusCode,
				"response_body":  string(body),
				"correlation_id": correlationID,
			}).Error("Failed to parse error response from Clicksign")
			return "", fmt.Errorf("Clicksign API error (status %d): %s", resp.StatusCode, string(body))
		}

		s.logger.WithFields(logrus.Fields{
			"error_type":     errorResp.Error.Type,
			"error_message":  errorResp.Error.Message,
			"error_code":     errorResp.Error.Code,
			"status_code":    resp.StatusCode,
			"correlation_id": correlationID,
		}).Error("Clicksign API returned error")

		return "", fmt.Errorf("Clicksign API error: %s - %s", errorResp.Error.Type, errorResp.Error.Message)
	}

	// Fazer parse da resposta de sucesso usando estrutura JSON API
	var createResponse dto.RequirementCreateResponseWrapper
	if err := json.Unmarshal(body, &createResponse); err != nil {
		s.logger.WithFields(logrus.Fields{
			"error":          err.Error(),
			"response_body":  string(body),
			"expected_format": "JSON API (data.type.attributes)",
			"correlation_id": correlationID,
		}).Error("Failed to parse JSON API response from Clicksign")
		return "", fmt.Errorf("failed to parse JSON API response from Clicksign: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"envelope_id":       envelopeID,
		"requirement_id":    createResponse.Data.ID,
		"action":           createResponse.Data.Attributes.Action,
		"role":             createResponse.Data.Attributes.Role,
		"auth":             createResponse.Data.Attributes.Auth,
		"response_type":    createResponse.Data.Type,
		"correlation_id":   correlationID,
	}).Info("Requirement created successfully in Clicksign envelope using JSON API format")

	return createResponse.Data.ID, nil
}

// CreateBulkRequirements cria múltiplos requisitos usando operações atômicas conforme JSON API spec
func (s *RequirementService) CreateBulkRequirements(ctx context.Context, envelopeID string, operations []BulkOperation) ([]string, error) {
	correlationID := ctx.Value("correlation_id")

	s.logger.WithFields(logrus.Fields{
		"envelope_id":        envelopeID,
		"operations_count":   len(operations),
		"correlation_id":     correlationID,
	}).Info("Creating bulk requirements in envelope using atomic operations")

	bulkRequest := s.mapBulkOperationsToBulkRequest(operations)

	s.logger.WithFields(logrus.Fields{
		"atomic_operations_count": len(bulkRequest.AtomicOperations),
		"correlation_id":         correlationID,
	}).Debug("Atomic operations request structure prepared")

	// Fazer chamada para API do Clicksign usando o endpoint de bulk requirements
	endpoint := fmt.Sprintf("/api/v3/envelopes/%s/bulk_requirements", envelopeID)
	resp, err := s.clicksignClient.Post(ctx, endpoint, bulkRequest)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"error":          err.Error(),
			"envelope_id":    envelopeID,
			"correlation_id": correlationID,
		}).Error("Failed to create bulk requirements in Clicksign envelope")
		return nil, fmt.Errorf("failed to create bulk requirements in Clicksign envelope: %w", err)
	}
	defer resp.Body.Close()

	// Ler resposta
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"error":          err.Error(),
			"envelope_id":    envelopeID,
			"correlation_id": correlationID,
		}).Error("Failed to read response from Clicksign")
		return nil, fmt.Errorf("failed to read response from Clicksign: %w", err)
	}

	// Verificar se houve erro na resposta
	if resp.StatusCode >= 400 {
		var errorResp dto.ClicksignErrorResponse
		if err := json.Unmarshal(body, &errorResp); err != nil {
			s.logger.WithFields(logrus.Fields{
				"error":          err.Error(),
				"status_code":    resp.StatusCode,
				"response_body":  string(body),
				"correlation_id": correlationID,
			}).Error("Failed to parse error response from Clicksign")
			return nil, fmt.Errorf("Clicksign API error (status %d): %s", resp.StatusCode, string(body))
		}

		s.logger.WithFields(logrus.Fields{
			"error_type":     errorResp.Error.Type,
			"error_message":  errorResp.Error.Message,
			"error_code":     errorResp.Error.Code,
			"status_code":    resp.StatusCode,
			"correlation_id": correlationID,
		}).Error("Clicksign API returned error")

		return nil, fmt.Errorf("Clicksign API error: %s - %s", errorResp.Error.Type, errorResp.Error.Message)
	}

	// Fazer parse da resposta de sucesso usando estrutura JSON API para bulk operations
	var bulkResponse dto.BulkRequirementsResponseWrapper
	if err := json.Unmarshal(body, &bulkResponse); err != nil {
		s.logger.WithFields(logrus.Fields{
			"error":          err.Error(),
			"response_body":  string(body),
			"expected_format": "JSON API atomic:results",
			"correlation_id": correlationID,
		}).Error("Failed to parse JSON API bulk response from Clicksign")
		return nil, fmt.Errorf("failed to parse JSON API bulk response from Clicksign: %w", err)
	}

	// Extrair IDs dos requisitos criados
	var createdIDs []string
	for _, result := range bulkResponse.AtomicResults {
		if result.Data != nil {
			createdIDs = append(createdIDs, result.Data.ID)
		}
	}

	s.logger.WithFields(logrus.Fields{
		"envelope_id":       envelopeID,
		"created_count":     len(createdIDs),
		"requirement_ids":   createdIDs,
		"correlation_id":    correlationID,
	}).Info("Bulk requirements created successfully in Clicksign envelope using atomic operations")

	return createdIDs, nil
}

// mapRequirementDataToCreateRequest mapeia os dados do requisito para a estrutura JSON API
func (s *RequirementService) mapRequirementDataToCreateRequest(reqData RequirementData) *dto.RequirementCreateRequestWrapper {
	req := &dto.RequirementCreateRequestWrapper{
		Data: dto.RequirementCreateData{
			Type: "requirements",
			Attributes: dto.RequirementCreateAttributes{
				Action: reqData.Action,
			},
		},
	}

	// Adicionar role se fornecido (usado para qualificação)
	if reqData.Role != "" {
		req.Data.Attributes.Role = reqData.Role
	}

	// Adicionar auth se fornecido (usado para autenticação)
	if reqData.Auth != "" {
		req.Data.Attributes.Auth = reqData.Auth
	}

	// Adicionar relacionamentos se fornecidos
	if reqData.DocumentID != "" || reqData.SignerID != "" {
		req.Data.Relationships = &dto.RequirementRelationships{}

		if reqData.DocumentID != "" {
			req.Data.Relationships.Document = &dto.RequirementRelationship{
				Data: dto.RequirementRelationshipData{
					Type: "documents",
					ID:   reqData.DocumentID,
				},
			}
		}

		if reqData.SignerID != "" {
			req.Data.Relationships.Signer = &dto.RequirementRelationship{
				Data: dto.RequirementRelationshipData{
					Type: "signers",
					ID:   reqData.SignerID,
				},
			}
		}
	}

	return req
}

// mapBulkOperationsToBulkRequest mapeia operações em lote para a estrutura de atomic operations
func (s *RequirementService) mapBulkOperationsToBulkRequest(operations []BulkOperation) *dto.BulkRequirementsRequestWrapper {
	atomicOps := make([]dto.AtomicOperation, len(operations))

	for i, op := range operations {
		atomicOp := dto.AtomicOperation{
			Op: op.Operation,
		}

		if op.Operation == "remove" && op.RequirementID != "" {
			atomicOp.Ref = &dto.AtomicOperationRef{
				Type: "requirements",
				ID:   op.RequirementID,
			}
		} else if op.Operation == "add" && op.RequirementData != nil {
			atomicOp.Data = &dto.RequirementCreateData{
				Type: "requirements",
				Attributes: dto.RequirementCreateAttributes{
					Action: op.RequirementData.Action,
					Role:   op.RequirementData.Role,
					Auth:   op.RequirementData.Auth,
				},
			}

			// Adicionar relacionamentos para operação add
			if op.RequirementData.DocumentID != "" || op.RequirementData.SignerID != "" {
				atomicOp.Data.Relationships = &dto.RequirementRelationships{}

				if op.RequirementData.DocumentID != "" {
					atomicOp.Data.Relationships.Document = &dto.RequirementRelationship{
						Data: dto.RequirementRelationshipData{
							Type: "documents",
							ID:   op.RequirementData.DocumentID,
						},
					}
				}

				if op.RequirementData.SignerID != "" {
					atomicOp.Data.Relationships.Signer = &dto.RequirementRelationship{
						Data: dto.RequirementRelationshipData{
							Type: "signers",
							ID:   op.RequirementData.SignerID,
						},
					}
				}
			}
		}

		atomicOps[i] = atomicOp
	}

	return &dto.BulkRequirementsRequestWrapper{
		AtomicOperations: atomicOps,
	}
}

// RequirementData representa os dados necessários para criar um requisito
type RequirementData struct {
	Action     string // "agree" para qualificação ou "provide_evidence" para autenticação
	Role       string // "sign" para qualificação
	Auth       string // "email" ou "icp_brasil" para autenticação
	DocumentID string // ID do documento relacionado
	SignerID   string // ID do signatário relacionado
}

// BulkOperation representa uma operação em lote para requisitos
type BulkOperation struct {
	Operation       string           // "add" ou "remove"
	RequirementID   string           // ID do requisito para operação "remove"
	RequirementData *RequirementData // Dados do requisito para operação "add"
}