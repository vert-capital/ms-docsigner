package clicksign

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"app/config"
	"app/usecase/clicksign"
)

type ClicksignClient struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string
	logger     *logrus.Logger
}

func NewClicksignClient(envVars config.EnvironmentVars, logger *logrus.Logger) clicksign.ClicksignClientInterface {
	client := &http.Client{
		Timeout: time.Duration(envVars.CLICKSIGN_TIMEOUT) * time.Second,
	}

	return &ClicksignClient{
		httpClient: client,
		baseURL:    envVars.CLICKSIGN_BASE_URL,
		apiKey:     envVars.CLICKSIGN_API_KEY,
		logger:     logger,
	}
}

func (c *ClicksignClient) Get(ctx context.Context, endpoint string) (*http.Response, error) {
	return c.doRequest(ctx, "GET", endpoint, nil)
}

func (c *ClicksignClient) Post(ctx context.Context, endpoint string, body interface{}) (*http.Response, error) {
	return c.doRequest(ctx, "POST", endpoint, body)
}

func (c *ClicksignClient) Put(ctx context.Context, endpoint string, body interface{}) (*http.Response, error) {
	return c.doRequest(ctx, "PUT", endpoint, body)
}

func (c *ClicksignClient) Delete(ctx context.Context, endpoint string) (*http.Response, error) {
	return c.doRequest(ctx, "DELETE", endpoint, nil)
}

func (c *ClicksignClient) Patch(ctx context.Context, endpoint string, body interface{}) (*http.Response, error) {
	return c.doRequest(ctx, "PATCH", endpoint, body)
}

func (c *ClicksignClient) doRequest(ctx context.Context, method, endpoint string, body interface{}) (*http.Response, error) {
	url := fmt.Sprintf("%s%s", c.baseURL, endpoint)
	
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			c.logger.WithFields(logrus.Fields{
				"error":    err.Error(),
				"endpoint": endpoint,
				"method":   method,
			}).Error("Failed to marshal request body")
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		c.logger.WithFields(logrus.Fields{
			"error":    err.Error(),
			"endpoint": endpoint,
			"method":   method,
		}).Error("Failed to create HTTP request")
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	correlationID := ctx.Value("correlation_id")
	if correlationID != nil {
		req.Header.Set("X-Correlation-ID", correlationID.(string))
	}

	c.logger.WithFields(logrus.Fields{
		"method":         method,
		"url":            url,
		"correlation_id": correlationID,
	}).Info("Making HTTP request to Clicksign API")

	startTime := time.Now()
	resp, err := c.httpClient.Do(req)
	duration := time.Since(startTime)

	if err != nil {
		c.logger.WithFields(logrus.Fields{
			"error":          err.Error(),
			"method":         method,
			"url":            url,
			"duration_ms":    duration.Milliseconds(),
			"correlation_id": correlationID,
		}).Error("HTTP request failed")
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}

	c.logger.WithFields(logrus.Fields{
		"method":         method,
		"url":            url,
		"status_code":    resp.StatusCode,
		"duration_ms":    duration.Milliseconds(),
		"correlation_id": correlationID,
	}).Info("HTTP request completed")

	return resp, nil
}