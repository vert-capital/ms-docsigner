package dtos

import (
	"fmt"
	"time"
)

// EnvelopeV2CreateRequestDTO representa a estrutura de request para criação de envelope na v2
// Esta versão inclui o campo Provider obrigatório para seleção do provider
type EnvelopeV2CreateRequestDTO struct {
	Provider        string                       `json:"provider" binding:"required,oneof=clicksign vertc-assinaturas"`
	Name            string                       `json:"name" binding:"required,min=3,max=255"`
	Description     string                       `json:"description,omitempty" binding:"max=1000"`
	DocumentsIDs    []int                        `json:"documents_ids,omitempty"`
	Documents       []EnvelopeDocumentRequest    `json:"documents,omitempty"`
	SignatoryEmails []string                     `json:"signatory_emails,omitempty"`
	Signatories     []EnvelopeSignatoryRequest   `json:"signatories,omitempty"`
	Requirements    []EnvelopeRequirementRequest `json:"requirements,omitempty"`
	Qualifiers      []EnvelopeRequirementRequest `json:"qualifiers,omitempty"` // Qualificadores para o envelope, como "sign", "agree", etc.
	Message         string                       `json:"message,omitempty" binding:"max=500"`
	DeadlineAt      *time.Time                   `json:"deadline_at,omitempty"`
	RemindInterval  int                          `json:"remind_interval,omitempty" binding:"omitempty,min=1,max=30"`
	AutoClose       bool                         `json:"auto_close,omitempty"`
	Approved        bool                         `json:"approved,omitempty"` // Indica se o envelope foi aprovado
}

// Validate valida o DTO de criação de envelope v2
// Reutiliza a mesma lógica de validação do DTO v1, mas com provider obrigatório
func (dto *EnvelopeV2CreateRequestDTO) Validate() error {
	// Validar provider
	if dto.Provider == "" {
		return fmt.Errorf("provider é obrigatório")
	}

	if dto.Provider != "clicksign" && dto.Provider != "vertc-assinaturas" {
		return fmt.Errorf("provider inválido: %s. Providers suportados: clicksign, vertc-assinaturas", dto.Provider)
	}

	// Deve ter pelo menos um tipo de documento (IDs ou base64)
	if len(dto.DocumentsIDs) == 0 && len(dto.Documents) == 0 {
		return fmt.Errorf("deve fornecer pelo menos um documento (documents_ids ou documents)")
	}

	// Não pode ter ambos ao mesmo tempo
	if len(dto.DocumentsIDs) > 0 && len(dto.Documents) > 0 {
		return fmt.Errorf("não é possível fornecer documents_ids e documents ao mesmo tempo")
	}

	// Deve ter pelo menos um tipo de signatário (emails ou signatários estruturados)
	if len(dto.SignatoryEmails) == 0 && len(dto.Signatories) == 0 {
		return fmt.Errorf("deve fornecer pelo menos um signatário (signatory_emails ou signatories)")
	}

	// Não pode ter ambos ao mesmo tempo
	if len(dto.SignatoryEmails) > 0 && len(dto.Signatories) > 0 {
		return fmt.Errorf("não é possível fornecer signatory_emails e signatories ao mesmo tempo")
	}

	// Validar signatários se fornecidos
	if len(dto.Signatories) > 0 {
		emailsMap := make(map[string]int) // valor é o índice do primeiro signatário com este email
		for i, signatory := range dto.Signatories {
			// Verificar emails únicos primeiro (mais eficiente)
			if firstIndex, exists := emailsMap[signatory.Email]; exists {
				return fmt.Errorf("email duplicado encontrado nos signatários: %s (posições %d e %d)",
					signatory.Email, firstIndex+1, i+1)
			}
			emailsMap[signatory.Email] = i

			// Reutilizar validação da estrutura SignatoryCreateRequestDTO
			tempSignatory := &SignatoryCreateRequestDTO{
				Name:              signatory.Name,
				Email:             signatory.Email,
				EnvelopeID:        1, // Valor temporário para validação
				Birthday:          signatory.Birthday,
				PhoneNumber:       signatory.PhoneNumber,
				HasDocumentation:  signatory.HasDocumentation,
				Refusable:         signatory.Refusable,
				Group:             signatory.Group,
				CommunicateEvents: signatory.CommunicateEvents,
			}

			if err := tempSignatory.Validate(); err != nil {
				return fmt.Errorf("erro na validação do signatário %d (%s): %v", i+1, signatory.Email, err)
			}
		}
	}

	// Validar requirements se fornecidos
	if len(dto.Requirements) > 0 {
		for i, requirement := range dto.Requirements {
			if err := requirement.Validate(); err != nil {
				return fmt.Errorf("erro na validação do requirement %d: %v", i+1, err)
			}
		}
	}

	return nil
}



