package dtos

import (
	"app/entity"
	"fmt"
	"net/mail"
	"regexp"
	"time"
)

// SignatoryCreateRequestDTO representa a estrutura de request para criação de signatário
type SignatoryCreateRequestDTO struct {
	Name              string                        `json:"name" binding:"required,min=2,max=255"`
	Email             string                        `json:"email" binding:"required,email"`
	EnvelopeID        int                           `json:"envelope_id" binding:"required"`
	Birthday          *string                       `json:"birthday,omitempty"`
	PhoneNumber       *string                       `json:"phone_number,omitempty"`
	HasDocumentation  *bool                         `json:"has_documentation,omitempty"`
	Refusable         *bool                         `json:"refusable,omitempty"`
	Group             *int                          `json:"group,omitempty"`
	CommunicateEvents *SignatoryCommunicateEventsDTO `json:"communicate_events,omitempty"`
}

// SignatoryCommunicateEventsDTO representa as configurações de comunicação do signatário
type SignatoryCommunicateEventsDTO struct {
	DocumentSigned    string `json:"document_signed"`
	SignatureRequest  string `json:"signature_request"`
	SignatureReminder string `json:"signature_reminder"`
}

// Validate valida o DTO de criação de signatário
func (dto *SignatoryCreateRequestDTO) Validate() error {
	// Validar email
	if _, err := mail.ParseAddress(dto.Email); err != nil {
		return fmt.Errorf("invalid email format: %s", dto.Email)
	}

	// Validar birthday se fornecido
	if dto.Birthday != nil {
		birthdayRegex := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
		if !birthdayRegex.MatchString(*dto.Birthday) {
			return fmt.Errorf("birthday must be in YYYY-MM-DD format, got: %s", *dto.Birthday)
		}

		_, err := time.Parse("2006-01-02", *dto.Birthday)
		if err != nil {
			return fmt.Errorf("invalid birthday date: %s", *dto.Birthday)
		}
	}

	// Validar phone number se fornecido
	if dto.PhoneNumber != nil {
		phoneRegex := regexp.MustCompile(`^\+\d{8,15}$`)
		if !phoneRegex.MatchString(*dto.PhoneNumber) {
			return fmt.Errorf("phone number must be in international format (+xxxxxxxx), got: %s", *dto.PhoneNumber)
		}
	}

	// Validar group se fornecido
	if dto.Group != nil && *dto.Group <= 0 {
		return fmt.Errorf("group must be a positive integer, got: %d", *dto.Group)
	}

	// Validar communicate events se fornecido
	if dto.CommunicateEvents != nil {
		validEvents := []string{"email", "sms", "none"}
		
		if !isValidEventType(dto.CommunicateEvents.DocumentSigned, validEvents) {
			return fmt.Errorf("invalid document_signed event type: %s", dto.CommunicateEvents.DocumentSigned)
		}

		if !isValidEventType(dto.CommunicateEvents.SignatureRequest, validEvents) {
			return fmt.Errorf("invalid signature_request event type: %s", dto.CommunicateEvents.SignatureRequest)
		}

		if !isValidEventType(dto.CommunicateEvents.SignatureReminder, validEvents) {
			return fmt.Errorf("invalid signature_reminder event type: %s", dto.CommunicateEvents.SignatureReminder)
		}
	}

	return nil
}

func isValidEventType(eventType string, validTypes []string) bool {
	for _, validType := range validTypes {
		if eventType == validType {
			return true
		}
	}
	return false
}

// ToEntity converte o DTO para a entidade Signatory
func (dto *SignatoryCreateRequestDTO) ToEntity() entity.EntitySignatory {
	var communicateEvents *entity.CommunicateEvents
	if dto.CommunicateEvents != nil {
		communicateEvents = &entity.CommunicateEvents{
			DocumentSigned:    dto.CommunicateEvents.DocumentSigned,
			SignatureRequest:  dto.CommunicateEvents.SignatureRequest,
			SignatureReminder: dto.CommunicateEvents.SignatureReminder,
		}
	}

	return entity.EntitySignatory{
		Name:              dto.Name,
		Email:             dto.Email,
		EnvelopeID:        dto.EnvelopeID,
		Birthday:          dto.Birthday,
		PhoneNumber:       dto.PhoneNumber,
		HasDocumentation:  dto.HasDocumentation,
		Refusable:         dto.Refusable,
		Group:             dto.Group,
		CommunicateEvents: communicateEvents,
	}
}

