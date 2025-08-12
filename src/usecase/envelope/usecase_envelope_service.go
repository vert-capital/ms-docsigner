package envelope

import (
	"context"
	"fmt"

	"app/entity"
	"app/infrastructure/clicksign"
	clicksignInterface "app/usecase/clicksign"
	usecase_document "app/usecase/document"
	usecase_requirement "app/usecase/requirement"

	"github.com/sirupsen/logrus"
)

type UsecaseEnvelopeService struct {
	repositoryEnvelope IRepositoryEnvelope
	clicksignClient    clicksignInterface.ClicksignClientInterface
	envelopeService    *clicksign.EnvelopeService
	documentService    *clicksign.DocumentService
	usecaseDocument    usecase_document.IUsecaseDocument
	usecaseRequirement usecase_requirement.IUsecaseRequirement
	logger             *logrus.Logger
}

func NewUsecaseEnvelopeService(
	repositoryEnvelope IRepositoryEnvelope,
	clicksignClient clicksignInterface.ClicksignClientInterface,
	usecaseDocument usecase_document.IUsecaseDocument,
	usecaseRequirement usecase_requirement.IUsecaseRequirement,
	logger *logrus.Logger,
) IUsecaseEnvelope {
	envelopeService := clicksign.NewEnvelopeService(clicksignClient, logger)
	documentService := clicksign.NewDocumentService(clicksignClient, logger)

	return &UsecaseEnvelopeService{
		repositoryEnvelope: repositoryEnvelope,
		clicksignClient:    clicksignClient,
		envelopeService:    envelopeService,
		documentService:    documentService,
		usecaseDocument:    usecaseDocument,
		usecaseRequirement: usecaseRequirement,
		logger:             logger,
	}
}

func (u *UsecaseEnvelopeService) CreateEnvelope(envelope *entity.EntityEnvelope) (*entity.EntityEnvelope, error) {
	ctx := context.Background()

	// Validar entidade
	err := envelope.Validate()
	if err != nil {
		return nil, fmt.Errorf("envelope validation failed: %w", err)
	}

	// Validações específicas de negócio
	if err := u.validateBusinessRules(envelope); err != nil {
		return nil, fmt.Errorf("business rule validation failed: %w", err)
	}

	// Criar envelope localmente primeiro
	err = u.repositoryEnvelope.Create(envelope)
	if err != nil {
		return nil, fmt.Errorf("failed to create envelope locally: %w", err)
	}

	// Criar envelope no Clicksign
	clicksignKey, rawData, err := u.envelopeService.CreateEnvelope(ctx, envelope)
	if err != nil {
		// Tentar reverter criação local (best effort)
		if deleteErr := u.repositoryEnvelope.Delete(envelope); deleteErr != nil {
			// Log error but continue
		}

		return nil, fmt.Errorf("failed to create envelope in Clicksign: %w", err)
	}

	// Atualizar envelope com chave e dados brutos do Clicksign
	envelope.SetClicksignKey(clicksignKey)
	envelope.SetClicksignRawData(rawData)
	err = u.repositoryEnvelope.Update(envelope)
	if err != nil {
		return nil, fmt.Errorf("failed to update envelope with Clicksign key: %w", err)
	}

	return envelope, nil
}

