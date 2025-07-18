package envelope

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"app/entity"
	"app/mocks"
	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestUsecaseEnvelopeService_CreateEnvelope(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockIRepositoryEnvelope(ctrl)
	mockClicksignClient := mocks.NewMockClicksignClientInterface(ctrl)
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Reduce log noise in tests

	service := NewUsecaseEnvelopeService(mockRepo, mockClicksignClient, logger)

	t.Run("should create envelope successfully", func(t *testing.T) {
		// Arrange
		envelope := &entity.EntityEnvelope{
			ID:               1,
			Name:             "Test Envelope",
			Description:      "Test description",
			Status:           "draft",
			DocumentsIDs:     []int{1, 2},
			SignatoryEmails:  []string{"test@example.com"},
			Message:          "Please sign",
			RemindInterval:   3,
			AutoClose:        true,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}

		// Mock expectations
		mockRepo.EXPECT().
			Create(envelope).
			Return(nil)

		mockClicksignClient.EXPECT().
			Post(gomock.Any(), "/api/v3/envelopes", gomock.Any()).
			Return(mockSuccessResponse(), nil)

		mockRepo.EXPECT().
			Update(envelope).
			Return(nil)

		// Act
		result, err := service.CreateEnvelope(envelope)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, envelope.Name, result.Name)
		assert.NotEmpty(t, result.ClicksignKey)
	})

	t.Run("should fail validation with invalid envelope", func(t *testing.T) {
		// Arrange
		envelope := &entity.EntityEnvelope{
			Name:            "Te", // Too short
			DocumentsIDs:    []int{1},
			SignatoryEmails: []string{"test@example.com"},
		}

		// Act
		result, err := service.CreateEnvelope(envelope)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "validation failed")
	})

	t.Run("should fail business rules with no documents", func(t *testing.T) {
		// Arrange
		envelope := &entity.EntityEnvelope{
			Name:            "Test Envelope",
			DocumentsIDs:    []int{1}, // Valid for entity but will fail business rules
			SignatoryEmails: []string{"test@example.com"},
			Status:          "sent", // Invalid status for business rules
			RemindInterval:  3,
		}

		// Act
		result, err := service.CreateEnvelope(envelope)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "business rule validation failed")
	})

	t.Run("should fail business rules with duplicate emails", func(t *testing.T) {
		// Arrange
		envelope := &entity.EntityEnvelope{
			Name:            "Test Envelope",
			DocumentsIDs:    []int{1},
			SignatoryEmails: []string{"test@example.com", "test@example.com"}, // Duplicate
			Status:          "draft",
			RemindInterval:  3,
		}

		// Act
		result, err := service.CreateEnvelope(envelope)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "duplicate signatory email")
	})

	t.Run("should fail when repository create fails", func(t *testing.T) {
		// Arrange
		envelope := &entity.EntityEnvelope{
			Name:            "Test Envelope",
			DocumentsIDs:    []int{1},
			SignatoryEmails: []string{"test@example.com"},
			Status:          "draft",
			RemindInterval:  3,
		}

		mockRepo.EXPECT().
			Create(envelope).
			Return(errors.New("database error"))

		// Act
		result, err := service.CreateEnvelope(envelope)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to create envelope locally")
	})

	t.Run("should rollback when clicksign fails", func(t *testing.T) {
		// Arrange
		envelope := &entity.EntityEnvelope{
			Name:            "Test Envelope",
			DocumentsIDs:    []int{1},
			SignatoryEmails: []string{"test@example.com"},
			Status:          "draft",
			RemindInterval:  3,
		}

		mockRepo.EXPECT().
			Create(envelope).
			Return(nil)

		mockClicksignClient.EXPECT().
			Post(gomock.Any(), "/api/v3/envelopes", gomock.Any()).
			Return(nil, errors.New("clicksign error"))

		mockRepo.EXPECT().
			Delete(envelope).
			Return(nil)

		// Act
		result, err := service.CreateEnvelope(envelope)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to create envelope in Clicksign")
	})
}

func TestUsecaseEnvelopeService_GetEnvelope(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockIRepositoryEnvelope(ctrl)
	mockClicksignClient := mocks.NewMockClicksignClientInterface(ctrl)
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	service := NewUsecaseEnvelopeService(mockRepo, mockClicksignClient, logger)

	t.Run("should get envelope successfully", func(t *testing.T) {
		// Arrange
		envelope := &entity.EntityEnvelope{
			ID:   1,
			Name: "Test Envelope",
		}

		mockRepo.EXPECT().
			GetByID(1).
			Return(envelope, nil)

		// Act
		result, err := service.GetEnvelope(1)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, envelope.ID, result.ID)
		assert.Equal(t, envelope.Name, result.Name)
	})

	t.Run("should fail when envelope not found", func(t *testing.T) {
		// Arrange
		mockRepo.EXPECT().
			GetByID(999).
			Return(nil, errors.New("not found"))

		// Act
		result, err := service.GetEnvelope(999)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "envelope not found")
	})
}

