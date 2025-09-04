package dtos

import (
	"time"
)

// AutoSignatureTermCreateRequestDTO representa a estrutura de request para criação de termo de assinatura automática
type AutoSignatureTermCreateRequestDTO struct {
	Signer     SignerInfoDTO `json:"signer" binding:"required"`
	AdminEmail string        `json:"admin_email" binding:"required,email"`
	APIEmail   string        `json:"api_email" binding:"required,email"`
}

// SignerInfoDTO representa as informações do signatário no DTO
type SignerInfoDTO struct {
	Documentation string `json:"documentation" binding:"required,min=11,max=14"`
	Birthday      string `json:"birthday" binding:"required"`
	Email         string `json:"email" binding:"required,email"`
	Name          string `json:"name" binding:"required,min=2,max=255"`
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
