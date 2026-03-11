package dtos

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnvelopeSignatoryRequest_ResolveAuthMethod(t *testing.T) {
	t.Run("should resolve auth_method when present", func(t *testing.T) {
		authMethod := "auto_signature"

		signatory := EnvelopeSignatoryRequest{
			Email:      "assinante@empresa.com",
			AuthMethod: &authMethod,
		}

		resolved, err := signatory.ResolveAuthMethod()

		require.NoError(t, err)
		assert.Equal(t, "auto_signature", resolved)
	})

	t.Run("should default to email when auth_method is omitted", func(t *testing.T) {
		signatory := EnvelopeSignatoryRequest{
			Email: "assinante@empresa.com",
		}

		resolved, err := signatory.ResolveAuthMethod()

		require.NoError(t, err)
		assert.Equal(t, "email", resolved)
	})
}

func TestEnvelopeV2CreateRequestDTO_Validate_VertSignAuthRules(t *testing.T) {
	t.Run("should accept auto_signature for vert-sign signatory", func(t *testing.T) {
		authMethod := "auto_signature"

		request := EnvelopeV2CreateRequestDTO{
			Provider: "vert-sign",
			Name:     "Envelope VertSign Auto Signature",
			Documents: []EnvelopeDocumentRequest{
				{
					Name:    "Contrato.pdf",
					FileURL: "https://example.com/contrato.pdf",
				},
			},
			Signatories: []EnvelopeSignatoryRequest{
				{
					Name:       "Assinante",
					Email:      "assinante@empresa.com",
					AuthMethod: &authMethod,
				},
			},
		}

		err := request.Validate()

		require.NoError(t, err)
	})

	t.Run("should reject icp_brasil for vert-sign signatory", func(t *testing.T) {
		authMethod := "icp_brasil"

		request := EnvelopeV2CreateRequestDTO{
			Provider: "vert-sign",
			Name:     "Envelope VertSign ICP",
			Documents: []EnvelopeDocumentRequest{
				{
					Name:    "Contrato.pdf",
					FileURL: "https://example.com/contrato.pdf",
				},
			},
			Signatories: []EnvelopeSignatoryRequest{
				{
					Name:       "Assinante",
					Email:      "assinante@empresa.com",
					AuthMethod: &authMethod,
				},
			},
		}

		err := request.Validate()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "não suporta")
	})
}
