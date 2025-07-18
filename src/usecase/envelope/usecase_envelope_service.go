package envelope

import (
	"context"
	"fmt"

	"app/entity"
	"app/infrastructure/clicksign"
	clicksignInterface "app/usecase/clicksign"
	"github.com/sirupsen/logrus"
)

type UsecaseEnvelopeService struct {
	repositoryEnvelope IRepositoryEnvelope
	clicksignClient    clicksignInterface.ClicksignClientInterface
	envelopeService    *clicksign.EnvelopeService
	logger             *logrus.Logger
}

func NewUsecaseEnvelopeService(
	repositoryEnvelope IRepositoryEnvelope,
	clicksignClient clicksignInterface.ClicksignClientInterface,
	logger *logrus.Logger,
) IUsecaseEnvelope {
	envelopeService := clicksign.NewEnvelopeService(clicksignClient, logger)

	return &UsecaseEnvelopeService{
		repositoryEnvelope: repositoryEnvelope,
		clicksignClient:    clicksignClient,
		envelopeService:    envelopeService,
		logger:             logger,
	}
}

func (u *UsecaseEnvelopeService) CreateEnvelope(envelope *entity.EntityEnvelope) (*entity.EntityEnvelope, error) {
	ctx := context.Background()
	correlationID := fmt.Sprintf("envelope_%d_%d", envelope.ID, envelope.CreatedAt.Unix())
	ctx = context.WithValue(ctx, "correlation_id", correlationID)

	u.logger.WithFields(logrus.Fields{
		"envelope_name":   envelope.Name,
		"documents_count": len(envelope.DocumentsIDs),
		"signers_count":   len(envelope.SignatoryEmails),
		"correlation_id":  correlationID,
	}).Info("Starting envelope creation process")

	// Validar entidade
	err := envelope.Validate()
	if err != nil {
		u.logger.WithFields(logrus.Fields{
			"error":          err.Error(),
			"envelope_name":  envelope.Name,
			"correlation_id": correlationID,
		}).Error("Envelope validation failed")
		return nil, fmt.Errorf("envelope validation failed: %w", err)
	}

	// Validações específicas de negócio
	if err := u.validateBusinessRules(envelope); err != nil {
		u.logger.WithFields(logrus.Fields{
			"error":          err.Error(),
			"envelope_name":  envelope.Name,
			"correlation_id": correlationID,
		}).Error("Business rule validation failed")
		return nil, fmt.Errorf("business rule validation failed: %w", err)
	}

	// Criar envelope localmente primeiro
	err = u.repositoryEnvelope.Create(envelope)
	if err != nil {
		u.logger.WithFields(logrus.Fields{
			"error":          err.Error(),
			"envelope_name":  envelope.Name,
			"correlation_id": correlationID,
		}).Error("Failed to create envelope locally")
		return nil, fmt.Errorf("failed to create envelope locally: %w", err)
	}

	u.logger.WithFields(logrus.Fields{
		"envelope_id":    envelope.ID,
		"envelope_name":  envelope.Name,
		"correlation_id": correlationID,
	}).Info("Envelope created locally, now creating in Clicksign")

	// Criar envelope no Clicksign
	clicksignKey, err := u.envelopeService.CreateEnvelope(ctx, envelope)
	if err != nil {
		u.logger.WithFields(logrus.Fields{
			"error":          err.Error(),
			"envelope_id":    envelope.ID,
			"envelope_name":  envelope.Name,
			"correlation_id": correlationID,
		}).Error("Failed to create envelope in Clicksign")

		// Tentar reverter criação local (best effort)
		if deleteErr := u.repositoryEnvelope.Delete(envelope); deleteErr != nil {
			u.logger.WithFields(logrus.Fields{
				"error":          deleteErr.Error(),
				"envelope_id":    envelope.ID,
				"correlation_id": correlationID,
			}).Error("Failed to rollback local envelope creation")
		}

		return nil, fmt.Errorf("failed to create envelope in Clicksign: %w", err)
	}

	// Atualizar envelope com chave do Clicksign
	envelope.SetClicksignKey(clicksignKey)
	err = u.repositoryEnvelope.Update(envelope)
	if err != nil {
		u.logger.WithFields(logrus.Fields{
			"error":          err.Error(),
			"envelope_id":    envelope.ID,
			"clicksign_key":  clicksignKey,
			"correlation_id": correlationID,
		}).Error("Failed to update envelope with Clicksign key")
		return nil, fmt.Errorf("failed to update envelope with Clicksign key: %w", err)
	}

	u.logger.WithFields(logrus.Fields{
		"envelope_id":    envelope.ID,
		"envelope_name":  envelope.Name,
		"clicksign_key":  clicksignKey,
		"correlation_id": correlationID,
	}).Info("Envelope created successfully")

	return envelope, nil
}

func (u *UsecaseEnvelopeService) GetEnvelope(id int) (*entity.EntityEnvelope, error) {
	envelope, err := u.repositoryEnvelope.GetByID(id)
	if err != nil {
		u.logger.WithFields(logrus.Fields{
			"error":       err.Error(),
			"envelope_id": id,
		}).Error("Failed to get envelope")
		return nil, fmt.Errorf("envelope not found: %w", err)
	}

	return envelope, nil
}

func (u *UsecaseEnvelopeService) GetEnvelopes(filters entity.EntityEnvelopeFilters) ([]entity.EntityEnvelope, error) {
	envelopes, err := u.repositoryEnvelope.GetEnvelopes(filters)
	if err != nil {
		u.logger.WithFields(logrus.Fields{
			"error":   err.Error(),
			"filters": filters,
		}).Error("Failed to get envelopes")
		return nil, fmt.Errorf("failed to get envelopes: %w", err)
	}

	return envelopes, nil
}

