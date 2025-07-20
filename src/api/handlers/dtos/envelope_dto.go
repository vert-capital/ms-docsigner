package dtos

import (
	"fmt"
	"time"
)

// EnvelopeCreateRequestDTO representa a estrutura de request para criação de envelope
type EnvelopeCreateRequestDTO struct {
	Name            string                          `json:"name" binding:"required,min=3,max=255"`
	Description     string                          `json:"description,omitempty" binding:"max=1000"`
	DocumentsIDs    []int                           `json:"documents_ids,omitempty"`
	Documents       []EnvelopeDocumentRequest       `json:"documents,omitempty"`
	SignatoryEmails []string                        `json:"signatory_emails,omitempty"`
	Signatories     []EnvelopeSignatoryRequest      `json:"signatories,omitempty"`
	Requirements    []EnvelopeRequirementRequest    `json:"requirements,omitempty"`
	Message         string                          `json:"message,omitempty" binding:"max=500"`
	DeadlineAt      *time.Time                      `json:"deadline_at,omitempty"`
	RemindInterval  int                             `json:"remind_interval,omitempty" binding:"omitempty,min=1,max=30"`
	AutoClose       bool                            `json:"auto_close,omitempty"`
}

// EnvelopeDocumentRequest representa um documento a ser criado junto com o envelope
type EnvelopeDocumentRequest struct {
	Name               string `json:"name" binding:"required,min=3,max=255"`
	FileContentBase64  string `json:"file_content_base64" binding:"required"`
	Description        string `json:"description,omitempty"`
}

// EnvelopeSignatoryRequest representa um signatário a ser criado junto com o envelope
type EnvelopeSignatoryRequest struct {
	Name              string                        `json:"name" binding:"required,min=2,max=255"`
	Email             string                        `json:"email" binding:"required,email"`
	Birthday          *string                       `json:"birthday,omitempty"`
	PhoneNumber       *string                       `json:"phone_number,omitempty"`
	HasDocumentation  *bool                         `json:"has_documentation,omitempty"`
	Refusable         *bool                         `json:"refusable,omitempty"`
	Group             *int                          `json:"group,omitempty"`
	CommunicateEvents *SignatoryCommunicateEventsDTO `json:"communicate_events,omitempty"`
}

// EnvelopeRequirementRequest representa um requirement a ser criado junto com o envelope
type EnvelopeRequirementRequest struct {
	Action     string  `json:"action" binding:"required,oneof=agree sign provide_evidence"`
	Role       string  `json:"role,omitempty" binding:"omitempty,oneof=sign"`
	Auth       *string `json:"auth,omitempty" binding:"omitempty,oneof=email icp_brasil"`
	DocumentID *string `json:"document_id,omitempty"`
	SignerID   *string `json:"signer_id,omitempty"`
}

// ToSignatoryCreateRequestDTO converte EnvelopeSignatoryRequest para SignatoryCreateRequestDTO
func (esr *EnvelopeSignatoryRequest) ToSignatoryCreateRequestDTO(envelopeID int) SignatoryCreateRequestDTO {
	return SignatoryCreateRequestDTO{
		Name:              esr.Name,
		Email:             esr.Email,
		EnvelopeID:        envelopeID,
		Birthday:          esr.Birthday,
		PhoneNumber:       esr.PhoneNumber,
		HasDocumentation:  esr.HasDocumentation,
		Refusable:         esr.Refusable,
		Group:             esr.Group,
		CommunicateEvents: esr.CommunicateEvents,
	}
}

// ToRequirementCreateRequestDTO converte EnvelopeRequirementRequest para RequirementCreateRequestDTO
func (err *EnvelopeRequirementRequest) ToRequirementCreateRequestDTO() RequirementCreateRequestDTO {
	return RequirementCreateRequestDTO{
		Action:     err.Action,
		Role:       err.Role,
		Auth:       err.Auth,
		DocumentID: err.DocumentID,
		SignerID:   err.SignerID,
	}
}

// Validate valida o requirement
func (err *EnvelopeRequirementRequest) Validate() error {
	// Para action "provide_evidence", auth é obrigatório
	if err.Action == "provide_evidence" && (err.Auth == nil || *err.Auth == "") {
		return fmt.Errorf("auth é obrigatório para action 'provide_evidence'")
	}

	return nil
}

// Validate valida o DTO de criação de envelope
func (dto *EnvelopeCreateRequestDTO) Validate() error {
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
// O campo clicksign_raw_data contém os dados brutos retornados pela API do Clicksign
type EnvelopeResponseDTO struct {
	ID               int                     `json:"id"`
	Name             string                  `json:"name"`
	Description      string                  `json:"description"`
	Status           string                  `json:"status"`
	ClicksignKey     string                  `json:"clicksign_key"`
	ClicksignRawData *string                 `json:"clicksign_raw_data,omitempty"`
	DocumentsIDs     []int                   `json:"documents_ids"`
	SignatoryEmails  []string                `json:"signatory_emails"`
	Signatories      []SignatoryResponseDTO  `json:"signatories,omitempty"`
	Message          string                  `json:"message"`
	DeadlineAt       *time.Time              `json:"deadline_at"`
	RemindInterval   int                     `json:"remind_interval"`
	AutoClose        bool                    `json:"auto_close"`
	CreatedAt        time.Time               `json:"created_at"`
	UpdatedAt        time.Time               `json:"updated_at"`
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
