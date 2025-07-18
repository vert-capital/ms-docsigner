package entity

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDocument(t *testing.T) {
	// Create a temporary file for testing
	tempFile, err := os.CreateTemp("", "test_document*.pdf")
	require.NoError(t, err)
	defer os.Remove(tempFile.Name())

	_, err = tempFile.Write([]byte("test content"))
	require.NoError(t, err)
	tempFile.Close()

	fileInfo, err := os.Stat(tempFile.Name())
	require.NoError(t, err)

	tests := []struct {
		name        string
		docParam    EntityDocument
		shouldError bool
		errorMsg    string
	}{
		{
			name: "Valid document",
			docParam: EntityDocument{
				Name:        "Test Document",
				FilePath:    tempFile.Name(),
				FileSize:    fileInfo.Size(),
				MimeType:    "application/pdf",
				Description: "Test description",
			},
			shouldError: false,
		},
		{
			name: "Valid document with custom status",
			docParam: EntityDocument{
				Name:        "Test Document",
				FilePath:    tempFile.Name(),
				FileSize:    fileInfo.Size(),
				MimeType:    "application/pdf",
				Status:      "ready",
				Description: "Test description",
			},
			shouldError: false,
		},
		{
			name: "Invalid - empty name",
			docParam: EntityDocument{
				Name:        "",
				FilePath:    tempFile.Name(),
				FileSize:    fileInfo.Size(),
				MimeType:    "application/pdf",
				Description: "Test description",
			},
			shouldError: true,
		},
		{
			name: "Invalid - name too short",
			docParam: EntityDocument{
				Name:        "ab",
				FilePath:    tempFile.Name(),
				FileSize:    fileInfo.Size(),
				MimeType:    "application/pdf",
				Description: "Test description",
			},
			shouldError: true,
		},
		{
			name: "Invalid - file does not exist",
			docParam: EntityDocument{
				Name:        "Test Document",
				FilePath:    "/nonexistent/file.pdf",
				FileSize:    1000,
				MimeType:    "application/pdf",
				Description: "Test description",
			},
			shouldError: true,
			errorMsg:    "file does not exist",
		},
		{
			name: "Invalid - zero file size",
			docParam: EntityDocument{
				Name:        "Test Document",
				FilePath:    tempFile.Name(),
				FileSize:    0,
				MimeType:    "application/pdf",
				Description: "Test description",
			},
			shouldError: true,
		},
		{
			name: "Invalid - invalid mime type",
			docParam: EntityDocument{
				Name:        "Test Document",
				FilePath:    tempFile.Name(),
				FileSize:    fileInfo.Size(),
				MimeType:    "application/msword",
				Description: "Test description",
			},
			shouldError: true,
			errorMsg:    "invalid mime type",
		},
		{
			name: "Invalid - invalid status",
			docParam: EntityDocument{
				Name:        "Test Document",
				FilePath:    tempFile.Name(),
				FileSize:    fileInfo.Size(),
				MimeType:    "application/pdf",
				Status:      "invalid_status",
				Description: "Test description",
			},
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := NewDocument(tt.docParam)

			if tt.shouldError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				assert.Nil(t, doc)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, doc)
				assert.Equal(t, tt.docParam.Name, doc.Name)
				assert.Equal(t, tt.docParam.FilePath, doc.FilePath)
				assert.Equal(t, tt.docParam.FileSize, doc.FileSize)
				assert.Equal(t, tt.docParam.MimeType, doc.MimeType)
				assert.Equal(t, tt.docParam.Description, doc.Description)

				// Check default status if not provided
				if tt.docParam.Status == "" {
					assert.Equal(t, "draft", doc.Status)
				} else {
					assert.Equal(t, tt.docParam.Status, doc.Status)
				}

				// Check timestamps
				assert.False(t, doc.CreatedAt.IsZero())
				assert.False(t, doc.UpdatedAt.IsZero())
			}
		})
	}
}

