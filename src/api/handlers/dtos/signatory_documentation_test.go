package dtos

import (
	"testing"

	"app/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSignatoryDocumentationValidation(t *testing.T) {
	t.Run("should accept alphanumeric cnpj on create without changing input", func(t *testing.T) {
		documentation := "12.abc.345/01de-35"
		dto := SignatoryCreateRequestDTO{
			Name:          "John Doe",
			Email:         "john.doe@example.com",
			EnvelopeID:    1,
			Documentation: &documentation,
		}

		err := dto.Validate()
		entityData := dto.ToEntity()

		require.NoError(t, err)
		require.NotNil(t, entityData.Documentation)
		assert.Equal(t, "12.abc.345/01de-35", *entityData.Documentation)
	})

	t.Run("should reject invalid alphanumeric cnpj on update", func(t *testing.T) {
		documentation := "12.ABC.345/01DE-36"
		dto := SignatoryUpdateRequestDTO{
			Documentation: &documentation,
		}

		err := dto.Validate()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "alphanumeric CNPJ")
	})

	t.Run("should apply original documentation to entity without normalization", func(t *testing.T) {
		documentation := "12.abc.345/01de-35"
		signatory := &entity.EntitySignatory{}
		dto := SignatoryUpdateRequestDTO{
			Documentation: &documentation,
		}

		dto.ApplyToEntity(signatory)

		require.NotNil(t, signatory.Documentation)
		assert.Equal(t, "12.abc.345/01de-35", *signatory.Documentation)
	})
}
