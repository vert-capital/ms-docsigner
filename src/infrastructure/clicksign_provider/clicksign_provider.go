package clicksign_provider

import (
	"context"

	"app/entity"
	"app/infrastructure/clicksign"
	"app/infrastructure/provider"

	"github.com/sirupsen/logrus"
)

// ClicksignProvider implementa a interface EnvelopeProvider para o provider Clicksign
type ClicksignProvider struct {
	envelopeService    *clicksign.EnvelopeService
	documentService    *clicksign.DocumentService
	signerService      *clicksign.SignerService
	requirementService *clicksign.RequirementService
	logger             *logrus.Logger
}

// NewClicksignProvider cria uma nova instância do ClicksignProvider
func NewClicksignProvider(
	clicksignClient clicksign.ClicksignClientInterface,
	logger *logrus.Logger,
) provider.EnvelopeProvider {
	envelopeService := clicksign.NewEnvelopeService(clicksignClient, logger)
	documentService := clicksign.NewDocumentService(clicksignClient, logger)
	signerService := clicksign.NewSignerService(clicksignClient, logger)
	requirementService := clicksign.NewRequirementService(clicksignClient, logger)

	return &ClicksignProvider{
		envelopeService:    envelopeService,
		documentService:    documentService,
		signerService:      signerService,
		requirementService: requirementService,
		logger:             logger,
	}
}

// CreateEnvelope cria um envelope no Clicksign
func (p *ClicksignProvider) CreateEnvelope(ctx context.Context, envelope *entity.EntityEnvelope) (string, string, error) {
	return p.envelopeService.CreateEnvelope(ctx, envelope)
}

// CreateDocument cria um documento dentro de um envelope no Clicksign
func (p *ClicksignProvider) CreateDocument(ctx context.Context, envelopeKey string, document *entity.EntityDocument, internalEnvelopeID int) (string, error) {
	return p.documentService.CreateDocument(ctx, envelopeKey, document, internalEnvelopeID)
}

// CreateSigner cria um signatário no envelope do Clicksign
func (p *ClicksignProvider) CreateSigner(ctx context.Context, envelopeKey string, signerData provider.SignerData) (string, error) {
	// Converter SignerData genérico para SignerData do Clicksign
	clicksignSignerData := clicksign.SignerData{
		Name:              signerData.Name,
		Email:             signerData.Email,
		Birthday:          signerData.Birthday,
		Documentation:     signerData.Documentation,
		PhoneNumber:       signerData.PhoneNumber,
		HasDocumentation:  signerData.HasDocumentation,
		Refusable:         signerData.Refusable,
		Group:             signerData.Group,
	}

	// Mapear CommunicateEvents se fornecido
	if signerData.CommunicateEvents != nil {
		clicksignSignerData.CommunicateEvents = &clicksign.SignerCommunicateEventsData{
			DocumentSigned:    signerData.CommunicateEvents.DocumentSigned,
			SignatureRequest:  signerData.CommunicateEvents.SignatureRequest,
			SignatureReminder: signerData.CommunicateEvents.SignatureReminder,
		}
	}

	return p.signerService.CreateSigner(ctx, envelopeKey, clicksignSignerData)
}

// CreateRequirement cria um requisito de assinatura no envelope do Clicksign
func (p *ClicksignProvider) CreateRequirement(ctx context.Context, envelopeKey string, reqData provider.RequirementData) (string, error) {
	// Converter RequirementData genérico para RequirementData do Clicksign
	clicksignReqData := clicksign.RequirementData{
		Action:     reqData.Action,
		Role:       reqData.Role,
		Auth:       reqData.Auth,
		DocumentID: reqData.DocumentID,
		SignerID:   reqData.SignerID,
	}

	return p.requirementService.CreateRequirement(ctx, envelopeKey, clicksignReqData)
}

// ActivateEnvelope ativa um envelope no Clicksign
func (p *ClicksignProvider) ActivateEnvelope(ctx context.Context, envelopeKey string) error {
	return p.envelopeService.ActivateEnvelope(ctx, envelopeKey)
}

// NotifyEnvelope envia uma notificação para os signatários de um envelope no Clicksign
func (p *ClicksignProvider) NotifyEnvelope(ctx context.Context, envelopeKey string, message string) error {
	return p.envelopeService.NotifyEnvelope(ctx, envelopeKey, message)
}




