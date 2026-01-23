package usecase_envelope

import (
	"context"
	"fmt"

	"app/entity"
	"app/infrastructure/provider"
	usecase_document "app/usecase/document"
	usecase_requirement "app/usecase/requirement"

	"github.com/sirupsen/logrus"
)

// UsecaseEnvelopeProviderService é um serviço que usa EnvelopeProvider para operações de envelope
// Esta implementação é agnóstica ao provider específico, permitindo alternar entre Clicksign, vertc-assinaturas, etc.
type UsecaseEnvelopeProviderService struct {
	repositoryEnvelope IRepositoryEnvelope
	envelopeProvider   provider.EnvelopeProvider
	usecaseDocument    usecase_document.IUsecaseDocument
	usecaseRequirement usecase_requirement.IUsecaseRequirement
	logger             *logrus.Logger
}

// NewUsecaseEnvelopeProviderService cria uma nova instância do UsecaseEnvelopeProviderService
func NewUsecaseEnvelopeProviderService(
	repositoryEnvelope IRepositoryEnvelope,
	envelopeProvider provider.EnvelopeProvider,
	usecaseDocument usecase_document.IUsecaseDocument,
	usecaseRequirement usecase_requirement.IUsecaseRequirement,
	logger *logrus.Logger,
) *UsecaseEnvelopeProviderService {
	return &UsecaseEnvelopeProviderService{
		repositoryEnvelope: repositoryEnvelope,
		envelopeProvider:   envelopeProvider,
		usecaseDocument:    usecaseDocument,
		usecaseRequirement: usecaseRequirement,
		logger:             logger,
	}
}

// CreateEnvelope cria um envelope usando o provider
func (u *UsecaseEnvelopeProviderService) CreateEnvelope(envelope *entity.EntityEnvelope) (*entity.EntityEnvelope, error) {
	ctx := context.Background()

	// Validar entidade
	err := envelope.Validate()
	if err != nil {
		return nil, fmt.Errorf("envelope validation failed: %w", err)
	}

	// Validações específicas de negócio
	if err := u.ValidateBusinessRules(envelope); err != nil {
		return nil, fmt.Errorf("business rule validation failed: %w", err)
	}

	// Criar envelope localmente primeiro
	err = u.repositoryEnvelope.Create(envelope)
	if err != nil {
		return nil, fmt.Errorf("failed to create envelope locally: %w", err)
	}

	// Criar envelope no provider
	providerKey, rawData, err := u.envelopeProvider.CreateEnvelope(ctx, envelope)
	if err != nil {
		// Tentar reverter criação local (best effort)
		if deleteErr := u.repositoryEnvelope.Delete(envelope); deleteErr != nil {
			// Log error but continue
		}

		return nil, fmt.Errorf("failed to create envelope in provider: %w", err)
	}

	// Atualizar envelope com chave e dados brutos do provider
	// Nota: Por enquanto usamos ClicksignKey e ClicksignRawData mesmo para outros providers
	// Isso pode ser abstraído no futuro
	envelope.SetClicksignKey(providerKey)
	envelope.SetClicksignRawData(rawData)
	err = u.repositoryEnvelope.Update(envelope)
	if err != nil {
		return nil, fmt.Errorf("failed to update envelope with provider key: %w", err)
	}

	return envelope, nil
}

// CreateDocument cria um documento dentro de um envelope usando o provider
func (u *UsecaseEnvelopeProviderService) CreateDocument(ctx context.Context, envelopeKey string, document *entity.EntityDocument, internalEnvelopeID int) (string, error) {
	return u.envelopeProvider.CreateDocument(ctx, envelopeKey, document, internalEnvelopeID)
}

