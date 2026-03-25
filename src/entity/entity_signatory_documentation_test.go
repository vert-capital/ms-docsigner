package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSignatoryDocumentation(t *testing.T) {
	t.Run("should accept valid alphanumeric cnpj", func(t *testing.T) {
		documentation := "12.ABC.345/01DE-35"
		signatory := &EntitySignatory{
			Documentation: &documentation,
		}

		err := signatory.validateDocumentation()

		require.NoError(t, err)
		assert.Equal(t, "12.ABC.345/01DE-35", *signatory.Documentation)
	})

	t.Run("should accept legacy cpf compatibility", func(t *testing.T) {
		documentation := "123.456.789-01"
		signatory := &EntitySignatory{
			Documentation: &documentation,
		}

		err := signatory.validateDocumentation()

		require.NoError(t, err)
	})

	t.Run("should reject invalid alphanumeric cnpj dv", func(t *testing.T) {
		documentation := "12.ABC.345/01DE-36"
		signatory := &EntitySignatory{
			Documentation: &documentation,
		}

		err := signatory.validateDocumentation()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "alphanumeric CNPJ")
	})

	t.Run("should preserve original value on set documentation", func(t *testing.T) {
		signatory := &EntitySignatory{}

		err := signatory.SetDocumentation("12.abc.345/01de-35")

		require.NoError(t, err)
		require.NotNil(t, signatory.Documentation)
		assert.Equal(t, "12.abc.345/01de-35", *signatory.Documentation)
	})
}
