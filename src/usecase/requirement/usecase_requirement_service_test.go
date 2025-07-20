package requirement

import (
	"context"
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

// Helper function to create mock HTTP response
func createMockHTTPResponse(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}
}

func TestUsecaseRequirementService_CreateRequirement(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepoRequirement := mocks.NewMockIRepositoryRequirement(ctrl)
	mockRepoEnvelope := mocks.NewMockIRepositoryEnvelope(ctrl)
	mockClicksignClient := mocks.NewMockClicksignClientInterface(ctrl)
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Reduce log noise in tests

	service := NewUsecaseRequirementService(mockRepoRequirement, mockRepoEnvelope, mockClicksignClient, logger)
	ctx := context.Background()

	t.Run("should create requirement successfully", func(t *testing.T) {
		// Arrange
		auth := "email"
		docID := "doc123"
		signerID := "signer123"

		requirement := &entity.EntityRequirement{
			EnvelopeID: 1,
			Action:     "sign",
			Role:       "sign",
			Auth:       &auth,
			DocumentID: &docID,
			SignerID:   &signerID,
			Status:     "pending",
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		envelope := &entity.EntityEnvelope{
			ID:           1,
			Name:         "Test Envelope",
			ClicksignKey: "envelope123",
		}

		createdRequirement := &entity.EntityRequirement{
			ID:         1,
			EnvelopeID: 1,
			Action:     "sign",
			Role:       "sign",
			Auth:       &auth,
			DocumentID: &docID,
			SignerID:   &signerID,
			Status:     "pending",
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		updatedRequirement := &entity.EntityRequirement{
			ID:           1,
			EnvelopeID:   1,
			ClicksignKey: "req123",
			Action:       "sign",
			Role:         "sign",
			Auth:         &auth,
			DocumentID:   &docID,
			SignerID:     &signerID,
			Status:       "pending",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		// Mock expectations
		mockRepoEnvelope.EXPECT().
			GetByID(1).
			Return(envelope, nil)

		mockRepoRequirement.EXPECT().
			Create(gomock.Any(), requirement).
			Return(createdRequirement, nil)

		mockClicksignClient.EXPECT().
			Post(gomock.Any(), "/api/v3/envelopes/envelope123/requirements", gomock.Any()).
			Return(createMockHTTPResponse(200, `{"data":{"id":"req123","type":"requirements","attributes":{"action":"sign","role":"sign","auth":"email"}}}`), nil)

		mockRepoRequirement.EXPECT().
			Update(gomock.Any(), gomock.Any()).
			Return(updatedRequirement, nil)

		// Act
		result, err := service.CreateRequirement(ctx, requirement)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "req123", result.ClicksignKey)
		assert.Equal(t, 1, result.ID)
	})

	t.Run("should fail when envelope not found", func(t *testing.T) {
		// Arrange
		requirement := &entity.EntityRequirement{
			EnvelopeID: 999,
			Action:     "sign",
			Role:       "sign",
			Status:     "pending",
		}

		mockRepoEnvelope.EXPECT().
			GetByID(999).
			Return(nil, errors.New("envelope not found"))

		// Act
		result, err := service.CreateRequirement(ctx, requirement)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "envelope not found")
	})

	t.Run("should fail when envelope has no Clicksign key", func(t *testing.T) {
		// Arrange
		requirement := &entity.EntityRequirement{
			EnvelopeID: 1,
			Action:     "sign",
			Role:       "sign",
			Status:     "pending",
		}

		envelope := &entity.EntityEnvelope{
			ID:           1,
			Name:         "Test Envelope",
			ClicksignKey: "", // No Clicksign key
		}

		mockRepoEnvelope.EXPECT().
			GetByID(1).
			Return(envelope, nil)

		// Act
		result, err := service.CreateRequirement(ctx, requirement)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "envelope must be created in Clicksign")
	})

	t.Run("should rollback when Clicksign creation fails", func(t *testing.T) {
		// Arrange
		requirement := &entity.EntityRequirement{
			EnvelopeID: 1,
			Action:     "sign",
			Role:       "sign",
			Status:     "pending",
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		envelope := &entity.EntityEnvelope{
			ID:           1,
			Name:         "Test Envelope",
			ClicksignKey: "envelope123",
		}

		createdRequirement := &entity.EntityRequirement{
			ID:         1,
			EnvelopeID: 1,
			Action:     "sign",
			Role:       "sign",
			Status:     "pending",
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		mockRepoEnvelope.EXPECT().
			GetByID(1).
			Return(envelope, nil)

		mockRepoRequirement.EXPECT().
			Create(gomock.Any(), requirement).
			Return(createdRequirement, nil)

		mockClicksignClient.EXPECT().
			Post(gomock.Any(), "/api/v3/envelopes/envelope123/requirements", gomock.Any()).
			Return(nil, errors.New("Clicksign API error"))

		mockRepoRequirement.EXPECT().
			Delete(gomock.Any(), createdRequirement).
			Return(nil)

		// Act
		result, err := service.CreateRequirement(ctx, requirement)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to create requirement in Clicksign")
	})
}

func TestUsecaseRequirementService_GetRequirementsByEnvelopeID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepoRequirement := mocks.NewMockIRepositoryRequirement(ctrl)
	mockRepoEnvelope := mocks.NewMockIRepositoryEnvelope(ctrl)
	mockClicksignClient := mocks.NewMockClicksignClientInterface(ctrl)
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	service := NewUsecaseRequirementService(mockRepoRequirement, mockRepoEnvelope, mockClicksignClient, logger)
	ctx := context.Background()

	t.Run("should get requirements successfully", func(t *testing.T) {
		// Arrange
		envelopeID := 1
		envelope := &entity.EntityEnvelope{
			ID:   1,
			Name: "Test Envelope",
		}

		expectedRequirements := []entity.EntityRequirement{
			{
				ID:         1,
				EnvelopeID: 1,
				Action:     "sign",
				Status:     "pending",
			},
			{
				ID:         2,
				EnvelopeID: 1,
				Action:     "agree",
				Status:     "pending",
			},
		}

		mockRepoEnvelope.EXPECT().
			GetByID(envelopeID).
			Return(envelope, nil)

		mockRepoRequirement.EXPECT().
			GetByEnvelopeID(gomock.Any(), envelopeID).
			Return(expectedRequirements, nil)

		// Act
		result, err := service.GetRequirementsByEnvelopeID(ctx, envelopeID)

		// Assert
		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, expectedRequirements, result)
	})

	t.Run("should fail when envelope not found", func(t *testing.T) {
		// Arrange
		envelopeID := 999

		mockRepoEnvelope.EXPECT().
			GetByID(envelopeID).
			Return(nil, errors.New("envelope not found"))

		// Act
		result, err := service.GetRequirementsByEnvelopeID(ctx, envelopeID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "envelope not found")
	})
}