func TestEntityDocument_Validate(t *testing.T) {
	// Create a temporary file for testing
	tempFile, err := os.CreateTemp("", "test_document*.pdf")
	require.NoError(t, err)
	defer os.Remove(tempFile.Name())

	_, err = tempFile.Write([]byte("test content"))
	require.NoError(t, err)
	tempFile.Close()

	fileInfo, err := os.Stat(tempFile.Name())
	require.NoError(t, err)

	validDoc := &EntityDocument{
		Name:        "Test Document",
		FilePath:    tempFile.Name(),
		FileSize:    fileInfo.Size(),
		MimeType:    "application/pdf",
		Status:      "draft",
		Description: "Test description",
	}

	err = validDoc.Validate()
	assert.NoError(t, err)
}

func TestEntityDocument_ValidateMimeType(t *testing.T) {
	doc := &EntityDocument{}

	tests := []struct {
		name        string
		mimeType    string
		shouldError bool
	}{
		{"Valid PDF", "application/pdf", false},
		{"Valid JPEG", "image/jpeg", false},
		{"Valid JPG", "image/jpg", false},
		{"Valid PNG", "image/png", false},
		{"Valid GIF", "image/gif", false},
		{"Case insensitive", "APPLICATION/PDF", false},
		{"Invalid Word", "application/msword", true},
		{"Invalid Excel", "application/vnd.ms-excel", true},
		{"Invalid text", "text/plain", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc.MimeType = tt.mimeType
			err := doc.validateMimeType()

			if tt.shouldError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "invalid mime type")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestEntityDocument_PrepareForSigning(t *testing.T) {
	tests := []struct {
		name        string
		status      string
		shouldError bool
		errorMsg    string
	}{
		{
			name:        "Valid ready status",
			status:      "ready",
			shouldError: false,
		},
		{
			name:        "Invalid draft status",
			status:      "draft",
			shouldError: true,
			errorMsg:    "document must be in 'ready' status",
		},
		{
			name:        "Invalid processing status",
			status:      "processing",
			shouldError: true,
			errorMsg:    "document must be in 'ready' status",
		},
		{
			name:        "Invalid sent status",
			status:      "sent",
			shouldError: true,
			errorMsg:    "document must be in 'ready' status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := &EntityDocument{
				Status:    tt.status,
				UpdatedAt: time.Now().Add(-1 * time.Hour),
			}

			err := doc.PrepareForSigning()

			if tt.shouldError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, "processing", doc.Status)
				assert.True(t, doc.UpdatedAt.After(time.Now().Add(-1*time.Minute)))
			}
		})
	}
}

func TestEntityDocument_SetStatus(t *testing.T) {
	doc := &EntityDocument{
		Status:    "draft",
		UpdatedAt: time.Now().Add(-1 * time.Hour),
	}

	tests := []struct {
		name        string
		status      string
		shouldError bool
	}{
		{"Valid draft", "draft", false},
		{"Valid ready", "ready", false},
		{"Valid processing", "processing", false},
		{"Valid sent", "sent", false},
		{"Invalid status", "invalid", true},
		{"Empty status", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldUpdatedAt := doc.UpdatedAt
			err := doc.SetStatus(tt.status)

			if tt.shouldError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "invalid status")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.status, doc.Status)
				assert.True(t, doc.UpdatedAt.After(oldUpdatedAt))
			}
		})
	}
}

func TestEntityDocument_SetClicksignKey(t *testing.T) {
	doc := &EntityDocument{
		ClicksignKey: "",
		UpdatedAt:    time.Now().Add(-1 * time.Hour),
	}

	oldUpdatedAt := doc.UpdatedAt
	testKey := "test-clicksign-key-123"

	doc.SetClicksignKey(testKey)

	assert.Equal(t, testKey, doc.ClicksignKey)
	assert.True(t, doc.UpdatedAt.After(oldUpdatedAt))
}
