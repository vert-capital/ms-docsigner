package clicksign

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"app/entity"
	"app/infrastructure/clicksign/dto"
	"app/mocks"
	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestEnvelopeService_CreateEnvelope(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockClicksignClientInterface(ctrl)
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Reduce log noise in tests

	service := NewEnvelopeService(mockClient, logger)

	t.Run("should create envelope with JSON API format", func(t *testing.T) {
		// Arrange
		envelope := &entity.EntityEnvelope{
			Name:           "Test Envelope",
			Message:        "Please sign this document",
			AutoClose:      true,
			RemindInterval: 3,
			DeadlineAt:     &time.Time{},
		}

		ctx := context.WithValue(context.Background(), "correlation_id", "test-correlation-123")

		// Mock response in JSON API format
		jsonAPIResponse := `{
			"data": {
				"type": "envelopes",
				"id": "envelope-123",
				"attributes": {
					"name": "Test Envelope",
					"status": "draft",
					"locale": "pt-BR",
					"auto_close": true,
					"remind_interval": 3,
					"block_after_refusal": true,
					"created_at": "2024-01-01T00:00:00Z",
					"updated_at": "2024-01-01T00:00:00Z"
				}
			}
		}`

		mockResponse := &http.Response{
			StatusCode: 201,
			Body:       io.NopCloser(strings.NewReader(jsonAPIResponse)),
		}

		// Expect POST call with JSON API structure
		mockClient.EXPECT().
			Post(ctx, "/api/v3/envelopes", gomock.Any()).
			DoAndReturn(func(ctx context.Context, endpoint string, body interface{}) (*http.Response, error) {
				// Validate that body is JSON API format
				wrapper, ok := body.(*dto.EnvelopeCreateRequestWrapper)
				assert.True(t, ok, "Body should be EnvelopeCreateRequestWrapper")
				assert.Equal(t, "envelopes", wrapper.Data.Type)
				assert.Equal(t, "Test Envelope", wrapper.Data.Attributes.Name)
				assert.Equal(t, "pt-BR", wrapper.Data.Attributes.Locale)
				assert.True(t, wrapper.Data.Attributes.AutoClose)
				assert.Equal(t, 3, wrapper.Data.Attributes.RemindInterval)
				assert.True(t, wrapper.Data.Attributes.BlockAfterRefusal)
				assert.Equal(t, "Please sign this document", wrapper.Data.Attributes.DefaultSubject)
				
				return mockResponse, nil
			})

		// Act
		envelopeID, err := service.CreateEnvelope(ctx, envelope)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "envelope-123", envelopeID)
	})

	t.Run("should handle JSON API error response", func(t *testing.T) {
		// Arrange
		envelope := &entity.EntityEnvelope{
			Name:           "Test Envelope",
			AutoClose:      true,
			RemindInterval: 3,
		}

		ctx := context.WithValue(context.Background(), "correlation_id", "test-correlation-123")

		errorResponse := `{
			"error": {
				"type": "validation_error",
				"message": "Name is required",
				"code": "INVALID_FIELD"
			}
		}`

		mockResponse := &http.Response{
			StatusCode: 400,
			Body:       io.NopCloser(strings.NewReader(errorResponse)),
		}

		mockClient.EXPECT().
			Post(ctx, "/api/v3/envelopes", gomock.Any()).
			Return(mockResponse, nil)

		// Act
		envelopeID, err := service.CreateEnvelope(ctx, envelope)

		// Assert
		assert.Error(t, err)
		assert.Empty(t, envelopeID)
		assert.Contains(t, err.Error(), "validation_error")
	})

	t.Run("should fail when client returns error", func(t *testing.T) {
		// Arrange
		envelope := &entity.EntityEnvelope{
			Name:           "Test Envelope",
			AutoClose:      true,
			RemindInterval: 3,
		}

		ctx := context.WithValue(context.Background(), "correlation_id", "test-correlation-123")

		mockClient.EXPECT().
			Post(ctx, "/api/v3/envelopes", gomock.Any()).
			Return(nil, assert.AnError)

		// Act
		envelopeID, err := service.CreateEnvelope(ctx, envelope)

		// Assert
		assert.Error(t, err)
		assert.Empty(t, envelopeID)
		assert.Contains(t, err.Error(), "failed to create envelope in Clicksign")
	})
}

