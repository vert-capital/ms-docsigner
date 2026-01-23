package provider_factory

import (
	"fmt"

	"app/config"
	"app/infrastructure/clicksign"
	"app/infrastructure/clicksign_provider"
	"app/infrastructure/provider"

	"github.com/sirupsen/logrus"
)

// ProviderFactory cria instâncias de EnvelopeProvider baseado no nome do provider
type ProviderFactory struct {
	clicksignClient clicksign.ClicksignClientInterface
	logger          *logrus.Logger
}

// NewProviderFactory cria uma nova instância do ProviderFactory
func NewProviderFactory(envVars config.EnvironmentVars, logger *logrus.Logger) *ProviderFactory {
	clicksignClient := clicksign.NewClicksignClient(envVars, logger)

	return &ProviderFactory{
		clicksignClient: clicksignClient,
		logger:          logger,
	}
}

// GetProvider retorna uma instância do provider baseado no nome
// Retorna erro se o provider não for suportado ou não estiver implementado
func (f *ProviderFactory) GetProvider(providerName string) (provider.EnvelopeProvider, error) {
	switch providerName {
	case "clicksign":
		return clicksign_provider.NewClicksignProvider(f.clicksignClient, f.logger), nil
	case "vertc-assinaturas":
		// Provider ainda não implementado
		return nil, fmt.Errorf("provider '%s' is not yet implemented", providerName)
	default:
		return nil, fmt.Errorf("unsupported provider: '%s'. Supported providers: clicksign, vertc-assinaturas", providerName)
	}
}

// IsProviderSupported verifica se um provider é suportado (mesmo que não implementado)
func (f *ProviderFactory) IsProviderSupported(providerName string) bool {
	return providerName == "clicksign" || providerName == "vertc-assinaturas"
}

// IsProviderImplemented verifica se um provider está implementado
func (f *ProviderFactory) IsProviderImplemented(providerName string) bool {
	return providerName == "clicksign"
}



