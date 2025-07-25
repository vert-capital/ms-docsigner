package signatory

import (
	"context"
	"fmt"

	"app/entity"
	"app/infrastructure/clicksign"
	clicksignInterface "app/usecase/clicksign"
	usecase_envelope "app/usecase/envelope"
	"github.com/sirupsen/logrus"
)

type UsecaseSignatoryService struct {
	repositorySignatory IRepositorySignatory
	repositoryEnvelope  usecase_envelope.IRepositoryEnvelope
	clicksignClient     clicksignInterface.ClicksignClientInterface
	signerService       *clicksign.SignerService
	signatoryMapper     *clicksign.SignatoryMapper
	logger              *logrus.Logger
}

func NewUsecaseSignatoryService(
	repositorySignatory IRepositorySignatory,
	repositoryEnvelope usecase_envelope.IRepositoryEnvelope,
	clicksignClient clicksignInterface.ClicksignClientInterface,
	logger *logrus.Logger,
) IUsecaseSignatory {
	signerService := clicksign.NewSignerService(clicksignClient, logger)
	signatoryMapper := clicksign.NewSignatoryMapper()

	return &UsecaseSignatoryService{
		repositorySignatory: repositorySignatory,
		repositoryEnvelope:  repositoryEnvelope,
		clicksignClient:     clicksignClient,
		signerService:       signerService,
		signatoryMapper:     signatoryMapper,
		logger:              logger,
	}
}

func (u *UsecaseSignatoryService) CreateSignatory(signatory *entity.EntitySignatory) (*entity.EntitySignatory, error) {
	// Validar entidade
	err := signatory.Validate()
	if err != nil {
		return nil, fmt.Errorf("signatory validation failed: %w", err)
	}

	// Validações específicas de negócio
	if err := u.validateBusinessRules(signatory); err != nil {
		return nil, fmt.Errorf("business rule validation failed: %w", err)
	}

	// Criar signatário localmente primeiro
	err = u.repositorySignatory.Create(signatory)
	if err != nil {
		return nil, fmt.Errorf("failed to create signatory locally: %w", err)
	}

	// Obter envelope para pegar a chave do Clicksign
	envelope, err := u.repositoryEnvelope.GetByID(signatory.EnvelopeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get envelope for Clicksign integration: %w", err)
	}

	// Verificar se envelope tem chave do Clicksign
	if envelope.ClicksignKey == "" {
		return nil, fmt.Errorf("envelope has no Clicksign key, cannot create signatory in Clicksign")
	}

	// Validar dados para Clicksign
	if err := u.signatoryMapper.ValidateForClicksign(signatory); err != nil {
		return nil, fmt.Errorf("signatory validation failed for Clicksign: %w", err)
	}

	// Mapear para estrutura do Clicksign
	clicksignRequest := u.signatoryMapper.ToClicksignCreateRequest(signatory)

	// Criar contexto para chamada do Clicksign
	ctx := context.Background()

	// Mapear para SignerData
	signerData := clicksign.SignerData{
		Name:             clicksignRequest.Data.Attributes.Name,
		Email:            clicksignRequest.Data.Attributes.Email,
		Birthday:         clicksignRequest.Data.Attributes.Birthday,
		PhoneNumber:      clicksignRequest.Data.Attributes.PhoneNumber,
		HasDocumentation: clicksignRequest.Data.Attributes.HasDocumentation,
		Refusable:        clicksignRequest.Data.Attributes.Refusable,
		Group:            clicksignRequest.Data.Attributes.Group,
	}

	// Mapear communicate events se fornecidos
	if clicksignRequest.Data.Attributes.CommunicateEvents != nil {
		signerData.CommunicateEvents = &clicksign.SignerCommunicateEventsData{
			DocumentSigned:    clicksignRequest.Data.Attributes.CommunicateEvents.DocumentSigned,
			SignatureRequest:  clicksignRequest.Data.Attributes.CommunicateEvents.SignatureRequest,
			SignatureReminder: clicksignRequest.Data.Attributes.CommunicateEvents.SignatureReminder,
		}
	}

	// Criar signatário no Clicksign
	clicksignSignerID, err := u.signerService.CreateSigner(ctx, envelope.ClicksignKey, signerData)
	if err != nil {
		// Tentar reverter criação local (best effort)
		if deleteErr := u.repositorySignatory.Delete(signatory); deleteErr != nil {
			// Log error but continue
		}

		return nil, fmt.Errorf("failed to create signatory in Clicksign: %w", err)
	}

	// Armazenar chave do Clicksign no signatário (se necessário)
	signatory.SetClicksignKey(clicksignSignerID)
	if err := u.repositorySignatory.Update(signatory); err != nil {
		return nil, fmt.Errorf("failed to update signatory with Clicksign key: %w", err)
	}

	return signatory, nil
}

