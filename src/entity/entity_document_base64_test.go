package entity

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEntityDocument_Base64Support(t *testing.T) {
	t.Run("Valid document from base64", func(t *testing.T) {
		// Criar arquivo temporário para simular base64
		tempFile, err := os.CreateTemp("", "test_base64_*")
		require.NoError(t, err)
		defer os.Remove(tempFile.Name())
		tempFile.Close()

		doc := &EntityDocument{
			Name:         "Test Document",
			FilePath:     tempFile.Name(),
			FileSize:     1024,
			MimeType:     "application/pdf",
			Status:       "draft",
			IsFromBase64: true,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		err = doc.Validate()
		assert.NoError(t, err)
	})

	t.Run("Valid document from file_path", func(t *testing.T) {
		// Criar arquivo temporário para simular file_path
		tempFile, err := os.CreateTemp("", "test_filepath_*")
		require.NoError(t, err)
		defer os.Remove(tempFile.Name())
		tempFile.Close()

		doc := &EntityDocument{
			Name:         "Test Document",
			FilePath:     tempFile.Name(),
			FileSize:     1024,
			MimeType:     "application/pdf",
			Status:       "draft",
			IsFromBase64: false,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		err = doc.Validate()
		assert.NoError(t, err)
	})

	t.Run("Base64 document should skip file existence validation", func(t *testing.T) {
		doc := &EntityDocument{
			Name:         "Test Document",
			FilePath:     "/non/existent/path.pdf",
			FileSize:     1024,
			MimeType:     "application/pdf",
			Status:       "draft",
			IsFromBase64: true,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		err := doc.Validate()
		assert.NoError(t, err, "Base64 documents should not validate file existence")
	})

	t.Run("File path document should validate file existence", func(t *testing.T) {
		doc := &EntityDocument{
			Name:         "Test Document",
			FilePath:     "/non/existent/path.pdf",
			FileSize:     1024,
			MimeType:     "application/pdf",
			Status:       "draft",
			IsFromBase64: false,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		err := doc.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "file does not exist")
	})

	t.Run("Invalid MIME type should fail for both base64 and file_path", func(t *testing.T) {
		// Test base64
		docBase64 := &EntityDocument{
			Name:         "Test Document",
			FilePath:     "/some/path.txt",
			FileSize:     1024,
			MimeType:     "text/plain",
			Status:       "draft",
			IsFromBase64: true,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		err := docBase64.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid mime type")

		// Test file_path
		tempFile, err := os.CreateTemp("", "test_invalid_mime_*")
		require.NoError(t, err)
		defer os.Remove(tempFile.Name())
		tempFile.Close()

		docFilePath := &EntityDocument{
			Name:         "Test Document",
			FilePath:     tempFile.Name(),
			FileSize:     1024,
			MimeType:     "text/plain",
			Status:       "draft",
			IsFromBase64: false,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		err = docFilePath.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid mime type")
	})
}

func TestEntityDocument_ProcessBase64Document(t *testing.T) {
	t.Run("ProcessBase64Document should update fields correctly", func(t *testing.T) {
		doc := &EntityDocument{
			Name:      "Test Document",
			Status:    "draft",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		originalUpdateTime := doc.UpdatedAt
		time.Sleep(1 * time.Millisecond) // Garantir que o tempo seja diferente

		tempPath := "/tmp/test_file.pdf"
		mimeType := "application/pdf"
		size := int64(2048)

		doc.ProcessBase64Document(tempPath, mimeType, size)

		assert.Equal(t, tempPath, doc.FilePath)
		assert.Equal(t, mimeType, doc.MimeType)
		assert.Equal(t, size, doc.FileSize)
		assert.True(t, doc.IsFromBase64)
		assert.True(t, doc.UpdatedAt.After(originalUpdateTime))
	})
}

func TestNewDocument_Base64Support(t *testing.T) {
	t.Run("Create new document with base64 flag", func(t *testing.T) {
		// Criar arquivo temporário
		tempFile, err := os.CreateTemp("", "test_new_doc_*")
		require.NoError(t, err)
		defer os.Remove(tempFile.Name())
		tempFile.Close()

		docParam := EntityDocument{
			Name:         "Test Document",
			FilePath:     tempFile.Name(),
			FileSize:     1024,
			MimeType:     "application/pdf",
			IsFromBase64: true,
		}

		doc, err := NewDocument(docParam)
		require.NoError(t, err)
		assert.NotNil(t, doc)
		assert.True(t, doc.IsFromBase64)
		assert.Equal(t, "draft", doc.Status)
	})

	t.Run("Create new document without base64 flag", func(t *testing.T) {
		// Criar arquivo temporário
		tempFile, err := os.CreateTemp("", "test_new_doc_*")
		require.NoError(t, err)
		defer os.Remove(tempFile.Name())
		tempFile.Close()

		docParam := EntityDocument{
			Name:         "Test Document",
			FilePath:     tempFile.Name(),
			FileSize:     1024,
			MimeType:     "application/pdf",
			IsFromBase64: false,
		}

		doc, err := NewDocument(docParam)
		require.NoError(t, err)
		assert.NotNil(t, doc)
		assert.False(t, doc.IsFromBase64)
		assert.Equal(t, "draft", doc.Status)
	})
}

func TestEntityDocument_MimeTypeValidation(t *testing.T) {
	validMimeTypes := []string{
		"application/pdf",
		"image/jpeg",
		"image/jpg",
		"image/png",
		"image/gif",
	}

	invalidMimeTypes := []string{
		"text/plain",
		"application/xml",
		"video/mp4",
		"audio/mp3",
		"application/zip",
	}

	for _, mimeType := range validMimeTypes {
		t.Run("Valid MIME type: "+mimeType, func(t *testing.T) {
			doc := &EntityDocument{
				Name:         "Test Document",
				FilePath:     "/some/path",
				FileSize:     1024,
				MimeType:     mimeType,
				Status:       "draft",
				IsFromBase64: true, // Skip file existence check
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			}

			err := doc.validateMimeType()
			assert.NoError(t, err)
		})
	}

	for _, mimeType := range invalidMimeTypes {
		t.Run("Invalid MIME type: "+mimeType, func(t *testing.T) {
			doc := &EntityDocument{
				Name:         "Test Document",
				FilePath:     "/some/path",
				FileSize:     1024,
				MimeType:     mimeType,
				Status:       "draft",
				IsFromBase64: true, // Skip file existence check
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			}

			err := doc.validateMimeType()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid mime type")
		})
	}
}