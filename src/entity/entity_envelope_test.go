package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewEnvelope(t *testing.T) {
	t.Run("should create envelope with valid data", func(t *testing.T) {
		// Arrange
		validEnvelope := EntityEnvelope{
			Name:             "Test Envelope",
			Description:      "Test description",
			DocumentsIDs:     []int{1, 2},
			SignatoryEmails:  []string{"test@example.com", "user@example.com"},
			Message:          "Please sign this document",
			AutoClose:        true,
		}

		// Act
		envelope, err := NewEnvelope(validEnvelope)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, envelope)
		assert.Equal(t, "Test Envelope", envelope.Name)
		assert.Equal(t, "Test description", envelope.Description)
		assert.Equal(t, "draft", envelope.Status)
		assert.Equal(t, 3, envelope.RemindInterval) // Default value
		assert.True(t, envelope.AutoClose)
		assert.NotZero(t, envelope.CreatedAt)
		assert.NotZero(t, envelope.UpdatedAt)
	})

	t.Run("should fail with invalid name", func(t *testing.T) {
		// Arrange
		invalidEnvelope := EntityEnvelope{
			Name:            "ab", // Too short
			DocumentsIDs:    []int{1},
			SignatoryEmails: []string{"test@example.com"},
		}

		// Act
		envelope, err := NewEnvelope(invalidEnvelope)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, envelope)
	})

	t.Run("should fail with no documents", func(t *testing.T) {
		// Arrange
		invalidEnvelope := EntityEnvelope{
			Name:            "Test Envelope",
			DocumentsIDs:    []int{}, // Empty
			SignatoryEmails: []string{"test@example.com"},
		}

		// Act
		envelope, err := NewEnvelope(invalidEnvelope)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, envelope)
	})

	t.Run("should fail with no signatories", func(t *testing.T) {
		// Arrange
		invalidEnvelope := EntityEnvelope{
			Name:            "Test Envelope",
			DocumentsIDs:    []int{1},
			SignatoryEmails: []string{}, // Empty
		}

		// Act
		envelope, err := NewEnvelope(invalidEnvelope)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, envelope)
	})

	t.Run("should fail with invalid email", func(t *testing.T) {
		// Arrange
		invalidEnvelope := EntityEnvelope{
			Name:            "Test Envelope",
			DocumentsIDs:    []int{1},
			SignatoryEmails: []string{"invalid-email"}, // Invalid email
		}

		// Act
		envelope, err := NewEnvelope(invalidEnvelope)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, envelope)
	})
}

func TestEnvelopeValidateEmails(t *testing.T) {
	t.Run("should validate correct emails", func(t *testing.T) {
		// Arrange
		envelope := &EntityEnvelope{
			SignatoryEmails: []string{"test@example.com", "user@domain.co.uk"},
		}

		// Act
		err := envelope.validateEmails()

		// Assert
		assert.NoError(t, err)
	})

	t.Run("should fail with invalid email format", func(t *testing.T) {
		// Arrange
		envelope := &EntityEnvelope{
			SignatoryEmails: []string{"invalid-email", "another@invalid"},
		}

		// Act
		err := envelope.validateEmails()

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid email format")
	})
}

func TestEnvelopeValidateDeadline(t *testing.T) {
	t.Run("should accept future deadline", func(t *testing.T) {
		// Arrange
		futureDate := time.Now().Add(10 * 24 * time.Hour)
		envelope := &EntityEnvelope{
			DeadlineAt: &futureDate,
		}

		// Act
		err := envelope.validateDeadline()

		// Assert
		assert.NoError(t, err)
	})

	t.Run("should accept nil deadline", func(t *testing.T) {
		// Arrange
		envelope := &EntityEnvelope{
			DeadlineAt: nil,
		}

		// Act
		err := envelope.validateDeadline()

		// Assert
		assert.NoError(t, err)
	})

	t.Run("should fail with past deadline", func(t *testing.T) {
		// Arrange
		pastDate := time.Now().Add(-1 * time.Hour)
		envelope := &EntityEnvelope{
			DeadlineAt: &pastDate,
		}

		// Act
		err := envelope.validateDeadline()

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "deadline must be in the future")
	})

	t.Run("should fail with deadline too far in future", func(t *testing.T) {
		// Arrange
		farFutureDate := time.Now().Add(100 * 24 * time.Hour)
		envelope := &EntityEnvelope{
			DeadlineAt: &farFutureDate,
		}

		// Act
		err := envelope.validateDeadline()

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "deadline cannot be more than 90 days from now")
	})
}

func TestEnvelopeSetStatus(t *testing.T) {
	t.Run("should set valid status", func(t *testing.T) {
		// Arrange
		envelope := &EntityEnvelope{
			Status: "draft",
		}

		// Act
		err := envelope.SetStatus("sent")

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "sent", envelope.Status)
	})

	t.Run("should fail with invalid status", func(t *testing.T) {
		// Arrange
		envelope := &EntityEnvelope{
			Status: "draft",
		}

		// Act
		err := envelope.SetStatus("invalid")

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid status")
	})
}

