package dto

import "time"

// EnvelopeCreateRequestWrapper representa a estrutura JSON API para criação de envelope na API do Clicksign
type EnvelopeCreateRequestWrapper struct {
	Data EnvelopeCreateData `json:"data"`
}

// EnvelopeCreateData representa a seção "data" da estrutura JSON API
type EnvelopeCreateData struct {
	Type       string                   `json:"type"`
	Attributes EnvelopeCreateAttributes `json:"attributes"`
}

// EnvelopeCreateAttributes representa os atributos do envelope dentro da estrutura JSON API
type EnvelopeCreateAttributes struct {
	Name              string     `json:"name"`
	Locale            string     `json:"locale,omitempty"`
	AutoClose         bool       `json:"auto_close,omitempty"`
	RemindInterval    int        `json:"remind_interval,omitempty"`
	BlockAfterRefusal bool       `json:"block_after_refusal,omitempty"`
	DeadlineAt        *time.Time `json:"deadline_at,omitempty"`
	DefaultSubject    string     `json:"default_subject,omitempty"`
}

// EnvelopeCreateRequest representa a estrutura legada para criação de envelope (mantida para compatibilidade)
type EnvelopeCreateRequest struct {
	Name           string     `json:"name"`
	Locale         string     `json:"locale,omitempty"`
	AutoClose      bool       `json:"auto_close,omitempty"`
	RemindInterval int        `json:"remind_interval,omitempty"`
	DeadlineAt     *time.Time `json:"deadline_at,omitempty"`
	DefaultSubject string     `json:"default_subject,omitempty"`
}

// EnvelopeCreateResponseWrapper representa a estrutura JSON API para resposta de criação de envelope
type EnvelopeCreateResponseWrapper struct {
	Data EnvelopeCreateResponseData `json:"data"`
}

// EnvelopeCreateResponseData representa a seção "data" da resposta JSON API
type EnvelopeCreateResponseData struct {
	Type       string                           `json:"type"`
	ID         string                           `json:"id"`
	Attributes EnvelopeCreateResponseAttributes `json:"attributes"`
}