func (u *UsecaseEnvelopeService) UpdateEnvelope(envelope *entity.EntityEnvelope) error {
	err := envelope.Validate()
	if err != nil {
		u.logger.WithFields(logrus.Fields{
			"error":         err.Error(),
			"envelope_id":   envelope.ID,
			"envelope_name": envelope.Name,
		}).Error("Envelope validation failed")
		return fmt.Errorf("envelope validation failed: %w", err)
	}

	existingEnvelope, err := u.repositoryEnvelope.GetByID(envelope.ID)
	if err != nil {
		u.logger.WithFields(logrus.Fields{
			"error":       err.Error(),
			"envelope_id": envelope.ID,
		}).Error("Envelope not found")
		return fmt.Errorf("envelope not found: %w", err)
	}

	if existingEnvelope.Status == "completed" || existingEnvelope.Status == "cancelled" {
		u.logger.WithFields(logrus.Fields{
			"envelope_id": envelope.ID,
			"status":      existingEnvelope.Status,
		}).Warn("Cannot update envelope in final status")
		return fmt.Errorf("cannot update envelope in '%s' status", existingEnvelope.Status)
	}

	err = u.repositoryEnvelope.Update(envelope)
	if err != nil {
		u.logger.WithFields(logrus.Fields{
			"error":       err.Error(),
			"envelope_id": envelope.ID,
		}).Error("Failed to update envelope")
		return fmt.Errorf("failed to update envelope: %w", err)
	}

	u.logger.WithFields(logrus.Fields{
		"envelope_id":   envelope.ID,
		"envelope_name": envelope.Name,
	}).Info("Envelope updated successfully")

	return nil
}

func (u *UsecaseEnvelopeService) DeleteEnvelope(id int) error {
	envelope, err := u.repositoryEnvelope.GetByID(id)
	if err != nil {
		u.logger.WithFields(logrus.Fields{
			"error":       err.Error(),
			"envelope_id": id,
		}).Error("Envelope not found")
		return fmt.Errorf("envelope not found: %w", err)
	}

	if envelope.Status == "sent" || envelope.Status == "pending" {
		u.logger.WithFields(logrus.Fields{
			"envelope_id": id,
			"status":      envelope.Status,
		}).Warn("Cannot delete envelope in active status")
		return fmt.Errorf("cannot delete envelope in '%s' status", envelope.Status)
	}

	err = u.repositoryEnvelope.Delete(envelope)
	if err != nil {
		u.logger.WithFields(logrus.Fields{
			"error":       err.Error(),
			"envelope_id": id,
		}).Error("Failed to delete envelope")
		return fmt.Errorf("failed to delete envelope: %w", err)
	}

	u.logger.WithFields(logrus.Fields{
		"envelope_id":   id,
		"envelope_name": envelope.Name,
	}).Info("Envelope deleted successfully")

	return nil
}

func (u *UsecaseEnvelopeService) ActivateEnvelope(id int) (*entity.EntityEnvelope, error) {
	ctx := context.Background()
	correlationID := fmt.Sprintf("activate_%d", id)
	ctx = context.WithValue(ctx, "correlation_id", correlationID)

	envelope, err := u.repositoryEnvelope.GetByID(id)
	if err != nil {
		u.logger.WithFields(logrus.Fields{
			"error":          err.Error(),
			"envelope_id":    id,
			"correlation_id": correlationID,
		}).Error("Envelope not found")
		return nil, fmt.Errorf("envelope not found: %w", err)
	}

	if envelope.ClicksignKey == "" {
		u.logger.WithFields(logrus.Fields{
			"envelope_id":    id,
			"correlation_id": correlationID,
		}).Error("Envelope has no Clicksign key")
		return nil, fmt.Errorf("envelope has no Clicksign key")
	}

	// Ativar envelope localmente
	err = envelope.ActivateEnvelope()
	if err != nil {
		u.logger.WithFields(logrus.Fields{
			"error":          err.Error(),
			"envelope_id":    id,
			"correlation_id": correlationID,
		}).Error("Failed to activate envelope locally")
		return nil, fmt.Errorf("failed to activate envelope locally: %w", err)
	}

	// Ativar envelope no Clicksign
	err = u.envelopeService.ActivateEnvelope(ctx, envelope.ClicksignKey)
	if err != nil {
		u.logger.WithFields(logrus.Fields{
			"error":          err.Error(),
			"envelope_id":    id,
			"clicksign_key":  envelope.ClicksignKey,
			"correlation_id": correlationID,
		}).Error("Failed to activate envelope in Clicksign")
		return nil, fmt.Errorf("failed to activate envelope in Clicksign: %w", err)
	}

	// Atualizar envelope no banco
	err = u.repositoryEnvelope.Update(envelope)
	if err != nil {
		u.logger.WithFields(logrus.Fields{
			"error":          err.Error(),
			"envelope_id":    id,
			"correlation_id": correlationID,
		}).Error("Failed to update envelope status")
		return nil, fmt.Errorf("failed to update envelope status: %w", err)
	}

	u.logger.WithFields(logrus.Fields{
		"envelope_id":    id,
		"envelope_name":  envelope.Name,
		"clicksign_key":  envelope.ClicksignKey,
		"correlation_id": correlationID,
	}).Info("Envelope activated successfully")

	return envelope, nil
}

func (u *UsecaseEnvelopeService) validateBusinessRules(envelope *entity.EntityEnvelope) error {
	// Validar que os documentos existem
	if len(envelope.DocumentsIDs) == 0 {
		return fmt.Errorf("envelope must have at least one document")
	}

	// Validar que os signatários existem
	if len(envelope.SignatoryEmails) == 0 {
		return fmt.Errorf("envelope must have at least one signatory")
	}

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
