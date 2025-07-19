package dtos

import (
	"testing"
	"time"

	"app/entity"
	"github.com/stretchr/testify/assert"
)

func TestSignatoryCreateRequestDTO_Validate(t *testing.T) {
	t.Run("should validate correct DTO", func(t *testing.T) {
		// Arrange
		birthday := "1990-01-01"
		phoneNumber := "+5511999999999"
		hasDoc := true
		refusable := false
		group := 2

		dto := SignatoryCreateRequestDTO{
			Name:             "John Doe",
			Email:            "john.doe@example.com",
			EnvelopeID:       1,
			Birthday:         &birthday,
			PhoneNumber:      &phoneNumber,
			HasDocumentation: &hasDoc,
			Refusable:        &refusable,
			Group:            &group,
			CommunicateEvents: &SignatoryCommunicateEventsDTO{
				DocumentSigned:    "email",
				SignatureRequest:  "sms",
				SignatureReminder: "none",
			},
		}

		// Act
		err := dto.Validate()

		// Assert
		assert.NoError(t, err)
	})

	t.Run("should fail with invalid email", func(t *testing.T) {
		// Arrange
		dto := SignatoryCreateRequestDTO{
			Name:       "John Doe",
			Email:      "invalid-email",
			EnvelopeID: 1,
		}

		// Act
		err := dto.Validate()

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid email format")
	})

	t.Run("should fail with invalid birthday format", func(t *testing.T) {
		// Arrange
		invalidBirthday := "01/01/1990"
		dto := SignatoryCreateRequestDTO{
			Name:       "John Doe",
			Email:      "john.doe@example.com",
			EnvelopeID: 1,
			Birthday:   &invalidBirthday,
		}

		// Act
		err := dto.Validate()

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "birthday must be in YYYY-MM-DD format")
	})

	t.Run("should fail with invalid birthday date", func(t *testing.T) {
		// Arrange
		invalidDate := "1990-13-32"
		dto := SignatoryCreateRequestDTO{
			Name:       "John Doe",
			Email:      "john.doe@example.com",
			EnvelopeID: 1,
			Birthday:   &invalidDate,
		}

		// Act
		err := dto.Validate()

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid birthday date")
	})

	t.Run("should fail with invalid phone number", func(t *testing.T) {
		// Arrange
		invalidPhone := "11999999999"
		dto := SignatoryCreateRequestDTO{
			Name:        "John Doe",
			Email:       "john.doe@example.com",
			EnvelopeID:  1,
			PhoneNumber: &invalidPhone,
		}

		// Act
		err := dto.Validate()

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "phone number must be in international format")
	})

	t.Run("should fail with invalid group", func(t *testing.T) {
		// Arrange
		invalidGroup := 0
		dto := SignatoryCreateRequestDTO{
			Name:       "John Doe",
			Email:      "john.doe@example.com",
			EnvelopeID: 1,
			Group:      &invalidGroup,
		}

		// Act
		err := dto.Validate()

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "group must be a positive integer")
	})

	t.Run("should fail with invalid communicate events", func(t *testing.T) {
		// Arrange
		dto := SignatoryCreateRequestDTO{
			Name:       "John Doe",
			Email:      "john.doe@example.com",
			EnvelopeID: 1,
			CommunicateEvents: &SignatoryCommunicateEventsDTO{
				DocumentSigned:    "invalid",
				SignatureRequest:  "email",
				SignatureReminder: "email",
			},
		}

		// Act
		err := dto.Validate()

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid document_signed event type")
	})
}

func TestSignatoryCreateRequestDTO_ToEntity(t *testing.T) {
	t.Run("should convert DTO to entity correctly", func(t *testing.T) {
		// Arrange
		birthday := "1990-01-01"
		phoneNumber := "+5511999999999"
		hasDoc := true
		refusable := false
		group := 2

		dto := SignatoryCreateRequestDTO{
			Name:             "John Doe",
			Email:            "john.doe@example.com",
			EnvelopeID:       1,
			Birthday:         &birthday,
			PhoneNumber:      &phoneNumber,
			HasDocumentation: &hasDoc,
			Refusable:        &refusable,
			Group:            &group,
			CommunicateEvents: &SignatoryCommunicateEventsDTO{
				DocumentSigned:    "email",
				SignatureRequest:  "sms",
				SignatureReminder: "none",
			},
		}

		// Act
		entity := dto.ToEntity()

		// Assert
		assert.Equal(t, "John Doe", entity.Name)
		assert.Equal(t, "john.doe@example.com", entity.Email)
		assert.Equal(t, 1, entity.EnvelopeID)
		assert.Equal(t, "1990-01-01", *entity.Birthday)
		assert.Equal(t, "+5511999999999", *entity.PhoneNumber)
		assert.Equal(t, true, *entity.HasDocumentation)
		assert.Equal(t, false, *entity.Refusable)
		assert.Equal(t, 2, *entity.Group)
		assert.NotNil(t, entity.CommunicateEvents)
		assert.Equal(t, "email", entity.CommunicateEvents.DocumentSigned)
		assert.Equal(t, "sms", entity.CommunicateEvents.SignatureRequest)
		assert.Equal(t, "none", entity.CommunicateEvents.SignatureReminder)
	})

	t.Run("should handle nil communicate events", func(t *testing.T) {
		// Arrange
		dto := SignatoryCreateRequestDTO{
			Name:              "John Doe",
			Email:             "john.doe@example.com",
			EnvelopeID:        1,
			CommunicateEvents: nil,
		}

		// Act
		entity := dto.ToEntity()

		// Assert
		assert.Equal(t, "John Doe", entity.Name)
		assert.Equal(t, "john.doe@example.com", entity.Email)
		assert.Equal(t, 1, entity.EnvelopeID)
		assert.Nil(t, entity.CommunicateEvents)
	})
}

