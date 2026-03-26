package dtos

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnvelopeDocumentationValidation(t *testing.T) {
	t.Run("should validate embedded signatory documentation on v1", func(t *testing.T) {
		documentation := "12.ABC.345/01DE-36"
		request := EnvelopeCreateRequestDTO{
			Name: "Envelope Test",
			Documents: []EnvelopeDocumentRequest{
				{
					Name:    "Contrato.pdf",
					FileURL: "https://example.com/contrato.pdf",
				},
			},
			Signatories: []EnvelopeSignatoryRequest{
				{
					Name:          "Assinante",
					Email:         "assinante@empresa.com",
					Documentation: &documentation,
				},
			},
		}

		err := request.Validate()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "alphanumeric CNPJ")
	})

	t.Run("should validate embedded signatory documentation on v2", func(t *testing.T) {
		documentation := "12.abc.345/01de-35"
		request := EnvelopeV2CreateRequestDTO{
			Provider: "clicksign",
			Name:     "Envelope Test",
			Documents: []EnvelopeDocumentRequest{
				{
					Name:    "Contrato.pdf",
					FileURL: "https://example.com/contrato.pdf",
				},
			},
			Signatories: []EnvelopeSignatoryRequest{
				{
					Name:          "Assinante",
					Email:         "assinante@empresa.com",
					Documentation: &documentation,
				},
			},
		}

		err := request.Validate()

		require.NoError(t, err)
	})
}
