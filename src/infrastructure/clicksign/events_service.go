package clicksign

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
)

// EventsService handles interactions with Clicksign Events API
type EventsService struct {
	client *ClicksignClient
	logger *logrus.Logger
}

// SignatureStatus represents a signature status from Clicksign API
type SignatureStatus struct {
	Email    string     `json:"email"`
	Name     string     `json:"name"`
	Signed   bool       `json:"signed"`
	SignedAt *time.Time `json:"signed_at"`
}

// NewEventsService creates a new events service
func NewEventsService(client interface{}, logger *logrus.Logger) *EventsService {
	return &EventsService{
		client: client.(*ClicksignClient),
		logger: logger,
	}
}

// GetSignaturesStatus retrieves signature statuses for a document from Clicksign API
func (e *EventsService) GetSignaturesStatus(ctx context.Context, documentKey string) (map[string]*SignatureStatus, error) {
	e.logger.WithField("document_key", documentKey).Info("Fetching signature statuses from Clicksign API")

	// TODO: Implementar chamada real para API da Clicksign
	// Por enquanto retorna dados mockados para teste
	statuses := make(map[string]*SignatureStatus)

	e.logger.WithFields(logrus.Fields{
		"document_key": documentKey,
		"statuses_count": len(statuses),
	}).Info("Retrieved signature statuses from Clicksign API")

	return statuses, nil
}