package provider

// SignerData representa os dados necessários para criar um signatário
// Esta estrutura é genérica e pode ser mapeada para o formato específico de cada provider
type SignerData struct {
	Name              string
	Email             string
	Birthday          string
	Documentation     *string
	PhoneNumber       *string
	HasDocumentation  bool
	Refusable         bool
	Group             int
	CommunicateEvents *SignerCommunicateEventsData
}

// SignerCommunicateEventsData representa as configurações de comunicação para signatários
type SignerCommunicateEventsData struct {
	DocumentSigned    string
	SignatureRequest  string
	SignatureReminder string
}

// RequirementData representa os dados necessários para criar um requisito
// Esta estrutura é genérica e pode ser mapeada para o formato específico de cada provider
type RequirementData struct {
	Action     string // "agree", "sign", "provide_evidence"
	Role       string // "sign" para qualificação
	Auth       string // "email", "icp_brasil", "auto_signature" para autenticação
	DocumentID string // ID do documento relacionado no provider
	SignerID   string // ID do signatário relacionado no provider
}



