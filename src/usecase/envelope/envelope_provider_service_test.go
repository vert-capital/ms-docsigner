package usecase_envelope_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"app/entity"
	"app/mocks"
	usecase_envelope "app/usecase/envelope"

	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestUsecaseEnvelopeProviderService_CreateEnvelope(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockIRepositoryEnvelope(ctrl)
	mockProvider := mocks.NewMockEnvelopeProvider(ctrl)
	mockDocumentUsecase := mocks.NewMockIUsecaseDocument(ctrl)
	mockRequirementUsecase := mocks.NewMockIUsecaseRequirement(ctrl)
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	service := usecase_envelope.NewUsecaseEnvelopeProviderService(
		mockRepo,
		mockProvider,
		mockDocumentUsecase,
		mockRequirementUsecase,
		logger,
	)

	t.Run("should create envelope successfully", func(t *testing.T) {
		// Arrange
		envelope := &entity.EntityEnvelope{
			ID:              1,
			Name:            "Test Envelope",
			Description:     "Test description",
			Status:          "draft",
			DocumentsIDs:    []int{1, 2},
			SignatoryEmails: []string{"test@example.com"},
			Message:         "Please sign",
			RemindInterval:  3,
			AutoClose:       true,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}

		// Mock expectations
		mockRepo.EXPECT().
			Create(envelope).
			Return(nil)

		mockProvider.EXPECT().
			CreateEnvelope(gomock.Any(), envelope).
			Return("provider-key-123", "raw-data", nil)

		mockRepo.EXPECT().
			Update(envelope).
			Return(nil)

		// Act
		result, err := service.CreateEnvelope(envelope)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, envelope.Name, result.Name)
		assert.Equal(t, "provider-key-123", result.ClicksignKey)
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

	t.Run("should rollback on provider error", func(t *testing.T) {
		// Arrange
		envelope := &entity.EntityEnvelope{
			ID:              1,
			Name:            "Test Envelope",
			Description:     "Test description",
			Status:          "draft",
			DocumentsIDs:    []int{1},
			SignatoryEmails: []string{"test@example.com"},
			RemindInterval:  3,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}

		// Mock expectations
		mockRepo.EXPECT().
			Create(envelope).
			Return(nil)

		mockProvider.EXPECT().
			CreateEnvelope(gomock.Any(), envelope).
			Return("", "", errors.New("provider error"))

		mockRepo.EXPECT().
			Delete(envelope).
			Return(nil) // Best effort rollback

		// Act
		result, err := service.CreateEnvelope(envelope)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "provider")
	})
}

func TestUsecaseEnvelopeProviderService_ActivateEnvelope(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockIRepositoryEnvelope(ctrl)
	mockProvider := mocks.NewMockEnvelopeProvider(ctrl)
	mockDocumentUsecase := mocks.NewMockIUsecaseDocument(ctrl)
	mockRequirementUsecase := mocks.NewMockIUsecaseRequirement(ctrl)
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	service := usecase_envelope.NewUsecaseEnvelopeProviderService(
		mockRepo,
		mockProvider,
		mockDocumentUsecase,
		mockRequirementUsecase,
		logger,
	)

	t.Run("should activate envelope successfully", func(t *testing.T) {
		// Arrange
		envelope := &entity.EntityEnvelope{
			ID:           1,
			Name:         "Test Envelope",
			Status:       "draft",
			ClicksignKey: "provider-key-123",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		// Mock expectations
		mockRepo.EXPECT().
			GetByID(1).
			Return(envelope, nil)

		mockProvider.EXPECT().
			ActivateEnvelope(gomock.Any(), "provider-key-123").
			Return(nil)

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

	t.Run("should fail if envelope not found", func(t *testing.T) {
		// Mock expectations
		mockRepo.EXPECT().
			GetByID(999).
			Return(nil, errors.New("not found"))

		// Act
		result, err := service.ActivateEnvelope(999)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("should fail if envelope has no provider key", func(t *testing.T) {
		// Arrange
		envelope := &entity.EntityEnvelope{
			ID:        1,
			Name:      "Test Envelope",
			Status:    "draft",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// Mock expectations
		mockRepo.EXPECT().
			GetByID(1).
			Return(envelope, nil)

		// Act
		result, err := service.ActivateEnvelope(1)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "provider key")
	})
}

func TestUsecaseEnvelopeProviderService_ValidateBusinessRules(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockIRepositoryEnvelope(ctrl)
	mockProvider := mocks.NewMockEnvelopeProvider(ctrl)
	mockDocumentUsecase := mocks.NewMockIUsecaseDocument(ctrl)
	mockRequirementUsecase := mocks.NewMockIUsecaseRequirement(ctrl)
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	service := usecase_envelope.NewUsecaseEnvelopeProviderService(
		mockRepo,
		mockProvider,
		mockDocumentUsecase,
		mockRequirementUsecase,
		logger,
	)

	t.Run("should fail with too many signatories", func(t *testing.T) {
		// Arrange
		envelope := &entity.EntityEnvelope{
			Name:            "Test Envelope",
			Status:          "draft",
			SignatoryEmails: make([]string, 51), // More than 50
		}

		// Act
		err := service.ValidateBusinessRules(envelope)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "50 signatories")
	})

	t.Run("should fail with duplicate emails", func(t *testing.T) {
		// Arrange
		envelope := &entity.EntityEnvelope{
			Name:            "Test Envelope",
			Status:          "draft",
			SignatoryEmails: []string{"test@example.com", "test@example.com"},
		}

		// Act
		err := service.ValidateBusinessRules(envelope)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "duplicate")
	})

	t.Run("should pass with valid envelope", func(t *testing.T) {
		// Arrange
		envelope := &entity.EntityEnvelope{
			Name:            "Test Envelope",
			Status:          "draft",
			SignatoryEmails: []string{"test@example.com"},
		}

		// Act
		err := service.ValidateBusinessRules(envelope)

		// Assert
		assert.NoError(t, err)
	})
}

func TestUsecaseEnvelopeProviderService_CreateDocument(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockIRepositoryEnvelope(ctrl)
	mockProvider := mocks.NewMockEnvelopeProvider(ctrl)
	mockDocumentUsecase := mocks.NewMockIUsecaseDocument(ctrl)
	mockRequirementUsecase := mocks.NewMockIUsecaseRequirement(ctrl)
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	service := usecase_envelope.NewUsecaseEnvelopeProviderService(
		mockRepo,
		mockProvider,
		mockDocumentUsecase,
		mockRequirementUsecase,
		logger,
	)

	t.Run("should create document successfully", func(t *testing.T) {
		// Arrange
		document := &entity.EntityDocument{
			ID:       1,
			Name:     "Test Document",
			FilePath: "/tmp/test.pdf",
			MimeType: "application/pdf",
		}

		// Mock expectations
		mockProvider.EXPECT().
			CreateDocument(gomock.Any(), "envelope-key", document, 1).
			Return("doc-key-123", nil)

		// Act
		docKey, err := service.CreateDocument(context.Background(), "envelope-key", document, 1)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "doc-key-123", docKey)
	})
}

