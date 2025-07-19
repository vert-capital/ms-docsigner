# Guia de Primeiros Passos

Este guia fornece um tutorial completo para começar a usar a API do ms-docsigner, desde a configuração inicial até o envio de seu primeiro documento para assinatura.

## Pré-requisitos

### 1. Acesso à API
- **Token JWT**: Necessário para autenticação em todos os endpoints
- **URL Base**: `https://api.ms-docsigner.com` (produção) ou `https://api-dev.ms-docsigner.com` (desenvolvimento)

### 2. Ferramentas Recomendadas
- **curl** ou **Postman** para testes
- Editor de texto para preparar payloads JSON
- Ferramenta para codificação base64 (se necessário)

### 3. Conhecimentos Básicos
- APIs REST e métodos HTTP
- Formato JSON
- Autenticação via Bearer Token

---

## Configuração Inicial

### 1. Teste de Conectividade

Primeiro, vamos verificar se você consegue acessar a API:

```bash
curl -X GET https://api.ms-docsigner.com/health \
  -H "Accept: application/json"
```

**Response esperado:**
```json
{
  "status": "ok",
  "timestamp": "2025-07-19T10:00:00Z"
}
```

### 2. Configuração da Autenticação

Todos os endpoints requerem autenticação via JWT. Configure seu token:

```bash
export API_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
export API_BASE_URL="https://api.ms-docsigner.com"
```

### 3. Teste de Autenticação

Verifique se seu token está funcionando:

```bash
curl -X GET $API_BASE_URL/api/v1/documents \
  -H "Authorization: Bearer $API_TOKEN" \
  -H "Content-Type: application/json"
```

---

## Tutorial Passo a Passo

### Cenário: Envio de Contrato para Assinatura

Vamos simular um cenário real onde você precisa enviar um contrato PDF para duas pessoas assinarem.

#### Passo 1: Preparar o Documento

Primeiro, você precisa converter seu arquivo PDF para base64:

```bash
# No Linux/Mac
base64 -i contrato.pdf > contrato_base64.txt

# No Windows (PowerShell)
[Convert]::ToBase64String([IO.File]::ReadAllBytes("contrato.pdf")) > contrato_base64.txt
```

#### Passo 2: Criar o Documento na API

```bash
curl -X POST $API_BASE_URL/api/v1/documents \
  -H "Authorization: Bearer $API_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Contrato de Prestação de Serviços - Cliente ABC",
    "file_content_base64": "'$(cat contrato_base64.txt)'",
    "description": "Contrato de desenvolvimento de software para cliente ABC"
  }'
```

**Response de sucesso:**
```json
{
  "id": 1,
  "name": "Contrato de Prestação de Serviços - Cliente ABC",
  "file_path": "/tmp/temp_12345.pdf",
  "file_size": 1048576,
  "mime_type": "application/pdf",
  "status": "draft",
  "clicksign_key": "",
  "description": "Contrato de desenvolvimento de software para cliente ABC",
  "created_at": "2025-07-19T10:00:00Z",
  "updated_at": "2025-07-19T10:00:00Z"
}
```

**💡 Dica:** Anote o `id` retornado (neste exemplo: `1`), você precisará dele no próximo passo.

#### Passo 3: Criar o Envelope

Agora vamos criar um envelope que associa o documento aos signatários:

```bash
curl -X POST $API_BASE_URL/api/v1/envelopes \
  -H "Authorization: Bearer $API_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Envelope - Contrato Cliente ABC",
    "description": "Contrato de prestação de serviços para assinatura",
    "documents_ids": [1],
    "signatory_emails": [
      "empresa@exemplo.com",
      "cliente@abc.com"
    ],
    "message": "Prezados, favor assinar o contrato de prestação de serviços conforme acordado. Em caso de dúvidas, entrem em contato.",
    "deadline_at": "2025-08-15T23:59:59Z",
    "remind_interval": 3,
    "auto_close": true
  }'
```

**Response de sucesso:**
```json
{
  "id": 123,
  "name": "Envelope - Contrato Cliente ABC",
  "description": "Contrato de prestação de serviços para assinatura",
  "status": "draft",
  "clicksign_key": "12345678-1234-1234-1234-123456789012",
  "documents_ids": [1],
  "signatory_emails": ["empresa@exemplo.com", "cliente@abc.com"],
  "message": "Prezados, favor assinar o contrato de prestação de serviços conforme acordado. Em caso de dúvidas, entrem em contato.",
  "deadline_at": "2025-08-15T23:59:59Z",
  "remind_interval": 3,
  "auto_close": true,
  "created_at": "2025-07-19T10:05:00Z",
  "updated_at": "2025-07-19T10:05:00Z"
}
```

**💡 Dica:** Anote o `id` do envelope (neste exemplo: `123`) e o `clicksign_key`.

#### Passo 4: Ativar o Envelope

Por segurança, envelopes são criados no status `draft`. Para iniciar o processo de assinatura, você precisa ativá-lo:

```bash
curl -X POST $API_BASE_URL/api/v1/envelopes/123/activate \
  -H "Authorization: Bearer $API_TOKEN" \
  -H "Content-Type: application/json"
```

**Response de sucesso:**
```json
{
  "id": 123,
  "name": "Envelope - Contrato Cliente ABC",
  "description": "Contrato de prestação de serviços para assinatura",
  "status": "running",
  "clicksign_key": "12345678-1234-1234-1234-123456789012",
  "documents_ids": [1],
  "signatory_emails": ["empresa@exemplo.com", "cliente@abc.com"],
  "message": "Prezados, favor assinar o contrato de prestação de serviços conforme acordado. Em caso de dúvidas, entrem em contato.",
  "deadline_at": "2025-08-15T23:59:59Z",
  "remind_interval": 3,
  "auto_close": true,
  "created_at": "2025-07-19T10:05:00Z",
  "updated_at": "2025-07-19T10:10:00Z"
}
```

