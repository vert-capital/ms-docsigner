package dtos

import (
	"errors"
	"time"
)

// DocumentCreateRequestDTO representa a estrutura de request para criação de documento
type DocumentCreateRequestDTO struct {
	Name              string `json:"name" binding:"required,min=3,max=255" example:"Contrato de Prestação de Serviços" doc:"Nome do documento"`
	FilePath          string `json:"file_path,omitempty" example:"/path/to/document.pdf" doc:"Caminho absoluto do arquivo (usar OU file_content_base64)"`
	FileContentBase64 string `json:"file_content_base64,omitempty" example:"JVBERi0xLjQKM..." doc:"Conteúdo do arquivo em base64 (usar OU file_path). Máximo 7.5MB após decodificação"`
	FileSize          int64  `json:"file_size,omitempty" example:"2048576" doc:"Tamanho do arquivo em bytes (obrigatório com file_path, opcional com base64)"`
	MimeType          string `json:"mime_type,omitempty" example:"application/pdf" doc:"Tipo MIME (obrigatório com file_path, opcional com base64). Tipos suportados: application/pdf, image/jpeg, image/png, image/gif"`
	Description       string `json:"description,omitempty" binding:"max=1000" example:"Documento para assinatura digital" doc:"Descrição opcional do documento"`
}

// Validate realiza validação customizada do DTO
func (dto *DocumentCreateRequestDTO) Validate() error {
	// Garantir que apenas um dos campos seja fornecido
	hasFilePath := dto.FilePath != ""
	hasBase64 := dto.FileContentBase64 != ""

	if !hasFilePath && !hasBase64 {
		return errors.New("é necessário fornecer file_path ou file_content_base64")
	}

	if hasFilePath && hasBase64 {
		return errors.New("forneça apenas file_path OU file_content_base64, não ambos")
	}

	// Validações específicas para file_path
	if hasFilePath {
		if dto.FileSize <= 0 {
			return errors.New("file_size é obrigatório quando file_path é fornecido")
		}
		if dto.MimeType == "" {
			return errors.New("mime_type é obrigatório quando file_path é fornecido")
		}
	}

	return nil
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
