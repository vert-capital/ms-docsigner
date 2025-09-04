package dto

// AutoSignatureTermRequest representa o request para criação de termo de assinatura automática
type AutoSignatureTermRequest struct {
	Data AutoSignatureTermData `json:"data"`
}

// AutoSignatureTermData representa os dados do termo
type AutoSignatureTermData struct {
	Type       string                      `json:"type"`
	Attributes AutoSignatureTermAttributes `json:"attributes"`
}

// AutoSignatureTermAttributes representa os atributos do termo
type AutoSignatureTermAttributes struct {
	Signer     SignerInfo `json:"signer"`
	AdminEmail string     `json:"admin_email"`
	APIEmail   string     `json:"api_email"`
}

// SignerInfo representa as informações do signatário
type SignerInfo struct {
	Documentation string `json:"documentation"`
	Birthday      string `json:"birthday"`
	Email         string `json:"email"`
	Name          string `json:"name"`
}

// AutoSignatureTermResponse representa a resposta do Clicksign para criação de termo
type AutoSignatureTermResponse struct {
	Data AutoSignatureTermResponseData `json:"data"`
}

// AutoSignatureTermResponseData representa os dados da resposta
type AutoSignatureTermResponseData struct {
	ID         string                      `json:"id"`
	Type       string                      `json:"type"`
	Attributes AutoSignatureTermAttributes `json:"attributes"`
}
