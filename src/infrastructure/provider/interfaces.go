package provider

import (
	"context"

	"app/entity"
)

// EnvelopeProvider define a interface comum para todos os providers de envelope
// Esta interface abstrai as operações específicas de cada provider (Clicksign, vertc-assinaturas, etc.)
type EnvelopeProvider interface {
	// CreateEnvelope cria um envelope no provider e retorna a chave do provider e dados brutos
	CreateEnvelope(ctx context.Context, envelope *entity.EntityEnvelope) (providerKey string, rawData string, err error)

	// CreateDocument cria um documento dentro de um envelope no provider
	CreateDocument(ctx context.Context, envelopeKey string, document *entity.EntityDocument, internalEnvelopeID int) (documentKey string, err error)

	// CreateSigner cria um signatário no envelope do provider
	// signerData contém os dados necessários para criar o signatário
	CreateSigner(ctx context.Context, envelopeKey string, signerData SignerData) (signerKey string, err error)

	// CreateRequirement cria um requisito de assinatura no envelope do provider
	// reqData contém os dados necessários para criar o requisito
	CreateRequirement(ctx context.Context, envelopeKey string, reqData RequirementData) (requirementKey string, err error)

	// ActivateEnvelope ativa um envelope no provider para iniciar o processo de assinatura
	ActivateEnvelope(ctx context.Context, envelopeKey string) error

	// NotifyEnvelope envia uma notificação para os signatários de um envelope
	NotifyEnvelope(ctx context.Context, envelopeKey string, message string) error
}




