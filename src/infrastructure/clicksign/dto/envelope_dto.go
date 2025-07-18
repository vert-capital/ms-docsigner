package dto

import "time"

// EnvelopeCreateRequestWrapper representa a estrutura JSON API para criação de envelope na API do Clicksign
type EnvelopeCreateRequestWrapper struct {
	Data EnvelopeCreateData `json:"data"`
}

// EnvelopeCreateData representa a seção "data" da estrutura JSON API
type EnvelopeCreateData struct {
	Type       string                    `json:"type"`
	Attributes EnvelopeCreateAttributes  `json:"attributes"`
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
	Type       string                         `json:"type"`
	ID         string                         `json:"id"`
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

// ClicksignErrorResponse representa a estrutura de erro da API do Clicksign
type ClicksignErrorResponse struct {
	Error struct {
		Type       string                 `json:"type"`
		Message    string                 `json:"message"`
		Details    map[string]interface{} `json:"details,omitempty"`
		Code       string                 `json:"code,omitempty"`
		StatusCode int                    `json:"status_code,omitempty"`
	} `json:"error"`
}
