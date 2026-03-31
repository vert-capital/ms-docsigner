package entity

// EnvelopeCheckEventsResult representa o retorno da verificação manual de eventos.
type EnvelopeCheckEventsResult struct {
	Events         []EnvelopeSignatureEventData `json:"events"`
	ProcessedCount int                          `json:"processed_count"`
	EnvelopeKey    string                       `json:"envelope_key"`
}

// EnvelopeSignatureEventData representa um evento de assinatura identificado pela API do provider.
type EnvelopeSignatureEventData struct {
	SignerKey string      `json:"signer_key"`
	Email     string      `json:"email"`
	Name      string      `json:"name"`
	SignedAt  interface{} `json:"signed_at"`
}
