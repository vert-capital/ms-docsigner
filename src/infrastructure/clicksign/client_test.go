package clicksign

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"app/config"
)

func TestNewClicksignClient(t *testing.T) {
	envVars := config.EnvironmentVars{
		CLICKSIGN_API_KEY:        "test-api-key",
		CLICKSIGN_BASE_URL:       "https://api.clicksign.com",
		CLICKSIGN_TIMEOUT:        30,
		CLICKSIGN_RETRY_ATTEMPTS: 3,
	}

	logger := logrus.New()
	client := NewClicksignClient(envVars, logger)

	assert.NotNil(t, client)
	assert.IsType(t, &ClicksignClient{}, client)
}

func TestClicksignClient_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/test", r.URL.Path)
		assert.Equal(t, "Bearer test-api-key", r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "application/json", r.Header.Get("Accept"))
		
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "success"}`))
	}))
	defer server.Close()

	envVars := config.EnvironmentVars{
		CLICKSIGN_API_KEY:        "test-api-key",
		CLICKSIGN_BASE_URL:       server.URL,
		CLICKSIGN_TIMEOUT:        30,
		CLICKSIGN_RETRY_ATTEMPTS: 3,
	}

	logger := logrus.New()
	client := NewClicksignClient(envVars, logger)

	ctx := context.Background()
	resp, err := client.Get(ctx, "/test")

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestClicksignClient_Post(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/test", r.URL.Path)
		assert.Equal(t, "Bearer test-api-key", r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"message": "created"}`))
	}))
	defer server.Close()

	envVars := config.EnvironmentVars{
		CLICKSIGN_API_KEY:        "test-api-key",
		CLICKSIGN_BASE_URL:       server.URL,
		CLICKSIGN_TIMEOUT:        30,
		CLICKSIGN_RETRY_ATTEMPTS: 3,
	}

	logger := logrus.New()
	client := NewClicksignClient(envVars, logger)

	ctx := context.Background()
	body := map[string]string{"key": "value"}
	resp, err := client.Post(ctx, "/test", body)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
}

func TestClicksignClient_Put(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/test", r.URL.Path)
		assert.Equal(t, "Bearer test-api-key", r.Header.Get("Authorization"))
		
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "updated"}`))
	}))
	defer server.Close()

	envVars := config.EnvironmentVars{
		CLICKSIGN_API_KEY:        "test-api-key",
		CLICKSIGN_BASE_URL:       server.URL,
		CLICKSIGN_TIMEOUT:        30,
		CLICKSIGN_RETRY_ATTEMPTS: 3,
	}

	logger := logrus.New()
	client := NewClicksignClient(envVars, logger)

	ctx := context.Background()
	body := map[string]string{"key": "value"}
	resp, err := client.Put(ctx, "/test", body)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestClicksignClient_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/test", r.URL.Path)
		assert.Equal(t, "Bearer test-api-key", r.Header.Get("Authorization"))
		
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	envVars := config.EnvironmentVars{
		CLICKSIGN_API_KEY:        "test-api-key",
		CLICKSIGN_BASE_URL:       server.URL,
		CLICKSIGN_TIMEOUT:        30,
		CLICKSIGN_RETRY_ATTEMPTS: 3,
	}

	logger := logrus.New()
	client := NewClicksignClient(envVars, logger)

	ctx := context.Background()
	resp, err := client.Delete(ctx, "/test")

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestClicksignClient_Patch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PATCH", r.Method)
		assert.Equal(t, "/test", r.URL.Path)
		assert.Equal(t, "Bearer test-api-key", r.Header.Get("Authorization"))
		
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "patched"}`))
	}))
	defer server.Close()

	envVars := config.EnvironmentVars{
		CLICKSIGN_API_KEY:        "test-api-key",
		CLICKSIGN_BASE_URL:       server.URL,
		CLICKSIGN_TIMEOUT:        30,
		CLICKSIGN_RETRY_ATTEMPTS: 3,
	}

	logger := logrus.New()
	client := NewClicksignClient(envVars, logger)

	ctx := context.Background()
	body := map[string]string{"key": "value"}
	resp, err := client.Patch(ctx, "/test", body)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestClicksignClient_ErrorHandling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal server error"}`))
	}))
	defer server.Close()

	envVars := config.EnvironmentVars{
		CLICKSIGN_API_KEY:        "test-api-key",
		CLICKSIGN_BASE_URL:       server.URL,
		CLICKSIGN_TIMEOUT:        30,
		CLICKSIGN_RETRY_ATTEMPTS: 3,
	}

	logger := logrus.New()
	client := NewClicksignClient(envVars, logger)

	ctx := context.Background()
	resp, err := client.Get(ctx, "/test")

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

func TestClicksignClient_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	envVars := config.EnvironmentVars{
		CLICKSIGN_API_KEY:        "test-api-key",
		CLICKSIGN_BASE_URL:       server.URL,
		CLICKSIGN_TIMEOUT:        1,
		CLICKSIGN_RETRY_ATTEMPTS: 3,
	}

	logger := logrus.New()
	client := NewClicksignClient(envVars, logger)

	ctx := context.Background()
	resp, err := client.Get(ctx, "/test")

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "context deadline exceeded")
}

func TestClicksignClient_CorrelationID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "test-correlation-id", r.Header.Get("X-Correlation-ID"))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	envVars := config.EnvironmentVars{
		CLICKSIGN_API_KEY:        "test-api-key",
		CLICKSIGN_BASE_URL:       server.URL,
		CLICKSIGN_TIMEOUT:        30,
		CLICKSIGN_RETRY_ATTEMPTS: 3,
	}

	logger := logrus.New()
	client := NewClicksignClient(envVars, logger)

	ctx := context.WithValue(context.Background(), "correlation_id", "test-correlation-id")
	resp, err := client.Get(ctx, "/test")

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestClicksignClient_InvalidJSON(t *testing.T) {
	envVars := config.EnvironmentVars{
		CLICKSIGN_API_KEY:        "test-api-key",
		CLICKSIGN_BASE_URL:       "https://api.clicksign.com",
		CLICKSIGN_TIMEOUT:        30,
		CLICKSIGN_RETRY_ATTEMPTS: 3,
	}

	logger := logrus.New()
	client := NewClicksignClient(envVars, logger)

	ctx := context.Background()
	invalidBody := make(chan int)
	resp, err := client.Post(ctx, "/test", invalidBody)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "failed to marshal request body")
}