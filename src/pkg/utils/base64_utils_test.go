package utils

import (
	"encoding/base64"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateBase64(t *testing.T) {
	tests := []struct {
		name      string
		data      string
		expectErr bool
	}{
		{
			name:      "Valid base64",
			data:      "SGVsbG8gV29ybGQ=", // "Hello World"
			expectErr: false,
		},
		{
			name:      "Invalid base64",
			data:      "invalid base64!",
			expectErr: true,
		},
		{
			name:      "Empty string",
			data:      "",
			expectErr: true,
		},
		{
			name:      "Too large string",
			data:      generateLargeBase64(MaxBase64Size + 1),
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBase64(tt.data)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDecodeBase64File(t *testing.T) {
	tests := []struct {
		name      string
		data      string
		expectErr bool
	}{
		{
			name:      "Valid PDF base64",
			data:      createValidPDFBase64(),
			expectErr: false,
		},
		{
			name:      "Valid JPEG base64",
			data:      createValidJPEGBase64(),
			expectErr: false,
		},
		{
			name:      "Invalid base64",
			data:      "invalid base64!",
			expectErr: true,
		},
		{
			name:      "Empty string",
			data:      "",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileInfo, err := DecodeBase64File(tt.data)
			
			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, fileInfo)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, fileInfo)
				assert.NotEmpty(t, fileInfo.TempPath)
				assert.Greater(t, fileInfo.Size, int64(0))
				assert.NotEmpty(t, fileInfo.MimeType)
				assert.NotEmpty(t, fileInfo.DecodedData)
				
				// Verificar se arquivo temporário foi criado
				_, err := os.Stat(fileInfo.TempPath)
				assert.NoError(t, err)
				
				// Limpar arquivo temporário
				err = CleanupTempFile(fileInfo.TempPath)
				assert.NoError(t, err)
				
				// Verificar se arquivo foi removido
				_, err = os.Stat(fileInfo.TempPath)
				assert.True(t, os.IsNotExist(err))
			}
		})
	}
}

func TestValidateMimeType(t *testing.T) {
	tests := []struct {
		name      string
		mimeType  string
		expectErr bool
	}{
		{
			name:      "Valid PDF",
			mimeType:  "application/pdf",
			expectErr: false,
		},
		{
			name:      "Valid JPEG",
			mimeType:  "image/jpeg",
			expectErr: false,
		},
		{
			name:      "Valid JPG",
			mimeType:  "image/jpg",
			expectErr: false,
		},
		{
			name:      "Valid PNG",
			mimeType:  "image/png",
			expectErr: false,
		},
		{
			name:      "Valid GIF",
			mimeType:  "image/gif",
			expectErr: false,
		},
		{
			name:      "Invalid MIME type",
			mimeType:  "application/xml",
			expectErr: true,
		},
		{
			name:      "Invalid text type",
			mimeType:  "text/plain",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMimeType(tt.mimeType)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetFileExtensionFromMimeType(t *testing.T) {
	tests := []struct {
		name      string
		mimeType  string
		expected  string
	}{
		{
			name:     "PDF",
			mimeType: "application/pdf",
			expected: ".pdf",
		},
		{
			name:     "JPEG",
			mimeType: "image/jpeg",
			expected: ".jpg",
		},
		{
			name:     "JPG",
			mimeType: "image/jpg",
			expected: ".jpg",
		},
		{
			name:     "PNG",
			mimeType: "image/png",
			expected: ".png",
		},
		{
			name:     "GIF",
			mimeType: "image/gif",
			expected: ".gif",
		},
		{
			name:     "Unknown",
			mimeType: "application/unknown",
			expected: ".bin",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ext := GetFileExtensionFromMimeType(tt.mimeType)
			assert.Equal(t, tt.expected, ext)
		})
	}
}

func TestCleanupTempFile(t *testing.T) {
	// Criar arquivo temporário
	tempFile, err := os.CreateTemp("", "test_cleanup_*")
	require.NoError(t, err)
	tempPath := tempFile.Name()
	tempFile.Close()

	// Verificar se arquivo existe
	_, err = os.Stat(tempPath)
	assert.NoError(t, err)

	// Limpar arquivo
	err = CleanupTempFile(tempPath)
	assert.NoError(t, err)

	// Verificar se arquivo foi removido
	_, err = os.Stat(tempPath)
	assert.True(t, os.IsNotExist(err))

	// Tentar limpar arquivo que não existe (não deve dar erro)
	err = CleanupTempFile(tempPath)
	assert.NoError(t, err)

	// Tentar limpar string vazia (não deve dar erro)
	err = CleanupTempFile("")
	assert.NoError(t, err)
}

// Helper functions

func generateLargeBase64(size int) string {
	data := make([]byte, size)
	for i := range data {
		data[i] = 'A'
	}
	return base64.StdEncoding.EncodeToString(data)
}

func createValidPDFBase64() string {
	// Conteúdo mínimo de um PDF válido
	pdfContent := "%PDF-1.4\n1 0 obj<</Type/Catalog/Pages 2 0 R>>endobj\n2 0 obj<</Type/Pages/Kids[3 0 R]/Count 1>>endobj\n3 0 obj<</Type/Page/Parent 2 0 R/MediaBox[0 0 612 792]>>endobj\nxref\n0 4\n0000000000 65535 f \n0000000010 00000 n \n0000000053 00000 n \n0000000100 00000 n \ntrailer<</Size 4/Root 1 0 R>>\nstartxref\n149\n%%EOF"
	return base64.StdEncoding.EncodeToString([]byte(pdfContent))
}

func createValidJPEGBase64() string {
	// Header mínimo de um JPEG (alguns bytes do header JPEG)
	jpegHeader := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46}
	return base64.StdEncoding.EncodeToString(jpegHeader)
}