func (u *UsecaseSignatoryService) GetSignatory(id int) (*entity.EntitySignatory, error) {
	signatory, err := u.repositorySignatory.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("signatory not found: %w", err)
	}

	return signatory, nil
}

func (u *UsecaseSignatoryService) GetSignatories(filters entity.EntitySignatoryFilters) ([]entity.EntitySignatory, error) {
	signatories, err := u.repositorySignatory.GetSignatories(filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get signatories: %w", err)
	}

	return signatories, nil
}

func (u *UsecaseSignatoryService) GetSignatoriesByEnvelope(envelopeID int) ([]entity.EntitySignatory, error) {
	// Validar se envelope existe
	_, err := u.repositoryEnvelope.GetByID(envelopeID)
	if err != nil {
		return nil, fmt.Errorf("envelope not found: %w", err)
	}

	signatories, err := u.repositorySignatory.GetByEnvelopeID(envelopeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get signatories by envelope: %w", err)
	}

	return signatories, nil
}

func (u *UsecaseSignatoryService) UpdateSignatory(signatory *entity.EntitySignatory) error {
	err := signatory.Validate()
	if err != nil {
		return fmt.Errorf("signatory validation failed: %w", err)
	}

	existingSignatory, err := u.repositorySignatory.GetByID(signatory.ID)
	if err != nil {
		return fmt.Errorf("signatory not found: %w", err)
	}

	// Verificar se envelope ainda permite modificações
	envelope, err := u.repositoryEnvelope.GetByID(existingSignatory.EnvelopeID)
	if err != nil {
		return fmt.Errorf("envelope not found for signatory: %w", err)
	}

	if envelope.Status == "completed" || envelope.Status == "cancelled" {
		return fmt.Errorf("cannot update signatory in envelope with '%s' status", envelope.Status)
	}

	// Validar mudança de envelope (se aplicável)
	if signatory.EnvelopeID != existingSignatory.EnvelopeID {
		if err := u.validateEnvelopeChange(signatory, existingSignatory.EnvelopeID); err != nil {
			return err
		}
	}

	err = u.repositorySignatory.Update(signatory)
	if err != nil {
		return fmt.Errorf("failed to update signatory: %w", err)
	}

	return nil
}

func (u *UsecaseSignatoryService) DeleteSignatory(id int) error {
	signatory, err := u.repositorySignatory.GetByID(id)
	if err != nil {
		return fmt.Errorf("signatory not found: %w", err)
	}

	// Verificar se envelope permite remoção
	envelope, err := u.repositoryEnvelope.GetByID(signatory.EnvelopeID)
	if err != nil {
		return fmt.Errorf("envelope not found for signatory: %w", err)
	}

	if envelope.Status == "sent" || envelope.Status == "pending" || envelope.Status == "completed" {
		return fmt.Errorf("cannot delete signatory from envelope in '%s' status", envelope.Status)
	}

	// Se o signatário foi criado no Clicksign, deletar de lá primeiro
	if signatory.ClicksignKey != "" && envelope.ClicksignKey != "" {
		ctx := context.Background()
		err = u.signerService.DeleteSigner(ctx, envelope.ClicksignKey, signatory.ClicksignKey)
		if err != nil {
			u.logger.WithFields(logrus.Fields{
				"signatory_id":      signatory.ID,
				"envelope_id":       envelope.ID,
				"clicksign_key":     signatory.ClicksignKey,
				"envelope_key":      envelope.ClicksignKey,
				"error":            err.Error(),
			}).Error("Failed to delete signatory from Clicksign")
			return fmt.Errorf("failed to delete signatory from Clicksign: %w", err)
		}

		u.logger.WithFields(logrus.Fields{
			"signatory_id":      signatory.ID,
			"envelope_id":       envelope.ID,
			"clicksign_key":     signatory.ClicksignKey,
			"envelope_key":      envelope.ClicksignKey,
		}).Info("Signatory successfully deleted from Clicksign")
	}

	// Deletar do banco local apenas se conseguiu deletar da Clicksign (ou se não estava na Clicksign)
	err = u.repositorySignatory.Delete(signatory)
	if err != nil {
		return fmt.Errorf("failed to delete signatory: %w", err)
	}

	u.logger.WithFields(logrus.Fields{
		"signatory_id": signatory.ID,
		"envelope_id":  envelope.ID,
		"email":        signatory.Email,
		"name":         signatory.Name,
	}).Info("Signatory successfully deleted locally")

	return nil
}

