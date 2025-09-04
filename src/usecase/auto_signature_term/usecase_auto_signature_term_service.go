package auto_signature_term

import (
	"app/entity"
	"app/infrastructure/clicksign"
	"app/infrastructure/clicksign/dto"
	"context"
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"
)

type UsecaseAutoSignatureTermService struct {
	repository           IRepositoryAutoSignatureTerm
	clicksignClient      *clicksign.ClicksignClient
	autoSignatureService *clicksign.AutoSignatureService
	logger               *logrus.Logger
}

func NewUsecaseAutoSignatureTermService(
	repository IRepositoryAutoSignatureTerm,
	clicksignClient *clicksign.ClicksignClient,
	logger *logrus.Logger,
) *UsecaseAutoSignatureTermService {
	return &UsecaseAutoSignatureTermService{
		repository:           repository,
		clicksignClient:      clicksignClient,
		autoSignatureService: clicksign.NewAutoSignatureService(clicksignClient, logger),
		logger:               logger,
	}
}

func (u *UsecaseAutoSignatureTermService) CreateAutoSignatureTerm(ctx context.Context, term *entity.EntityAutoSignatureTerm) (*entity.EntityAutoSignatureTerm, error) {
	u.logger.WithFields(logrus.Fields{
		"signer_email": term.SignerEmail,
		"admin_email":  term.AdminEmail,
		"api_email":    term.APIEmail,
	}).Info("Creating auto signature term")

	// Criar o termo no Clicksign
	clicksignResponse, err := u.createTermInClicksign(term)
	if err != nil {
		u.logger.WithError(err).Error("Failed to create term in Clicksign")
		return nil, fmt.Errorf("failed to create term in Clicksign: %w", err)
	}

	// Salvar a resposta do Clicksign
	clicksignRawData, _ := json.Marshal(clicksignResponse)
	clicksignRawDataStr := string(clicksignRawData)
	term.ClicksignRawData = &clicksignRawDataStr

	// Extrair a chave do Clicksign da resposta
	if clicksignResponse.Data.ID != "" {
		term.ClicksignKey = clicksignResponse.Data.ID
	}

	// Salvar no banco de dados
	err = u.repository.Create(term)
	if err != nil {
		u.logger.WithError(err).Error("Failed to save term to database")
		return nil, fmt.Errorf("failed to save term to database: %w", err)
	}

	u.logger.WithFields(logrus.Fields{
		"term_id":       term.ID,
		"clicksign_key": term.ClicksignKey,
	}).Info("Auto signature term created successfully")

	return term, nil
}

func (u *UsecaseAutoSignatureTermService) GetAutoSignatureTerm(id int) (*entity.EntityAutoSignatureTerm, error) {
	term, err := u.repository.GetByID(id)
	if err != nil {
		u.logger.WithError(err).WithField("term_id", id).Error("Failed to get auto signature term")
		return nil, fmt.Errorf("failed to get auto signature term: %w", err)
	}
	return term, nil
}

func (u *UsecaseAutoSignatureTermService) GetAutoSignatureTermByClicksignKey(key string) (*entity.EntityAutoSignatureTerm, error) {
	term, err := u.repository.GetByClicksignKey(key)
	if err != nil {
		u.logger.WithError(err).WithField("clicksign_key", key).Error("Failed to get auto signature term by Clicksign key")
		return nil, fmt.Errorf("failed to get auto signature term by Clicksign key: %w", err)
	}
	return term, nil
}

func (u *UsecaseAutoSignatureTermService) GetAllAutoSignatureTerms() ([]entity.EntityAutoSignatureTerm, error) {
	terms, err := u.repository.GetAll()
	if err != nil {
		u.logger.WithError(err).Error("Failed to get all auto signature terms")
		return nil, fmt.Errorf("failed to get all auto signature terms: %w", err)
	}
	return terms, nil
}

func (u *UsecaseAutoSignatureTermService) UpdateAutoSignatureTerm(term *entity.EntityAutoSignatureTerm) error {
	err := u.repository.Update(term)
	if err != nil {
		u.logger.WithError(err).WithField("term_id", term.ID).Error("Failed to update auto signature term")
		return fmt.Errorf("failed to update auto signature term: %w", err)
	}
	return nil
}

func (u *UsecaseAutoSignatureTermService) DeleteAutoSignatureTerm(id int) error {
	term, err := u.repository.GetByID(id)
	if err != nil {
		u.logger.WithError(err).WithField("term_id", id).Error("Failed to get auto signature term for deletion")
		return fmt.Errorf("failed to get auto signature term for deletion: %w", err)
	}

	err = u.repository.Delete(term)
	if err != nil {
		u.logger.WithError(err).WithField("term_id", id).Error("Failed to delete auto signature term")
		return fmt.Errorf("failed to delete auto signature term: %w", err)
	}

	u.logger.WithField("term_id", id).Info("Auto signature term deleted successfully")
	return nil
}

func (u *UsecaseAutoSignatureTermService) createTermInClicksign(term *entity.EntityAutoSignatureTerm) (*dto.AutoSignatureTermResponse, error) {
	// Preparar o payload para o Clicksign
	payload := dto.AutoSignatureTermRequest{
		Data: dto.AutoSignatureTermData{
			Type: "auto_signature_terms",
			Attributes: dto.AutoSignatureTermAttributes{
				Signer: dto.SignerInfo{
					Documentation: term.SignerDocumentation,
					Birthday:      term.SignerBirthday,
					Email:         term.SignerEmail,
					Name:          term.SignerName,
				},
				AdminEmail: term.AdminEmail,
				APIEmail:   term.APIEmail,
			},
		},
	}

	// Fazer a requisição para o Clicksign
	response, err := u.autoSignatureService.CreateAutoSignatureTerm(payload)
	if err != nil {
		return nil, err
	}

	return response, nil
}