func TestSignatoryUpdateRequestDTO_Validate(t *testing.T) {
	t.Run("should validate correct update DTO", func(t *testing.T) {
		// Arrange
		name := "John Smith"
		email := "john.smith@example.com"
		birthday := "1990-01-01"
		phoneNumber := "+5511999999999"
		group := 3

		dto := SignatoryUpdateRequestDTO{
			Name:        &name,
			Email:       &email,
			Birthday:    &birthday,
			PhoneNumber: &phoneNumber,
			Group:       &group,
			CommunicateEvents: &SignatoryCommunicateEventsDTO{
				DocumentSigned:    "email",
				SignatureRequest:  "email",
				SignatureReminder: "email",
			},
		}

		// Act
		err := dto.Validate()

		// Assert
		assert.NoError(t, err)
	})

	t.Run("should fail with invalid email", func(t *testing.T) {
		// Arrange
		invalidEmail := "invalid-email"
		dto := SignatoryUpdateRequestDTO{
			Email: &invalidEmail,
		}

		// Act
		err := dto.Validate()

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid email format")
	})
}

func TestSignatoryUpdateRequestDTO_ApplyToEntity(t *testing.T) {
	t.Run("should apply updates to entity correctly", func(t *testing.T) {
		// Arrange
		originalSignatory := &entity.EntitySignatory{
			ID:         1,
			Name:       "John Doe",
			Email:      "john.doe@example.com",
			EnvelopeID: 1,
			CreatedAt:  time.Now().Add(-time.Hour),
			UpdatedAt:  time.Now().Add(-time.Hour),
		}

		newName := "John Smith"
		newEmail := "john.smith@example.com"
		newEnvelopeID := 2
		birthday := "1990-01-01"
		group := 5

		dto := SignatoryUpdateRequestDTO{
			Name:       &newName,
			Email:      &newEmail,
			EnvelopeID: &newEnvelopeID,
			Birthday:   &birthday,
			Group:      &group,
			CommunicateEvents: &SignatoryCommunicateEventsDTO{
				DocumentSigned:    "sms",
				SignatureRequest:  "email",
				SignatureReminder: "none",
			},
		}

		// Act
		dto.ApplyToEntity(originalSignatory)

		// Assert
		assert.Equal(t, "John Smith", originalSignatory.Name)
		assert.Equal(t, "john.smith@example.com", originalSignatory.Email)
		assert.Equal(t, 2, originalSignatory.EnvelopeID)
		assert.Equal(t, "1990-01-01", *originalSignatory.Birthday)
		assert.Equal(t, 5, *originalSignatory.Group)
		assert.NotNil(t, originalSignatory.CommunicateEvents)
		assert.Equal(t, "sms", originalSignatory.CommunicateEvents.DocumentSigned)
		assert.Equal(t, "email", originalSignatory.CommunicateEvents.SignatureRequest)
		assert.Equal(t, "none", originalSignatory.CommunicateEvents.SignatureReminder)
		// UpdatedAt should be updated
		assert.True(t, originalSignatory.UpdatedAt.After(originalSignatory.CreatedAt))
	})

	t.Run("should not modify entity when fields are nil", func(t *testing.T) {
		// Arrange
		originalSignatory := &entity.EntitySignatory{
			ID:         1,
			Name:       "John Doe",
			Email:      "john.doe@example.com",
			EnvelopeID: 1,
			CreatedAt:  time.Now().Add(-time.Hour),
			UpdatedAt:  time.Now().Add(-time.Hour),
		}

		originalName := originalSignatory.Name
		originalEmail := originalSignatory.Email
		originalEnvelopeID := originalSignatory.EnvelopeID

		dto := SignatoryUpdateRequestDTO{
			// All fields are nil
		}

		// Act
		dto.ApplyToEntity(originalSignatory)

		// Assert
		assert.Equal(t, originalName, originalSignatory.Name)
		assert.Equal(t, originalEmail, originalSignatory.Email)
		assert.Equal(t, originalEnvelopeID, originalSignatory.EnvelopeID)
		// UpdatedAt should still be updated
		assert.True(t, originalSignatory.UpdatedAt.After(originalSignatory.CreatedAt))
	})
}

