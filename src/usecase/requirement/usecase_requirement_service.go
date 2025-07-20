package requirement

import (
	"context"
	"fmt"

	"app/entity"
	"app/infrastructure/clicksign"
	clicksignInterface "app/usecase/clicksign"
	"github.com/sirupsen/logrus"
)

type UsecaseRequirementService struct {
	repositoryRequirement IRepositoryRequirement
	repositoryEnvelope    interface {
		GetByID(id int) (*entity.EntityEnvelope, error)
	}
	clicksignClient     clicksignInterface.ClicksignClientInterface
	requirementService  *clicksign.RequirementService
	logger              *logrus.Logger
}

func NewUsecaseRequirementService(
	repositoryRequirement IRepositoryRequirement,
	repositoryEnvelope interface {
		GetByID(id int) (*entity.EntityEnvelope, error)
	},
	clicksignClient clicksignInterface.ClicksignClientInterface,
	logger *logrus.Logger,
) IUsecaseRequirement {
	requirementService := clicksign.NewRequirementService(clicksignClient, logger)

	return &UsecaseRequirementService{
		repositoryRequirement: repositoryRequirement,
		repositoryEnvelope:    repositoryEnvelope,
		clicksignClient:       clicksignClient,
		requirementService:    requirementService,
		logger:                logger,
	}
}

func (u *UsecaseRequirementService) CreateRequirement(ctx context.Context, requirement *entity.EntityRequirement) (*entity.EntityRequirement, error) {
	correlationID := ctx.Value("correlation_id")
	if correlationID == nil {
		correlationID = fmt.Sprintf("requirement_%d_%d", requirement.EnvelopeID, requirement.CreatedAt.Unix())
		ctx = context.WithValue(ctx, "correlation_id", correlationID)
	}

	u.logger.WithFields(logrus.Fields{
		"envelope_id":         requirement.EnvelopeID,
		"requirement_action":  requirement.Action,
		"requirement_role":    requirement.Role,
		"requirement_auth":    requirement.Auth,
		"correlation_id":      correlationID,
		"step":                "requirement_creation_start",
	}).Info("Starting requirement creation process")

	// Verificar se o envelope existe
	envelope, err := u.repositoryEnvelope.GetByID(requirement.EnvelopeID)
	if err != nil {
		u.logger.WithFields(logrus.Fields{
			"envelope_id":    requirement.EnvelopeID,
			"error":          err.Error(),
			"correlation_id": correlationID,
			"step":           "envelope_validation",
		}).Error("Failed to validate envelope existence")
		return nil, fmt.Errorf("envelope not found: %w", err)
	}

	u.logger.WithFields(logrus.Fields{
		"envelope_id":       requirement.EnvelopeID,
		"envelope_name":     envelope.Name,
		"envelope_status":   envelope.Status,
		"clicksign_key":     envelope.ClicksignKey,
		"correlation_id":    correlationID,
		"step":              "envelope_validation_success",
	}).Debug("Envelope validated successfully")

	// Verificar se o envelope tem ClicksignKey
	if envelope.ClicksignKey == "" {
		u.logger.WithFields(logrus.Fields{
			"envelope_id":    requirement.EnvelopeID,
			"envelope_name":  envelope.Name,
			"correlation_id": correlationID,
			"step":           "clicksign_key_validation",
		}).Error("Envelope does not have Clicksign key")
		return nil, fmt.Errorf("envelope must be created in Clicksign before adding requirements")
	}

	// Salvar requirement no banco local primeiro
	createdRequirement, err := u.repositoryRequirement.Create(ctx, requirement)
	if err != nil {
		u.logger.WithFields(logrus.Fields{
			"envelope_id":    requirement.EnvelopeID,
			"error":          err.Error(),
			"correlation_id": correlationID,
			"step":           "local_creation",
		}).Error("Failed to create requirement in local database")
		return nil, fmt.Errorf("failed to create requirement locally: %w", err)
	}

	u.logger.WithFields(logrus.Fields{
		"envelope_id":     requirement.EnvelopeID,
		"requirement_id":  createdRequirement.ID,
		"correlation_id":  correlationID,
		"step":            "local_creation_success",
	}).Debug("Requirement created successfully in local database")

	// Criar requirement no Clicksign
	reqData := clicksign.RequirementData{
		Action: requirement.Action,
		Role:   requirement.Role,
	}

	if requirement.Auth != nil {
		reqData.Auth = *requirement.Auth
	}
	if requirement.DocumentID != nil {
		reqData.DocumentID = *requirement.DocumentID
	}
	if requirement.SignerID != nil {
		reqData.SignerID = *requirement.SignerID
	}

	clicksignKey, err := u.requirementService.CreateRequirement(ctx, envelope.ClicksignKey, reqData)
	if err != nil {
		// Rollback: remover requirement do banco local
		deleteErr := u.repositoryRequirement.Delete(ctx, createdRequirement)
		if deleteErr != nil {
			u.logger.WithFields(logrus.Fields{
				"envelope_id":      requirement.EnvelopeID,
				"requirement_id":   createdRequirement.ID,
				"original_error":   err.Error(),
				"rollback_error":   deleteErr.Error(),
				"correlation_id":   correlationID,
				"step":             "rollback_failed",
			}).Error("Failed to rollback local requirement creation after Clicksign failure")
		}

		u.logger.WithFields(logrus.Fields{
			"envelope_id":      requirement.EnvelopeID,
			"requirement_id":   createdRequirement.ID,
			"clicksign_key":    envelope.ClicksignKey,
			"error":            err.Error(),
			"correlation_id":   correlationID,
			"step":             "clicksign_creation",
		}).Error("Failed to create requirement in Clicksign")
		return nil, fmt.Errorf("failed to create requirement in Clicksign: %w", err)
	}

	// Atualizar requirement com ClicksignKey
	createdRequirement.SetClicksignKey(clicksignKey)
	updatedRequirement, err := u.repositoryRequirement.Update(ctx, createdRequirement)
	if err != nil {
		u.logger.WithFields(logrus.Fields{
			"envelope_id":       requirement.EnvelopeID,
			"requirement_id":    createdRequirement.ID,
			"clicksign_key":     clicksignKey,
			"error":             err.Error(),
			"correlation_id":    correlationID,
			"step":              "clicksign_key_update",
		}).Warn("Failed to update requirement with Clicksign key, but requirement was created successfully")
		// Não retornar erro aqui pois o requirement foi criado com sucesso
		updatedRequirement = createdRequirement
	}

	u.logger.WithFields(logrus.Fields{
		"envelope_id":         requirement.EnvelopeID,
		"requirement_id":      updatedRequirement.ID,
		"clicksign_key":       updatedRequirement.ClicksignKey,
		"requirement_action":  updatedRequirement.Action,
		"requirement_role":    updatedRequirement.Role,
		"requirement_auth":    updatedRequirement.Auth,
		"correlation_id":      correlationID,
		"step":                "requirement_creation_complete",
	}).Info("Requirement created successfully in both local database and Clicksign")

	return updatedRequirement, nil
}

