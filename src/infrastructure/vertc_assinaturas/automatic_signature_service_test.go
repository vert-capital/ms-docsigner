package vertc_assinaturas

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAutomaticSignatureService_CheckSignedTermByEmail(t *testing.T) {
	t.Run("should return signed=true when permission is active and signed", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/api/v1/auth/login":
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"access_token":"token-1"}`))
			case "/api/v1/automatic-signature":
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`[
					{"id":"perm-1","recipientUser":{"email":"investidor@exemplo.com"},"contractStatus":"signed","isActive":true}
				]`))
			default:
				http.NotFound(w, r)
			}
		}))
		defer server.Close()

		client := &VertcAssinaturasClient{
			httpClient: server.Client(),
			baseURL:    server.URL,
			email:      "service@vert.com",
			password:   "secret",
			logger:     logrus.New(),
		}

		service := NewAutomaticSignatureService(client, logrus.New())
		result, err := service.CheckSignedTermByEmail(context.Background(), "investidor@exemplo.com")

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.HasSignedTerm)
		assert.True(t, result.PermissionFound)
		assert.Equal(t, "perm-1", result.PermissionID)
		assert.Equal(t, "signed", result.ContractStatus)
		require.NotNil(t, result.IsActive)
		assert.True(t, *result.IsActive)
	})

	t.Run("should return permission found but signed=false for pending term", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/api/v1/auth/login":
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"access_token":"token-1"}`))
			case "/api/v1/automatic-signature":
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{
					"data":[
						{"id":"perm-2","recipient_user":{"email":"investidor@exemplo.com"},"contract_status":"pending_signature","is_active":false}
					]
				}`))
			default:
				http.NotFound(w, r)
			}
		}))
		defer server.Close()

		client := &VertcAssinaturasClient{
			httpClient: server.Client(),
			baseURL:    server.URL,
			email:      "service@vert.com",
			password:   "secret",
			logger:     logrus.New(),
		}

		service := NewAutomaticSignatureService(client, logrus.New())
		result, err := service.CheckSignedTermByEmail(context.Background(), "investidor@exemplo.com")

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.HasSignedTerm)
		assert.True(t, result.PermissionFound)
		assert.Equal(t, "perm-2", result.PermissionID)
		assert.Equal(t, "pending_signature", result.ContractStatus)
		require.NotNil(t, result.IsActive)
		assert.False(t, *result.IsActive)
	})

	t.Run("should return permission found false when no permission exists for e-mail", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/api/v1/auth/login":
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"access_token":"token-1"}`))
			case "/api/v1/automatic-signature":
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`[]`))
			default:
				http.NotFound(w, r)
			}
		}))
		defer server.Close()

		client := &VertcAssinaturasClient{
			httpClient: server.Client(),
			baseURL:    server.URL,
			email:      "service@vert.com",
			password:   "secret",
			logger:     logrus.New(),
		}

		service := NewAutomaticSignatureService(client, logrus.New())
		result, err := service.CheckSignedTermByEmail(context.Background(), "naoexiste@exemplo.com")

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.HasSignedTerm)
		assert.False(t, result.PermissionFound)
		assert.Empty(t, result.PermissionID)
		assert.Nil(t, result.IsActive)
	})
}