func TestSignatoryResponseDTO_FromEntity(t *testing.T) {
	t.Run("should convert entity to response DTO correctly", func(t *testing.T) {
		// Arrange
		birthday := "1990-01-01"
		phoneNumber := "+5511999999999"
		hasDoc := true
		refusable := false
		group := 2
		createdAt := time.Now().Add(-time.Hour)
		updatedAt := time.Now()

		signatory := &entity.EntitySignatory{
			ID:               1,
			Name:             "John Doe",
			Email:            "john.doe@example.com",
			EnvelopeID:       1,
			Birthday:         &birthday,
			PhoneNumber:      &phoneNumber,
			HasDocumentation: &hasDoc,
			Refusable:        &refusable,
			Group:            &group,
			CommunicateEvents: &entity.CommunicateEvents{
				DocumentSigned:    "email",
				SignatureRequest:  "sms",
				SignatureReminder: "none",
			},
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		}

		dto := &SignatoryResponseDTO{}

		// Act
		dto.FromEntity(signatory)

		// Assert
		assert.Equal(t, 1, dto.ID)
		assert.Equal(t, "John Doe", dto.Name)
		assert.Equal(t, "john.doe@example.com", dto.Email)
		assert.Equal(t, 1, dto.EnvelopeID)
		assert.Equal(t, "1990-01-01", *dto.Birthday)
		assert.Equal(t, "+5511999999999", *dto.PhoneNumber)
		assert.Equal(t, true, *dto.HasDocumentation)
		assert.Equal(t, false, *dto.Refusable)
		assert.Equal(t, 2, *dto.Group)
		assert.NotNil(t, dto.CommunicateEvents)
		assert.Equal(t, "email", dto.CommunicateEvents.DocumentSigned)
		assert.Equal(t, "sms", dto.CommunicateEvents.SignatureRequest)
		assert.Equal(t, "none", dto.CommunicateEvents.SignatureReminder)
		assert.Equal(t, createdAt, dto.CreatedAt)
		assert.Equal(t, updatedAt, dto.UpdatedAt)
	})

	t.Run("should handle nil communicate events", func(t *testing.T) {
		// Arrange
		signatory := &entity.EntitySignatory{
			ID:                1,
			Name:              "John Doe",
			Email:             "john.doe@example.com",
			EnvelopeID:        1,
			CommunicateEvents: nil,
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		}

		dto := &SignatoryResponseDTO{}

		// Act
		dto.FromEntity(signatory)

		// Assert
		assert.Equal(t, 1, dto.ID)
		assert.Equal(t, "John Doe", dto.Name)
		assert.Equal(t, "john.doe@example.com", dto.Email)
		assert.Equal(t, 1, dto.EnvelopeID)
		assert.Nil(t, dto.CommunicateEvents)
	})
}

func TestSignatoryFiltersDTO_ToEntityFilters(t *testing.T) {
	t.Run("should convert filters DTO to entity filters correctly", func(t *testing.T) {
		// Arrange
		dto := SignatoryFiltersDTO{
			IDs:        []uint{1, 2, 3},
			EnvelopeID: 5,
			Email:      "test@example.com",
			Name:       "John",
		}

		// Act
		entityFilters := dto.ToEntityFilters()

		// Assert
		assert.Equal(t, []uint{1, 2, 3}, entityFilters.IDs)
		assert.Equal(t, 5, entityFilters.EnvelopeID)
		assert.Equal(t, "test@example.com", entityFilters.Email)
		assert.Equal(t, "John", entityFilters.Name)
	})

	t.Run("should handle empty filters", func(t *testing.T) {
		// Arrange
		dto := SignatoryFiltersDTO{}

		// Act
		entityFilters := dto.ToEntityFilters()

		// Assert
		assert.Empty(t, entityFilters.IDs)
		assert.Equal(t, 0, entityFilters.EnvelopeID)
		assert.Equal(t, "", entityFilters.Email)
		assert.Equal(t, "", entityFilters.Name)
	})
}

func TestIsValidEventType(t *testing.T) {
	t.Run("should validate correct event types", func(t *testing.T) {
		validTypes := []string{"email", "sms", "none"}

		assert.True(t, isValidEventType("email", validTypes))
		assert.True(t, isValidEventType("sms", validTypes))
		assert.True(t, isValidEventType("none", validTypes))
	})

	t.Run("should reject invalid event types", func(t *testing.T) {
		validTypes := []string{"email", "sms", "none"}

		assert.False(t, isValidEventType("invalid", validTypes))
		assert.False(t, isValidEventType("", validTypes))
		assert.False(t, isValidEventType("push", validTypes))
	})
}