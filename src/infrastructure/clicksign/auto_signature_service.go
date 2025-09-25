package clicksign

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"app/infrastructure/clicksign/dto"

	"github.com/sirupsen/logrus"
)

type ClicksignClientInterface interface {
	Get(ctx context.Context, endpoint string) (*http.Response, error)
	Post(ctx context.Context, endpoint string, body interface{}) (*http.Response, error)
	Put(ctx context.Context, endpoint string, body interface{}) (*http.Response, error)
	Delete(ctx context.Context, endpoint string) (*http.Response, error)
	Patch(ctx context.Context, endpoint string, body interface{}) (*http.Response, error)
}

// AutoSignatureService gerencia operações relacionadas ao termo de assinatura automática
type AutoSignatureService struct {
	client ClicksignClientInterface
	logger *logrus.Logger
}

// NewAutoSignatureService cria uma nova instância do AutoSignatureService
func NewAutoSignatureService(client ClicksignClientInterface, logger *logrus.Logger) *AutoSignatureService {
	return &AutoSignatureService{
		client: client,
		logger: logger,
	}
}

// CreateAutoSignatureTerm cria um termo de assinatura automática no Clicksign
func (s *AutoSignatureService) CreateAutoSignatureTerm(request dto.AutoSignatureTermRequest) (*dto.AutoSignatureTermResponse, error) {
	ctx := context.Background()

	resp, err := s.client.Post(ctx, "/api/v3/auto_signature/terms", request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, &ClicksignError{
			Type:       s.categorizeHTTPError(resp.StatusCode),
			Message:    fmt.Sprintf("failed to create auto signature term, status: %d, body: %s", resp.StatusCode, string(bodyBytes)),
			StatusCode: resp.StatusCode,
		}
	}

	var response dto.AutoSignatureTermResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, &ClicksignError{
			Type:     ErrorTypeSerialization,
			Message:  "failed to decode response",
			Original: err,
		}
	}

	return &response, nil
}

// categorizeHTTPError categoriza erros baseados no status code HTTP
func (s *AutoSignatureService) categorizeHTTPError(statusCode int) string {
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
