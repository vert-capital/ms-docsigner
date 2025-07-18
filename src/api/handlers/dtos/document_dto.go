package dtos

import "time"

// DocumentCreateRequestDTO representa a estrutura de request para criação de documento
type DocumentCreateRequestDTO struct {
	Name        string `json:"name" binding:"required,min=3,max=255"`
	FilePath    string `json:"file_path" binding:"required"`
	FileSize    int64  `json:"file_size" binding:"required,gt=0"`
	MimeType    string `json:"mime_type" binding:"required"`
	Description string `json:"description,omitempty" binding:"max=1000"`
}

// DocumentUpdateRequestDTO representa a estrutura de request para atualização de documento
type DocumentUpdateRequestDTO struct {
	Name        *string `json:"name,omitempty" binding:"omitempty,min=3,max=255"`
	Description *string `json:"description,omitempty" binding:"omitempty,max=1000"`
	Status      *string `json:"status,omitempty" binding:"omitempty,oneof=draft ready processing sent"`
}

// DocumentResponseDTO representa a estrutura de response para documento
type DocumentResponseDTO struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	FilePath     string    `json:"file_path"`
	FileSize     int64     `json:"file_size"`
	MimeType     string    `json:"mime_type"`
	Status       string    `json:"status"`
	ClicksignKey string    `json:"clicksign_key"`
	Description  string    `json:"description"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// DocumentListResponseDTO representa a estrutura de response para lista de documentos
type DocumentListResponseDTO struct {
	Documents []DocumentResponseDTO `json:"documents"`
	Total     int                   `json:"total"`
}
