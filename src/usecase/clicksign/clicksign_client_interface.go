package clicksign

import (
	"context"
	"net/http"
)

type ClicksignClientInterface interface {
	Get(ctx context.Context, endpoint string) (*http.Response, error)
	Post(ctx context.Context, endpoint string, body interface{}) (*http.Response, error)
	Put(ctx context.Context, endpoint string, body interface{}) (*http.Response, error)
	Delete(ctx context.Context, endpoint string) (*http.Response, error)
	Patch(ctx context.Context, endpoint string, body interface{}) (*http.Response, error)
}
