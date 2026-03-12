package vertc_assinaturas

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"

	"app/config"

	"github.com/sirupsen/logrus"
)

// VertcAssinaturasError representa tipos específicos de erros do vertc-assinaturas
type VertcAssinaturasError struct {
	Type       string
	Message    string
	StatusCode int
	Original   error
}

func (e *VertcAssinaturasError) Error() string {
	if e.Original != nil {
		return fmt.Sprintf("vertc-assinaturas %s error: %s (original: %v)", e.Type, e.Message, e.Original)
	}
	return fmt.Sprintf("vertc-assinaturas %s error: %s", e.Type, e.Message)
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

// LoginResponse representa a resposta do endpoint de login
type LoginResponse struct {
	AccessToken string `json:"access_token"`
}

// LoginRequest representa a requisição de login
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// VertcAssinaturasClient é o cliente HTTP para comunicação com a API vertc-assinaturas
type VertcAssinaturasClient struct {
	httpClient *http.Client
	baseURL    string
	email      string
	password   string
	logger     *logrus.Logger
}

// NewVertcAssinaturasClient cria uma nova instância do VertcAssinaturasClient
func NewVertcAssinaturasClient(envVars config.EnvironmentVars, logger *logrus.Logger) *VertcAssinaturasClient {
	client := &http.Client{
		Timeout: time.Duration(envVars.VERTC_ASSINATURAS_TIMEOUT) * time.Second,
	}

	return &VertcAssinaturasClient{
		httpClient: client,
		baseURL:    envVars.VERTC_ASSINATURAS_BASE_URL,
		email:      envVars.VERTC_ASSINATURAS_EMAIL,
		password:   envVars.VERTC_ASSINATURAS_PASSWORD,
		logger:     logger,
	}
}

// Login faz login na API e retorna o access token
func (c *VertcAssinaturasClient) Login(ctx context.Context) (string, error) {
	url := fmt.Sprintf("%s/api/v1/auth/login", c.baseURL)

	loginReq := LoginRequest{
		Email:    c.email,
		Password: c.password,
	}

	bodyBytes, err := json.Marshal(loginReq)
	if err != nil {
		return "", &VertcAssinaturasError{
			Type:     ErrorTypeSerialization,
			Message:  "failed to marshal login request",
			Original: err,
		}
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", &VertcAssinaturasError{
			Type:     ErrorTypeClient,
			Message:  "failed to create HTTP request",
			Original: err,
		}
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	c.logger.Debugf("Fazendo login no vertc-assinaturas: %s", url)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		errorType := c.categorizeError(err)
		return "", &VertcAssinaturasError{
			Type:     errorType,
			Message:  "HTTP request failed during login",
			Original: err,
		}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", &VertcAssinaturasError{
			Type:     ErrorTypeClient,
			Message:  "failed to read login response",
			Original: err,
		}
	}

	// Aceitar tanto 200 (OK) quanto 201 (Created) como sucesso
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", &VertcAssinaturasError{
			Type:       ErrorTypeAuthentication,
			Message:    fmt.Sprintf("login failed with status %d: %s", resp.StatusCode, string(body)),
			StatusCode: resp.StatusCode,
		}
	}

	var loginResp LoginResponse
	if err := json.Unmarshal(body, &loginResp); err != nil {
		return "", &VertcAssinaturasError{
			Type:     ErrorTypeSerialization,
			Message:  "failed to unmarshal login response",
			Original: err,
		}
	}

	if loginResp.AccessToken == "" {
		return "", &VertcAssinaturasError{
			Type:    ErrorTypeAuthentication,
			Message: "access token is empty in login response",
		}
	}

	c.logger.Debug("Login realizado com sucesso no vertc-assinaturas")
	return loginResp.AccessToken, nil
}

// Get faz uma requisição GET autenticada
func (c *VertcAssinaturasClient) Get(ctx context.Context, endpoint string) (*http.Response, error) {
	return c.doAuthenticatedRequest(ctx, http.MethodGet, endpoint, nil, "")
}

