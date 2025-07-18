package signatory

import (
	"fmt"

	"app/entity"
	usecase_envelope "app/usecase/envelope"
	"github.com/sirupsen/logrus"
)

type UsecaseSignatoryService struct {
	repositorySignatory IRepositorySignatory
	repositoryEnvelope  usecase_envelope.IRepositoryEnvelope
	logger              *logrus.Logger
}

func NewUsecaseSignatoryService(
	repositorySignatory IRepositorySignatory,
	repositoryEnvelope usecase_envelope.IRepositoryEnvelope,
	logger *logrus.Logger,
) IUsecaseSignatory {
	return &UsecaseSignatoryService{
		repositorySignatory: repositorySignatory,
		repositoryEnvelope:  repositoryEnvelope,
		logger:              logger,
	}
}

func (u *UsecaseSignatoryService) CreateSignatory(signatory *entity.EntitySignatory) (*entity.EntitySignatory, error) {
	u.logger.WithFields(logrus.Fields{
		"signatory_name":  signatory.Name,
		"signatory_email": signatory.Email,
		"envelope_id":     signatory.EnvelopeID,
	}).Info("Starting signatory creation process")

	// Validar entidade
	err := signatory.Validate()
	if err != nil {
		u.logger.WithFields(logrus.Fields{
			"error":           err.Error(),
			"signatory_email": signatory.Email,
			"envelope_id":     signatory.EnvelopeID,
		}).Error("Signatory validation failed")
		return nil, fmt.Errorf("signatory validation failed: %w", err)
	}

	// Validações específicas de negócio
	if err := u.validateBusinessRules(signatory); err != nil {
		u.logger.WithFields(logrus.Fields{
			"error":           err.Error(),
			"signatory_email": signatory.Email,
			"envelope_id":     signatory.EnvelopeID,
		}).Error("Business rule validation failed")
		return nil, fmt.Errorf("business rule validation failed: %w", err)
	}

	// Criar signatário localmente
	err = u.repositorySignatory.Create(signatory)
	if err != nil {
		u.logger.WithFields(logrus.Fields{
			"error":           err.Error(),
			"signatory_email": signatory.Email,
			"envelope_id":     signatory.EnvelopeID,
		}).Error("Failed to create signatory locally")
		return nil, fmt.Errorf("failed to create signatory locally: %w", err)
	}

	u.logger.WithFields(logrus.Fields{
		"signatory_id":    signatory.ID,
		"signatory_name":  signatory.Name,
		"signatory_email": signatory.Email,
		"envelope_id":     signatory.EnvelopeID,
	}).Info("Signatory created successfully")

	return signatory, nil
}

func (u *UsecaseSignatoryService) GetSignatory(id int) (*entity.EntitySignatory, error) {
	signatory, err := u.repositorySignatory.GetByID(id)
	if err != nil {
		u.logger.WithFields(logrus.Fields{
			"error":        err.Error(),
			"signatory_id": id,
		}).Error("Failed to get signatory")
		return nil, fmt.Errorf("signatory not found: %w", err)
	}

	return signatory, nil
}

func (u *UsecaseSignatoryService) GetSignatories(filters entity.EntitySignatoryFilters) ([]entity.EntitySignatory, error) {
	signatories, err := u.repositorySignatory.GetSignatories(filters)
	if err != nil {
		u.logger.WithFields(logrus.Fields{
			"error":   err.Error(),
			"filters": filters,
		}).Error("Failed to get signatories")
		return nil, fmt.Errorf("failed to get signatories: %w", err)
	}

	return signatories, nil
}

func (u *UsecaseSignatoryService) GetSignatoriesByEnvelope(envelopeID int) ([]entity.EntitySignatory, error) {
	// Validar se envelope existe
	_, err := u.repositoryEnvelope.GetByID(envelopeID)
	if err != nil {
		u.logger.WithFields(logrus.Fields{
			"error":       err.Error(),
			"envelope_id": envelopeID,
		}).Error("Envelope not found")
		return nil, fmt.Errorf("envelope not found: %w", err)
	}

	signatories, err := u.repositorySignatory.GetByEnvelopeID(envelopeID)
	if err != nil {
		u.logger.WithFields(logrus.Fields{
			"error":       err.Error(),
			"envelope_id": envelopeID,
		}).Error("Failed to get signatories by envelope")
		return nil, fmt.Errorf("failed to get signatories by envelope: %w", err)
	}

	return signatories, nil
}

