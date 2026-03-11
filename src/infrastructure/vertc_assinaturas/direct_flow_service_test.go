package vertc_assinaturas

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"app/entity"
	"app/infrastructure/provider"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDirectFlowService_CreateEnvelopeWithDocumentsAndSigners(t *testing.T) {
	t.Run("should create auto-signature signer without required methods", func(t *testing.T) {
		var receivedSigners []directSignerRequest
		var sendRequest directSendRequest
		documentsUploaded := 0

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/api/v1/auth/login":
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"access_token":"token-1"}`))
			case "/api/v1/envelopes":
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte(`{"id":"env-123","name":"Envelope Teste","status":"draft"}`))
			case "/api/v1/documents/env-123":
				assert.Contains(t, r.Header.Get("Content-Type"), "multipart/form-data")
				documentsUploaded++
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte(`{"files":[{"id":"doc-1"}]}`))
			case "/api/v1/signers":
				require.NoError(t, json.NewDecoder(r.Body).Decode(&receivedSigners))
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte(`[{"id":"sig-1"},{"id":"sig-2"}]`))
			case "/api/v1/envelopes/env-123/send":
				require.NoError(t, json.NewDecoder(r.Body).Decode(&sendRequest))
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"status":"processing"}`))
			default:
				http.NotFound(w, r)
			}
		}))
		defer server.Close()

		tempFile, err := os.CreateTemp("", "vertsign-direct-flow-*.pdf")
		require.NoError(t, err)
		defer func() {
			_ = os.Remove(tempFile.Name())
		}()

		_, err = tempFile.Write([]byte("%PDF-1.4 test"))
		require.NoError(t, err)
		require.NoError(t, tempFile.Close())

		client := &VertcAssinaturasClient{
			httpClient: server.Client(),
			baseURL:    server.URL,
			email:      "service@vert.com",
			password:   "secret",
			logger:     logrus.New(),
		}

		service := NewDirectFlowService(client, logrus.New())
		deadline := time.Now().Add(24 * time.Hour)

		result, err := service.CreateEnvelopeWithDocumentsAndSigners(context.Background(), QuickSendData{
			Envelope: &entity.EntityEnvelope{
				Name:        "Envelope Teste",
				Description: "Mensagem do envelope",
				DeadlineAt:  &deadline,
			},
			Documents: []*entity.EntityDocument{
				{
					Name:         "Contrato.pdf",
					FilePath:     tempFile.Name(),
					FileSize:     13,
					MimeType:     "application/pdf",
					IsFromBase64: true,
				},
			},
			Signers: []provider.SignerData{
				{
					Name:       "Assinante Auto",
					Email:      "auto@empresa.com",
					Refusable:  false,
					AuthMethod: "auto_signature",
				},
				{
					Name:       "Assinante Email",
					Email:      "email@empresa.com",
					Refusable:  false,
					AuthMethod: "email",
				},
			},
		})

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "env-123", result.EnvelopeID)
		assert.Equal(t, "sent", result.EnvelopeStatus)
		assert.Equal(t, 1, result.DocumentsUploaded)
		assert.Equal(t, 2, result.SignersCreated)
		assert.True(t, result.NotificationTriggered)

		require.Len(t, receivedSigners, 2)
		assert.Equal(t, "auto@empresa.com", receivedSigners[0].Email)
		assert.Nil(t, receivedSigners[0].RequiredMethods)
		assert.Equal(t, "email@empresa.com", receivedSigners[1].Email)
		assert.Equal(t, []string{"code_email"}, receivedSigners[1].RequiredMethods)
		assert.Equal(t, "Envelope Teste", sendRequest.Subject)
		assert.Equal(t, "Mensagem do envelope", sendRequest.Message)
		assert.Equal(t, 1, documentsUploaded)
	})

	t.Run("should fail for unsupported auth method", func(t *testing.T) {
		client := &VertcAssinaturasClient{
			httpClient: http.DefaultClient,
			baseURL:    "http://localhost",
			email:      "service@vert.com",
			password:   "secret",
			logger:     logrus.New(),
		}

		service := NewDirectFlowService(client, logrus.New())

		_, err := service.createSigners(context.Background(), "env-123", []provider.SignerData{
			{
				Name:       "Assinante ICP",
				Email:      "icp@empresa.com",
				AuthMethod: "icp_brasil",
			},
		})

		require.Error(t, err)
		assert.True(t, strings.Contains(err.Error(), "unsupported auth method"))
	})
}