func TestUsecaseEnvelopeService_GetEnvelopes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockIRepositoryEnvelope(ctrl)
	mockClicksignClient := mocks.NewMockClicksignClientInterface(ctrl)
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	service := NewUsecaseEnvelopeService(mockRepo, mockClicksignClient, logger)

	t.Run("should get envelopes successfully", func(t *testing.T) {
		// Arrange
		envelopes := []entity.EntityEnvelope{
			{ID: 1, Name: "Envelope 1"},
			{ID: 2, Name: "Envelope 2"},
		}

		filters := entity.EntityEnvelopeFilters{
			Status: "draft",
		}

		mockRepo.EXPECT().
			GetEnvelopes(filters).
			Return(envelopes, nil)

		// Act
		result, err := service.GetEnvelopes(filters)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result, 2)
		assert.Equal(t, envelopes[0].ID, result[0].ID)
		assert.Equal(t, envelopes[1].ID, result[1].ID)
	})

	t.Run("should fail when repository fails", func(t *testing.T) {
		// Arrange
		filters := entity.EntityEnvelopeFilters{}

		mockRepo.EXPECT().
			GetEnvelopes(filters).
			Return(nil, errors.New("database error"))

		// Act
		result, err := service.GetEnvelopes(filters)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to get envelopes")
	})
}

func TestUsecaseEnvelopeService_UpdateEnvelope(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockIRepositoryEnvelope(ctrl)
	mockClicksignClient := mocks.NewMockClicksignClientInterface(ctrl)
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	service := NewUsecaseEnvelopeService(mockRepo, mockClicksignClient, logger)

	t.Run("should update envelope successfully", func(t *testing.T) {
		// Arrange
		envelope := &entity.EntityEnvelope{
			ID:              1,
			Name:            "Updated Envelope",
			DocumentsIDs:    []int{1},
			SignatoryEmails: []string{"test@example.com"},
			Status:          "draft",
			RemindInterval:  3,
		}

		mockRepo.EXPECT().
			GetByID(1).
			Return(envelope, nil)

		mockRepo.EXPECT().
			Update(envelope).
			Return(nil)

		// Act
		err := service.UpdateEnvelope(envelope)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("should fail when envelope not found", func(t *testing.T) {
		// Arrange
		envelope := &entity.EntityEnvelope{
			ID:              999,
			Name:            "Updated Envelope",
			DocumentsIDs:    []int{1},
			SignatoryEmails: []string{"test@example.com"},
			Status:          "draft",
			RemindInterval:  3,
		}

		mockRepo.EXPECT().
			GetByID(999).
			Return(nil, errors.New("not found"))

		// Act
		err := service.UpdateEnvelope(envelope)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "envelope not found")
	})

	t.Run("should fail when envelope is completed", func(t *testing.T) {
		// Arrange
		envelope := &entity.EntityEnvelope{
			ID:              1,
			Name:            "Updated Envelope",
			DocumentsIDs:    []int{1},
			SignatoryEmails: []string{"test@example.com"},
			Status:          "completed",
			RemindInterval:  3,
		}

		existingEnvelope := &entity.EntityEnvelope{
			ID:             1,
			Status:         "completed",
			RemindInterval: 3,
		}

		mockRepo.EXPECT().
			GetByID(1).
			Return(existingEnvelope, nil)

		// Act
		err := service.UpdateEnvelope(envelope)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot update envelope in 'completed' status")
	})
}

func TestUsecaseEnvelopeService_DeleteEnvelope(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockIRepositoryEnvelope(ctrl)
	mockClicksignClient := mocks.NewMockClicksignClientInterface(ctrl)
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	service := NewUsecaseEnvelopeService(mockRepo, mockClicksignClient, logger)

	t.Run("should delete envelope successfully", func(t *testing.T) {
		// Arrange
		envelope := &entity.EntityEnvelope{
			ID:     1,
			Name:   "Test Envelope",
			Status: "draft",
		}

		mockRepo.EXPECT().
			GetByID(1).
			Return(envelope, nil)

		mockRepo.EXPECT().
			Delete(envelope).
			Return(nil)

		// Act
		err := service.DeleteEnvelope(1)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("should fail when envelope not found", func(t *testing.T) {
		// Arrange
		mockRepo.EXPECT().
			GetByID(999).
			Return(nil, errors.New("not found"))

		// Act
		err := service.DeleteEnvelope(999)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "envelope not found")
	})

	t.Run("should fail when envelope is sent", func(t *testing.T) {
		// Arrange
		envelope := &entity.EntityEnvelope{
			ID:     1,
			Status: "sent",
		}

		mockRepo.EXPECT().
			GetByID(1).
			Return(envelope, nil)

		// Act
		err := service.DeleteEnvelope(1)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot delete envelope in 'sent' status")
	})
}

func TestUsecaseEnvelopeService_ActivateEnvelope(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockIRepositoryEnvelope(ctrl)
	mockClicksignClient := mocks.NewMockClicksignClientInterface(ctrl)
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	service := NewUsecaseEnvelopeService(mockRepo, mockClicksignClient, logger)

	t.Run("should activate envelope successfully", func(t *testing.T) {
		// Arrange
		envelope := &entity.EntityEnvelope{
			ID:           1,
			Name:         "Test Envelope",
			Status:       "draft",
			ClicksignKey: "test-key-123",
		}

		mockRepo.EXPECT().
			GetByID(1).
			Return(envelope, nil)

		mockClicksignClient.EXPECT().
			Patch(gomock.Any(), "/api/v3/envelopes/test-key-123", gomock.Any()).
			Return(mockSuccessResponsePatch(), nil)

		mockRepo.EXPECT().
			Update(envelope).
			Return(nil)

		// Act
		result, err := service.ActivateEnvelope(1)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "sent", result.Status)
	})

	t.Run("should fail when envelope not found", func(t *testing.T) {
		// Arrange
		mockRepo.EXPECT().
			GetByID(999).
			Return(nil, errors.New("not found"))

		// Act
		result, err := service.ActivateEnvelope(999)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "envelope not found")
	})

	t.Run("should fail when envelope has no clicksign key", func(t *testing.T) {
		// Arrange
		envelope := &entity.EntityEnvelope{
			ID:           1,
			Status:       "draft",
			ClicksignKey: "", // Empty
		}

		mockRepo.EXPECT().
			GetByID(1).
			Return(envelope, nil)

		// Act
		result, err := service.ActivateEnvelope(1)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "envelope has no Clicksign key")
	})
}

