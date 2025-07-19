package entity

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type EntityDocumentFilters struct {
	IDs          []uint `json:"ids"`
	Search       string `json:"search"`
	Status       string `json:"status"`
	ClicksignKey string `json:"clicksign_key"`
}

type EntityDocument struct {
	ID           int       `json:"id" gorm:"primaryKey"`
	Name         string    `json:"name" gorm:"not null" validate:"required,min=3,max=255"`
	FilePath     string    `json:"file_path" gorm:"not null" validate:"required"`
	FileSize     int64     `json:"file_size" gorm:"not null" validate:"required,gt=0"`
	MimeType     string    `json:"mime_type" gorm:"not null" validate:"required"`
	Status       string    `json:"status" gorm:"not null;default:'draft'" validate:"required,oneof=draft ready processing sent"`
	ClicksignKey string    `json:"clicksign_key" gorm:"index"`
	Description  string    `json:"description" validate:"max=1000"`
	IsFromBase64 bool      `json:"is_from_base64" gorm:"default:false"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TableName sets the table name for GORM
func (EntityDocument) TableName() string {
	return "documents"
}

func NewDocument(docParam EntityDocument) (*EntityDocument, error) {
	now := time.Now()

	if docParam.Status == "" {
		docParam.Status = "draft"
	}

	d := &EntityDocument{
		Name:         docParam.Name,
		FilePath:     docParam.FilePath,
		FileSize:     docParam.FileSize,
		MimeType:     docParam.MimeType,
		Status:       docParam.Status,
		ClicksignKey: docParam.ClicksignKey,
		Description:  docParam.Description,
		IsFromBase64: docParam.IsFromBase64,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	err := d.Validate()
	if err != nil {
		return nil, err
	}

	return d, nil
}

func (d *EntityDocument) Validate() error {
	err := validate.Struct(d)
	if err != nil {
		return err
	}

	// Validar se arquivo existe apenas quando não for base64
	if !d.IsFromBase64 {
		if err := d.validateFileExists(); err != nil {
			return err
		}
	}

	if err := d.validateMimeType(); err != nil {
		return err
	}

	return nil
}

func (d *EntityDocument) validateFileExists() error {
	if _, err := os.Stat(d.FilePath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", d.FilePath)
	}
	return nil
}

func (d *EntityDocument) validateMimeType() error {
	validMimeTypes := []string{
		"application/pdf",
		"image/jpeg",
		"image/jpg",
		"image/png",
		"image/gif",
	}

	for _, validType := range validMimeTypes {
		if strings.EqualFold(d.MimeType, validType) {
			return nil
		}
	}

	return fmt.Errorf("invalid mime type: %s. Allowed types: %s", d.MimeType, strings.Join(validMimeTypes, ", "))
}

func (d *EntityDocument) PrepareForSigning() error {
	if d.Status != "ready" {
		return fmt.Errorf("document must be in 'ready' status to prepare for signing, current status: %s", d.Status)
	}

	d.Status = "processing"
	d.UpdatedAt = time.Now()

	return nil
}

func (d *EntityDocument) SetStatus(status string) error {
	validStatuses := []string{"draft", "ready", "processing", "sent"}

	for _, validStatus := range validStatuses {
		if status == validStatus {
			d.Status = status
			d.UpdatedAt = time.Now()
			return nil
		}
	}

	return fmt.Errorf("invalid status: %s. Valid statuses: %s", status, strings.Join(validStatuses, ", "))
}

func (d *EntityDocument) SetClicksignKey(key string) {
	d.ClicksignKey = key
	d.UpdatedAt = time.Now()
}

// ProcessBase64Document processa um documento que veio de base64
func (d *EntityDocument) ProcessBase64Document(tempPath, mimeType string, size int64) {
	d.FilePath = tempPath
	d.MimeType = mimeType
	d.FileSize = size
	d.IsFromBase64 = true
	d.UpdatedAt = time.Now()
}