func (u *UsecaseSignatoryService) UpdateSignatory(signatory *entity.EntitySignatory) error {
	err := signatory.Validate()
	if err != nil {
		u.logger.WithFields(logrus.Fields{
			"error":        err.Error(),
			"signatory_id": signatory.ID,
			"email":        signatory.Email,
		}).Error("Signatory validation failed")
		return fmt.Errorf("signatory validation failed: %w", err)
	}

	existingSignatory, err := u.repositorySignatory.GetByID(signatory.ID)
	if err != nil {
		u.logger.WithFields(logrus.Fields{
			"error":        err.Error(),
			"signatory_id": signatory.ID,
		}).Error("Signatory not found")
		return fmt.Errorf("signatory not found: %w", err)
	}

	// Verificar se envelope ainda permite modificações
	envelope, err := u.repositoryEnvelope.GetByID(existingSignatory.EnvelopeID)
	if err != nil {
		u.logger.WithFields(logrus.Fields{
			"error":       err.Error(),
			"envelope_id": existingSignatory.EnvelopeID,
		}).Error("Envelope not found for signatory")
		return fmt.Errorf("envelope not found for signatory: %w", err)
	}

	if envelope.Status == "completed" || envelope.Status == "cancelled" {
		u.logger.WithFields(logrus.Fields{
			"signatory_id": signatory.ID,
			"envelope_id":  envelope.ID,
			"status":       envelope.Status,
		}).Warn("Cannot update signatory in envelope with final status")
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
		u.logger.WithFields(logrus.Fields{
			"error":        err.Error(),
			"signatory_id": signatory.ID,
		}).Error("Failed to update signatory")
		return fmt.Errorf("failed to update signatory: %w", err)
	}

	u.logger.WithFields(logrus.Fields{
		"signatory_id": signatory.ID,
		"email":        signatory.Email,
	}).Info("Signatory updated successfully")

	return nil
}

func (u *UsecaseSignatoryService) DeleteSignatory(id int) error {
	signatory, err := u.repositorySignatory.GetByID(id)
	if err != nil {
		u.logger.WithFields(logrus.Fields{
			"error":        err.Error(),
			"signatory_id": id,
		}).Error("Signatory not found")
		return fmt.Errorf("signatory not found: %w", err)
	}

	// Verificar se envelope permite remoção
	envelope, err := u.repositoryEnvelope.GetByID(signatory.EnvelopeID)
	if err != nil {
		u.logger.WithFields(logrus.Fields{
			"error":       err.Error(),
			"envelope_id": signatory.EnvelopeID,
		}).Error("Envelope not found for signatory")
		return fmt.Errorf("envelope not found for signatory: %w", err)
	}

	if envelope.Status == "sent" || envelope.Status == "pending" || envelope.Status == "completed" {
		u.logger.WithFields(logrus.Fields{
			"signatory_id": id,
			"envelope_id":  envelope.ID,
			"status":       envelope.Status,
		}).Warn("Cannot delete signatory from envelope in active/final status")
		return fmt.Errorf("cannot delete signatory from envelope in '%s' status", envelope.Status)
	}

	err = u.repositorySignatory.Delete(signatory)
	if err != nil {
		u.logger.WithFields(logrus.Fields{
			"error":        err.Error(),
			"signatory_id": id,
		}).Error("Failed to delete signatory")
		return fmt.Errorf("failed to delete signatory: %w", err)
	}

	u.logger.WithFields(logrus.Fields{
		"signatory_id": id,
		"email":        signatory.Email,
	}).Info("Signatory deleted successfully")

	return nil
}

func (u *UsecaseSignatoryService) AssociateToEnvelope(signatoryID int, envelopeID int) error {
	signatory, err := u.repositorySignatory.GetByID(signatoryID)
	if err != nil {
		u.logger.WithFields(logrus.Fields{
			"error":        err.Error(),
			"signatory_id": signatoryID,
		}).Error("Signatory not found")
		return fmt.Errorf("signatory not found: %w", err)
	}

	// Validar se envelope existe
	envelope, err := u.repositoryEnvelope.GetByID(envelopeID)
	if err != nil {
		u.logger.WithFields(logrus.Fields{
			"error":       err.Error(),
			"envelope_id": envelopeID,
		}).Error("Envelope not found")
		return fmt.Errorf("envelope not found: %w", err)
	}

	// Verificar se envelope permite associação
	if envelope.Status == "completed" || envelope.Status == "cancelled" {
		u.logger.WithFields(logrus.Fields{
			"signatory_id": signatoryID,
			"envelope_id":  envelopeID,
			"status":       envelope.Status,
		}).Warn("Cannot associate signatory to envelope with final status")
		return fmt.Errorf("cannot associate signatory to envelope with '%s' status", envelope.Status)
	}

	// Verificar se já existe signatário com mesmo email no envelope
	existingSignatory, err := u.repositorySignatory.GetByEmailAndEnvelopeID(signatory.Email, envelopeID)
	if err == nil && existingSignatory.ID != signatoryID {
		u.logger.WithFields(logrus.Fields{
			"signatory_id":          signatoryID,
			"existing_signatory_id": existingSignatory.ID,
			"email":                 signatory.Email,
			"envelope_id":           envelopeID,
		}).Warn("Email already exists in this envelope")
		return fmt.Errorf("email '%s' already exists in envelope %d", signatory.Email, envelopeID)
	}

	// Atualizar envelope_id do signatário
	signatory.EnvelopeID = envelopeID
	err = u.repositorySignatory.Update(signatory)
	if err != nil {
		u.logger.WithFields(logrus.Fields{
			"error":        err.Error(),
			"signatory_id": signatoryID,
			"envelope_id":  envelopeID,
		}).Error("Failed to associate signatory to envelope")
		return fmt.Errorf("failed to associate signatory to envelope: %w", err)
	}

	u.logger.WithFields(logrus.Fields{
		"signatory_id": signatoryID,
		"envelope_id":  envelopeID,
		"email":        signatory.Email,
	}).Info("Signatory associated to envelope successfully")

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