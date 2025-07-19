package dtos

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDocumentCreateRequestDTO_Validate(t *testing.T) {
	tests := []struct {
		name      string
		dto       DocumentCreateRequestDTO
		expectErr bool
		errMsg    string
	}{
		{
			name: "Valid with file_path",
			dto: DocumentCreateRequestDTO{
				Name:     "Test Document",
				FilePath: "/path/to/file.pdf",
				FileSize: 1024,
				MimeType: "application/pdf",
			},
			expectErr: false,
		},
		{
			name: "Valid with file_content_base64",
			dto: DocumentCreateRequestDTO{
				Name:              "Test Document",
				FileContentBase64: "SGVsbG8gV29ybGQ=",
			},
			expectErr: false,
		},
		{
			name: "Invalid - both fields provided",
			dto: DocumentCreateRequestDTO{
				Name:              "Test Document",
				FilePath:          "/path/to/file.pdf",
				FileContentBase64: "SGVsbG8gV29ybGQ=",
				FileSize:          1024,
				MimeType:          "application/pdf",
			},
			expectErr: true,
			errMsg:    "forneça apenas file_path OU file_content_base64, não ambos",
		},
		{
			name: "Invalid - neither field provided",
			dto: DocumentCreateRequestDTO{
				Name: "Test Document",
			},
			expectErr: true,
			errMsg:    "é necessário fornecer file_path ou file_content_base64",
		},
		{
			name: "Invalid - file_path without file_size",
			dto: DocumentCreateRequestDTO{
				Name:     "Test Document",
				FilePath: "/path/to/file.pdf",
				MimeType: "application/pdf",
			},
			expectErr: true,
			errMsg:    "file_size é obrigatório quando file_path é fornecido",
		},
		{
			name: "Invalid - file_path without mime_type",
			dto: DocumentCreateRequestDTO{
				Name:     "Test Document",
				FilePath: "/path/to/file.pdf",
				FileSize: 1024,
			},
			expectErr: true,
			errMsg:    "mime_type é obrigatório quando file_path é fornecido",
		},
		{
			name: "Invalid - file_path with zero file_size",
			dto: DocumentCreateRequestDTO{
				Name:     "Test Document",
				FilePath: "/path/to/file.pdf",
				FileSize: 0,
				MimeType: "application/pdf",
			},
			expectErr: true,
			errMsg:    "file_size é obrigatório quando file_path é fornecido",
		},
		{
			name: "Valid - base64 without file_size and mime_type",
			dto: DocumentCreateRequestDTO{
				Name:              "Test Document",
				FileContentBase64: "SGVsbG8gV29ybGQ=",
				Description:       "Test description",
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.dto.Validate()
			
			if tt.expectErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDocumentCreateRequestDTO_ValidateFields(t *testing.T) {
	t.Run("Valid file_path scenario", func(t *testing.T) {
		dto := DocumentCreateRequestDTO{
			Name:        "Test Document",
			FilePath:    "/valid/path/document.pdf",
			FileSize:    2048,
			MimeType:    "application/pdf",
			Description: "Test description",
		}

		err := dto.Validate()
		assert.NoError(t, err)
	})

	t.Run("Valid base64 scenario", func(t *testing.T) {
		dto := DocumentCreateRequestDTO{
			Name:              "Test Document",
			FileContentBase64: "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8/5+hHgAHggJ/PchI7wAAAABJRU5ErkJggg==",
			Description:       "Test description",
		}

		err := dto.Validate()
		assert.NoError(t, err)
	})

	t.Run("Edge case - empty strings", func(t *testing.T) {
		dto := DocumentCreateRequestDTO{
			Name:              "Test Document",
			FilePath:          "",
			FileContentBase64: "",
		}

		err := dto.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "é necessário fornecer file_path ou file_content_base64")
	})

	t.Run("Edge case - whitespace strings", func(t *testing.T) {
		dto := DocumentCreateRequestDTO{
			Name:              "Test Document",
			FilePath:          "   ",
			FileContentBase64: "   ",
		}

		// Note: Since we're checking for empty strings with !=, whitespace counts as non-empty
		err := dto.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "forneça apenas file_path OU file_content_base64, não ambos")
	})
}