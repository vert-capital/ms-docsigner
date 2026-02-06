package vertc_assinaturas_provider

import (
	"context"
	"encoding/json"
	"fmt"

	"app/entity"
	"app/infrastructure/provider"
	"app/infrastructure/vertc_assinaturas"

	"github.com/sirupsen/logrus"
)

// VertcAssinaturasProvider implementa a interface EnvelopeProvider para o provider vertc-assinaturas
type VertcAssinaturasProvider struct {
	quickSendService *vertc_assinaturas.QuickSendService
	logger           *logrus.Logger
}

// NewVertcAssinaturasProvider cria uma nova instância do VertcAssinaturasProvider
func NewVertcAssinaturasProvider(
	quickSendService *vertc_assinaturas.QuickSendService,
	logger *logrus.Logger,
) provider.EnvelopeProvider {
	return &VertcAssinaturasProvider{
		quickSendService: quickSendService,
		logger:           logger,
	}
}

// CreateEnvelope cria um envelope usando quick-send
// Para vertc-assinaturas, o quick-send cria envelope, documentos e signatários de uma vez
// Os documentos e signatários devem estar no contexto via QuickSendData
func (p *VertcAssinaturasProvider) CreateEnvelope(ctx context.Context, envelope *entity.EntityEnvelope) (string, string, error) {
	// Extrair dados do contexto
	quickSendData, ok := vertc_assinaturas.GetQuickSendDataFromContext(ctx)
	if !ok {
		return "", "", fmt.Errorf("quick-send data not found in context. Documents and signers must be provided via context for vertc-assinaturas provider")
	}

	// Validar que temos documentos
	if len(quickSendData.Documents) == 0 {
		return "", "", fmt.Errorf("at least one document is required for quick-send")
	}

	// Validar que temos signatários
	if len(quickSendData.Signers) == 0 {
		return "", "", fmt.Errorf("at least one signer is required for quick-send")
	}

	// Preparar dados para quick-send
	data := vertc_assinaturas.QuickSendData{
		Envelope:  envelope,
		Documents: quickSendData.Documents,
		Signers:   quickSendData.Signers,
	}

	// Chamar quick-send
	response, err := p.quickSendService.QuickSend(ctx, data)
	if err != nil {
		return "", "", fmt.Errorf("failed to create envelope via quick-send: %w", err)
	}

	// Serializar resposta para rawData
	rawDataBytes, err := json.Marshal(response)
	if err != nil {
		p.logger.Warnf("Failed to marshal quick-send response: %v", err)
		rawDataBytes = []byte("{}")
	}

	return response.EnvelopeID, string(rawDataBytes), nil
}

// CreateDocument não é necessário para vertc-assinaturas (quick-send já cria)
func (p *VertcAssinaturasProvider) CreateDocument(ctx context.Context, envelopeKey string, document *entity.EntityDocument, internalEnvelopeID int) (string, error) {
	return "", fmt.Errorf("CreateDocument is not supported for vertc-assinaturas provider. Use quick-send to create envelope with documents")
}

// CreateSigner não é necessário para vertc-assinaturas (quick-send já cria)
func (p *VertcAssinaturasProvider) CreateSigner(ctx context.Context, envelopeKey string, signerData provider.SignerData) (string, error) {
	return "", fmt.Errorf("CreateSigner is not supported for vertc-assinaturas provider. Use quick-send to create envelope with signers")
}

// CreateRequirement não é necessário para vertc-assinaturas (quick-send já cria)
func (p *VertcAssinaturasProvider) CreateRequirement(ctx context.Context, envelopeKey string, reqData provider.RequirementData) (string, error) {
	return "", fmt.Errorf("CreateRequirement is not supported for vertc-assinaturas provider. Requirements are configured via requiredMethods in quick-send")
}

// ActivateEnvelope não é necessário para vertc-assinaturas (quick-send já ativa)
func (p *VertcAssinaturasProvider) ActivateEnvelope(ctx context.Context, envelopeKey string) error {
	return fmt.Errorf("ActivateEnvelope is not supported for vertc-assinaturas provider. Quick-send automatically activates the envelope")
}

// NotifyEnvelope envia uma notificação para os signatários de um envelope
// TODO: Implementar quando a API vertc-assinaturas tiver endpoint de notificação
func (p *VertcAssinaturasProvider) NotifyEnvelope(ctx context.Context, envelopeKey string, message string) error {
	return fmt.Errorf("NotifyEnvelope is not yet implemented for vertc-assinaturas provider")
}