**🎉 Sucesso!** O envelope agora está no status `running` e os signatários receberão e-mails para assinar o documento.

#### Passo 5: Monitorar o Status

Você pode acompanhar o progresso do envelope:

```bash
curl -X GET $API_BASE_URL/api/v1/envelopes/123 \
  -H "Authorization: Bearer $API_TOKEN"
```

---

## Fluxo Básico Completo

### Resumo dos Endpoints Utilizados

1. **POST /api/v1/documents** - Criar documento
2. **POST /api/v1/envelopes** - Criar envelope
3. **POST /api/v1/envelopes/{id}/activate** - Ativar envelope
4. **GET /api/v1/envelopes/{id}** - Monitorar status

### Script Bash Completo

Aqui está um script que automatiza todo o processo:

```bash
#!/bin/bash

# Configuração
API_TOKEN="seu-token-aqui"
API_BASE_URL="https://api.ms-docsigner.com"
DOCUMENT_FILE="contrato.pdf"

# Função para exibir erros
check_response() {
  if [ $? -ne 0 ]; then
    echo "❌ Erro na requisição"
    exit 1
  fi
}

echo "🚀 Iniciando processo de envio de documento..."

# 1. Converter arquivo para base64
echo "📄 Convertendo documento para base64..."
DOCUMENT_BASE64=$(base64 -i $DOCUMENT_FILE)

# 2. Criar documento
echo "📤 Criando documento na API..."
DOCUMENT_RESPONSE=$(curl -s -X POST $API_BASE_URL/api/v1/documents \
  -H "Authorization: Bearer $API_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Contrato de Prestação de Serviços",
    "file_content_base64": "'$DOCUMENT_BASE64'",
    "description": "Contrato para assinatura"
  }')

check_response
DOCUMENT_ID=$(echo $DOCUMENT_RESPONSE | jq -r '.id')
echo "✅ Documento criado com ID: $DOCUMENT_ID"

# 3. Criar envelope
echo "📧 Criando envelope..."
ENVELOPE_RESPONSE=$(curl -s -X POST $API_BASE_URL/api/v1/envelopes \
  -H "Authorization: Bearer $API_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Envelope - Contrato",
    "documents_ids": ['$DOCUMENT_ID'],
    "signatory_emails": ["empresa@exemplo.com", "cliente@exemplo.com"],
    "message": "Favor assinar o contrato conforme acordado.",
    "deadline_at": "2025-08-15T23:59:59Z"
  }')

check_response
ENVELOPE_ID=$(echo $ENVELOPE_RESPONSE | jq -r '.id')
echo "✅ Envelope criado com ID: $ENVELOPE_ID"

# 4. Ativar envelope
echo "🔥 Ativando envelope..."
curl -s -X POST $API_BASE_URL/api/v1/envelopes/$ENVELOPE_ID/activate \
  -H "Authorization: Bearer $API_TOKEN" > /dev/null

check_response
echo "✅ Envelope ativado com sucesso!"

echo "🎉 Processo concluído! Os signatários receberão e-mails em breve."
echo "📊 Para monitorar: curl -X GET $API_BASE_URL/api/v1/envelopes/$ENVELOPE_ID"
```

---

## Casos de Uso Comuns

### 1. Múltiplos Documentos em um Envelope

```bash
# Criar vários documentos
DOCUMENT_ID_1=1  # ID do primeiro documento
DOCUMENT_ID_2=2  # ID do segundo documento

# Criar envelope com múltiplos documentos
curl -X POST $API_BASE_URL/api/v1/envelopes \
  -H "Authorization: Bearer $API_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Contrato Completo - Múltiplos Anexos",
    "documents_ids": ['$DOCUMENT_ID_1', '$DOCUMENT_ID_2'],
    "signatory_emails": ["parte1@email.com", "parte2@email.com"]
  }'
```

### 2. Envelope com Prazo Urgente

```bash
# Prazo de 24 horas com lembretes a cada 6 horas
DEADLINE=$(date -d "+1 day" -u +"%Y-%m-%dT%H:%M:%SZ")

curl -X POST $API_BASE_URL/api/v1/envelopes \
  -H "Authorization: Bearer $API_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Documento Urgente",
    "documents_ids": [1],
    "signatory_emails": ["urgente@email.com"],
    "deadline_at": "'$DEADLINE'",
    "remind_interval": 1
  }'
```

### 3. Buscar Envelopes por Status

```bash
# Listar apenas envelopes ativos
curl -X GET "$API_BASE_URL/api/v1/envelopes?status=running" \
  -H "Authorization: Bearer $API_TOKEN"

# Buscar envelopes por nome
curl -X GET "$API_BASE_URL/api/v1/envelopes?search=contrato" \
  -H "Authorization: Bearer $API_TOKEN"
```

---

## Próximos Passos

Agora que você concluiu o tutorial básico, explore:

1. **[API de Documentos](documents.md)** - Documentação completa dos endpoints de documentos
2. **[API de Envelopes](clicksign-envelopes.md)** - Documentação completa dos endpoints de envelopes
3. **[Autenticação](authentication.md)** - Detalhes sobre sistema de autenticação
4. **[Exemplos Práticos](examples/)** - Mais casos de uso e exemplos
5. **[Error Handling](error-handling.md)** - Guia de troubleshooting

---

## Suporte

- **Documentação**: Consulte os arquivos específicos na pasta `/docs/api/`
- **Logs**: Use o header `X-Correlation-ID` para rastreamento
- **Status da API**: Verifique `/health` para status do serviço

**Boa sorte com sua integração! 🚀**