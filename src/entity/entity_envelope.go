package entity

import (
	"fmt"
	"net/mail"
	"strings"
	"time"
)

type EntityEnvelopeFilters struct {
	IDs          []uint `json:"ids"`
	Search       string `json:"search"`
	Status       string `json:"status"`
	ClicksignKey string `json:"clicksign_key"`
}

type EntityEnvelope struct {
	ID              int        `json:"id" gorm:"primaryKey"`
	Name            string     `json:"name" gorm:"not null" validate:"required,min=3,max=255"`
	Description     string     `json:"description" validate:"max=1000"`
	Status          string     `json:"status" gorm:"not null;default:'draft'" validate:"required,oneof=draft sent pending completed cancelled"`
	ClicksignKey    string     `json:"clicksign_key" gorm:"index"`
	DocumentsIDs    []int      `json:"documents_ids" gorm:"serializer:json" validate:"required,min=1"`
	SignatoryEmails []string   `json:"signatory_emails" gorm:"serializer:json"`
	Message         string     `json:"message" validate:"max=500"`
	DeadlineAt      *time.Time `json:"deadline_at"`
	RemindInterval  int        `json:"remind_interval" validate:"min=1,max=30"`
	AutoClose       bool       `json:"auto_close" gorm:"default:true"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// TableName sets the table name for GORM
func (EntityEnvelope) TableName() string {
	return "envelopes"
}

func NewEnvelope(envelopeParam EntityEnvelope) (*EntityEnvelope, error) {
	now := time.Now()

	if envelopeParam.Status == "" {
		envelopeParam.Status = "draft"
	}

	if envelopeParam.RemindInterval == 0 {
		envelopeParam.RemindInterval = 3
	}

	e := &EntityEnvelope{
		Name:            envelopeParam.Name,
		Description:     envelopeParam.Description,
		Status:          envelopeParam.Status,
		ClicksignKey:    envelopeParam.ClicksignKey,
		DocumentsIDs:    envelopeParam.DocumentsIDs,
		SignatoryEmails: envelopeParam.SignatoryEmails,
		Message:         envelopeParam.Message,
		DeadlineAt:      envelopeParam.DeadlineAt,
		RemindInterval:  envelopeParam.RemindInterval,
		AutoClose:       envelopeParam.AutoClose,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	err := e.Validate()
	if err != nil {
		return nil, err
	}

	return e, nil
}

func (e *EntityEnvelope) Validate() error {
	err := validate.Struct(e)
	if err != nil {
		return err
	}

	if err := e.validateEmails(); err != nil {
		return err
	}

	if err := e.validateDeadline(); err != nil {
		return err
	}

	return nil
}

func (e *EntityEnvelope) validateEmails() error {
	// Validar apenas se houver emails na lista
	for _, email := range e.SignatoryEmails {
		if _, err := mail.ParseAddress(email); err != nil {
			return fmt.Errorf("invalid email format: %s", email)
		}
	}
	return nil
}

func (e *EntityEnvelope) validateDeadline() error {
	if e.DeadlineAt != nil {
		now := time.Now()
		if e.DeadlineAt.Before(now) {
			return fmt.Errorf("deadline must be in the future")
		}

		maxDeadline := now.Add(90 * 24 * time.Hour)
		if e.DeadlineAt.After(maxDeadline) {
			return fmt.Errorf("deadline cannot be more than 90 days from now")
		}
	}
	return nil
}

func (e *EntityEnvelope) SetStatus(status string) error {
	validStatuses := []string{"draft", "sent", "pending", "completed", "cancelled"}

	for _, validStatus := range validStatuses {
		if status == validStatus {
			e.Status = status
			e.UpdatedAt = time.Now()
			return nil
		}
	}

	return fmt.Errorf("invalid status: %s. Valid statuses: %s", status, strings.Join(validStatuses, ", "))
}

func (e *EntityEnvelope) SetClicksignKey(key string) {
	e.ClicksignKey = key
	e.UpdatedAt = time.Now()
}

func (e *EntityEnvelope) ActivateEnvelope() error {
	if e.Status != "draft" {
		return fmt.Errorf("envelope must be in 'draft' status to activate, current status: %s", e.Status)
	}

	e.Status = "sent"
	e.UpdatedAt = time.Now()

	return nil
}

func (e *EntityEnvelope) AddDocument(documentID int) {
	e.DocumentsIDs = append(e.DocumentsIDs, documentID)
	e.UpdatedAt = time.Now()
}

func (e *EntityEnvelope) RemoveDocument(documentID int) {
	for i, id := range e.DocumentsIDs {
		if id == documentID {
			e.DocumentsIDs = append(e.DocumentsIDs[:i], e.DocumentsIDs[i+1:]...)
			e.UpdatedAt = time.Now()
			break
		}
	}
}

func (e *EntityEnvelope) AddSignatory(email string) error {
	if _, err := mail.ParseAddress(email); err != nil {
		return fmt.Errorf("invalid email format: %s", email)
	}

	e.SignatoryEmails = append(e.SignatoryEmails, email)
	e.UpdatedAt = time.Now()
	return nil
}

func (e *EntityEnvelope) RemoveSignatory(email string) {
	for i, signatoryEmail := range e.SignatoryEmails {
		if signatoryEmail == email {
			e.SignatoryEmails = append(e.SignatoryEmails[:i], e.SignatoryEmails[i+1:]...)
			e.UpdatedAt = time.Now()
			break
		}
	}
}
