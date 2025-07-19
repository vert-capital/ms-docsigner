package entity

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSignatory(t *testing.T) {
	t.Run("should create signatory with valid data", func(t *testing.T) {
		// Arrange
		birthday := "1990-01-01"
		phoneNumber := "+5511999999999"
		hasDoc := true
		refusable := false
		group := 2

		validSignatory := EntitySignatory{
			Name:             "John Doe",
			Email:            "john.doe@example.com",
			EnvelopeID:       1,
			Birthday:         &birthday,
			PhoneNumber:      &phoneNumber,
			HasDocumentation: &hasDoc,
			Refusable:        &refusable,
			Group:            &group,
		}

		// Act
		signatory, err := NewSignatory(validSignatory)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, signatory)
		assert.Equal(t, "John Doe", signatory.Name)
		assert.Equal(t, "john.doe@example.com", signatory.Email)
		assert.Equal(t, 1, signatory.EnvelopeID)
		assert.Equal(t, "1990-01-01", *signatory.Birthday)
		assert.Equal(t, "+5511999999999", *signatory.PhoneNumber)
		assert.Equal(t, true, *signatory.HasDocumentation)
		assert.Equal(t, false, *signatory.Refusable)
		assert.Equal(t, 2, *signatory.Group)
		assert.NotNil(t, signatory.CommunicateEvents)
		assert.Equal(t, "email", signatory.CommunicateEvents.DocumentSigned)
		assert.NotZero(t, signatory.CreatedAt)
		assert.NotZero(t, signatory.UpdatedAt)
	})

	t.Run("should create signatory with default values", func(t *testing.T) {
		// Arrange
		validSignatory := EntitySignatory{
			Name:       "Jane Doe",
			Email:      "jane.doe@example.com",
			EnvelopeID: 1,
		}

		// Act
		signatory, err := NewSignatory(validSignatory)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, signatory)
		assert.False(t, *signatory.HasDocumentation)
		assert.True(t, *signatory.Refusable)
		assert.Equal(t, 1, *signatory.Group)
		assert.NotNil(t, signatory.CommunicateEvents)
		assert.Equal(t, "email", signatory.CommunicateEvents.DocumentSigned)
		assert.Equal(t, "email", signatory.CommunicateEvents.SignatureRequest)
		assert.Equal(t, "email", signatory.CommunicateEvents.SignatureReminder)
	})

	t.Run("should fail with invalid name", func(t *testing.T) {
		// Arrange
		invalidSignatory := EntitySignatory{
			Name:       "J", // Too short
			Email:      "john.doe@example.com",
			EnvelopeID: 1,
		}

		// Act
		signatory, err := NewSignatory(invalidSignatory)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, signatory)
	})

	t.Run("should fail with invalid email", func(t *testing.T) {
		// Arrange
		invalidSignatory := EntitySignatory{
			Name:       "John Doe",
			Email:      "invalid-email", // Invalid email format
			EnvelopeID: 1,
		}

		// Act
		signatory, err := NewSignatory(invalidSignatory)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, signatory)
		// Check for either validation tag error or custom email format error
		assert.True(t, 
			strings.Contains(err.Error(), "invalid email format") ||
			strings.Contains(err.Error(), "Field validation for 'Email' failed on the 'email' tag"))
	})

	t.Run("should fail with invalid birthday format", func(t *testing.T) {
		// Arrange
		invalidBirthday := "01/01/1990" // Wrong format
		invalidSignatory := EntitySignatory{
			Name:       "John Doe",
			Email:      "john.doe@example.com",
			EnvelopeID: 1,
			Birthday:   &invalidBirthday,
		}

		// Act
		signatory, err := NewSignatory(invalidSignatory)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, signatory)
		assert.Contains(t, err.Error(), "birthday must be in YYYY-MM-DD format")
	})

	t.Run("should fail with invalid phone number", func(t *testing.T) {
		// Arrange
		invalidPhone := "11999999999" // Missing + prefix
		invalidSignatory := EntitySignatory{
			Name:        "John Doe",
			Email:       "john.doe@example.com",
			EnvelopeID:  1,
			PhoneNumber: &invalidPhone,
		}

		// Act
		signatory, err := NewSignatory(invalidSignatory)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, signatory)
		assert.Contains(t, err.Error(), "phone number must be in international format")
	})

	t.Run("should fail with invalid group", func(t *testing.T) {
		// Arrange
		invalidGroup := 0 // Zero is not valid
		invalidSignatory := EntitySignatory{
			Name:       "John Doe",
			Email:      "john.doe@example.com",
			EnvelopeID: 1,
			Group:      &invalidGroup,
		}

		// Act
		signatory, err := NewSignatory(invalidSignatory)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, signatory)
		assert.Contains(t, err.Error(), "group must be a positive integer")
	})
}