func TestEnvelopeService_mapEntityToCreateRequest(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	service := &EnvelopeService{
		logger: logger,
	}

	t.Run("should map entity to JSON API format correctly", func(t *testing.T) {
		// Arrange
		deadline := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)
		envelope := &entity.EntityEnvelope{
			Name:           "Test Envelope",
			Message:        "Please sign",
			AutoClose:      true,
			RemindInterval: 5,
			DeadlineAt:     &deadline,
		}

		// Act
		request := service.mapEntityToCreateRequest(envelope)

		// Assert
		assert.NotNil(t, request)
		assert.Equal(t, "envelopes", request.Data.Type)
		
		attrs := request.Data.Attributes
		assert.Equal(t, "Test Envelope", attrs.Name)
		assert.Equal(t, "pt-BR", attrs.Locale)
		assert.True(t, attrs.AutoClose)
		assert.Equal(t, 5, attrs.RemindInterval)
		assert.True(t, attrs.BlockAfterRefusal)
		assert.Equal(t, "Please sign", attrs.DefaultSubject)
		assert.Equal(t, &deadline, attrs.DeadlineAt)
	})

	t.Run("should map entity without message correctly", func(t *testing.T) {
		// Arrange
		envelope := &entity.EntityEnvelope{
			Name:           "Test Envelope",
			AutoClose:      false,
			RemindInterval: 7,
		}

		// Act
		request := service.mapEntityToCreateRequest(envelope)

		// Assert
		assert.NotNil(t, request)
		assert.Equal(t, "envelopes", request.Data.Type)
		
		attrs := request.Data.Attributes
		assert.Equal(t, "Test Envelope", attrs.Name)
		assert.Equal(t, "pt-BR", attrs.Locale)
		assert.False(t, attrs.AutoClose)
		assert.Equal(t, 7, attrs.RemindInterval)
		assert.True(t, attrs.BlockAfterRefusal)
		assert.Empty(t, attrs.DefaultSubject) // Empty when no message
		assert.Nil(t, attrs.DeadlineAt)
	})
}

func TestEnvelopeService_JSONAPIStructure(t *testing.T) {
	t.Run("should validate JSON API wrapper structure", func(t *testing.T) {
		// Test that our DTO structures follow JSON API spec correctly
		wrapper := &dto.EnvelopeCreateRequestWrapper{
			Data: dto.EnvelopeCreateData{
				Type: "envelopes",
				Attributes: dto.EnvelopeCreateAttributes{
					Name:              "Test",
					Locale:            "pt-BR",
					AutoClose:         true,
					RemindInterval:    3,
					BlockAfterRefusal: true,
				},
			},
		}

		// Validate structure
		assert.Equal(t, "envelopes", wrapper.Data.Type)
		assert.Equal(t, "Test", wrapper.Data.Attributes.Name)
		assert.Equal(t, "pt-BR", wrapper.Data.Attributes.Locale)
		assert.True(t, wrapper.Data.Attributes.AutoClose)
		assert.Equal(t, 3, wrapper.Data.Attributes.RemindInterval)
		assert.True(t, wrapper.Data.Attributes.BlockAfterRefusal)
	})

	t.Run("should validate JSON API response structure", func(t *testing.T) {
		// Test response structure
		response := &dto.EnvelopeCreateResponseWrapper{
			Data: dto.EnvelopeCreateResponseData{
				Type: "envelopes",
				ID:   "envelope-123",
				Attributes: dto.EnvelopeCreateResponseAttributes{
					Name:   "Test Envelope",
					Status: "draft",
					Locale: "pt-BR",
				},
			},
		}

		// Validate structure
		assert.Equal(t, "envelopes", response.Data.Type)
		assert.Equal(t, "envelope-123", response.Data.ID)
		assert.Equal(t, "Test Envelope", response.Data.Attributes.Name)
		assert.Equal(t, "draft", response.Data.Attributes.Status)
		assert.Equal(t, "pt-BR", response.Data.Attributes.Locale)
	})
}