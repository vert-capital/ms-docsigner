package signatory

import (
	"testing"

	"app/entity"
	"app/mocks"
	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestUsecaseSignatoryService_CreateSignatory(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSignatoryRepo := mocks.NewMockIRepositorySignatory(ctrl)
	mockEnvelopeRepo := mocks.NewMockIRepositoryEnvelope(ctrl)
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Reduce log noise in tests

	service := NewUsecaseSignatoryService(mockSignatoryRepo, mockEnvelopeRepo, logger)

	t.Run("should create signatory successfully", func(t *testing.T) {
		// Arrange
		envelope := &entity.EntityEnvelope{
			ID:     1,
			Name:   "Test Envelope",
			Status: "draft",
		}

		signatory := &entity.EntitySignatory{
			Name:       "John Doe",
			Email:      "john.doe@example.com",
			EnvelopeID: 1,
		}

		// Mock expectations
		mockEnvelopeRepo.EXPECT().
			GetByID(1).
			Return(envelope, nil)

		mockSignatoryRepo.EXPECT().
			GetByEmailAndEnvelopeID("john.doe@example.com", 1).
			Return(nil, gorm.ErrRecordNotFound)

		mockSignatoryRepo.EXPECT().
			GetByEnvelopeID(1).
			Return([]entity.EntitySignatory{}, nil)

		mockSignatoryRepo.EXPECT().
			Create(gomock.Any()).
			Return(nil)

		// Act
		result, err := service.CreateSignatory(signatory)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "John Doe", result.Name)
		assert.Equal(t, "john.doe@example.com", result.Email)
	})

	t.Run("should fail when envelope not found", func(t *testing.T) {
		// Arrange
		signatory := &entity.EntitySignatory{
			Name:       "John Doe",
			Email:      "john.doe@example.com",
			EnvelopeID: 999,
		}

		// Mock expectations
		mockEnvelopeRepo.EXPECT().
			GetByID(999).
			Return(nil, gorm.ErrRecordNotFound)

		// Act
		result, err := service.CreateSignatory(signatory)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "envelope not found")
	})

	t.Run("should fail when envelope status is completed", func(t *testing.T) {
		// Arrange
		envelope := &entity.EntityEnvelope{
			ID:     1,
			Name:   "Test Envelope",
			Status: "completed",
		}

		signatory := &entity.EntitySignatory{
			Name:       "John Doe",
			Email:      "john.doe@example.com",
			EnvelopeID: 1,
		}

		// Mock expectations
		mockEnvelopeRepo.EXPECT().
			GetByID(1).
			Return(envelope, nil)

		// Act
		result, err := service.CreateSignatory(signatory)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "cannot add signatory to envelope in 'completed' status")
	})

	t.Run("should fail when email already exists in envelope", func(t *testing.T) {
		// Arrange
		envelope := &entity.EntityEnvelope{
			ID:     1,
			Name:   "Test Envelope",
			Status: "draft",
		}

		existingSignatory := &entity.EntitySignatory{
			ID:         2,
			Name:       "Jane Doe",
			Email:      "john.doe@example.com",
			EnvelopeID: 1,
		}

		signatory := &entity.EntitySignatory{
			Name:       "John Doe",
			Email:      "john.doe@example.com",
			EnvelopeID: 1,
		}

		// Mock expectations
		mockEnvelopeRepo.EXPECT().
			GetByID(1).
			Return(envelope, nil)

		mockSignatoryRepo.EXPECT().
			GetByEmailAndEnvelopeID("john.doe@example.com", 1).
			Return(existingSignatory, nil)

		// Act
		result, err := service.CreateSignatory(signatory)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "email 'john.doe@example.com' already exists in envelope 1")
	})

	t.Run("should fail when envelope has too many signatories", func(t *testing.T) {
		// Arrange
		envelope := &entity.EntityEnvelope{
			ID:     1,
			Name:   "Test Envelope",
			Status: "draft",
		}

		// Create 50 existing signatories
		existingSignatories := make([]entity.EntitySignatory, 50)
		for i := 0; i < 50; i++ {
			existingSignatories[i] = entity.EntitySignatory{
				ID:         i + 1,
				Name:       "User",
				Email:      "user@example.com",
				EnvelopeID: 1,
			}
		}

		signatory := &entity.EntitySignatory{
			Name:       "John Doe",
			Email:      "john.doe@example.com",
			EnvelopeID: 1,
		}

		// Mock expectations
		mockEnvelopeRepo.EXPECT().
			GetByID(1).
			Return(envelope, nil)

		mockSignatoryRepo.EXPECT().
			GetByEmailAndEnvelopeID("john.doe@example.com", 1).
			Return(nil, gorm.ErrRecordNotFound)

		mockSignatoryRepo.EXPECT().
			GetByEnvelopeID(1).
			Return(existingSignatories, nil)

		// Act
		result, err := service.CreateSignatory(signatory)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "envelope cannot have more than 50 signatories")
	})
}

