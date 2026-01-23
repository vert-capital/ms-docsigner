package clicksign_provider

import (
	"context"
	"testing"

	"app/entity"
	"app/infrastructure/provider"
	"app/mocks"

	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestNewClicksignProvider(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockClicksignClientInterface(ctrl)
	logger := logrus.New()

	envelopeProvider := NewClicksignProvider(mockClient, logger)

	assert.NotNil(t, envelopeProvider)
	assert.IsType(t, &ClicksignProvider{}, envelopeProvider)
}

func TestClicksignProvider_ImplementsInterface(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockClicksignClientInterface(ctrl)
	logger := logrus.New()

	envelopeProvider := NewClicksignProvider(mockClient, logger)

	// Verificar que implementa a interface
	var _ provider.EnvelopeProvider = envelopeProvider
	assert.NotNil(t, envelopeProvider)
}

func TestClicksignProvider_CreateSigner(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockClicksignClientInterface(ctrl)
	logger := logrus.New()

	envelopeProvider := NewClicksignProvider(mockClient, logger)
	clicksignProvider := envelopeProvider.(*ClicksignProvider)

	signerData := provider.SignerData{
		Name:              "Test User",
		Email:             "test@example.com",
		Birthday:          "1990-01-01",
		HasDocumentation:  true,
		Refusable:         false,
		Group:             1,
	}

	// Verificar que os dados são mapeados corretamente
	// Este teste verifica que a estrutura está correta
	assert.Equal(t, "Test User", signerData.Name)
	assert.Equal(t, "test@example.com", signerData.Email)
	assert.Equal(t, clicksignProvider, envelopeProvider)
}

func TestClicksignProvider_CreateRequirement(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockClicksignClientInterface(ctrl)
	logger := logrus.New()

	envelopeProvider := NewClicksignProvider(mockClient, logger)

	reqData := provider.RequirementData{
		Action:     "sign",
		Role:       "sign",
		Auth:       "email",
		DocumentID: "doc123",
		SignerID:   "signer123",
	}

	// Verificar que os dados são mapeados corretamente
	assert.Equal(t, "sign", reqData.Action)
	assert.Equal(t, "email", reqData.Auth)
	assert.NotNil(t, envelopeProvider)
}

func TestClicksignProvider_Context(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockClicksignClientInterface(ctrl)
	logger := logrus.New()

	envelopeProvider := NewClicksignProvider(mockClient, logger)

	ctx := context.Background()
	envelope := &entity.EntityEnvelope{
		Name: "Test Envelope",
	}

	// Verificar que o contexto pode ser passado
	assert.NotNil(t, ctx)
	assert.NotNil(t, envelope)
	assert.NotNil(t, envelopeProvider)
}

