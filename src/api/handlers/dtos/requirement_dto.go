package dtos

import (
	"fmt"
	"time"

	"app/entity"
)

// RequirementCreateRequestDTO representa a estrutura de request para criação de requirement
type RequirementCreateRequestDTO struct {
	Action     string  `json:"action" binding:"required,oneof=agree sign provide_evidence"`
	Role       string  `json:"role,omitempty" binding:"omitempty,oneof=sign"`
	Auth       *string `json:"auth,omitempty" binding:"omitempty,oneof=email icp_brasil auto_signature"`
	DocumentID *string `json:"document_id,omitempty"`
	SignerID   *string `json:"signer_id,omitempty"`
}

// RequirementResponseDTO representa a estrutura de response para requirement
type RequirementResponseDTO struct {
	ID           int       `json:"id"`
	EnvelopeID   int       `json:"envelope_id"`
	ClicksignKey string    `json:"clicksign_key,omitempty"`
	Action       string    `json:"action"`
	Role         string    `json:"role"`
	Auth         *string   `json:"auth,omitempty"`
	DocumentID   *string   `json:"document_id,omitempty"`
	SignerID     *string   `json:"signer_id,omitempty"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// RequirementListResponseDTO representa a lista de requirements
type RequirementListResponseDTO struct {
	Requirements []RequirementResponseDTO `json:"requirements"`
	Total        int                      `json:"total"`
}

// RequirementUpdateRequestDTO representa a estrutura de request para atualização de requirement
type RequirementUpdateRequestDTO struct {
	Status string `json:"status,omitempty" binding:"omitempty,oneof=pending completed"`
}

// Validate valida o DTO de criação de requirement
func (dto *RequirementCreateRequestDTO) Validate() error {
	// Para action "provide_evidence", auth é obrigatório
	if dto.Action == "provide_evidence" && (dto.Auth == nil || *dto.Auth == "") {
		return fmt.Errorf("auth é obrigatório para action 'provide_evidence'")
	}

	return nil
}

// ToEntity converte RequirementCreateRequestDTO para EntityRequirement
func (dto *RequirementCreateRequestDTO) ToEntity(envelopeID int) *entity.EntityRequirement {
	now := time.Now()

	role := "sign" // valor padrão
	if dto.Role != "" {
		role = dto.Role
	}

	return &entity.EntityRequirement{
		EnvelopeID: envelopeID,
		Action:     dto.Action,
		Role:       role,
		Auth:       dto.Auth,
		DocumentID: dto.DocumentID,
		SignerID:   dto.SignerID,
		Status:     "pending", // status padrão
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// FromEntity converte EntityRequirement para RequirementResponseDTO
func (dto *RequirementResponseDTO) FromEntity(requirement *entity.EntityRequirement) *RequirementResponseDTO {
	return &RequirementResponseDTO{
		ID:           requirement.ID,
		EnvelopeID:   requirement.EnvelopeID,
		ClicksignKey: requirement.ClicksignKey,
		Action:       requirement.Action,
		Role:         requirement.Role,
		Auth:         requirement.Auth,
		DocumentID:   requirement.DocumentID,
		SignerID:     requirement.SignerID,
		Status:       requirement.Status,
		CreatedAt:    requirement.CreatedAt,
		UpdatedAt:    requirement.UpdatedAt,
	}
}

// FromEntityList converte lista de EntityRequirement para RequirementListResponseDTO
func (dto *RequirementListResponseDTO) FromEntityList(requirements []entity.EntityRequirement) *RequirementListResponseDTO {
	data := make([]RequirementResponseDTO, len(requirements))
	for i, requirement := range requirements {
		responseDTO := RequirementResponseDTO{}
		data[i] = *responseDTO.FromEntity(&requirement)
	}

	return &RequirementListResponseDTO{
		Requirements: data,
		Total:        len(requirements),
	}
}