// UpdateEnvelope atualiza um envelope
func (u *UsecaseEnvelopeProviderService) UpdateEnvelope(envelope *entity.EntityEnvelope) error {
	err := envelope.Validate()
	if err != nil {
		return fmt.Errorf("envelope validation failed: %w", err)
	}

	existingEnvelope, err := u.repositoryEnvelope.GetByID(envelope.ID)
	if err != nil {
		return fmt.Errorf("envelope not found: %w", err)
	}

	if existingEnvelope.Status == "completed" || existingEnvelope.Status == "cancelled" {
		return fmt.Errorf("cannot update envelope in '%s' status", existingEnvelope.Status)
	}

	err = u.repositoryEnvelope.Update(envelope)
	if err != nil {
		return fmt.Errorf("failed to update envelope: %w", err)
	}

	return nil
}

// ActivateEnvelope ativa um envelope usando o provider
func (u *UsecaseEnvelopeProviderService) ActivateEnvelope(id int) (*entity.EntityEnvelope, error) {
	ctx := context.Background()

	envelope, err := u.repositoryEnvelope.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("envelope not found: %w", err)
	}

	if envelope.ClicksignKey == "" {
		return nil, fmt.Errorf("envelope has no provider key")
	}

	// Ativar envelope localmente
	err = envelope.ActivateEnvelope()
	if err != nil {
		return nil, fmt.Errorf("failed to activate envelope locally: %w", err)
	}

	// Ativar envelope no provider
	err = u.envelopeProvider.ActivateEnvelope(ctx, envelope.ClicksignKey)
	if err != nil {
		return nil, fmt.Errorf("failed to activate envelope in provider: %w", err)
	}

	// Atualizar envelope no banco
	err = u.repositoryEnvelope.Update(envelope)
	if err != nil {
		return nil, fmt.Errorf("failed to update envelope status: %w", err)
	}

	return envelope, nil
}

// NotifyEnvelope envia uma notificação para os signatários de um envelope usando o provider
func (u *UsecaseEnvelopeProviderService) NotifyEnvelope(ctx context.Context, envelopeID int, message string) error {
	// Buscar envelope no banco de dados
	envelope, err := u.repositoryEnvelope.GetByID(envelopeID)
	if err != nil {
		return fmt.Errorf("failed to get envelope: %w", err)
	}

	// Verificar se o envelope existe
	if envelope == nil {
		return fmt.Errorf("envelope not found")
	}

	// Verificar se o envelope tem chave do provider
	if envelope.ClicksignKey == "" {
		return fmt.Errorf("envelope does not have provider key")
	}

	// Enviar notificação para o provider
	err = u.envelopeProvider.NotifyEnvelope(ctx, envelope.ClicksignKey, message)
	if err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}

	return nil
}

// GetEnvelope obtém um envelope por ID
func (u *UsecaseEnvelopeProviderService) GetEnvelope(id int) (*entity.EntityEnvelope, error) {
	envelope, err := u.repositoryEnvelope.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("envelope not found: %w", err)
	}

	return envelope, nil
}

// GetEnvelopes obtém uma lista de envelopes com filtros
func (u *UsecaseEnvelopeProviderService) GetEnvelopes(filters entity.EntityEnvelopeFilters) ([]entity.EntityEnvelope, error) {
	envelopes, err := u.repositoryEnvelope.GetEnvelopes(filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get envelopes: %w", err)
	}

	return envelopes, nil
}

// ValidateBusinessRules valida regras de negócio para envelopes
func (u *UsecaseEnvelopeProviderService) ValidateBusinessRules(envelope *entity.EntityEnvelope) error {
	// Validar limite de signatários (exemplo: máximo 50)
	if len(envelope.SignatoryEmails) > 50 {
		return fmt.Errorf("envelope cannot have more than 50 signatories")
	}

	// Validar duplicatas de email
	emailMap := make(map[string]bool)
	for _, email := range envelope.SignatoryEmails {
		if emailMap[email] {
			return fmt.Errorf("duplicate signatory email: %s", email)
		}
		emailMap[email] = true
	}

	// Validar que apenas envelopes em draft podem ser criados
	if envelope.Status != "draft" && envelope.Status != "" {
		return fmt.Errorf("new envelopes must be in 'draft' status")
	}

	return nil
}
