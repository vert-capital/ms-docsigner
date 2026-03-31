package entity

import (
	"app/pkg/utils/brdoc"
	"fmt"
	"net/mail"
	"time"
)

// EntityAutoSignatureTerm representa um termo de assinatura automática
type EntityAutoSignatureTerm struct {
	ID                  int       `json:"id" gorm:"primaryKey"`
	SignerDocumentation string    `json:"signer_documentation" gorm:"column:signer_documentation;not null" validate:"required"`
	SignerBirthday      string    `json:"signer_birthday" gorm:"column:signer_birthday;not null" validate:"required"`
	SignerEmail         string    `json:"signer_email" gorm:"column:signer_email;not null" validate:"required,email"`
	SignerName          string    `json:"signer_name" gorm:"column:signer_name;not null" validate:"required,min=2,max=255"`
	AdminEmail          string    `json:"admin_email" gorm:"not null" validate:"required,email"`
	APIEmail            string    `json:"api_email" gorm:"not null" validate:"required,email"`
	ClicksignKey        string    `json:"clicksign_key" gorm:"index"`
	ClicksignRawData    *string   `json:"clicksign_raw_data" gorm:"type:text"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// SignerInfo representa as informações do signatário (para DTOs)
type SignerInfo struct {
	Documentation string `json:"documentation" validate:"required"`
	Birthday      string `json:"birthday" validate:"required"`
	Email         string `json:"email" validate:"required,email"`
	Name          string `json:"name" validate:"required,min=2,max=255"`
}

// TableName sets the table name for GORM
func (EntityAutoSignatureTerm) TableName() string {
	return "auto_signature_terms"
}

func NewAutoSignatureTerm(termParam EntityAutoSignatureTerm) (*EntityAutoSignatureTerm, error) {
	now := time.Now()

	term := &EntityAutoSignatureTerm{
		SignerDocumentation: termParam.SignerDocumentation,
		SignerBirthday:      termParam.SignerBirthday,
		SignerEmail:         termParam.SignerEmail,
		SignerName:          termParam.SignerName,
		AdminEmail:          termParam.AdminEmail,
		APIEmail:            termParam.APIEmail,
		ClicksignKey:        termParam.ClicksignKey,
		ClicksignRawData:    termParam.ClicksignRawData,
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	err := term.Validate()
	if err != nil {
		return nil, err
	}

	return term, nil
}

func (t *EntityAutoSignatureTerm) Validate() error {
	err := validate.Struct(t)
	if err != nil {
		return err
	}

	if err := t.validateEmails(); err != nil {
		return err
	}

	if err := t.validateDocumentation(); err != nil {
		return err
	}

	if err := t.validateBirthday(); err != nil {
		return err
	}

	return nil
}

func (t *EntityAutoSignatureTerm) validateEmails() error {
	// Validar email do signatário
	if _, err := mail.ParseAddress(t.SignerEmail); err != nil {
		return fmt.Errorf("invalid signer email format: %s", t.SignerEmail)
	}

	// Validar admin_email
	if _, err := mail.ParseAddress(t.AdminEmail); err != nil {
		return fmt.Errorf("invalid admin email format: %s", t.AdminEmail)
	}

	// Validar api_email
	if _, err := mail.ParseAddress(t.APIEmail); err != nil {
		return fmt.Errorf("invalid api email format: %s", t.APIEmail)
	}

	return nil
}

func (t *EntityAutoSignatureTerm) validateDocumentation() error {
	if _, err := brdoc.Validate(t.SignerDocumentation); err != nil {
		return err
	}

	return nil
}

func (t *EntityAutoSignatureTerm) validateBirthday() error {
	// Validar formato da data de nascimento (YYYY-MM-DD)
	_, err := time.Parse("2006-01-02", t.SignerBirthday)
	if err != nil {
		return fmt.Errorf("invalid birthday format, expected YYYY-MM-DD: %s", t.SignerBirthday)
	}

	return nil
}
