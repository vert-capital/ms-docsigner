package dto

import "time"

// EnvelopeCreateRequest representa a estrutura para criação de envelope na API do Clicksign
type EnvelopeCreateRequest struct {
	Name           string     `json:"name"`
	Locale         string     `json:"locale,omitempty"`
	AutoClose      bool       `json:"auto_close,omitempty"`
	RemindInterval int        `json:"remind_interval,omitempty"`
	DeadlineAt     *time.Time `json:"deadline_at,omitempty"`
	DefaultSubject string     `json:"default_subject,omitempty"`
}

// EnvelopeCreateResponse representa a resposta da API do Clicksign para criação de envelope
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
