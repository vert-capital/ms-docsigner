package brdoc

import (
	"fmt"
	"strings"
	"unicode"
)

type Kind string

const (
	KindUnknown          Kind = "unknown"
	KindCPF              Kind = "cpf"
	KindCNPJ             Kind = "cnpj"
	KindCNPJAlphanumeric Kind = "cnpj_alphanumeric"
)

type ValidationResult struct {
	Kind       Kind
	Normalized string
}

func Validate(value string) (*ValidationResult, error) {
	normalized, err := normalizeForValidation(value)
	if err != nil {
		return nil, err
	}

	switch {
	case len(normalized) == 11:
		if !isDigitsOnly(normalized) {
			return nil, invalidDocumentationError(value)
		}

		return &ValidationResult{
			Kind:       KindCPF,
			Normalized: normalized,
		}, nil
	case len(normalized) == 14 && isDigitsOnly(normalized):
		return &ValidationResult{
			Kind:       KindCNPJ,
			Normalized: normalized,
		}, nil
	case len(normalized) == 14:
		if !isValidAlphanumericCNPJFormat(normalized) {
			return nil, invalidDocumentationError(value)
		}

		if !validateAlphanumericCNPJDV(normalized) {
			return nil, invalidDocumentationError(value)
		}

		return &ValidationResult{
			Kind:       KindCNPJAlphanumeric,
			Normalized: normalized,
		}, nil
	default:
		return nil, invalidDocumentationError(value)
	}
}

func normalizeForValidation(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", invalidDocumentationError(value)
	}

	var builder strings.Builder
	builder.Grow(len(trimmed))

	for _, r := range trimmed {
		switch {
		case unicode.IsSpace(r), r == '.', r == '/', r == '-':
			continue
		case r >= '0' && r <= '9':
			builder.WriteRune(r)
		case unicode.IsLetter(r):
			builder.WriteRune(unicode.ToUpper(r))
		default:
			return "", invalidDocumentationError(value)
		}
	}

	normalized := builder.String()
	if normalized == "" {
		return "", invalidDocumentationError(value)
	}

	return normalized, nil
}

func isValidAlphanumericCNPJFormat(value string) bool {
	if len(value) != 14 {
		return false
	}

	hasLetter := false
	for i, r := range value {
		switch {
		case i < 12 && r >= '0' && r <= '9':
		case i < 12 && r >= 'A' && r <= 'Z':
			hasLetter = true
		case i >= 12 && r >= '0' && r <= '9':
		default:
			return false
		}
	}

	return hasLetter
}

func validateAlphanumericCNPJDV(value string) bool {
	firstDigit := calculateCNPJDigit(value[:12], []int{5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2})
	secondDigit := calculateCNPJDigit(value[:12]+string(rune('0'+firstDigit)), []int{6, 5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2})

	return int(value[12]-'0') == firstDigit && int(value[13]-'0') == secondDigit
}

func calculateCNPJDigit(value string, weights []int) int {
	sum := 0
	for i, r := range value {
		sum += cnpjCharValue(r) * weights[i]
	}

	remainder := sum % 11
	if remainder == 0 || remainder == 1 {
		return 0
	}

	return 11 - remainder
}

func cnpjCharValue(r rune) int {
	if r >= '0' && r <= '9' {
		return int(r - '0')
	}

	return int(r) - 48
}

func isDigitsOnly(value string) bool {
	for _, r := range value {
		if r < '0' || r > '9' {
			return false
		}
	}

	return true
}

func invalidDocumentationError(value string) error {
	return fmt.Errorf("documentation must be a valid CPF, numeric CNPJ, or alphanumeric CNPJ, got: %s", value)
}