func TestSignatoryValidateEmail(t *testing.T) {
	t.Run("should validate correct email", func(t *testing.T) {
		// Arrange
		signatory := &EntitySignatory{
			Email: "test@example.com",
		}

		// Act
		err := signatory.validateEmail()

		// Assert
		assert.NoError(t, err)
	})

	t.Run("should fail with invalid email format", func(t *testing.T) {
		// Arrange
		signatory := &EntitySignatory{
			Email: "invalid-email",
		}

		// Act
		err := signatory.validateEmail()

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid email format")
	})
}

func TestSignatoryValidateBirthday(t *testing.T) {
	t.Run("should validate correct birthday", func(t *testing.T) {
		// Arrange
		birthday := "1990-12-25"
		signatory := &EntitySignatory{
			Birthday: &birthday,
		}

		// Act
		err := signatory.validateBirthday()

		// Assert
		assert.NoError(t, err)
	})

	t.Run("should accept nil birthday", func(t *testing.T) {
		// Arrange
		signatory := &EntitySignatory{
			Birthday: nil,
		}

		// Act
		err := signatory.validateBirthday()

		// Assert
		assert.NoError(t, err)
	})

	t.Run("should fail with invalid birthday format", func(t *testing.T) {
		// Arrange
		invalidBirthday := "25/12/1990"
		signatory := &EntitySignatory{
			Birthday: &invalidBirthday,
		}

		// Act
		err := signatory.validateBirthday()

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "birthday must be in YYYY-MM-DD format")
	})

	t.Run("should fail with invalid date", func(t *testing.T) {
		// Arrange
		invalidDate := "1990-13-32" // Invalid month and day
		signatory := &EntitySignatory{
			Birthday: &invalidDate,
		}

		// Act
		err := signatory.validateBirthday()

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid birthday date")
	})
}

func TestSignatoryValidatePhoneNumber(t *testing.T) {
	t.Run("should validate correct phone number", func(t *testing.T) {
		// Arrange
		phoneNumber := "+5511999999999"
		signatory := &EntitySignatory{
			PhoneNumber: &phoneNumber,
		}

		// Act
		err := signatory.validatePhoneNumber()

		// Assert
		assert.NoError(t, err)
	})

	t.Run("should accept nil phone number", func(t *testing.T) {
		// Arrange
		signatory := &EntitySignatory{
			PhoneNumber: nil,
		}

		// Act
		err := signatory.validatePhoneNumber()

		// Assert
		assert.NoError(t, err)
	})

	t.Run("should fail without + prefix", func(t *testing.T) {
		// Arrange
		invalidPhone := "5511999999999"
		signatory := &EntitySignatory{
			PhoneNumber: &invalidPhone,
		}

		// Act
		err := signatory.validatePhoneNumber()

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "phone number must be in international format")
	})

	t.Run("should fail with too short number", func(t *testing.T) {
		// Arrange
		shortPhone := "+1234567"
		signatory := &EntitySignatory{
			PhoneNumber: &shortPhone,
		}

		// Act
		err := signatory.validatePhoneNumber()

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "phone number must be in international format")
	})
}

func TestSignatoryValidateGroup(t *testing.T) {
	t.Run("should validate positive group", func(t *testing.T) {
		// Arrange
		group := 5
		signatory := &EntitySignatory{
			Group: &group,
		}

		// Act
		err := signatory.validateGroup()

		// Assert
		assert.NoError(t, err)
	})

	t.Run("should accept nil group", func(t *testing.T) {
		// Arrange
		signatory := &EntitySignatory{
			Group: nil,
		}

		// Act
		err := signatory.validateGroup()

		// Assert
		assert.NoError(t, err)
	})

	t.Run("should fail with zero group", func(t *testing.T) {
		// Arrange
		group := 0
		signatory := &EntitySignatory{
			Group: &group,
		}

		// Act
		err := signatory.validateGroup()

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "group must be a positive integer")
	})

	t.Run("should fail with negative group", func(t *testing.T) {
		// Arrange
		group := -1
		signatory := &EntitySignatory{
			Group: &group,
		}

		// Act
		err := signatory.validateGroup()

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "group must be a positive integer")
	})
}

func TestSignatoryValidateCommunicateEvents(t *testing.T) {
	t.Run("should validate correct communicate events", func(t *testing.T) {
		// Arrange
		signatory := &EntitySignatory{
			CommunicateEvents: &CommunicateEvents{
				DocumentSigned:    "email",
				SignatureRequest:  "sms",
				SignatureReminder: "none",
			},
		}

		// Act
		err := signatory.validateCommunicateEvents()

		// Assert
		assert.NoError(t, err)
	})

	t.Run("should accept nil communicate events", func(t *testing.T) {
		// Arrange
		signatory := &EntitySignatory{
			CommunicateEvents: nil,
		}

		// Act
		err := signatory.validateCommunicateEvents()

		// Assert
		assert.NoError(t, err)
	})

	t.Run("should fail with invalid document_signed event", func(t *testing.T) {
		// Arrange
		signatory := &EntitySignatory{
			CommunicateEvents: &CommunicateEvents{
				DocumentSigned:    "invalid",
				SignatureRequest:  "email",
				SignatureReminder: "email",
			},
		}

		// Act
		err := signatory.validateCommunicateEvents()

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid document_signed event type")
	})
}