func (u *UsecaseSignatoryService) AssociateToEnvelope(signatoryID int, envelopeID int) error {
	signatory, err := u.repositorySignatory.GetByID(signatoryID)
	if err != nil {
		return fmt.Errorf("signatory not found: %w", err)
	}

	// Validar se envelope existe
	envelope, err := u.repositoryEnvelope.GetByID(envelopeID)
	if err != nil {
		return fmt.Errorf("envelope not found: %w", err)
	}

	// Verificar se envelope permite associação
	if envelope.Status == "completed" || envelope.Status == "cancelled" {
		return fmt.Errorf("cannot associate signatory to envelope with '%s' status", envelope.Status)
	}

	// Verificar se já existe signatário com mesmo email no envelope
	existingSignatory, err := u.repositorySignatory.GetByEmailAndEnvelopeID(signatory.Email, envelopeID)
	if err == nil && existingSignatory.ID != signatoryID {
		return fmt.Errorf("email '%s' already exists in envelope %d", signatory.Email, envelopeID)
	}

	// Atualizar envelope_id do signatário
	signatory.EnvelopeID = envelopeID
	err = u.repositorySignatory.Update(signatory)
	if err != nil {
		return fmt.Errorf("failed to associate signatory to envelope: %w", err)
	}

	return nil
}

func (u *UsecaseSignatoryService) validateBusinessRules(signatory *entity.EntitySignatory) error {
	// Validar se envelope existe
	envelope, err := u.repositoryEnvelope.GetByID(signatory.EnvelopeID)
	if err != nil {
		return fmt.Errorf("envelope not found: %w", err)
	}

	// Validar que apenas envelopes em draft ou sent permitem novos signatários
	if envelope.Status != "draft" && envelope.Status != "sent" {
		return fmt.Errorf("cannot add signatory to envelope in '%s' status", envelope.Status)
	}

	// Verificar se já existe signatário com mesmo email no envelope
	existingSignatory, err := u.repositorySignatory.GetByEmailAndEnvelopeID(signatory.Email, signatory.EnvelopeID)
	if err == nil && existingSignatory.ID != signatory.ID {
		return fmt.Errorf("email '%s' already exists in envelope %d", signatory.Email, signatory.EnvelopeID)
	}

	// Validar limite de signatários por envelope (máximo 50)
	existingSignatories, err := u.repositorySignatory.GetByEnvelopeID(signatory.EnvelopeID)
	if err != nil {
		return fmt.Errorf("failed to check existing signatories: %w", err)
	}

	if len(existingSignatories) >= 50 {
		return fmt.Errorf("envelope cannot have more than 50 signatories")
	}

	return nil
}

func (u *UsecaseSignatoryService) validateEnvelopeChange(signatory *entity.EntitySignatory, oldEnvelopeID int) error {
	// Validar se novo envelope existe
	newEnvelope, err := u.repositoryEnvelope.GetByID(signatory.EnvelopeID)
	if err != nil {
		return fmt.Errorf("new envelope not found: %w", err)
	}

	// Validar se antigo envelope permite remoção
	oldEnvelope, err := u.repositoryEnvelope.GetByID(oldEnvelopeID)
	if err != nil {
		return fmt.Errorf("old envelope not found: %w", err)
	}

	if oldEnvelope.Status == "sent" || oldEnvelope.Status == "pending" || oldEnvelope.Status == "completed" {
		return fmt.Errorf("cannot remove signatory from envelope in '%s' status", oldEnvelope.Status)
	}

	// Validar se novo envelope permite adição
	if newEnvelope.Status != "draft" && newEnvelope.Status != "sent" {
		return fmt.Errorf("cannot add signatory to envelope in '%s' status", newEnvelope.Status)
	}

	// Verificar duplicação de email no novo envelope
	existingSignatory, err := u.repositorySignatory.GetByEmailAndEnvelopeID(signatory.Email, signatory.EnvelopeID)
	if err == nil && existingSignatory.ID != signatory.ID {
		return fmt.Errorf("email '%s' already exists in new envelope %d", signatory.Email, signatory.EnvelopeID)
	}

	return nil
}