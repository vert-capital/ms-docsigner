package entity

import (
	"fmt"
	"strings"
	"time"
)

type EntityRequirement struct {
	ID           int       `json:"id" gorm:"primaryKey"`
	EnvelopeID   int       `json:"envelope_id" gorm:"not null;index" validate:"required"`
	ClicksignKey string    `json:"clicksign_key" gorm:"index"`
	Action       string    `json:"action" gorm:"not null" validate:"required,oneof=agree sign provide_evidence"`
	Role         string    `json:"role" gorm:"not null;validate:"omitempty,oneof=sign"`
	Auth         *string   `json:"auth" validate:"omitempty,oneof=email icp_brasil"`
	DocumentID   *string   `json:"document_id" gorm:"index"`
	SignerID     *string   `json:"signer_id" gorm:"index"`
	Status       string    `json:"status" gorm:"not null;default:'pending'" validate:"required,oneof=pending completed"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TableName sets the table name for GORM
func (EntityRequirement) TableName() string {
	return "requirements"
}

func NewRequirement(requirementParam EntityRequirement) (*EntityRequirement, error) {
	now := time.Now()

	// Set default values
	if requirementParam.Role == "" {
		requirementParam.Role = "sign"
	}

	if requirementParam.Status == "" {
		requirementParam.Status = "pending"
	}

	r := &EntityRequirement{
		EnvelopeID:   requirementParam.EnvelopeID,
		ClicksignKey: requirementParam.ClicksignKey,
		Action:       requirementParam.Action,
		Role:         requirementParam.Role,
		Auth:         requirementParam.Auth,
		DocumentID:   requirementParam.DocumentID,
		SignerID:     requirementParam.SignerID,
		Status:       requirementParam.Status,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	err := r.Validate()
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (r *EntityRequirement) Validate() error {
	err := validate.Struct(r)
	if err != nil {
		return err
	}

	if err := r.validateBusinessRules(); err != nil {
		return err
	}

	return nil
}

func (r *EntityRequirement) validateBusinessRules() error {
	// For action "provide_evidence", auth is required
	if r.Action == "provide_evidence" && (r.Auth == nil || *r.Auth == "") {
		return fmt.Errorf("auth is required for action 'provide_evidence'")
	}

	return nil
}

func (r *EntityRequirement) SetStatus(status string) error {
	validStatuses := []string{"pending", "completed"}

	for _, validStatus := range validStatuses {
		if status == validStatus {
			r.Status = status
			r.UpdatedAt = time.Now()
			return nil
		}
	}

	return fmt.Errorf("invalid status: %s. Valid statuses: %s", status, strings.Join(validStatuses, ", "))
}

func (r *EntityRequirement) SetClicksignKey(key string) {
	r.ClicksignKey = key
	r.UpdatedAt = time.Now()
}

func (r *EntityRequirement) Complete() error {
	if r.Status != "pending" {
		return fmt.Errorf("requirement must be in 'pending' status to complete, current status: %s", r.Status)
	}

	r.Status = "completed"
	r.UpdatedAt = time.Now()

	return nil
}

func (r *EntityRequirement) IsValidAction(action string) bool {
	validActions := []string{"agree", "sign", "provide_evidence"}
	for _, validAction := range validActions {
		if action == validAction {
			return true
		}
	}
	return false
}

func (r *EntityRequirement) IsValidAuth(auth string) bool {
	validAuths := []string{"email", "icp_brasil"}
	for _, validAuth := range validAuths {
		if auth == validAuth {
			return true
		}
	}
	return false
}
