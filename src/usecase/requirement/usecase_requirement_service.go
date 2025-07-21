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
	clicksignClient    clicksignInterface.ClicksignClientInterface
	requirementService *clicksign.RequirementService
	logger             *logrus.Logger
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
	// Verificar se o envelope existe
	envelope, err := u.repositoryEnvelope.GetByID(requirement.EnvelopeID)
	if err != nil {
		return nil, fmt.Errorf("envelope not found: %w", err)
	}

	// Verificar se o envelope tem ClicksignKey
	if envelope.ClicksignKey == "" {
		return nil, fmt.Errorf("envelope must be created in Clicksign before adding requirements")
	}

	// Salvar requirement no banco local primeiro
	createdRequirement, err := u.repositoryRequirement.Create(ctx, requirement)
	if err != nil {
		return nil, fmt.Errorf("failed to create requirement locally: %w", err)
	}

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
			// Log the rollback failure if needed, but continue with original error
		}

		return nil, fmt.Errorf("failed to create requirement in Clicksign: %w", err)
	}

	// Atualizar requirement com ClicksignKey
	createdRequirement.SetClicksignKey(clicksignKey)
	updatedRequirement, err := u.repositoryRequirement.Update(ctx, createdRequirement)
	if err != nil {
		// Não retornar erro aqui pois o requirement foi criado com sucesso
		updatedRequirement = createdRequirement
	}

	return updatedRequirement, nil
}

func (u *UsecaseRequirementService) GetRequirementsByEnvelopeID(ctx context.Context, envelopeID int) ([]entity.EntityRequirement, error) {
	// Verificar se o envelope existe
	_, err := u.repositoryEnvelope.GetByID(envelopeID)
	if err != nil {
		return nil, fmt.Errorf("envelope not found: %w", err)
	}

	// Buscar requirements no banco local
	requirements, err := u.repositoryRequirement.GetByEnvelopeID(ctx, envelopeID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch requirements: %w", err)
	}

	return requirements, nil
}

func (u *UsecaseRequirementService) GetRequirement(ctx context.Context, id int) (*entity.EntityRequirement, error) {
	requirement, err := u.repositoryRequirement.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch requirement: %w", err)
	}

	return requirement, nil
}

func (u *UsecaseRequirementService) UpdateRequirement(ctx context.Context, requirement *entity.EntityRequirement) (*entity.EntityRequirement, error) {
	updatedRequirement, err := u.repositoryRequirement.Update(ctx, requirement)
	if err != nil {
		return nil, fmt.Errorf("failed to update requirement: %w", err)
	}

	return updatedRequirement, nil
}

func (u *UsecaseRequirementService) DeleteRequirement(ctx context.Context, id int) error {
	// Buscar requirement para obter dados antes de deletar
	requirement, err := u.repositoryRequirement.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to fetch requirement for deletion: %w", err)
	}

	err = u.repositoryRequirement.Delete(ctx, requirement)
	if err != nil {
		return fmt.Errorf("failed to delete requirement: %w", err)
	}

	return nil
}