func TestSignatorySetCommunicateEvents(t *testing.T) {
	t.Run("should set valid communicate events", func(t *testing.T) {
		// Arrange
		signatory := &EntitySignatory{
			Name:       "John Doe",
			Email:      "john@example.com",
			EnvelopeID: 1,
		}
		events := CommunicateEvents{
			DocumentSigned:    "email",
			SignatureRequest:  "sms",
			SignatureReminder: "none",
		}

		// Act
		err := signatory.SetCommunicateEvents(events)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "email", signatory.CommunicateEvents.DocumentSigned)
		assert.Equal(t, "sms", signatory.CommunicateEvents.SignatureRequest)
		assert.Equal(t, "none", signatory.CommunicateEvents.SignatureReminder)
	})

	t.Run("should fail with invalid events", func(t *testing.T) {
		// Arrange
		signatory := &EntitySignatory{}
		events := CommunicateEvents{
			DocumentSigned:    "invalid",
			SignatureRequest:  "email",
			SignatureReminder: "email",
		}

		// Act
		err := signatory.SetCommunicateEvents(events)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid document_signed event type")
	})
}

func TestSignatorySetGroup(t *testing.T) {
	t.Run("should set valid group", func(t *testing.T) {
		// Arrange
		signatory := &EntitySignatory{}

		// Act
		err := signatory.SetGroup(5)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, 5, *signatory.Group)
	})

	t.Run("should fail with invalid group", func(t *testing.T) {
		// Arrange
		signatory := &EntitySignatory{}

		// Act
		err := signatory.SetGroup(0)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "group must be a positive integer")
	})
}

func TestSignatorySetBirthday(t *testing.T) {
	t.Run("should set valid birthday", func(t *testing.T) {
		// Arrange
		signatory := &EntitySignatory{}

		// Act
		err := signatory.SetBirthday("1990-01-01")

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "1990-01-01", *signatory.Birthday)
	})

	t.Run("should fail with invalid birthday", func(t *testing.T) {
		// Arrange
		signatory := &EntitySignatory{}

		// Act
		err := signatory.SetBirthday("01/01/1990")

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "birthday must be in YYYY-MM-DD format")
	})
}

func TestSignatorySetPhoneNumber(t *testing.T) {
	t.Run("should set valid phone number", func(t *testing.T) {
		// Arrange
		signatory := &EntitySignatory{}

		// Act
		err := signatory.SetPhoneNumber("+5511999999999")

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "+5511999999999", *signatory.PhoneNumber)
	})

	t.Run("should fail with invalid phone number", func(t *testing.T) {
		// Arrange
		signatory := &EntitySignatory{}

		// Act
		err := signatory.SetPhoneNumber("11999999999")

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "phone number must be in international format")
	})
}

func TestSignatoryJSON(t *testing.T) {
	t.Run("should convert to and from JSON", func(t *testing.T) {
		// Arrange
		birthday := "1990-01-01"
		phone := "+5511999999999"
		hasDoc := true
		refusable := false
		group := 2

		original := &EntitySignatory{
			Name:             "John Doe",
			Email:            "john@example.com",
			EnvelopeID:       1,
			Birthday:         &birthday,
			PhoneNumber:      &phone,
			HasDocumentation: &hasDoc,
			Refusable:        &refusable,
			Group:            &group,
			CommunicateEvents: &CommunicateEvents{
				DocumentSigned:    "email",
				SignatureRequest:  "sms",
				SignatureReminder: "none",
			},
		}

		// Act - Convert to JSON
		jsonData, err := original.ToJSON()
		assert.NoError(t, err)

		// Act - Convert from JSON
		restored := &EntitySignatory{}
		err = restored.FromJSON(jsonData)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, original.Name, restored.Name)
		assert.Equal(t, original.Email, restored.Email)
		assert.Equal(t, original.EnvelopeID, restored.EnvelopeID)
		assert.Equal(t, *original.Birthday, *restored.Birthday)
		assert.Equal(t, *original.PhoneNumber, *restored.PhoneNumber)
		assert.Equal(t, *original.HasDocumentation, *restored.HasDocumentation)
		assert.Equal(t, *original.Refusable, *restored.Refusable)
		assert.Equal(t, *original.Group, *restored.Group)
		assert.Equal(t, original.CommunicateEvents.DocumentSigned, restored.CommunicateEvents.DocumentSigned)
	})
}

func TestSignatoryTableName(t *testing.T) {
	t.Run("should return correct table name", func(t *testing.T) {
		// Arrange
		signatory := EntitySignatory{}

		// Act
		tableName := signatory.TableName()

		// Assert
		assert.Equal(t, "signatories", tableName)
	})
}