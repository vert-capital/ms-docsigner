package brdoc

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidate(t *testing.T) {
	t.Run("should accept legacy cpf", func(t *testing.T) {
		result, err := Validate("123.456.789-01")

		require.NoError(t, err)
		assert.Equal(t, KindCPF, result.Kind)
		assert.Equal(t, "12345678901", result.Normalized)
	})

	t.Run("should accept legacy numeric cnpj", func(t *testing.T) {
		result, err := Validate("12.345.678/0001-95")

		require.NoError(t, err)
		assert.Equal(t, KindCNPJ, result.Kind)
		assert.Equal(t, "12345678000195", result.Normalized)
	})

	t.Run("should accept valid alphanumeric cnpj", func(t *testing.T) {
		result, err := Validate("12.ABC.345/01DE-35")

		require.NoError(t, err)
		assert.Equal(t, KindCNPJAlphanumeric, result.Kind)
		assert.Equal(t, "12ABC34501DE35", result.Normalized)
	})

	t.Run("should accept lowercase alphanumeric cnpj", func(t *testing.T) {
		result, err := Validate("12.abc.345/01de-35")

		require.NoError(t, err)
		assert.Equal(t, KindCNPJAlphanumeric, result.Kind)
		assert.Equal(t, "12ABC34501DE35", result.Normalized)
	})

	t.Run("should preserve leading zeroes during validation", func(t *testing.T) {
		validCNPJ := buildMaskedAlphanumericCNPJ("00ABC34501DE")

		result, err := Validate(validCNPJ)

		require.NoError(t, err)
		assert.Equal(t, KindCNPJAlphanumeric, result.Kind)
		assert.Equal(t, "00ABC34501DE56", result.Normalized)
	})

	t.Run("should reject invalid alphanumeric cnpj dv", func(t *testing.T) {
		_, err := Validate("12.ABC.345/01DE-36")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "alphanumeric CNPJ")
	})

	t.Run("should reject invalid symbols", func(t *testing.T) {
		_, err := Validate("12.ABC.345/01DE-3@")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "alphanumeric CNPJ")
	})
}

func buildMaskedAlphanumericCNPJ(base string) string {
	firstDigit := calculateCNPJDigit(base, []int{5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2})
	secondDigit := calculateCNPJDigit(base+fmt.Sprintf("%d", firstDigit), []int{6, 5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2})

	return fmt.Sprintf("%s.%s.%s/%s-%d%d", base[:2], base[2:5], base[5:8], base[8:12], firstDigit, secondDigit)
}