func TestAutomaticSignatureService_CreateTermEnsuringUser(t *testing.T) {
	t.Run("should create permission without creating user when user already exists", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/api/v1/auth/login":
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"access_token":"token-1"}`))
			case "/api/v1/users":
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`[
					{"id":"user-1","email":"assinante@empresa.com","name":"Assinante"}
				]`))
			case "/api/v1/automatic-signature":
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte(`{
						"id":"perm-123",
						"envelopeId":"env-123",
						"contractStatus":"pending_signature",
						"isActive":false,
						"recipientUser":{"email":"assinante@empresa.com"}
					}`))
			case "/api/v1/documents/env-123":
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte(`{"files":[]}`))
			case "/api/v1/envelopes/env-123/send":
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"status":"processing"}`))
			default:
				http.NotFound(w, r)
			}
		}))
		defer server.Close()

		client := &VertcAssinaturasClient{
			httpClient: server.Client(),
			baseURL:    server.URL,
			email:      "service@vert.com",
			password:   "secret",
			logger:     logrus.New(),
		}

		service := NewAutomaticSignatureService(client, logrus.New())
		result, err := service.CreateTermEnsuringUser(context.Background(), "assinante@empresa.com", "Assinante")

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "perm-123", result.PermissionID)
		assert.Equal(t, "env-123", result.EnvelopeID)
		assert.Equal(t, "pending_signature", result.ContractStatus)
		assert.True(t, result.NotificationSent)
		assert.False(t, result.UserCreated)
		assert.True(t, result.UserExisted)
		require.NotNil(t, result.IsActive)
		assert.False(t, *result.IsActive)
	})

	t.Run("should create user first when not found in users list", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/api/v1/auth/login":
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"access_token":"token-1"}`))
			case "/api/v1/users":
				if r.Method == http.MethodGet {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(`[]`))
					return
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte(`{"id":"user-new","email":"novo@empresa.com","name":"Novo"}`))
			case "/api/v1/automatic-signature":
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte(`{
						"id":"perm-new",
						"envelopeId":"env-new",
						"contractStatus":"pending_signature",
						"isActive":false,
						"recipientUser":{"email":"novo@empresa.com"}
					}`))
			case "/api/v1/documents/env-new":
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte(`{"files":[]}`))
			case "/api/v1/envelopes/env-new/send":
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"status":"processing"}`))
			default:
				http.NotFound(w, r)
			}
		}))
		defer server.Close()

		client := &VertcAssinaturasClient{
			httpClient: server.Client(),
			baseURL:    server.URL,
			email:      "service@vert.com",
			password:   "secret",
			logger:     logrus.New(),
		}

		service := NewAutomaticSignatureService(client, logrus.New())
		result, err := service.CreateTermEnsuringUser(context.Background(), "novo@empresa.com", "Novo")

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "perm-new", result.PermissionID)
		assert.Equal(t, "env-new", result.EnvelopeID)
		assert.True(t, result.NotificationSent)
		assert.True(t, result.UserCreated)
		assert.False(t, result.UserExisted)
		require.NotNil(t, result.IsActive)
		assert.False(t, *result.IsActive)
	})

	t.Run("should keep permission creation successful when send fails due to missing documents", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/api/v1/auth/login":
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"access_token":"token-1"}`))
			case "/api/v1/users":
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`[
					{"id":"user-1","email":"semdoc@empresa.com","name":"Sem Documento"}
				]`))
			case "/api/v1/automatic-signature":
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte(`{
						"id":"perm-semdoc",
						"envelopeId":"env-semdoc",
						"contractStatus":"pending_signature",
						"isActive":false,
						"recipientUser":{"email":"semdoc@empresa.com"}
					}`))
			case "/api/v1/documents/env-semdoc":
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte(`{"files":[]}`))
			case "/api/v1/envelopes/env-semdoc/send":
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte(`{"message":"Nenhum documento encontrado para este envelope","error":"Not Found","statusCode":404}`))
			default:
				http.NotFound(w, r)
			}
		}))
		defer server.Close()

		client := &VertcAssinaturasClient{
			httpClient: server.Client(),
			baseURL:    server.URL,
			email:      "service@vert.com",
			password:   "secret",
			logger:     logrus.New(),
		}

		service := NewAutomaticSignatureService(client, logrus.New())
		result, err := service.CreateTermEnsuringUser(context.Background(), "semdoc@empresa.com", "Sem Documento")

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "perm-semdoc", result.PermissionID)
		assert.Equal(t, "env-semdoc", result.EnvelopeID)
		assert.False(t, result.NotificationSent)
		require.NotNil(t, result.NotificationError)
		assert.Contains(t, *result.NotificationError, "sem documentos")
	})
}
