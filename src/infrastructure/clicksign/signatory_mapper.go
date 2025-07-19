package clicksign

import (
	"app/entity"
	"app/infrastructure/clicksign/dto"
)

// SignatoryMapper fornece funções para mapear entre entidades internas e DTOs do Clicksign
type SignatoryMapper struct{}

// NewSignatoryMapper cria uma nova instância do mapeador
func NewSignatoryMapper() *SignatoryMapper {
	return &SignatoryMapper{}
}

// ToClicksignCreateRequest mapeia uma entidade Signatory para o DTO de criação do Clicksign
func (m *SignatoryMapper) ToClicksignCreateRequest(signatory *entity.EntitySignatory) *dto.SignerCreateRequestWrapper {
	attributes := dto.SignerCreateAttributes{
		Name:             signatory.Name,
		Email:            signatory.Email,
		HasDocumentation: false, // default
		Refusable:        true,  // default
		Group:            1,     // default
	}

	// Mapear campos opcionais
	if signatory.Birthday != nil && *signatory.Birthday != "" {
		attributes.Birthday = *signatory.Birthday
	}

	if signatory.PhoneNumber != nil {
		attributes.PhoneNumber = signatory.PhoneNumber
	}

	if signatory.HasDocumentation != nil {
		attributes.HasDocumentation = *signatory.HasDocumentation
	}

	if signatory.Refusable != nil {
		attributes.Refusable = *signatory.Refusable
	}

	if signatory.Group != nil {
		attributes.Group = *signatory.Group
	}

	// Mapear communicate events
	if signatory.CommunicateEvents != nil {
		attributes.CommunicateEvents = &dto.SignerCommunicateEvents{
			DocumentSigned:    signatory.CommunicateEvents.DocumentSigned,
			SignatureRequest:  signatory.CommunicateEvents.SignatureRequest,
			SignatureReminder: signatory.CommunicateEvents.SignatureReminder,
		}
	} else {
		// Definir valores padrão se não fornecidos
		attributes.CommunicateEvents = &dto.SignerCommunicateEvents{
			DocumentSigned:    "email",
			SignatureRequest:  "email",
			SignatureReminder: "email",
		}
	}

	return &dto.SignerCreateRequestWrapper{
		Data: dto.SignerCreateData{
			Type:       "signers",
			Attributes: attributes,
		},
	}
}

// FromClicksignCreateResponse mapeia a resposta de criação do Clicksign para campos da entidade
func (m *SignatoryMapper) FromClicksignCreateResponse(response *dto.SignerCreateResponseWrapper, signatory *entity.EntitySignatory) {
	if response == nil || signatory == nil {
		return
	}

	// Não precisamos atualizar campos da entidade pois eles já existem
	// A resposta do Clicksign confirma que o signatário foi criado com sucesso
	// O ID do Clicksign seria armazenado se necessário em um campo específico
}

// ValidateForClicksign valida se a entidade tem os dados necessários para integração com Clicksign
func (m *SignatoryMapper) ValidateForClicksign(signatory *entity.EntitySignatory) error {
	if signatory == nil {
		return &ClicksignValidationError{
			Field:   "signatory",
			Message: "signatory cannot be nil",
		}
	}

	if signatory.Name == "" {
		return &ClicksignValidationError{
			Field:   "name",
			Message: "name is required for Clicksign integration",
		}
	}

	if signatory.Email == "" {
		return &ClicksignValidationError{
			Field:   "email",
			Message: "email is required for Clicksign integration",
		}
	}

	// Validar formato de birthday se fornecido
	if signatory.Birthday != nil && *signatory.Birthday != "" {
		if !isValidBirthdayFormat(*signatory.Birthday) {
			return &ClicksignValidationError{
				Field:   "birthday",
				Message: "birthday must be in YYYY-MM-DD format for Clicksign integration",
			}
		}
	}

	return nil
}

// isValidBirthdayFormat verifica se o formato de data está correto
func isValidBirthdayFormat(birthday string) bool {
	// Validação básica do formato YYYY-MM-DD
	if len(birthday) != 10 {
		return false
	}
	
	if birthday[4] != '-' || birthday[7] != '-' {
		return false
	}
	
	// Verificar se os caracteres são dígitos nas posições corretas
	for i, char := range birthday {
		if i == 4 || i == 7 {
			continue // skip hyphens
		}
		if char < '0' || char > '9' {
			return false
		}
	}
	
	return true
}

// ClicksignValidationError representa um erro de validação específico para Clicksign
type ClicksignValidationError struct {
	Field   string
	Message string
}

func (e *ClicksignValidationError) Error() string {
	return e.Field + ": " + e.Message
}