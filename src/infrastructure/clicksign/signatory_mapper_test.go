package clicksign

import (
	"testing"

	"app/entity"
)

func TestSignatoryMapper_NormalizeNameForClicksign(t *testing.T) {
	mapper := NewSignatoryMapper()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Nome com sobrenome",
			input:    "João Silva",
			expected: "João Silva",
		},
		{
			name:     "Nome sem sobrenome",
			input:    "Belagrícola",
			expected: "Belagrícola N/A",
		},
		{
			name:     "Nome vazio",
			input:    "",
			expected: "Nome Não Informado",
		},
		{
			name:     "Nome com espaços extras",
			input:    "  Maria Santos  ",
			expected: "Maria Santos",
		},
		{
			name:     "Nome com múltiplos sobrenomes",
			input:    "João da Silva Santos",
			expected: "João da Silva Santos",
		},
		{
			name:     "Nome com caracteres especiais",
			input:    "José-Maria",
			expected: "José-Maria N/A",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapper.normalizeNameForClicksign(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeNameForClicksign(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSignatoryMapper_ToClicksignCreateRequest_WithNormalization(t *testing.T) {
	mapper := NewSignatoryMapper()

	// Criar uma entidade de teste com nome sem sobrenome
	signatory := &entity.EntitySignatory{
		Name:  "Belagrícola",
		Email: "test@example.com",
	}

	request := mapper.ToClicksignCreateRequest(signatory)

	expectedName := "Belagrícola N/A"
	if request.Data.Attributes.Name != expectedName {
		t.Errorf("Expected normalized name %q, got %q", expectedName, request.Data.Attributes.Name)
	}

	if request.Data.Attributes.Email != "test@example.com" {
		t.Errorf("Expected email %q, got %q", "test@example.com", request.Data.Attributes.Email)
	}
}