// EnvelopeCreateResponseAttributes representa os atributos do envelope na resposta JSON API
type EnvelopeCreateResponseAttributes struct {
	Name           string     `json:"name"`
	Status         string     `json:"status"`
	Locale         string     `json:"locale"`
	AutoClose      bool       `json:"auto_close"`
	RemindInterval int        `json:"remind_interval"`
	DeadlineAt     *time.Time `json:"deadline_at"`
	DefaultSubject string     `json:"default_subject"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// EnvelopeCreateResponse representa a resposta legada da API do Clicksign (mantida para compatibilidade)
type EnvelopeCreateResponse struct {
	ID             string     `json:"id"`
	Name           string     `json:"name"`
	Status         string     `json:"status"`
	Locale         string     `json:"locale"`
	AutoClose      bool       `json:"auto_close"`
	RemindInterval int        `json:"remind_interval"`
	DeadlineAt     *time.Time `json:"deadline_at"`
	DefaultSubject string     `json:"default_subject"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// EnvelopeUpdateRequest representa a estrutura para atualização de envelope na API do Clicksign
type EnvelopeUpdateRequest struct {
	Name           *string    `json:"name,omitempty"`
	Status         *string    `json:"status,omitempty"`
	AutoClose      *bool      `json:"auto_close,omitempty"`
	RemindInterval *int       `json:"remind_interval,omitempty"`
	DeadlineAt     *time.Time `json:"deadline_at,omitempty"`
	DefaultSubject *string    `json:"default_subject,omitempty"`
}

type EnvelopeUpdateRequestWrapper struct {
	Data EnvelopeUpdateRequestWrapperData `json:"data"`
}
type EnvelopeUpdateRequestWrapperAttributes struct {
	Status     string `json:"status"`
	DeadlineAt string `json:"deadline_at,omitempty"`
}
type EnvelopeUpdateRequestWrapperData struct {
	ID         string                                 `json:"id"`
	Type       string                                 `json:"type"`
	Attributes EnvelopeUpdateRequestWrapperAttributes `json:"attributes"`
}

// EnvelopeGetResponse representa a resposta da API do Clicksign para consulta de envelope
type EnvelopeGetResponse struct {
	ID             string     `json:"id"`
	Name           string     `json:"name"`
	Status         string     `json:"status"`
	Locale         string     `json:"locale"`
	AutoClose      bool       `json:"auto_close"`
	RemindInterval int        `json:"remind_interval"`
	DeadlineAt     *time.Time `json:"deadline_at"`
	DefaultSubject string     `json:"default_subject"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	DocumentsCount int        `json:"documents_count"`
	SignersCount   int        `json:"signers_count"`
}

// DocumentCreateRequestWrapper representa a estrutura JSON API para criação de documento
type DocumentCreateRequestWrapper struct {
	Data DocumentCreateData `json:"data"`
}

// DocumentCreateData representa a seção "data" da estrutura JSON API para criação de documento
type DocumentCreateData struct {
	Type       string                   `json:"type"`
	Attributes DocumentCreateAttributes `json:"attributes"`
}

// DocumentCreateAttributes representa os atributos do documento para criação conforme Postman Collection
type DocumentCreateAttributes struct {
	Filename      string            `json:"filename"`
	ContentBase64 string            `json:"content_base64,omitempty"`
	Template      *DocumentTemplate `json:"template,omitempty"`
	Metadata      *DocumentMetadata `json:"metadata,omitempty"`
}

// DocumentTemplate representa a estrutura para documentos criados via template
type DocumentTemplate struct {
	ID string `json:"id"`
}

// DocumentMetadata representa metadados do documento conforme especificação oficial
type DocumentMetadata struct {
	Type       string `json:"type"`
	ID         int    `json:"id"`
	User       int    `json:"user"`
	EnvelopeID int    `json:"envelope_id"`
}

// DocumentUploadRequestWrapper representa a estrutura JSON API para upload de documento (DEPRECATED)
type DocumentUploadRequestWrapper struct {
	Data DocumentUploadData `json:"data"`
}

// DocumentUploadData representa a seção "data" da estrutura JSON API para upload (DEPRECATED)
type DocumentUploadData struct {
	Type       string                   `json:"type"`
	Attributes DocumentUploadAttributes `json:"attributes"`
}

// DocumentUploadAttributes representa os atributos do documento para upload (DEPRECATED)
type DocumentUploadAttributes struct {
	Path          string `json:"path,omitempty"`
	ContentBase64 string `json:"content_base64,omitempty"`
	Filename      string `json:"filename,omitempty"`
}

// DocumentCreateResponseWrapper representa a resposta JSON API para criação de documento
type DocumentCreateResponseWrapper struct {
	Data DocumentCreateResponseData `json:"data"`
}

// DocumentCreateResponseData representa a seção "data" da resposta JSON API para criação
type DocumentCreateResponseData struct {
	Type       string                           `json:"type"`
	ID         string                           `json:"id"`
	Attributes DocumentCreateResponseAttributes `json:"attributes"`
}

// DocumentCreateResponseAttributes representa os atributos do documento na resposta de criação
type DocumentCreateResponseAttributes struct {
	Filename    string    `json:"filename"`
	ContentType string    `json:"content_type"`
	Filesize    int64     `json:"filesize"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// SignerCreateRequestWrapper representa a estrutura JSON API para criação de signatário
type SignerCreateRequestWrapper struct {
	Data SignerCreateData `json:"data"`
}

// SignerCreateData representa a seção "data" da estrutura JSON API para criação de signatário
type SignerCreateData struct {
	Type       string                 `json:"type"`
	Attributes SignerCreateAttributes `json:"attributes"`
}

// SignerCreateAttributes representa os atributos do signatário conforme Postman Collection
type SignerCreateAttributes struct {
	Name              string                   `json:"name"`
	Email             string                   `json:"email"`
	Birthday          string                   `json:"birthday,omitempty"`
	PhoneNumber       *string                  `json:"phone_number,omitempty"`
	HasDocumentation  bool                     `json:"has_documentation"`
	Refusable         bool                     `json:"refusable"`
	Group             int                      `json:"group"`
	CommunicateEvents *SignerCommunicateEvents `json:"communicate_events,omitempty"`
}

// SignerCommunicateEvents representa as configurações de comunicação do signatário
type SignerCommunicateEvents struct {
	DocumentSigned    string `json:"document_signed,omitempty"`
	SignatureRequest  string `json:"signature_request,omitempty"`
	SignatureReminder string `json:"signature_reminder,omitempty"`
}

// SignerCreateResponseWrapper representa a resposta JSON API para criação de signatário
type SignerCreateResponseWrapper struct {
	Data SignerCreateResponseData `json:"data"`
}

// SignerCreateResponseData representa a seção "data" da resposta JSON API para criação de signatário
type SignerCreateResponseData struct {
	Type       string                         `json:"type"`
	ID         string                         `json:"id"`
	Attributes SignerCreateResponseAttributes `json:"attributes"`
}

// SignerCreateResponseAttributes representa os atributos do signatário na resposta de criação
type SignerCreateResponseAttributes struct {
	Name              string                   `json:"name"`
	Email             string                   `json:"email"`
	Birthday          string                   `json:"birthday"`
	PhoneNumber       *string                  `json:"phone_number"`
	HasDocumentation  bool                     `json:"has_documentation"`
	Refusable         bool                     `json:"refusable"`
	Group             int                      `json:"group"`
	CommunicateEvents *SignerCommunicateEvents `json:"communicate_events"`
	CreatedAt         time.Time                `json:"created_at"`
	UpdatedAt         time.Time                `json:"updated_at"`
}

// DocumentUploadResponseWrapper representa a resposta JSON API para upload de documento (DEPRECATED)
type DocumentUploadResponseWrapper struct {
	Data DocumentUploadResponseData `json:"data"`
}

// DocumentUploadResponseData representa a seção "data" da resposta JSON API (DEPRECATED)
type DocumentUploadResponseData struct {
	Type       string                           `json:"type"`
	ID         string                           `json:"id"`
	Attributes DocumentUploadResponseAttributes `json:"attributes"`
}

// DocumentUploadResponseAttributes representa os atributos do documento na resposta (DEPRECATED)
type DocumentUploadResponseAttributes struct {
	Filename    string    `json:"filename"`
	ContentType string    `json:"content_type"`
	Filesize    int64     `json:"filesize"`
	Path        string    `json:"path"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// RequirementCreateRequestWrapper representa a estrutura JSON API para criação de requisito
type RequirementCreateRequestWrapper struct {
	Data RequirementCreateData `json:"data"`
}

// RequirementCreateData representa a seção "data" da estrutura JSON API para requisito
type RequirementCreateData struct {
	Type          string                      `json:"type"`
	Attributes    RequirementCreateAttributes `json:"attributes"`
	Relationships *RequirementRelationships   `json:"relationships,omitempty"`
}

// RequirementCreateAttributes representa os atributos do requisito conforme JSON API spec
type RequirementCreateAttributes struct {
	Action string `json:"action"`
	Role   string `json:"role,omitempty"`
	Auth   string `json:"auth,omitempty"`
}

// RequirementRelationships representa relacionamentos conforme JSON API spec
type RequirementRelationships struct {
	Document *RequirementRelationship `json:"document,omitempty"`
	Signer   *RequirementRelationship `json:"signer,omitempty"`
}

// RequirementRelationship representa um relacionamento individual
type RequirementRelationship struct {
	Data RequirementRelationshipData `json:"data"`
}

// RequirementRelationshipData representa os dados do relacionamento
type RequirementRelationshipData struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

// RequirementCreateResponseWrapper representa a resposta JSON API para criação de requisito
type RequirementCreateResponseWrapper struct {
	Data RequirementCreateResponseData `json:"data"`
}

// RequirementCreateResponseData representa a seção "data" da resposta JSON API para requisito
type RequirementCreateResponseData struct {
	Type          string                              `json:"type"`
	ID            string                              `json:"id"`
	Attributes    RequirementCreateResponseAttributes `json:"attributes"`
	Relationships *RequirementRelationships           `json:"relationships,omitempty"`
}

// RequirementCreateResponseAttributes representa os atributos do requisito na resposta
type RequirementCreateResponseAttributes struct {
	Action    string    `json:"action"`
	Role      string    `json:"role,omitempty"`
	Auth      string    `json:"auth,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BulkRequirementsRequestWrapper representa a estrutura para operações em massa
type BulkRequirementsRequestWrapper struct {
	AtomicOperations []AtomicOperation `json:"atomic:operations"`
}

// AtomicOperation representa uma operação atômica conforme JSON API spec
type AtomicOperation struct {
	Op   string                 `json:"op"`
	Ref  *AtomicOperationRef    `json:"ref,omitempty"`
	Data *RequirementCreateData `json:"data,omitempty"`
}

// AtomicOperationRef representa a referência para operações de remoção
type AtomicOperationRef struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

// BulkRequirementsResponseWrapper representa a resposta para operações em massa
type BulkRequirementsResponseWrapper struct {
	AtomicResults []AtomicResult `json:"atomic:results"`
}

// AtomicResult representa o resultado de uma operação atômica
type AtomicResult struct {
	Data *RequirementCreateResponseData `json:"data,omitempty"`
}

// ClicksignErrorResponse representa a estrutura de erro da API do Clicksign
type ClicksignErrorResponse struct {
	Error struct {
		Type       string         `json:"type"`
		Message    string         `json:"message"`
		Details    map[string]any `json:"details,omitempty"`
		Code       string         `json:"code,omitempty"`
		StatusCode int            `json:"status_code,omitempty"`
	} `json:"error"`
}
