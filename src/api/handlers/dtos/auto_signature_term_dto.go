package dtos

import (
	"time"
)

// AutoSignatureTermCreateRequestDTO representa a estrutura de request para criação de termo de assinatura automática
type AutoSignatureTermCreateRequestDTO struct {
	Signer     SignerInfoDTO `json:"signer" binding:"required"`
	AdminEmail string        `json:"admin_email" binding:"omitempty,email"`
	APIEmail   string        `json:"api_email" binding:"omitempty,email"`
}

// SignerInfoDTO representa as informações do signatário no DTO
type SignerInfoDTO struct {
	Documentation string `json:"documentation" binding:"omitempty,min=11,max=14"`
	Birthday      string `json:"birthday" binding:"omitempty"`
	Email         string `json:"email" binding:"required,email"`
	Name          string `json:"name" binding:"omitempty,min=2,max=255"`
}

// AutoSignatureTermResponseDTO representa a estrutura de response para termo de assinatura automática
type AutoSignatureTermResponseDTO struct {
	ID               int           `json:"id"`
	Signer           SignerInfoDTO `json:"signer"`
	AdminEmail       string        `json:"admin_email"`
	APIEmail         string        `json:"api_email"`
	ClicksignKey     string        `json:"clicksign_key,omitempty"`
	ClicksignRawData *string       `json:"clicksign_raw_data,omitempty"`
	CreatedAt        time.Time     `json:"created_at"`
	UpdatedAt        time.Time     `json:"updated_at"`
}

// AutoSignatureTermListResponseDTO representa a estrutura de response para lista de termos
type AutoSignatureTermListResponseDTO struct {
	Terms []AutoSignatureTermResponseDTO `json:"terms"`
	Total int                            `json:"total"`
}

// AutoSignatureTermStatusQueryDTO representa os query params para consulta de status por provider
type AutoSignatureTermStatusQueryDTO struct {
	Provider string `form:"provider" binding:"required"`
	Email    string `form:"email" binding:"required,email"`
}

// AutoSignatureTermStatusResponseDTO representa a resposta da consulta de status por provider
type AutoSignatureTermStatusResponseDTO struct {
	Provider        string `json:"provider"`
	Email           string `json:"email"`
	HasSignedTerm   bool   `json:"has_signed_term"`
	PermissionFound bool   `json:"permission_found"`
	PermissionID    string `json:"permission_id,omitempty"`
	ContractStatus  string `json:"contract_status,omitempty"`
	IsActive        *bool  `json:"is_active,omitempty"`
}

// AutoSignatureTermProviderCreateResponseDTO representa a resposta de criação do termo por provider
type AutoSignatureTermProviderCreateResponseDTO struct {
	Provider          string `json:"provider"`
	Email             string `json:"email"`
	PermissionID      string `json:"permission_id,omitempty"`
	EnvelopeID        string `json:"envelope_id,omitempty"`
	ContractStatus    string `json:"contract_status,omitempty"`
	IsActive          *bool  `json:"is_active,omitempty"`
	NotificationSent  bool   `json:"notification_sent"`
	NotificationError string `json:"notification_error,omitempty"`
	UserCreated       bool   `json:"user_created"`
	UserExisted       bool   `json:"user_existed"`
}