func TestUsecaseSignatoryService_GetSignatory(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSignatoryRepo := mocks.NewMockIRepositorySignatory(ctrl)
	mockEnvelopeRepo := mocks.NewMockIRepositoryEnvelope(ctrl)
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	service := NewUsecaseSignatoryService(mockSignatoryRepo, mockEnvelopeRepo, logger)

	t.Run("should get signatory successfully", func(t *testing.T) {
		// Arrange
		signatory := &entity.EntitySignatory{
			ID:         1,
			Name:       "John Doe",
			Email:      "john.doe@example.com",
			EnvelopeID: 1,
		}

		// Mock expectations
		mockSignatoryRepo.EXPECT().
			GetByID(1).
			Return(signatory, nil)

		// Act
		result, err := service.GetSignatory(1)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "John Doe", result.Name)
	})

	t.Run("should fail when signatory not found", func(t *testing.T) {
		// Arrange
		// Mock expectations
		mockSignatoryRepo.EXPECT().
			GetByID(999).
			Return(nil, gorm.ErrRecordNotFound)

		// Act
		result, err := service.GetSignatory(999)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "signatory not found")
	})
}

func TestUsecaseSignatoryService_GetSignatoriesByEnvelope(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSignatoryRepo := mocks.NewMockIRepositorySignatory(ctrl)
	mockEnvelopeRepo := mocks.NewMockIRepositoryEnvelope(ctrl)
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	service := NewUsecaseSignatoryService(mockSignatoryRepo, mockEnvelopeRepo, logger)

	t.Run("should get signatories by envelope successfully", func(t *testing.T) {
		// Arrange
		envelope := &entity.EntityEnvelope{
			ID:     1,
			Name:   "Test Envelope",
			Status: "draft",
		}

		signatories := []entity.EntitySignatory{
			{
				ID:         1,
				Name:       "John Doe",
				Email:      "john.doe@example.com",
				EnvelopeID: 1,
			},
			{
				ID:         2,
				Name:       "Jane Doe",
				Email:      "jane.doe@example.com",
				EnvelopeID: 1,
			},
		}

		// Mock expectations
		mockEnvelopeRepo.EXPECT().
			GetByID(1).
			Return(envelope, nil)

		mockSignatoryRepo.EXPECT().
			GetByEnvelopeID(1).
			Return(signatories, nil)

		// Act
		result, err := service.GetSignatoriesByEnvelope(1)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result, 2)
		assert.Equal(t, "John Doe", result[0].Name)
		assert.Equal(t, "Jane Doe", result[1].Name)
	})

	t.Run("should fail when envelope not found", func(t *testing.T) {
		// Arrange
		// Mock expectations
		mockEnvelopeRepo.EXPECT().
			GetByID(999).
			Return(nil, gorm.ErrRecordNotFound)

		// Act
		result, err := service.GetSignatoriesByEnvelope(999)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "envelope not found")
	})
}