// SignatoryUpdateRequestDTO representa a estrutura de request para atualização de signatário
type SignatoryUpdateRequestDTO struct {
	Name              *string                        `json:"name,omitempty" binding:"omitempty,min=2,max=255"`
	Email             *string                        `json:"email,omitempty" binding:"omitempty,email"`
	EnvelopeID        *int                           `json:"envelope_id,omitempty"`
	Birthday          *string                        `json:"birthday,omitempty"`
	PhoneNumber       *string                        `json:"phone_number,omitempty"`
	HasDocumentation  *bool                          `json:"has_documentation,omitempty"`
	Refusable         *bool                          `json:"refusable,omitempty"`
	Group             *int                           `json:"group,omitempty"`
	CommunicateEvents *SignatoryCommunicateEventsDTO `json:"communicate_events,omitempty"`
}

// Validate valida o DTO de atualização de signatário
func (dto *SignatoryUpdateRequestDTO) Validate() error {
	// Validar email se fornecido
	if dto.Email != nil {
		if _, err := mail.ParseAddress(*dto.Email); err != nil {
			return fmt.Errorf("invalid email format: %s", *dto.Email)
		}
	}

	// Validar birthday se fornecido
	if dto.Birthday != nil {
		birthdayRegex := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
		if !birthdayRegex.MatchString(*dto.Birthday) {
			return fmt.Errorf("birthday must be in YYYY-MM-DD format, got: %s", *dto.Birthday)
		}

		_, err := time.Parse("2006-01-02", *dto.Birthday)
		if err != nil {
			return fmt.Errorf("invalid birthday date: %s", *dto.Birthday)
		}
	}

	// Validar phone number se fornecido
	if dto.PhoneNumber != nil {
		phoneRegex := regexp.MustCompile(`^\+\d{8,15}$`)
		if !phoneRegex.MatchString(*dto.PhoneNumber) {
			return fmt.Errorf("phone number must be in international format (+xxxxxxxx), got: %s", *dto.PhoneNumber)
		}
	}

	// Validar group se fornecido
	if dto.Group != nil && *dto.Group <= 0 {
		return fmt.Errorf("group must be a positive integer, got: %d", *dto.Group)
	}

	// Validar communicate events se fornecido
	if dto.CommunicateEvents != nil {
		validEvents := []string{"email", "sms", "none"}
		
		if !isValidEventType(dto.CommunicateEvents.DocumentSigned, validEvents) {
			return fmt.Errorf("invalid document_signed event type: %s", dto.CommunicateEvents.DocumentSigned)
		}

		if !isValidEventType(dto.CommunicateEvents.SignatureRequest, validEvents) {
			return fmt.Errorf("invalid signature_request event type: %s", dto.CommunicateEvents.SignatureRequest)
		}

		if !isValidEventType(dto.CommunicateEvents.SignatureReminder, validEvents) {
			return fmt.Errorf("invalid signature_reminder event type: %s", dto.CommunicateEvents.SignatureReminder)
		}
	}

	return nil
}

