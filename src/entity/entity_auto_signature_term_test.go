package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAutoSignatureTermDocumentation(t *testing.T) {
	t.Run("should accept valid alphanumeric cnpj", func(t *testing.T) {
		term, err := NewAutoSignatureTerm(EntityAutoSignatureTerm{
			SignerDocumentation: "12.abc.345/01de-35",
			SignerBirthday:      "1990-01-01",
			SignerEmail:         "signer@example.com",
			SignerName:          "John Doe",
			AdminEmail:          "admin@example.com",
			APIEmail:            "api@example.com",
		})

		require.NoError(t, err)
		assert.Equal(t, "12.abc.345/01de-35", term.SignerDocumentation)
	})

	t.Run("should reject invalid alphanumeric cnpj", func(t *testing.T) {
		_, err := NewAutoSignatureTerm(EntityAutoSignatureTerm{
			SignerDocumentation: "12.ABC.345/01DE-36",
			SignerBirthday:      "1990-01-01",
			SignerEmail:         "signer@example.com",
			SignerName:          "John Doe",
			AdminEmail:          "admin@example.com",
			APIEmail:            "api@example.com",
		})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "alphanumeric CNPJ")
	})
}