func TestEnvelopeSetClicksignKey(t *testing.T) {
	t.Run("should set clicksign key", func(t *testing.T) {
		// Arrange
		envelope := &EntityEnvelope{}

		// Act
		envelope.SetClicksignKey("test-key-123")

		// Assert
		assert.Equal(t, "test-key-123", envelope.ClicksignKey)
	})
}

func TestEnvelopeActivateEnvelope(t *testing.T) {
	t.Run("should activate draft envelope", func(t *testing.T) {
		// Arrange
		envelope := &EntityEnvelope{
			Status: "draft",
		}

		// Act
		err := envelope.ActivateEnvelope()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "sent", envelope.Status)
	})

	t.Run("should fail to activate non-draft envelope", func(t *testing.T) {
		// Arrange
		envelope := &EntityEnvelope{
			Status: "sent",
		}

		// Act
		err := envelope.ActivateEnvelope()

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "envelope must be in 'draft' status to activate")
	})
}

func TestEnvelopeAddDocument(t *testing.T) {
	t.Run("should add document to envelope", func(t *testing.T) {
		// Arrange
		envelope := &EntityEnvelope{
			DocumentsIDs: []int{1, 2},
		}

		// Act
		envelope.AddDocument(3)

		// Assert
		assert.Contains(t, envelope.DocumentsIDs, 3)
		assert.Len(t, envelope.DocumentsIDs, 3)
	})
}

func TestEnvelopeRemoveDocument(t *testing.T) {
	t.Run("should remove document from envelope", func(t *testing.T) {
		// Arrange
		envelope := &EntityEnvelope{
			DocumentsIDs: []int{1, 2, 3},
		}

		// Act
		envelope.RemoveDocument(2)

		// Assert
		assert.NotContains(t, envelope.DocumentsIDs, 2)
		assert.Len(t, envelope.DocumentsIDs, 2)
		assert.Contains(t, envelope.DocumentsIDs, 1)
		assert.Contains(t, envelope.DocumentsIDs, 3)
	})

	t.Run("should not fail when removing non-existent document", func(t *testing.T) {
		// Arrange
		envelope := &EntityEnvelope{
			DocumentsIDs: []int{1, 2},
		}

		// Act
		envelope.RemoveDocument(999)

		// Assert
		assert.Len(t, envelope.DocumentsIDs, 2)
	})
}

func TestEnvelopeAddSignatory(t *testing.T) {
	t.Run("should add valid signatory", func(t *testing.T) {
		// Arrange
		envelope := &EntityEnvelope{
			SignatoryEmails: []string{"test@example.com"},
		}

		// Act
		err := envelope.AddSignatory("new@example.com")

		// Assert
		assert.NoError(t, err)
		assert.Contains(t, envelope.SignatoryEmails, "new@example.com")
		assert.Len(t, envelope.SignatoryEmails, 2)
	})

	t.Run("should fail with invalid email", func(t *testing.T) {
		// Arrange
		envelope := &EntityEnvelope{
			SignatoryEmails: []string{"test@example.com"},
		}

		// Act
		err := envelope.AddSignatory("invalid-email")

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid email format")
		assert.Len(t, envelope.SignatoryEmails, 1)
	})
}

func TestEnvelopeRemoveSignatory(t *testing.T) {
	t.Run("should remove signatory from envelope", func(t *testing.T) {
		// Arrange
		envelope := &EntityEnvelope{
			SignatoryEmails: []string{"test@example.com", "user@example.com", "admin@example.com"},
		}

		// Act
		envelope.RemoveSignatory("user@example.com")

		// Assert
		assert.NotContains(t, envelope.SignatoryEmails, "user@example.com")
		assert.Len(t, envelope.SignatoryEmails, 2)
		assert.Contains(t, envelope.SignatoryEmails, "test@example.com")
		assert.Contains(t, envelope.SignatoryEmails, "admin@example.com")
	})

	t.Run("should not fail when removing non-existent signatory", func(t *testing.T) {
		// Arrange
		envelope := &EntityEnvelope{
			SignatoryEmails: []string{"test@example.com"},
		}

		// Act
		envelope.RemoveSignatory("nonexistent@example.com")

		// Assert
		assert.Len(t, envelope.SignatoryEmails, 1)
	})
}

func TestEnvelopeTableName(t *testing.T) {
	t.Run("should return correct table name", func(t *testing.T) {
		// Arrange
		envelope := EntityEnvelope{}

		// Act
		tableName := envelope.TableName()

		// Assert
		assert.Equal(t, "envelopes", tableName)
	})
}