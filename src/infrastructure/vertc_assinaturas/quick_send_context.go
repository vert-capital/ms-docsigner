package vertc_assinaturas

import (
	"context"

	"app/entity"
	"app/infrastructure/provider"
)

// ContextKey é a chave usada para armazenar QuickSendData no contexto
type contextKey string

const (
	// QuickSendDataKey é a chave para acessar QuickSendData no contexto
	quickSendDataKey contextKey = "quick_send_data"
)

// GetQuickSendDataFromContext extrai QuickSendData do contexto
func GetQuickSendDataFromContext(ctx context.Context) (*QuickSendData, bool) {
	data, ok := ctx.Value(quickSendDataKey).(*QuickSendData)
	return data, ok
}

// WithQuickSendData adiciona QuickSendData ao contexto
func WithQuickSendData(ctx context.Context, data *QuickSendData) context.Context {
	return context.WithValue(ctx, quickSendDataKey, data)
}

// QuickSendData contém todos os dados necessários para quick-send
type QuickSendData struct {
	Envelope  *entity.EntityEnvelope
	Documents []*entity.EntityDocument
	Signers   []provider.SignerData
}