func TestUsecaseRequirementService_GetRequirement(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepoRequirement := mocks.NewMockIRepositoryRequirement(ctrl)
	mockRepoEnvelope := mocks.NewMockIRepositoryEnvelope(ctrl)
	mockClicksignClient := mocks.NewMockClicksignClientInterface(ctrl)
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	service := NewUsecaseRequirementService(mockRepoRequirement, mockRepoEnvelope, mockClicksignClient, logger)
	ctx := context.Background()

	t.Run("should get requirement successfully", func(t *testing.T) {
		// Arrange
		requirementID := 1
		expectedRequirement := &entity.EntityRequirement{
			ID:         1,
			EnvelopeID: 1,
			Action:     "sign",
			Status:     "pending",
		}

		mockRepoRequirement.EXPECT().
			GetByID(gomock.Any(), requirementID).
			Return(expectedRequirement, nil)

		// Act
		result, err := service.GetRequirement(ctx, requirementID)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, expectedRequirement, result)
	})

	t.Run("should fail when requirement not found", func(t *testing.T) {
		// Arrange
		requirementID := 999

		mockRepoRequirement.EXPECT().
			GetByID(gomock.Any(), requirementID).
			Return(nil, errors.New("requirement not found"))

		// Act
		result, err := service.GetRequirement(ctx, requirementID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to fetch requirement")
	})
}

func TestUsecaseRequirementService_UpdateRequirement(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepoRequirement := mocks.NewMockIRepositoryRequirement(ctrl)
	mockRepoEnvelope := mocks.NewMockIRepositoryEnvelope(ctrl)
	mockClicksignClient := mocks.NewMockClicksignClientInterface(ctrl)
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	service := NewUsecaseRequirementService(mockRepoRequirement, mockRepoEnvelope, mockClicksignClient, logger)
	ctx := context.Background()

	t.Run("should update requirement successfully", func(t *testing.T) {
		// Arrange
		requirement := &entity.EntityRequirement{
			ID:         1,
			EnvelopeID: 1,
			Action:     "sign",
			Status:     "completed",
		}

		mockRepoRequirement.EXPECT().
			Update(gomock.Any(), requirement).
			Return(requirement, nil)

		// Act
		result, err := service.UpdateRequirement(ctx, requirement)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, requirement, result)
	})
}

func TestUsecaseRequirementService_DeleteRequirement(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepoRequirement := mocks.NewMockIRepositoryRequirement(ctrl)
	mockRepoEnvelope := mocks.NewMockIRepositoryEnvelope(ctrl)
	mockClicksignClient := mocks.NewMockClicksignClientInterface(ctrl)
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	service := NewUsecaseRequirementService(mockRepoRequirement, mockRepoEnvelope, mockClicksignClient, logger)
	ctx := context.Background()

	t.Run("should delete requirement successfully", func(t *testing.T) {
		// Arrange
		requirementID := 1
		requirement := &entity.EntityRequirement{
			ID:         1,
			EnvelopeID: 1,
			Action:     "sign",
			Status:     "pending",
		}

		mockRepoRequirement.EXPECT().
			GetByID(gomock.Any(), requirementID).
			Return(requirement, nil)

		mockRepoRequirement.EXPECT().
			Delete(gomock.Any(), requirement).
			Return(nil)

		// Act
		err := service.DeleteRequirement(ctx, requirementID)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("should fail when requirement not found", func(t *testing.T) {
		// Arrange
		requirementID := 999

		mockRepoRequirement.EXPECT().
			GetByID(gomock.Any(), requirementID).
			Return(nil, errors.New("requirement not found"))

		// Act
		err := service.DeleteRequirement(ctx, requirementID)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to fetch requirement for deletion")
	})
}