func TestUsecaseEnvelopeService_ValidateBusinessRules(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockIRepositoryEnvelope(ctrl)
	mockClicksignClient := mocks.NewMockClicksignClientInterface(ctrl)
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	service := NewUsecaseEnvelopeService(mockRepo, mockClicksignClient, logger).(*UsecaseEnvelopeService)

	t.Run("should pass validation with valid envelope", func(t *testing.T) {
		// Arrange
		envelope := &entity.EntityEnvelope{
			DocumentsIDs:    []int{1, 2},
			SignatoryEmails: []string{"test@example.com", "user@example.com"},
			Status:          "draft",
		}

		// Act
		err := service.validateBusinessRules(envelope)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("should fail with no documents", func(t *testing.T) {
		// Arrange
		envelope := &entity.EntityEnvelope{
			DocumentsIDs:    []int{},
			SignatoryEmails: []string{"test@example.com"},
		}

		// Act
		err := service.validateBusinessRules(envelope)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "envelope must have at least one document")
	})

	t.Run("should fail with no signatories", func(t *testing.T) {
		// Arrange
		envelope := &entity.EntityEnvelope{
			DocumentsIDs:    []int{1},
			SignatoryEmails: []string{},
		}

		// Act
		err := service.validateBusinessRules(envelope)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "envelope must have at least one signatory")
	})

	t.Run("should fail with too many signatories", func(t *testing.T) {
		// Arrange
		signatories := make([]string, 51)
		for i := 0; i < 51; i++ {
			signatories[i] = "user" + string(rune(i)) + "@example.com"
		}

		envelope := &entity.EntityEnvelope{
			DocumentsIDs:    []int{1},
			SignatoryEmails: signatories,
		}

		// Act
		err := service.validateBusinessRules(envelope)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "envelope cannot have more than 50 signatories")
	})

	t.Run("should fail with duplicate emails", func(t *testing.T) {
		// Arrange
		envelope := &entity.EntityEnvelope{
			DocumentsIDs:    []int{1},
			SignatoryEmails: []string{"test@example.com", "test@example.com"},
		}

		// Act
		err := service.validateBusinessRules(envelope)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "duplicate signatory email")
	})

	t.Run("should fail with non-draft status", func(t *testing.T) {
		// Arrange
		envelope := &entity.EntityEnvelope{
			DocumentsIDs:    []int{1},
			SignatoryEmails: []string{"test@example.com"},
			Status:          "sent",
		}

		// Act
		err := service.validateBusinessRules(envelope)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "new envelopes must be in 'draft' status")
	})
}

// Helper functions for tests

func mockSuccessResponse() *http.Response {
	return &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(`{"id": "test-key-123", "status": "draft"}`)),
	}
}

func mockSuccessResponsePatch() *http.Response {
	return &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(`{"id": "test-key-123", "status": "running"}`)),
	}
}