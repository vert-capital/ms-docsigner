package dtos

import "time"

// EnvelopeCreateRequestDTO representa a estrutura de request para criação de envelope
type EnvelopeCreateRequestDTO struct {
	Name            string     `json:"name" binding:"required,min=3,max=255"`
	Description     string     `json:"description,omitempty" binding:"max=1000"`
	DocumentsIDs    []int      `json:"documents_ids" binding:"required,min=1"`
	SignatoryEmails []string   `json:"signatory_emails" binding:"required,min=1"`
	Message         string     `json:"message,omitempty" binding:"max=500"`
	DeadlineAt      *time.Time `json:"deadline_at,omitempty"`
	RemindInterval  int        `json:"remind_interval,omitempty" binding:"min=1,max=30"`
	AutoClose       bool       `json:"auto_close,omitempty"`
}

// EnvelopeUpdateRequestDTO representa a estrutura de request para atualização de envelope
type EnvelopeUpdateRequestDTO struct {
	Name            *string    `json:"name,omitempty" binding:"omitempty,min=3,max=255"`
	Description     *string    `json:"description,omitempty" binding:"omitempty,max=1000"`
	DocumentsIDs    *[]int     `json:"documents_ids,omitempty" binding:"omitempty,min=1"`
	SignatoryEmails *[]string  `json:"signatory_emails,omitempty" binding:"omitempty,min=1"`
	Message         *string    `json:"message,omitempty" binding:"omitempty,max=500"`
	DeadlineAt      *time.Time `json:"deadline_at,omitempty"`
	RemindInterval  *int       `json:"remind_interval,omitempty" binding:"omitempty,min=1,max=30"`
	AutoClose       *bool      `json:"auto_close,omitempty"`
}

// EnvelopeResponseDTO representa a estrutura de response para envelope
type EnvelopeResponseDTO struct {
	ID              int        `json:"id"`
	Name            string     `json:"name"`
	Description     string     `json:"description"`
	Status          string     `json:"status"`
	ClicksignKey    string     `json:"clicksign_key"`
	DocumentsIDs    []int      `json:"documents_ids"`
	SignatoryEmails []string   `json:"signatory_emails"`
	Message         string     `json:"message"`
	DeadlineAt      *time.Time `json:"deadline_at"`
	RemindInterval  int        `json:"remind_interval"`
	AutoClose       bool       `json:"auto_close"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// EnvelopeListResponseDTO representa a estrutura de response para lista de envelopes
type EnvelopeListResponseDTO struct {
	Envelopes []EnvelopeResponseDTO `json:"envelopes"`
	Total     int                   `json:"total"`
}

// EnvelopeActivateRequestDTO representa a estrutura de request para ativação de envelope
type EnvelopeActivateRequestDTO struct {
	// Pode ser vazio, pois a ativação é apenas mudança de status
}

// ErrorResponseDTO representa a estrutura de response para erros
type ErrorResponseDTO struct {
	Error   string                 `json:"error"`
	Message string                 `json:"message,omitempty"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// ValidationErrorResponseDTO representa a estrutura de response para erros de validação
type ValidationErrorResponseDTO struct {
	Error   string                  `json:"error"`
	Message string                  `json:"message"`
	Details []ValidationErrorDetail `json:"details"`
}

// ValidationErrorDetail representa um erro de validação específico
type ValidationErrorDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   string `json:"value,omitempty"`
}
