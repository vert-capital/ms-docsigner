package provider_factory

import (
	"testing"

	"app/config"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestNewProviderFactory(t *testing.T) {
	envVars := config.EnvironmentVars{
		CLICKSIGN_API_KEY:        "test-api-key",
		CLICKSIGN_BASE_URL:       "https://api.clicksign.com",
		CLICKSIGN_TIMEOUT:        30,
		CLICKSIGN_RETRY_ATTEMPTS: 3,
	}

	logger := logrus.New()
	factory := NewProviderFactory(envVars, logger)

	assert.NotNil(t, factory)
	assert.IsType(t, &ProviderFactory{}, factory)
}

func TestProviderFactory_GetProvider(t *testing.T) {
	envVars := config.EnvironmentVars{
		CLICKSIGN_API_KEY:        "test-api-key",
		CLICKSIGN_BASE_URL:       "https://api.clicksign.com",
		CLICKSIGN_TIMEOUT:        30,
		CLICKSIGN_RETRY_ATTEMPTS: 3,
	}

	logger := logrus.New()
	factory := NewProviderFactory(envVars, logger)

	t.Run("should return ClicksignProvider for clicksign", func(t *testing.T) {
		provider, err := factory.GetProvider("clicksign")
		assert.NoError(t, err)
		assert.NotNil(t, provider)
	})

	t.Run("should return VertSign provider for vert-sign", func(t *testing.T) {
		provider, err := factory.GetProvider("vert-sign")
		assert.NoError(t, err)
		assert.NotNil(t, provider)
	})

	t.Run("should return error for unsupported provider", func(t *testing.T) {
		provider, err := factory.GetProvider("invalid-provider")
		assert.Error(t, err)
		assert.Nil(t, provider)
		assert.Contains(t, err.Error(), "unsupported provider")
	})
}

func TestProviderFactory_IsProviderSupported(t *testing.T) {
	envVars := config.EnvironmentVars{
		CLICKSIGN_API_KEY:        "test-api-key",
		CLICKSIGN_BASE_URL:       "https://api.clicksign.com",
		CLICKSIGN_TIMEOUT:        30,
		CLICKSIGN_RETRY_ATTEMPTS: 3,
	}

	logger := logrus.New()
	factory := NewProviderFactory(envVars, logger)

	t.Run("should return true for clicksign", func(t *testing.T) {
		assert.True(t, factory.IsProviderSupported("clicksign"))
	})

	t.Run("should return true for vert-sign", func(t *testing.T) {
		assert.True(t, factory.IsProviderSupported("vert-sign"))
	})

	t.Run("should return false for invalid provider", func(t *testing.T) {
		assert.False(t, factory.IsProviderSupported("invalid-provider"))
	})
}

func TestProviderFactory_IsProviderImplemented(t *testing.T) {
	envVars := config.EnvironmentVars{
		CLICKSIGN_API_KEY:        "test-api-key",
		CLICKSIGN_BASE_URL:       "https://api.clicksign.com",
		CLICKSIGN_TIMEOUT:        30,
		CLICKSIGN_RETRY_ATTEMPTS: 3,
	}

	logger := logrus.New()
	factory := NewProviderFactory(envVars, logger)

	t.Run("should return true for clicksign", func(t *testing.T) {
		assert.True(t, factory.IsProviderImplemented("clicksign"))
	})

	t.Run("should return true for vert-sign", func(t *testing.T) {
		assert.True(t, factory.IsProviderImplemented("vert-sign"))
	})

	t.Run("should return false for invalid provider", func(t *testing.T) {
		assert.False(t, factory.IsProviderImplemented("invalid-provider"))
	})
}
