package clicksign

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"app/config"
	"app/usecase/clicksign"

	"github.com/sirupsen/logrus"
)

// ClicksignError representa tipos específicos de erros do Clicksign
type ClicksignError struct {
	Type       string
	Message    string
	StatusCode int
	Original   error
}

func (e *ClicksignError) Error() string {
	if e.Original != nil {
		return fmt.Sprintf("clicksign %s error: %s (original: %v)", e.Type, e.Message, e.Original)
	}
	return fmt.Sprintf("clicksign %s error: %s", e.Type, e.Message)
}

// Error types constants
const (
	ErrorTypeNetwork        = "network"
	ErrorTypeTimeout        = "timeout"
	ErrorTypeAuthentication = "authentication"
	ErrorTypeClient         = "client"
	ErrorTypeServer         = "server"
	ErrorTypeSerialization  = "serialization"
)

type ClicksignClient struct {
	httpClient    *http.Client
	baseURL       string
	apiKey        string
	logger        *logrus.Logger
	retryAttempts int
}

func NewClicksignClient(envVars config.EnvironmentVars, logger *logrus.Logger) clicksign.ClicksignClientInterface {
	client := &http.Client{
		Timeout: time.Duration(envVars.CLICKSIGN_TIMEOUT) * time.Second,
	}

	return &ClicksignClient{
		httpClient:    client,
		baseURL:       envVars.CLICKSIGN_BASE_URL,
		apiKey:        envVars.CLICKSIGN_API_KEY,
		logger:        logger,
		retryAttempts: envVars.CLICKSIGN_RETRY_ATTEMPTS,
	}
}

func (c *ClicksignClient) Get(ctx context.Context, endpoint string) (*http.Response, error) {
	return c.doRequest(ctx, "GET", endpoint, nil)
}

func (c *ClicksignClient) Post(ctx context.Context, endpoint string, body any) (*http.Response, error) {
	return c.doRequest(ctx, "POST", endpoint, body)
}

func (c *ClicksignClient) Put(ctx context.Context, endpoint string, body any) (*http.Response, error) {
	return c.doRequest(ctx, "PUT", endpoint, body)
}

func (c *ClicksignClient) Delete(ctx context.Context, endpoint string) (*http.Response, error) {
	return c.doRequest(ctx, "DELETE", endpoint, nil)
}

func (c *ClicksignClient) Patch(ctx context.Context, endpoint string, body any) (*http.Response, error) {
	return c.doRequest(ctx, "PATCH", endpoint, body)
}

func (c *ClicksignClient) doRequest(ctx context.Context, method, endpoint string, body any) (*http.Response, error) {
	url := fmt.Sprintf("%s%s", c.baseURL, endpoint)

	var bodyBytes []byte
	if body != nil {
		var err error
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return nil, &ClicksignError{
				Type:     ErrorTypeSerialization,
				Message:  "failed to marshal request body",
				Original: err,
			}
		}
	}

	// Implementar retry com backoff exponencial
	var lastErr error
	for attempt := 0; attempt <= c.retryAttempts; attempt++ {
		if attempt > 0 {
			// Backoff exponencial: 100ms, 200ms, 400ms, 800ms...
			backoffDuration := time.Duration(100*attempt*attempt) * time.Millisecond

			select {
			case <-ctx.Done():
				return nil, &ClicksignError{
					Type:     ErrorTypeTimeout,
					Message:  "context cancelled during retry backoff",
					Original: ctx.Err(),
				}
			case <-time.After(backoffDuration):
				// Continue with retry
			}
		}

		resp, err := c.executeRequest(ctx, method, url, bodyBytes)
		if err != nil {
			lastErr = err
			// Verificar se deve tentar novamente
			if !c.shouldRetry(err, attempt) {
				return nil, err
			}
			continue
		}

		// Verificar se deve tentar novamente baseado no status code
		if resp.StatusCode >= 500 && attempt < c.retryAttempts {
			resp.Body.Close() // Fechar o body antes de tentar novamente
			lastErr = &ClicksignError{
				Type:       ErrorTypeServer,
				Message:    "server error - retrying",
				StatusCode: resp.StatusCode,
			}
			continue
		}

		// Se chegou aqui e ainda é um erro 500, mas não há mais tentativas
		if resp.StatusCode >= 500 {
			resp.Body.Close()
			return nil, &ClicksignError{
				Type:       ErrorTypeServer,
				Message:    "server error - max retries exceeded",
				StatusCode: resp.StatusCode,
			}
		}

		return resp, nil
	}

	// Se chegou aqui, todos os attempts falharam
	if lastErr != nil {
		return nil, lastErr
	}

	return nil, &ClicksignError{
		Type:    ErrorTypeClient,
		Message: "all retry attempts failed",
	}
}

// executeRequest executa uma única requisição HTTP
func (c *ClicksignClient) executeRequest(ctx context.Context, method, url string, bodyBytes []byte) (*http.Response, error) {
	var bodyReader io.Reader
	if bodyBytes != nil {
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, &ClicksignError{
			Type:     ErrorTypeClient,
			Message:  "failed to create HTTP request",
			Original: err,
		}
	}

	req.Header.Set("Authorization", c.apiKey)
	req.Header.Set("Content-Type", "application/vnd.api+json")
	req.Header.Set("Accept", "application/vnd.api+json")

	correlationID := ctx.Value("correlation_id")
	if correlationID != nil {
		req.Header.Set("X-Correlation-ID", correlationID.(string))
	}

	resp, err := c.httpClient.Do(req)

	if err != nil {

		// Categorizar tipos de erro
		errorType := c.categorizeError(err)
		return nil, &ClicksignError{
			Type:     errorType,
			Message:  "HTTP request failed",
			Original: err,
		}
	}

	return resp, nil
}

// shouldRetry determina se deve tentar novamente baseado no erro
func (c *ClicksignClient) shouldRetry(err error, attempt int) bool {
	if attempt >= c.retryAttempts {
		return false
	}

	if clicksignErr, ok := err.(*ClicksignError); ok {
		switch clicksignErr.Type {
		case ErrorTypeTimeout, ErrorTypeNetwork:
			return true
		case ErrorTypeServer:
			return true
		case ErrorTypeAuthentication, ErrorTypeClient:
			return false
		default:
			return false
		}
	}

	return false
}

// categorizeError categoriza erros de rede/timeout
func (c *ClicksignClient) categorizeError(err error) string {
	errorStr := err.Error()
	switch {
	case contains(errorStr, "timeout", "context deadline exceeded"):
		return ErrorTypeTimeout
	case contains(errorStr, "connection refused", "no such host", "network"):
		return ErrorTypeNetwork
	default:
		return ErrorTypeClient
	}
}

// categorizeHTTPError categoriza erros baseados no status code HTTP
func (c *ClicksignClient) categorizeHTTPError(statusCode int) string {
	switch {
	case statusCode == 401 || statusCode == 403:
		return ErrorTypeAuthentication
	case statusCode >= 400 && statusCode < 500:
		return ErrorTypeClient
	case statusCode >= 500:
		return ErrorTypeServer
	default:
		return ErrorTypeClient
	}
}

// contains verifica se uma string contém qualquer uma das substrings fornecidas
func contains(s string, substrings ...string) bool {
	for _, substring := range substrings {
		if len(s) >= len(substring) {
			for i := 0; i <= len(s)-len(substring); i++ {
				if s[i:i+len(substring)] == substring {
					return true
				}
			}
		}
	}
	return false
}
