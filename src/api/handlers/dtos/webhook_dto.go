package dtos

import (
	"time"
)

// WebhookRequestDTO representa o payload recebido do Clicksign
type WebhookRequestDTO struct {
	Event    WebhookEventDTO    `json:"event" binding:"required"`
	Document WebhookDocumentDTO `json:"document" binding:"required"`
}

// WebhookEventDTO representa o evento do webhook
type WebhookEventDTO struct {
	Name       string                 `json:"name" binding:"required"`
	Data       map[string]interface{} `json:"data"`
	OccurredAt string                 `json:"occurred_at" binding:"required"`
}

// WebhookDocumentDTO representa o documento do webhook
type WebhookDocumentDTO struct {
	Key               string                 `json:"key" binding:"required"`
	AccountKey        string                 `json:"account_key" binding:"required"`
	Path              string                 `json:"path"`
	Filename          string                 `json:"filename"`
	UploadedAt        string                 `json:"uploaded_at"`
	UpdatedAt         string                 `json:"updated_at"`
	FinishedAt        string                 `json:"finished_at"`
	DeadlineAt        string                 `json:"deadline_at"`
	Status            string                 `json:"status" binding:"required"`
	AutoClose         bool                   `json:"auto_close"`
	Locale            string                 `json:"locale"`
	Metadata          map[string]interface{} `json:"metadata"`
	SequenceEnabled   bool                   `json:"sequence_enabled"`
	SignableGroup     *string                `json:"signable_group"`
	RemindInterval    int                    `json:"remind_interval"`
	BlockAfterRefusal bool                   `json:"block_after_refusal"`
	Preview           bool                   `json:"preview"`
	Downloads         WebhookDownloadsDTO    `json:"downloads"`
	Template          *string                `json:"template"`
	Signers           []WebhookSignerDTO     `json:"signers"`
	Events            []WebhookEventDTO      `json:"events"`
	Attachments       []interface{}          `json:"attachments"`
	Links             WebhookLinksDTO        `json:"links"`
}

// WebhookDownloadsDTO representa os downloads do documento
type WebhookDownloadsDTO struct {
	OriginalFileURL string `json:"original_file_url"`
}

// WebhookSignerDTO representa um signatário do webhook
type WebhookSignerDTO struct {
	SignAs                  string   `json:"sign_as"`
	ListKey                 string   `json:"list_key"`
	Key                     string   `json:"key"`
	Email                   string   `json:"email"`
	Name                    string   `json:"name"`
	Birthday                *string  `json:"birthday"`
	CreatedAt               string   `json:"created_at"`
	Documentation           *string  `json:"documentation"`
	HasDocumentation        bool     `json:"has_documentation"`
	Auths                   []string `json:"auths"`
	SelfieEnabled           bool     `json:"selfie_enabled"`
	HandwrittenEnabled      bool     `json:"handwritten_enabled"`
	OfficialDocumentEnabled bool     `json:"official_document_enabled"`
	LivenessEnabled         bool     `json:"liveness_enabled"`
	FacialBiometricsEnabled bool     `json:"facial_biometrics_enabled"`
	CommunicateBy           string   `json:"communicate_by"`
	PhoneNumber             *string  `json:"phone_number"`
	PhoneNumberHash         *string  `json:"phone_number_hash"`
	FederalDataValidation   *string  `json:"federal_data_validation"`
	URL                     string   `json:"url"`
	Address                 string   `json:"address"`
	Longitude               *string  `json:"longitude"`
	Latitude                *string  `json:"latitude"`
}

// WebhookLinksDTO representa os links do documento
type WebhookLinksDTO struct {
	Self string `json:"self"`
}

// WebhookResponseDTO representa a resposta do webhook
type WebhookResponseDTO struct {
	ID          int        `json:"id"`
	EventName   string     `json:"event_name"`
	DocumentKey string     `json:"document_key"`
	AccountKey  string     `json:"account_key"`
	Status      string     `json:"status"`
	ProcessedAt *time.Time `json:"processed_at,omitempty"`
	Error       *string    `json:"error,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// WebhookListResponseDTO representa a lista de webhooks
type WebhookListResponseDTO struct {
	Webhooks []WebhookResponseDTO `json:"webhooks"`
	Total    int                  `json:"total"`
}

// WebhookFiltersDTO representa os filtros para listagem de webhooks
type WebhookFiltersDTO struct {
	EventName   string `json:"event_name,omitempty"`
	DocumentKey string `json:"document_key,omitempty"`
	AccountKey  string `json:"account_key,omitempty"`
	Status      string `json:"status,omitempty"`
	Page        int    `json:"page,omitempty"`
	Limit       int    `json:"limit,omitempty"`
}

// WebhookProcessRequestDTO representa a requisição para processar um webhook
type WebhookProcessRequestDTO struct {
	WebhookID int `json:"webhook_id" binding:"required"`
}

// WebhookProcessResponseDTO representa a resposta do processamento
type WebhookProcessResponseDTO struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// IsAutoCloseEvent verifica se é um evento de fechamento automático
func (w *WebhookRequestDTO) IsAutoCloseEvent() bool {
	return w.Event.Name == "auto_close"
}

// IsSignEvent verifica se é um evento de assinatura
func (w *WebhookRequestDTO) IsSignEvent() bool {
	return w.Event.Name == "sign"
}

// IsSignatureStartedEvent verifica se é um evento de início de assinatura
func (w *WebhookRequestDTO) IsSignatureStartedEvent() bool {
	return w.Event.Name == "signature_started"
}

// IsAddSignerEvent verifica se é um evento de adição de signatário
func (w *WebhookRequestDTO) IsAddSignerEvent() bool {
	return w.Event.Name == "add_signer"
}

// IsUploadEvent verifica se é um evento de upload
func (w *WebhookRequestDTO) IsUploadEvent() bool {
	return w.Event.Name == "upload"
}

// GetDocumentStatus retorna o status do documento
func (w *WebhookRequestDTO) GetDocumentStatus() string {
	return w.Document.Status
}

// IsDocumentClosed verifica se o documento está fechado
func (w *WebhookRequestDTO) IsDocumentClosed() bool {
	return w.Document.Status == "closed"
}

// IsDocumentFinished verifica se o documento está finalizado
func (w *WebhookRequestDTO) IsDocumentFinished() bool {
	return w.Document.Status == "finished"
}

// IsDocumentCancelled verifica se o documento foi cancelado
func (w *WebhookRequestDTO) IsDocumentCancelled() bool {
	return w.Document.Status == "cancelled"
}
