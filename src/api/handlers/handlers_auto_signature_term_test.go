package handlers

import (
	"testing"

	"app/api/handlers/dtos"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAutoSignatureTermHandler_MapCreateRequestToEntity(t *testing.T) {
	handler := &AutoSignatureTermHandlers{}

	t.Run("should map valid alphanumeric cnpj without changing input", func(t *testing.T) {
		requestDTO := dtos.AutoSignatureTermCreateRequestDTO{
			Signer: dtos.SignerInfoDTO{
				Documentation: "12.abc.345/01de-35",
				Birthday:      "1990-01-01",
				Email:         "signer@example.com",
				Name:          "John Doe",
			},
			AdminEmail: "admin@example.com",
			APIEmail:   "api@example.com",
		}

		term, err := handler.mapCreateRequestToEntity(requestDTO)

		require.NoError(t, err)
		assert.Equal(t, "12.abc.345/01de-35", term.SignerDocumentation)
	})

	t.Run("should reject invalid alphanumeric cnpj", func(t *testing.T) {
		requestDTO := dtos.AutoSignatureTermCreateRequestDTO{
			Signer: dtos.SignerInfoDTO{
				Documentation: "12.ABC.345/01DE-36",
				Birthday:      "1990-01-01",
				Email:         "signer@example.com",
				Name:          "John Doe",
			},
			AdminEmail: "admin@example.com",
			APIEmail:   "api@example.com",
		}

		_, err := handler.mapCreateRequestToEntity(requestDTO)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "alphanumeric CNPJ")
	})
}