func TestUsecaseSignatoryService_UpdateSignatory(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSignatoryRepo := mocks.NewMockIRepositorySignatory(ctrl)
	mockEnvelopeRepo := mocks.NewMockIRepositoryEnvelope(ctrl)
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	service := NewUsecaseSignatoryService(mockSignatoryRepo, mockEnvelopeRepo, logger)

	t.Run("should update signatory successfully", func(t *testing.T) {
		// Arrange
		envelope := &entity.EntityEnvelope{
			ID:     1,
			Name:   "Test Envelope",
			Status: "draft",
		}

		existingSignatory := &entity.EntitySignatory{
			ID:         1,
			Name:       "John Doe",
			Email:      "john.doe@example.com",
			EnvelopeID: 1,
		}

		updatedSignatory := &entity.EntitySignatory{
			ID:         1,
			Name:       "John Smith",
			Email:      "john.smith@example.com",
			EnvelopeID: 1,
		}

		// Mock expectations
		mockSignatoryRepo.EXPECT().
			GetByID(1).
			Return(existingSignatory, nil)

		mockEnvelopeRepo.EXPECT().
			GetByID(1).
			Return(envelope, nil)

		mockSignatoryRepo.EXPECT().
			Update(gomock.Any()).
			Return(nil)

		// Act
		err := service.UpdateSignatory(updatedSignatory)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("should fail when signatory not found", func(t *testing.T) {
		// Arrange
		signatory := &entity.EntitySignatory{
			ID:         999,
			Name:       "John Doe",
			Email:      "john.doe@example.com",
			EnvelopeID: 1,
		}

		// Mock expectations
		mockSignatoryRepo.EXPECT().
			GetByID(999).
			Return(nil, gorm.ErrRecordNotFound)

		// Act
		err := service.UpdateSignatory(signatory)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "signatory not found")
	})

	t.Run("should fail when envelope is completed", func(t *testing.T) {
		// Arrange
		envelope := &entity.EntityEnvelope{
			ID:     1,
			Name:   "Test Envelope",
			Status: "completed",
		}

		existingSignatory := &entity.EntitySignatory{
			ID:         1,
			Name:       "John Doe",
			Email:      "john.doe@example.com",
			EnvelopeID: 1,
		}

		updatedSignatory := &entity.EntitySignatory{
			ID:         1,
			Name:       "John Smith",
			Email:      "john.smith@example.com",
			EnvelopeID: 1,
		}

		// Mock expectations
		mockSignatoryRepo.EXPECT().
			GetByID(1).
			Return(existingSignatory, nil)

		mockEnvelopeRepo.EXPECT().
			GetByID(1).
			Return(envelope, nil)

		// Act
		err := service.UpdateSignatory(updatedSignatory)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot update signatory in envelope with 'completed' status")
	})
}

func TestUsecaseSignatoryService_DeleteSignatory(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSignatoryRepo := mocks.NewMockIRepositorySignatory(ctrl)
	mockEnvelopeRepo := mocks.NewMockIRepositoryEnvelope(ctrl)
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	service := NewUsecaseSignatoryService(mockSignatoryRepo, mockEnvelopeRepo, logger)

	t.Run("should delete signatory successfully", func(t *testing.T) {
		// Arrange
		envelope := &entity.EntityEnvelope{
			ID:     1,
			Name:   "Test Envelope",
			Status: "draft",
		}

		signatory := &entity.EntitySignatory{
			ID:         1,
			Name:       "John Doe",
			Email:      "john.doe@example.com",
			EnvelopeID: 1,
		}

		// Mock expectations
		mockSignatoryRepo.EXPECT().
			GetByID(1).
			Return(signatory, nil)

		mockEnvelopeRepo.EXPECT().
			GetByID(1).
			Return(envelope, nil)

		mockSignatoryRepo.EXPECT().
			Delete(signatory).
			Return(nil)

		// Act
		err := service.DeleteSignatory(1)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("should fail when signatory not found", func(t *testing.T) {
		// Arrange
		// Mock expectations
		mockSignatoryRepo.EXPECT().
			GetByID(999).
			Return(nil, gorm.ErrRecordNotFound)

		// Act
		err := service.DeleteSignatory(999)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "signatory not found")
	})

	t.Run("should fail when envelope is sent", func(t *testing.T) {
		// Arrange
		envelope := &entity.EntityEnvelope{
			ID:     1,
			Name:   "Test Envelope",
			Status: "sent",
		}

		signatory := &entity.EntitySignatory{
			ID:         1,
			Name:       "John Doe",
			Email:      "john.doe@example.com",
			EnvelopeID: 1,
		}

		// Mock expectations
		mockSignatoryRepo.EXPECT().
			GetByID(1).
			Return(signatory, nil)

		mockEnvelopeRepo.EXPECT().
			GetByID(1).
			Return(envelope, nil)

		// Act
		err := service.DeleteSignatory(1)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot delete signatory from envelope in 'sent' status")
	})
}

func TestUsecaseSignatoryService_AssociateToEnvelope(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSignatoryRepo := mocks.NewMockIRepositorySignatory(ctrl)
	mockEnvelopeRepo := mocks.NewMockIRepositoryEnvelope(ctrl)
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	service := NewUsecaseSignatoryService(mockSignatoryRepo, mockEnvelopeRepo, logger)

	t.Run("should associate signatory to envelope successfully", func(t *testing.T) {
		// Arrange
		envelope := &entity.EntityEnvelope{
			ID:     2,
			Name:   "New Envelope",
			Status: "draft",
		}

		signatory := &entity.EntitySignatory{
			ID:         1,
			Name:       "John Doe",
			Email:      "john.doe@example.com",
			EnvelopeID: 1,
		}

		// Mock expectations
		mockSignatoryRepo.EXPECT().
			GetByID(1).
			Return(signatory, nil)

		mockEnvelopeRepo.EXPECT().
			GetByID(2).
			Return(envelope, nil)

		mockSignatoryRepo.EXPECT().
			GetByEmailAndEnvelopeID("john.doe@example.com", 2).
			Return(nil, gorm.ErrRecordNotFound)

		mockSignatoryRepo.EXPECT().
			Update(gomock.Any()).
			Return(nil)

		// Act
		err := service.AssociateToEnvelope(1, 2)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("should fail when signatory not found", func(t *testing.T) {
		// Arrange
		// Mock expectations
		mockSignatoryRepo.EXPECT().
			GetByID(999).
			Return(nil, gorm.ErrRecordNotFound)

		// Act
		err := service.AssociateToEnvelope(999, 2)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "signatory not found")
	})

	t.Run("should fail when envelope not found", func(t *testing.T) {
		// Arrange
		signatory := &entity.EntitySignatory{
			ID:         1,
			Name:       "John Doe",
			Email:      "john.doe@example.com",
			EnvelopeID: 1,
		}

		// Mock expectations
		mockSignatoryRepo.EXPECT().
			GetByID(1).
			Return(signatory, nil)

		mockEnvelopeRepo.EXPECT().
			GetByID(999).
			Return(nil, gorm.ErrRecordNotFound)

		// Act
		err := service.AssociateToEnvelope(1, 999)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "envelope not found")
	})

	t.Run("should fail when envelope is completed", func(t *testing.T) {
		// Arrange
		envelope := &entity.EntityEnvelope{
			ID:     2,
			Name:   "New Envelope",
			Status: "completed",
		}

		signatory := &entity.EntitySignatory{
			ID:         1,
			Name:       "John Doe",
			Email:      "john.doe@example.com",
			EnvelopeID: 1,
		}

		// Mock expectations
		mockSignatoryRepo.EXPECT().
			GetByID(1).
			Return(signatory, nil)

		mockEnvelopeRepo.EXPECT().
			GetByID(2).
			Return(envelope, nil)

		// Act
		err := service.AssociateToEnvelope(1, 2)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot associate signatory to envelope with 'completed' status")
	})

	t.Run("should fail when email already exists in target envelope", func(t *testing.T) {
		// Arrange
		envelope := &entity.EntityEnvelope{
			ID:     2,
			Name:   "New Envelope",
			Status: "draft",
		}

		signatory := &entity.EntitySignatory{
			ID:         1,
			Name:       "John Doe",
			Email:      "john.doe@example.com",
			EnvelopeID: 1,
		}

		existingSignatory := &entity.EntitySignatory{
			ID:         3,
			Name:       "Jane Doe",
			Email:      "john.doe@example.com",
			EnvelopeID: 2,
		}

		// Mock expectations
		mockSignatoryRepo.EXPECT().
			GetByID(1).
			Return(signatory, nil)

		mockEnvelopeRepo.EXPECT().
			GetByID(2).
			Return(envelope, nil)

		mockSignatoryRepo.EXPECT().
			GetByEmailAndEnvelopeID("john.doe@example.com", 2).
			Return(existingSignatory, nil)

		// Act
		err := service.AssociateToEnvelope(1, 2)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "email 'john.doe@example.com' already exists in envelope 2")
	})
}