// Post faz uma requisição POST autenticada
func (c *VertcAssinaturasClient) Post(ctx context.Context, endpoint string, body interface{}, idempotencyKey string) (*http.Response, error) {
	return c.doAuthenticatedRequest(ctx, http.MethodPost, endpoint, body, idempotencyKey)
}

// PostMultipartFile faz upload autenticado de um único arquivo multipart.
func (c *VertcAssinaturasClient) PostMultipartFile(
	ctx context.Context,
	endpoint string,
	fieldName string,
	fileName string,
	fileContent []byte,
) (*http.Response, error) {
	token, err := c.Login(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to login: %w", err)
	}

	url := fmt.Sprintf("%s%s", c.baseURL, endpoint)

	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	part, err := writer.CreateFormFile(fieldName, fileName)
	if err != nil {
		return nil, &VertcAssinaturasError{
			Type:     ErrorTypeClient,
			Message:  "failed to create multipart form file",
			Original: err,
		}
	}

	if _, err := part.Write(fileContent); err != nil {
		return nil, &VertcAssinaturasError{
			Type:     ErrorTypeClient,
			Message:  "failed to write multipart file content",
			Original: err,
		}
	}

	if err := writer.Close(); err != nil {
		return nil, &VertcAssinaturasError{
			Type:     ErrorTypeClient,
			Message:  "failed to finalize multipart payload",
			Original: err,
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &requestBody)
	if err != nil {
		return nil, &VertcAssinaturasError{
			Type:     ErrorTypeClient,
			Message:  "failed to create multipart HTTP request",
			Original: err,
		}
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Accept", "application/json")

	correlationID := ctx.Value("correlation_id")
	if correlationID != nil {
		req.Header.Set("X-Correlation-ID", correlationID.(string))
	}

	c.logger.Debugf("Fazendo requisição multipart POST para: %s", url)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		errorType := c.categorizeError(err)
		return nil, &VertcAssinaturasError{
			Type:     errorType,
			Message:  "multipart HTTP request failed",
			Original: err,
		}
	}

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		errorType := c.categorizeHTTPError(resp.StatusCode)
		return nil, &VertcAssinaturasError{
			Type:       errorType,
			Message:    fmt.Sprintf("API error (status %d): %s", resp.StatusCode, string(body)),
			StatusCode: resp.StatusCode,
		}
	}

	return resp, nil
}

func (c *VertcAssinaturasClient) doAuthenticatedRequest(ctx context.Context, method, endpoint string, body interface{}, idempotencyKey string) (*http.Response, error) {
	// Fazer login para obter token
	token, err := c.Login(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to login: %w", err)
	}

	url := fmt.Sprintf("%s%s", c.baseURL, endpoint)

	var bodyBytes []byte
	if body != nil {
		var err error
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return nil, &VertcAssinaturasError{
				Type:     ErrorTypeSerialization,
				Message:  "failed to marshal request body",
				Original: err,
			}
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, &VertcAssinaturasError{
			Type:     ErrorTypeClient,
			Message:  "failed to create HTTP request",
			Original: err,
		}
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	if idempotencyKey != "" {
		req.Header.Set("x-idempotency-key", idempotencyKey)
	}

	correlationID := ctx.Value("correlation_id")
	if correlationID != nil {
		req.Header.Set("X-Correlation-ID", correlationID.(string))
	}

	c.logger.Debugf("Fazendo requisição %s para: %s", method, url)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		errorType := c.categorizeError(err)
		return nil, &VertcAssinaturasError{
			Type:     errorType,
			Message:  "HTTP request failed",
			Original: err,
		}
	}

	// Verificar se houve erro na resposta
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		errorType := c.categorizeHTTPError(resp.StatusCode)
		return nil, &VertcAssinaturasError{
			Type:       errorType,
			Message:    fmt.Sprintf("API error (status %d): %s", resp.StatusCode, string(body)),
			StatusCode: resp.StatusCode,
		}
	}

	return resp, nil
}

// categorizeError categoriza erros de rede/timeout
func (c *VertcAssinaturasClient) categorizeError(err error) string {
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
func (c *VertcAssinaturasClient) categorizeHTTPError(statusCode int) string {
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