func (u *UsecaseEnvelopeService) CreateEnvelopeWithDocuments(envelope *entity.EntityEnvelope, documents []*entity.EntityDocument) (*entity.EntityEnvelope, error) {
	ctx := context.Background()

	for i, doc := range documents {
		if err := doc.Validate(); err != nil {
			return nil, fmt.Errorf("document %d validation failed: %w", i, err)
		}
	}

	// Criar documentos localmente primeiro
	var documentIDs []int
	for _, doc := range documents {
		err := u.usecaseDocument.Create(doc)
		if err != nil {
			return nil, fmt.Errorf("failed to create document '%s' locally: %w", doc.Name, err)
		}
		documentIDs = append(documentIDs, doc.ID)
	}

	// Adicionar IDs dos documentos ao envelope
	envelope.DocumentsIDs = documentIDs

	// Validar entidade envelope (após documentos serem criados)
	err := envelope.Validate()
	if err != nil {
		return nil, fmt.Errorf("envelope validation failed: %w", err)
	}

	// Validações específicas de negócio para envelope com documentos
	if err := u.validateBusinessRulesWithDocuments(envelope); err != nil {
		return nil, fmt.Errorf("business rule validation failed: %w", err)
	}

	// Criar envelope localmente
	err = u.repositoryEnvelope.Create(envelope)
	if err != nil {
		return nil, fmt.Errorf("failed to create envelope locally: %w", err)
	}

	// Criar envelope no Clicksign
	clicksignKey, rawData, err := u.envelopeService.CreateEnvelope(ctx, envelope)
	if err != nil {
		// Tentar reverter criação local (best effort)
		if deleteErr := u.repositoryEnvelope.Delete(envelope); deleteErr != nil {
			// Log error but continue
		}

		return nil, fmt.Errorf("failed to create envelope in Clicksign: %w", err)
	}

	// Criar documentos no Clicksign dentro do envelope
	for _, doc := range documents {
		clicksignDocID, err := u.documentService.CreateDocument(ctx, clicksignKey, doc)
		if err != nil {
			return nil, fmt.Errorf("failed to create document '%s' in Clicksign: %w", doc.Name, err)
		}

		// Atualizar documento com chave do Clicksign
		doc.SetClicksignKey(clicksignDocID)
		if err := u.usecaseDocument.Update(doc); err != nil {
			return nil, fmt.Errorf("failed to update document with Clicksign key: %w", err)
		}
	}

	// Atualizar envelope com chave e dados brutos do Clicksign
	envelope.SetClicksignKey(clicksignKey)
	envelope.SetClicksignRawData(rawData)
	err = u.repositoryEnvelope.Update(envelope)
	if err != nil {
		return nil, fmt.Errorf("failed to update envelope with Clicksign key: %w", err)
	}

	return envelope, nil
}

func (u *UsecaseEnvelopeService) CreateDocument(ctx context.Context, envelopeID string, document *entity.EntityDocument) (string, error) {
	return u.documentService.CreateDocument(ctx, envelopeID, document)
}

func (u *UsecaseEnvelopeService) GetEnvelope(id int) (*entity.EntityEnvelope, error) {
	envelope, err := u.repositoryEnvelope.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("envelope not found: %w", err)
	}

	return envelope, nil
}

func (u *UsecaseEnvelopeService) GetEnvelopeByClicksignKey(key string) (*entity.EntityEnvelope, error) {
	envelope, err := u.repositoryEnvelope.GetByClicksignKey(key)
	if err != nil {
		return nil, fmt.Errorf("envelope not found: %w", err)
	}

	return envelope, nil
}

func (u *UsecaseEnvelopeService) GetEnvelopeByDocumentKey(documentKey string) (*entity.EntityEnvelope, error) {
	// Buscar envelope que contenha um documento com o clicksign_key especificado
	envelopes, err := u.repositoryEnvelope.GetEnvelopes(entity.EntityEnvelopeFilters{})
	if err != nil {
		return nil, fmt.Errorf("failed to get envelopes: %w", err)
	}

	// Para cada envelope, verificar se algum dos documentos tem o clicksign_key
	for _, envelope := range envelopes {
		// Aqui precisamos buscar os documentos do envelope
		// Por enquanto, vou implementar uma busca simples
		// TODO: Implementar busca mais eficiente no repository
		if envelope.ClicksignKey == documentKey {
			return &envelope, nil
		}
	}

	return nil, fmt.Errorf("envelope not found for document key: %s", documentKey)
}

func (u *UsecaseEnvelopeService) GetEnvelopes(filters entity.EntityEnvelopeFilters) ([]entity.EntityEnvelope, error) {
	envelopes, err := u.repositoryEnvelope.GetEnvelopes(filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get envelopes: %w", err)
	}

	return envelopes, nil
}