func (u *UsecaseRequirementService) GetRequirementsByEnvelopeID(ctx context.Context, envelopeID int) ([]entity.EntityRequirement, error) {
	correlationID := ctx.Value("correlation_id")
	if correlationID == nil {
		correlationID = fmt.Sprintf("get_requirements_%d", envelopeID)
		ctx = context.WithValue(ctx, "correlation_id", correlationID)
	}

	u.logger.WithFields(logrus.Fields{
		"envelope_id":    envelopeID,
		"correlation_id": correlationID,
		"step":           "get_requirements_start",
	}).Info("Starting to get requirements by envelope ID")

	// Verificar se o envelope existe
	_, err := u.repositoryEnvelope.GetByID(envelopeID)
	if err != nil {
		u.logger.WithFields(logrus.Fields{
			"envelope_id":    envelopeID,
			"error":          err.Error(),
			"correlation_id": correlationID,
			"step":           "envelope_validation",
		}).Error("Failed to validate envelope existence")
		return nil, fmt.Errorf("envelope not found: %w", err)
	}

	// Buscar requirements no banco local
	requirements, err := u.repositoryRequirement.GetByEnvelopeID(ctx, envelopeID)
	if err != nil {
		u.logger.WithFields(logrus.Fields{
			"envelope_id":    envelopeID,
			"error":          err.Error(),
			"correlation_id": correlationID,
			"step":           "local_fetch",
		}).Error("Failed to fetch requirements from local database")
		return nil, fmt.Errorf("failed to fetch requirements: %w", err)
	}

	u.logger.WithFields(logrus.Fields{
		"envelope_id":        envelopeID,
		"requirements_count": len(requirements),
		"correlation_id":     correlationID,
		"step":               "get_requirements_complete",
	}).Info("Requirements fetched successfully from local database")

	return requirements, nil
}

