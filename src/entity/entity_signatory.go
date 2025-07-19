package entity

import (
	"encoding/json"
	"fmt"
	"net/mail"
	"regexp"
	"time"
)

type EntitySignatoryFilters struct {
	IDs        []uint `json:"ids"`
	EnvelopeID int    `json:"envelope_id"`
	Email      string `json:"email"`
	Name       string `json:"name"`
}

type CommunicateEvents struct {
	DocumentSigned     string `json:"document_signed"`
	SignatureRequest   string `json:"signature_request"`
	SignatureReminder  string `json:"signature_reminder"`
}

type EntitySignatory struct {
	ID                int                `json:"id" gorm:"primaryKey"`
	Name              string             `json:"name" gorm:"not null" validate:"required,min=2,max=255"`
	Email             string             `json:"email" gorm:"not null" validate:"required,email"`
	EnvelopeID        int                `json:"envelope_id" gorm:"not null" validate:"required"`
	Birthday          *string            `json:"birthday,omitempty"`
	PhoneNumber       *string            `json:"phone_number,omitempty"`
	HasDocumentation  *bool              `json:"has_documentation,omitempty"`
	Refusable         *bool              `json:"refusable,omitempty"`
	Group             *int               `json:"group,omitempty"`
	CommunicateEvents *CommunicateEvents `json:"communicate_events,omitempty" gorm:"serializer:json"`
	CreatedAt         time.Time          `json:"created_at"`
	UpdatedAt         time.Time          `json:"updated_at"`
}

// TableName sets the table name for GORM
func (EntitySignatory) TableName() string {
	return "signatories"
}

func NewSignatory(signatoryParam EntitySignatory) (*EntitySignatory, error) {
	now := time.Now()

	// Set default values for optional fields only when not explicitly provided
	if signatoryParam.HasDocumentation == nil {
		defaultHasDoc := false
		signatoryParam.HasDocumentation = &defaultHasDoc
	}

	if signatoryParam.Refusable == nil {
		defaultRefusable := true
		signatoryParam.Refusable = &defaultRefusable
	}

	if signatoryParam.Group == nil {
		defaultGroup := 1
		signatoryParam.Group = &defaultGroup
	}

	// Set default communicate events if not provided
	if signatoryParam.CommunicateEvents == nil {
		signatoryParam.CommunicateEvents = &CommunicateEvents{
			DocumentSigned:    "email",
			SignatureRequest:  "email",
			SignatureReminder: "email",
		}
	}

	s := &EntitySignatory{
		Name:              signatoryParam.Name,
		Email:             signatoryParam.Email,
		EnvelopeID:        signatoryParam.EnvelopeID,
		Birthday:          signatoryParam.Birthday,
		PhoneNumber:       signatoryParam.PhoneNumber,
		HasDocumentation:  signatoryParam.HasDocumentation,
		Refusable:         signatoryParam.Refusable,
		Group:             signatoryParam.Group,
		CommunicateEvents: signatoryParam.CommunicateEvents,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	err := s.Validate()
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (s *EntitySignatory) Validate() error {
	err := validate.Struct(s)
	if err != nil {
		return err
	}

	if err := s.validateEmail(); err != nil {
		return err
	}

	if err := s.validateBirthday(); err != nil {
		return err
	}

	if err := s.validatePhoneNumber(); err != nil {
		return err
	}

	if err := s.validateGroup(); err != nil {
		return err
	}

	if err := s.validateCommunicateEvents(); err != nil {
		return err
	}

	return nil
}

func (s *EntitySignatory) validateEmail() error {
	if _, err := mail.ParseAddress(s.Email); err != nil {
		return fmt.Errorf("invalid email format: %s", s.Email)
	}
	return nil
}

func (s *EntitySignatory) validateBirthday() error {
	if s.Birthday != nil {
		// Validate YYYY-MM-DD format
		birthdayRegex := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
		if !birthdayRegex.MatchString(*s.Birthday) {
			return fmt.Errorf("birthday must be in YYYY-MM-DD format, got: %s", *s.Birthday)
		}

		// Validate that it's a valid date
		_, err := time.Parse("2006-01-02", *s.Birthday)
		if err != nil {
			return fmt.Errorf("invalid birthday date: %s", *s.Birthday)
		}
	}
	return nil
}

func (s *EntitySignatory) validatePhoneNumber() error {
	if s.PhoneNumber != nil {
		// Basic international phone number validation (starts with + and contains digits)
		phoneRegex := regexp.MustCompile(`^\+\d{8,15}$`)
		if !phoneRegex.MatchString(*s.PhoneNumber) {
			return fmt.Errorf("phone number must be in international format (+xxxxxxxx), got: %s", *s.PhoneNumber)
		}
	}
	return nil
}

func (s *EntitySignatory) validateGroup() error {
	if s.Group != nil && *s.Group <= 0 {
		return fmt.Errorf("group must be a positive integer, got: %d", *s.Group)
	}
	return nil
}

func (s *EntitySignatory) validateCommunicateEvents() error {
	if s.CommunicateEvents != nil {
		validEvents := []string{"email", "sms", "none"}
		
		// Validate document_signed
		if !isValidEventType(s.CommunicateEvents.DocumentSigned, validEvents) {
			return fmt.Errorf("invalid document_signed event type: %s", s.CommunicateEvents.DocumentSigned)
		}

		// Validate signature_request
		if !isValidEventType(s.CommunicateEvents.SignatureRequest, validEvents) {
			return fmt.Errorf("invalid signature_request event type: %s", s.CommunicateEvents.SignatureRequest)
		}

		// Validate signature_reminder
		if !isValidEventType(s.CommunicateEvents.SignatureReminder, validEvents) {
			return fmt.Errorf("invalid signature_reminder event type: %s", s.CommunicateEvents.SignatureReminder)
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

func (s *EntitySignatory) SetCommunicateEvents(events CommunicateEvents) error {
	s.CommunicateEvents = &events
	s.UpdatedAt = time.Now()
	return s.validateCommunicateEvents()
}

func (s *EntitySignatory) SetGroup(group int) error {
	if group <= 0 {
		return fmt.Errorf("group must be a positive integer, got: %d", group)
	}
	s.Group = &group
	s.UpdatedAt = time.Now()
	return nil
}

func (s *EntitySignatory) SetBirthday(birthday string) error {
	s.Birthday = &birthday
	s.UpdatedAt = time.Now()
	return s.validateBirthday()
}

func (s *EntitySignatory) SetPhoneNumber(phoneNumber string) error {
	s.PhoneNumber = &phoneNumber
	s.UpdatedAt = time.Now()
	return s.validatePhoneNumber()
}

func (s *EntitySignatory) ToJSON() ([]byte, error) {
	return json.Marshal(s)
}

func (s *EntitySignatory) FromJSON(data []byte) error {
	return json.Unmarshal(data, s)
}