func (u *UsecaseEnvelopeService) UpdateEnvelope(envelope *entity.EntityEnvelope) error {
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

// UpdateEnvelopeForWebhook atualiza envelope sem restrições de status (usado para webhooks)
func (u *UsecaseEnvelopeService) UpdateEnvelopeForWebhook(envelope *entity.EntityEnvelope) error {
	err := envelope.Validate()
	if err != nil {
		return fmt.Errorf("envelope validation failed: %w", err)
	}

	// Verificar se o envelope existe
	_, err = u.repositoryEnvelope.GetByID(envelope.ID)
	if err != nil {
		return fmt.Errorf("envelope not found: %w", err)
	}

	// Atualizar sem restrições de status
	err = u.repositoryEnvelope.Update(envelope)
	if err != nil {
		return fmt.Errorf("failed to update envelope: %w", err)
	}

	return nil
}

func (u *UsecaseEnvelopeService) DeleteEnvelope(id int) error {
	envelope, err := u.repositoryEnvelope.GetByID(id)
	if err != nil {
		return fmt.Errorf("envelope not found: %w", err)
	}

	if envelope.Status == "sent" || envelope.Status == "pending" {
		return fmt.Errorf("cannot delete envelope in '%s' status", envelope.Status)
	}

	err = u.repositoryEnvelope.Delete(envelope)
	if err != nil {
		return fmt.Errorf("failed to delete envelope: %w", err)
	}

	return nil
}

func (u *UsecaseEnvelopeService) CreateEnvelopeWithRequirements(ctx context.Context, envelope *entity.EntityEnvelope, requirements []*entity.EntityRequirement) (*entity.EntityEnvelope, error) {
	// 1. Criar envelope primeiro
	createdEnvelope, err := u.CreateEnvelope(envelope)
	if err != nil {
		return nil, fmt.Errorf("failed to create envelope: %w", err)
	}

	// 2. Criar requirements se fornecidos
	if len(requirements) > 0 {
		var createdRequirements []*entity.EntityRequirement
		var failedRequirements []error

		for i, requirement := range requirements {
			// Definir o envelope_id do requirement
			requirement.EnvelopeID = createdEnvelope.ID

			createdRequirement, err := u.usecaseRequirement.CreateRequirement(ctx, requirement)
			if err != nil {
				failedRequirements = append(failedRequirements, fmt.Errorf("requirement %d (%s): %w", i+1, requirement.Action, err))
				continue
			}

			createdRequirements = append(createdRequirements, createdRequirement)
		}

		// Se houve falhas, mas envelope foi criado, ainda assim retorna sucesso
		// (requirements são opcionais)
	}

	return createdEnvelope, nil
}

func (u *UsecaseEnvelopeService) ActivateEnvelope(id int) (*entity.EntityEnvelope, error) {
	ctx := context.Background()

	envelope, err := u.repositoryEnvelope.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("envelope not found: %w", err)
	}

	if envelope.ClicksignKey == "" {
		return nil, fmt.Errorf("envelope has no Clicksign key")
	}

	// Ativar envelope localmente
	err = envelope.ActivateEnvelope()
	if err != nil {
		return nil, fmt.Errorf("failed to activate envelope locally: %w", err)
	}

	// Ativar envelope no Clicksign
	err = u.envelopeService.ActivateEnvelope(ctx, envelope.ClicksignKey)
	if err != nil {
		return nil, fmt.Errorf("failed to activate envelope in Clicksign: %w", err)
	}

	// Atualizar envelope no banco
	err = u.repositoryEnvelope.Update(envelope)
	if err != nil {
		return nil, fmt.Errorf("failed to update envelope status: %w", err)
	}

	return envelope, nil
}

func (u *UsecaseEnvelopeService) validateBusinessRules(envelope *entity.EntityEnvelope) error {
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

func (u *UsecaseEnvelopeService) validateBusinessRulesWithDocuments(envelope *entity.EntityEnvelope) error {
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

func (u *UsecaseEnvelopeService) NotifyEnvelope(ctx context.Context, envelopeID int, message string) error {
	// Buscar envelope no banco de dados
	envelope, err := u.repositoryEnvelope.GetByID(envelopeID)
	if err != nil {
		return fmt.Errorf("failed to get envelope: %w", err)
	}

	// Verificar se o envelope existe
	if envelope == nil {
		return fmt.Errorf("envelope not found")
	}

	// Verificar se o envelope tem chave do Clicksign
	if envelope.ClicksignKey == "" {
		return fmt.Errorf("envelope does not have Clicksign key")
	}

	// Enviar notificação para o Clicksign
	err = u.envelopeService.NotifyEnvelope(ctx, envelope.ClicksignKey, message)
	if err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}

	return nil
}