// ApplyToEntity aplica as mudanças do DTO à entidade existente
func (dto *SignatoryUpdateRequestDTO) ApplyToEntity(signatory *entity.EntitySignatory) {
	if dto.Name != nil {
		signatory.Name = *dto.Name
	}
	if dto.Email != nil {
		signatory.Email = *dto.Email
	}
	if dto.EnvelopeID != nil {
		signatory.EnvelopeID = *dto.EnvelopeID
	}
	if dto.Birthday != nil {
		signatory.Birthday = dto.Birthday
	}
	if dto.PhoneNumber != nil {
		signatory.PhoneNumber = dto.PhoneNumber
	}
	if dto.HasDocumentation != nil {
		signatory.HasDocumentation = dto.HasDocumentation
	}
	if dto.Refusable != nil {
		signatory.Refusable = dto.Refusable
	}
	if dto.Group != nil {
		signatory.Group = dto.Group
	}
	if dto.CommunicateEvents != nil {
		signatory.CommunicateEvents = &entity.CommunicateEvents{
			DocumentSigned:    dto.CommunicateEvents.DocumentSigned,
			SignatureRequest:  dto.CommunicateEvents.SignatureRequest,
			SignatureReminder: dto.CommunicateEvents.SignatureReminder,
		}
	}
	signatory.UpdatedAt = time.Now()
}

// SignatoryResponseDTO representa a estrutura de response para signatário
type SignatoryResponseDTO struct {
	ID                int                            `json:"id"`
	Name              string                         `json:"name"`
	Email             string                         `json:"email"`
	EnvelopeID        int                            `json:"envelope_id"`
	Birthday          *string                        `json:"birthday,omitempty"`
	PhoneNumber       *string                        `json:"phone_number,omitempty"`
	HasDocumentation  *bool                          `json:"has_documentation,omitempty"`
	Refusable         *bool                          `json:"refusable,omitempty"`
	Group             *int                           `json:"group,omitempty"`
	CommunicateEvents *SignatoryCommunicateEventsDTO `json:"communicate_events,omitempty"`
	CreatedAt         time.Time                      `json:"created_at"`
	UpdatedAt         time.Time                      `json:"updated_at"`
}

// FromEntity converte a entidade para o DTO de response
func (dto *SignatoryResponseDTO) FromEntity(signatory *entity.EntitySignatory) {
	dto.ID = signatory.ID
	dto.Name = signatory.Name
	dto.Email = signatory.Email
	dto.EnvelopeID = signatory.EnvelopeID
	dto.Birthday = signatory.Birthday
	dto.PhoneNumber = signatory.PhoneNumber
	dto.HasDocumentation = signatory.HasDocumentation
	dto.Refusable = signatory.Refusable
	dto.Group = signatory.Group
	
	if signatory.CommunicateEvents != nil {
		dto.CommunicateEvents = &SignatoryCommunicateEventsDTO{
			DocumentSigned:    signatory.CommunicateEvents.DocumentSigned,
			SignatureRequest:  signatory.CommunicateEvents.SignatureRequest,
			SignatureReminder: signatory.CommunicateEvents.SignatureReminder,
		}
	}
	
	dto.CreatedAt = signatory.CreatedAt
	dto.UpdatedAt = signatory.UpdatedAt
}

// SignatoryListResponseDTO representa a estrutura de response para lista de signatários
type SignatoryListResponseDTO struct {
	Signatories []SignatoryResponseDTO `json:"signatories"`
	Total       int                    `json:"total"`
}

// SignatoryAssociateRequestDTO representa a estrutura de request para associação de signatário a envelope
type SignatoryAssociateRequestDTO struct {
	EnvelopeID int `json:"envelope_id" binding:"required"`
}

// SignatoryFiltersDTO representa os filtros para consulta de signatários
type SignatoryFiltersDTO struct {
	IDs        []uint `json:"ids,omitempty"`
	EnvelopeID int    `json:"envelope_id,omitempty"`
	Email      string `json:"email,omitempty"`
	Name       string `json:"name,omitempty"`
}

// ToEntityFilters converte os filtros do DTO para a estrutura da entidade
func (dto *SignatoryFiltersDTO) ToEntityFilters() entity.EntitySignatoryFilters {
	return entity.EntitySignatoryFilters{
		IDs:        dto.IDs,
		EnvelopeID: dto.EnvelopeID,
		Email:      dto.Email,
		Name:       dto.Name,
	}
}