func (u *UsecaseRequirementService) GetRequirement(ctx context.Context, id int) (*entity.EntityRequirement, error) {
	correlationID := ctx.Value("correlation_id")
	if correlationID == nil {
		correlationID = fmt.Sprintf("get_requirement_%d", id)
		ctx = context.WithValue(ctx, "correlation_id", correlationID)
	}

	u.logger.WithFields(logrus.Fields{
		"requirement_id": id,
		"correlation_id": correlationID,
		"step":           "get_requirement_start",
	}).Debug("Starting to get requirement by ID")

	requirement, err := u.repositoryRequirement.GetByID(ctx, id)
	if err != nil {
		u.logger.WithFields(logrus.Fields{
			"requirement_id": id,
			"error":          err.Error(),
			"correlation_id": correlationID,
			"step":           "requirement_fetch",
		}).Error("Failed to fetch requirement from database")
		return nil, fmt.Errorf("failed to fetch requirement: %w", err)
	}

	u.logger.WithFields(logrus.Fields{
		"requirement_id":     requirement.ID,
		"envelope_id":        requirement.EnvelopeID,
		"requirement_action": requirement.Action,
		"correlation_id":     correlationID,
		"step":               "get_requirement_complete",
	}).Debug("Requirement fetched successfully")

	return requirement, nil
}

func (u *UsecaseRequirementService) UpdateRequirement(ctx context.Context, requirement *entity.EntityRequirement) (*entity.EntityRequirement, error) {
	correlationID := ctx.Value("correlation_id")
	if correlationID == nil {
		correlationID = fmt.Sprintf("update_requirement_%d", requirement.ID)
		ctx = context.WithValue(ctx, "correlation_id", correlationID)
	}

	u.logger.WithFields(logrus.Fields{
		"requirement_id": requirement.ID,
		"envelope_id":    requirement.EnvelopeID,
		"correlation_id": correlationID,
		"step":           "update_requirement_start",
	}).Info("Starting requirement update process")

	updatedRequirement, err := u.repositoryRequirement.Update(ctx, requirement)
	if err != nil {
		u.logger.WithFields(logrus.Fields{
			"requirement_id": requirement.ID,
			"error":          err.Error(),
			"correlation_id": correlationID,
			"step":           "requirement_update",
		}).Error("Failed to update requirement in database")
		return nil, fmt.Errorf("failed to update requirement: %w", err)
	}

	u.logger.WithFields(logrus.Fields{
		"requirement_id": updatedRequirement.ID,
		"envelope_id":    updatedRequirement.EnvelopeID,
		"correlation_id": correlationID,
		"step":           "update_requirement_complete",
	}).Info("Requirement updated successfully")

	return updatedRequirement, nil
}

func (u *UsecaseRequirementService) DeleteRequirement(ctx context.Context, id int) error {
	correlationID := ctx.Value("correlation_id")
	if correlationID == nil {
		correlationID = fmt.Sprintf("delete_requirement_%d", id)
		ctx = context.WithValue(ctx, "correlation_id", correlationID)
	}

	u.logger.WithFields(logrus.Fields{
		"requirement_id": id,
		"correlation_id": correlationID,
		"step":           "delete_requirement_start",
	}).Info("Starting requirement deletion process")

	// Buscar requirement para obter dados antes de deletar
	requirement, err := u.repositoryRequirement.GetByID(ctx, id)
	if err != nil {
		u.logger.WithFields(logrus.Fields{
			"requirement_id": id,
			"error":          err.Error(),
			"correlation_id": correlationID,
			"step":           "requirement_fetch",
		}).Error("Failed to fetch requirement for deletion")
		return fmt.Errorf("failed to fetch requirement for deletion: %w", err)
	}

	err = u.repositoryRequirement.Delete(ctx, requirement)
	if err != nil {
		u.logger.WithFields(logrus.Fields{
			"requirement_id": id,
			"envelope_id":    requirement.EnvelopeID,
			"error":          err.Error(),
			"correlation_id": correlationID,
			"step":           "requirement_deletion",
		}).Error("Failed to delete requirement from database")
		return fmt.Errorf("failed to delete requirement: %w", err)
	}

	u.logger.WithFields(logrus.Fields{
		"requirement_id": id,
		"envelope_id":    requirement.EnvelopeID,
		"correlation_id": correlationID,
		"step":           "delete_requirement_complete",
	}).Info("Requirement deleted successfully")

	return nil
}