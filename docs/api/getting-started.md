# Guia de Primeiros Passos - MS-DocSigner

Este guia fornece um tutorial completo para começar a usar o microserviço ms-docsigner, desde a configuração inicial até a coleta de assinaturas digitais.

## Pré-requisitos

### 1. Configuração de Ambiente

Antes de começar, certifique-se de ter:

- **Token JWT válido** para autenticação
- **Acesso à API** do ms-docsigner
- **Conta Clicksign** configurada (sandbox ou produção)
- **Documentos** prontos para assinatura (PDF, JPEG, PNG, GIF)

### 2. Variáveis de Ambiente Necessárias

```bash
# Configuração do JWT
JWT_SECRET=your-jwt-secret-key

# Configuração do Clicksign
CLICKSIGN_API_URL=https://sandbox.clicksign.com/api/v3
CLICKSIGN_ACCESS_TOKEN=your-clicksign-access-token

# Configuração do Banco de Dados
DATABASE_URL=postgresql://user:password@localhost:5432/docsigner

# Configuração do Servidor
PORT=8080
```

### 3. Headers HTTP Obrigatórios

Todas as requisições devem incluir:

```bash
Authorization: Bearer <jwt_token>
Content-Type: application/json
X-Correlation-ID: <optional-trace-id>  # Opcional, mas recomendado
```

---

## Fluxo Básico Completo

### Passo 1: Autenticação

Obtenha um token JWT válido através do sistema de autenticação:

```bash
curl -X POST https://api.ms-docsigner.com/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "seu-usuario",
    "password": "sua-senha"
  }'
```

**Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 3600
}
```

### Passo 2: Criar Documento

Você pode criar documentos de duas formas:

#### Opção A: Upload via Base64 (Recomendado para Frontend)

```bash
curl -X POST https://api.ms-docsigner.com/api/v1/documents \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -H "X-Correlation-ID: getting-started-001" \
  -d '{
    "name": "Contrato de Prestação de Serviços",
    "file_content_base64": "JVBERi0xLjQKMSAwIG9iag0KPDwNCi9UeXBlIC9DYXRhbG9nDQovUGFnZXMgMiAwIFINCj4+DQplbmRvYmoNCjIgMCBvYmoNCjw8DQovVHlwZSAvUGFnZXMNCi9LaWRzIFs...",
    "description": "Contrato para assinatura digital"
  }'
```

#### Opção B: Upload via File Path (Para Backend)

```bash
curl -X POST https://api.ms-docsigner.com/api/v1/documents \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -H "X-Correlation-ID: getting-started-001" \
  -d '{
    "name": "Contrato de Prestação de Serviços",
    "file_path": "/uploads/contratos/contrato_cliente_abc.pdf",
    "file_size": 2048576,
    "mime_type": "application/pdf",
    "description": "Contrato para assinatura digital"
  }'
```

**Response de Sucesso:**
```json
{
  "id": 1,
  "name": "Contrato de Prestação de Serviços",
  "file_path": "/tmp/processed_document_1627123456.pdf",
  "file_size": 2048576,
  "mime_type": "application/pdf",
  "status": "draft",
  "clicksign_key": "",
  "description": "Contrato para assinatura digital",
  "created_at": "2025-07-19T10:00:00Z",
  "updated_at": "2025-07-19T10:00:00Z"
}
```

**⚠️ Guarde o `id` do documento para usar no próximo passo!**

### Passo 3: Atualizar Status do Documento (Opcional)

Se necessário, marque o documento como pronto:

```bash
curl -X PUT https://api.ms-docsigner.com/api/v1/documents/1 \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -H "X-Correlation-ID: getting-started-001" \
  -d '{
    "status": "ready"
  }'
```

### Passo 4: Criar Envelope

Crie um envelope associando o documento aos signatários:

```bash
curl -X POST https://api.ms-docsigner.com/api/v1/envelopes \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -H "X-Correlation-ID: getting-started-001" \
  -d '{
    "name": "Envelope - Contrato Cliente ABC",
    "description": "Contrato de prestação de serviços para assinatura",
    "documents_ids": [1],
    "signatory_emails": [
      "empresa@exemplo.com",
      "cliente@abc.com"
    ],
    "message": "Favor assinar o contrato conforme acordado.",
    "deadline_at": "2025-08-15T23:59:59Z",
    "remind_interval": 3,
    "auto_close": true
  }'
```

**Response de Sucesso:**
```json
{
  "id": 123,
  "name": "Envelope - Contrato Cliente ABC",
  "description": "Contrato de prestação de serviços para assinatura",
  "status": "draft",
  "clicksign_key": "12345678-1234-1234-1234-123456789012",
  "documents_ids": [1],
  "signatory_emails": ["empresa@exemplo.com", "cliente@abc.com"],
  "message": "Favor assinar o contrato conforme acordado.",
  "deadline_at": "2025-08-15T23:59:59Z",
  "remind_interval": 3,
  "auto_close": true,
  "created_at": "2025-07-19T10:05:00Z",
  "updated_at": "2025-07-19T10:05:00Z"
}
```

**⚠️ Guarde o `id` do envelope para ativação!**

### Passo 5: Ativar Envelope para Assinatura

Ative o envelope para iniciar o processo de assinatura:

```bash
curl -X POST https://api.ms-docsigner.com/api/v1/envelopes/123/activate \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "X-Correlation-ID: getting-started-001"
```

**Response de Sucesso:**
```json
{
  "id": 123,
  "name": "Envelope - Contrato Cliente ABC",
  "description": "Contrato de prestação de serviços para assinatura",
  "status": "running",
  "clicksign_key": "12345678-1234-1234-1234-123456789012",
  "documents_ids": [1],
  "signatory_emails": ["empresa@exemplo.com", "cliente@abc.com"],
  "message": "Favor assinar o contrato conforme acordado.",
  "deadline_at": "2025-08-15T23:59:59Z",
  "remind_interval": 3,
  "auto_close": true,
  "created_at": "2025-07-19T10:05:00Z",
  "updated_at": "2025-07-19T10:10:00Z"
}
```

### Passo 6: Monitorar Status do Envelope

Consulte periodicamente o status do envelope:

```bash
curl -X GET https://api.ms-docsigner.com/api/v1/envelopes/123 \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "X-Correlation-ID: getting-started-001"
```

---

## Exemplo Completo End-to-End

### Cenário: Contrato de Trabalho

Vamos implementar um caso completo de contrato de trabalho:

```bash
#!/bin/bash

# Configurações
API_BASE="https://api.ms-docsigner.com"
TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
CORRELATION_ID="contract-workflow-$(date +%s)"

echo "🚀 Iniciando fluxo completo de assinatura de contrato..."

# Passo 1: Criar documento
echo "📄 Criando documento..."
DOCUMENT_RESPONSE=$(curl -s -X POST "$API_BASE/api/v1/documents" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "X-Correlation-ID: $CORRELATION_ID" \
  -d '{
    "name": "Contrato de Trabalho - João Silva",
    "file_content_base64": "JVBERi0xLjQKMSAwIG9iag0KPDwNCi9UeXBlIC9DYXRhbG9nDQovUGFnZXMgMiAwIFINCj4+DQplbmRvYmoNCjIgMCBvYmoNCjw8DQovVHlwZSAvUGFnZXMNCi9LaWRzIFs...",
    "description": "Contrato de trabalho para novo funcionário"
  }')

DOCUMENT_ID=$(echo $DOCUMENT_RESPONSE | jq -r '.id')
echo "✅ Documento criado com ID: $DOCUMENT_ID"

# Passo 2: Marcar documento como pronto
echo "🔄 Atualizando status do documento..."
curl -s -X PUT "$API_BASE/api/v1/documents/$DOCUMENT_ID" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "X-Correlation-ID: $CORRELATION_ID" \
  -d '{
    "status": "ready"
  }' > /dev/null

echo "✅ Documento marcado como pronto"

# Passo 3: Criar envelope
echo "📦 Criando envelope..."
ENVELOPE_RESPONSE=$(curl -s -X POST "$API_BASE/api/v1/envelopes" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "X-Correlation-ID: $CORRELATION_ID" \
  -d "{
    \"name\": \"Contrato de Trabalho - João Silva\",
    \"description\": \"Contrato de trabalho para assinatura do funcionário e RH\",
    \"documents_ids\": [$DOCUMENT_ID],
    \"signatory_emails\": [
      \"joao.silva@empresa.com\",
      \"rh@empresa.com\"
    ],
    \"message\": \"Favor assinar o contrato de trabalho. Em caso de dúvidas, entre em contato com o RH.\",
    \"deadline_at\": \"$(date -d '+7 days' -Iseconds)\",
    \"remind_interval\": 2,
    \"auto_close\": true
  }")

ENVELOPE_ID=$(echo $ENVELOPE_RESPONSE | jq -r '.id')
echo "✅ Envelope criado com ID: $ENVELOPE_ID"

# Passo 4: Ativar envelope
echo "🚀 Ativando envelope..."
ACTIVATE_RESPONSE=$(curl -s -X POST "$API_BASE/api/v1/envelopes/$ENVELOPE_ID/activate" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Correlation-ID: $CORRELATION_ID")

STATUS=$(echo $ACTIVATE_RESPONSE | jq -r '.status')
CLICKSIGN_KEY=$(echo $ACTIVATE_RESPONSE | jq -r '.clicksign_key')

echo "✅ Envelope ativado com sucesso!"
echo "📊 Status: $STATUS"
echo "🔑 Clicksign Key: $CLICKSIGN_KEY"

# Passo 5: Consultar status
echo "🔍 Consultando status final..."
FINAL_STATUS=$(curl -s -X GET "$API_BASE/api/v1/envelopes/$ENVELOPE_ID" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Correlation-ID: $CORRELATION_ID")

echo "📋 Status final do envelope:"
echo $FINAL_STATUS | jq '{id: .id, name: .name, status: .status, clicksign_key: .clicksign_key}'

echo "🎉 Fluxo concluído! Os signatários receberão e-mails para assinatura."
```

---

## Casos de Uso Comuns

### 1. NDA para Funcionários

```bash
# Criar documento padrão de NDA
curl -X POST https://api.ms-docsigner.com/api/v1/documents \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "name": "NDA Padrão Empresa",
    "file_content_base64": "...",
    "description": "Acordo de confidencialidade padrão"
  }'

# Criar envelope para múltiplos funcionários
curl -X POST https://api.ms-docsigner.com/api/v1/envelopes \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "name": "NDA - Novos Funcionários Julho 2025",
    "documents_ids": [2],
    "signatory_emails": [
      "funcionario1@empresa.com",
      "funcionario2@empresa.com",
      "funcionario3@empresa.com"
    ],
    "deadline_at": "2025-07-26T17:00:00Z",
    "remind_interval": 1
  }'
```

### 2. Contrato de Cliente

```bash
# Upload de contrato específico do cliente
curl -X POST https://api.ms-docsigner.com/api/v1/documents \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "name": "Contrato Cliente XYZ Corp",
    "file_path": "/contracts/xyz_corp_2025.pdf",
    "file_size": 3145728,
    "mime_type": "application/pdf"
  }'

# Envelope com prazo específico
curl -X POST https://api.ms-docsigner.com/api/v1/envelopes \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "name": "Contrato XYZ Corp - Projeto ABC",
    "documents_ids": [3],
    "signatory_emails": [
      "vendas@empresa.com",
      "contrato@xyzcorp.com",
      "juridico@xyzcorp.com"
    ],
    "deadline_at": "2025-08-01T23:59:59Z",
    "remind_interval": 3
  }'
```

### 3. Termo Médico Urgente

```bash
# Documento com prazo crítico
curl -X POST https://api.ms-docsigner.com/api/v1/envelopes \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "name": "Termo de Consentimento - Cirurgia Emergencial",
    "documents_ids": [4],
    "signatory_emails": [
      "paciente@email.com",
      "responsavel@email.com"
    ],
    "deadline_at": "2025-07-19T18:00:00Z",
    "remind_interval": 1,
    "auto_close": true
  }'
```

---

## Monitoramento e Debugging

### 1. Usar IDs de Correlação

Sempre inclua o header `X-Correlation-ID` para facilitar o rastreamento:

```bash
CORRELATION_ID="debug-session-$(date +%s)"

curl -X POST https://api.ms-docsigner.com/api/v1/documents \
  -H "X-Correlation-ID: $CORRELATION_ID" \
  # ... resto da requisição
```

### 2. Consultar Logs

Os logs podem ser consultados usando o correlation ID nos sistemas de monitoramento.

### 3. Validar Responses

Sempre verifique o status HTTP e o conteúdo da response:

```bash
RESPONSE=$(curl -s -w "%{http_code}" -X POST https://api.ms-docsigner.com/api/v1/documents \
  -H "Authorization: Bearer $TOKEN" \
  -d '{ ... }')

HTTP_CODE="${RESPONSE: -3}"
BODY="${RESPONSE%???}"

if [ "$HTTP_CODE" -eq 201 ]; then
  echo "✅ Sucesso: $BODY"
else
  echo "❌ Erro ($HTTP_CODE): $BODY"
fi
```

---

## Tratamento de Erros Comuns

### 1. Token Expirado (401)

```bash
# Renovar token
curl -X POST https://api.ms-docsigner.com/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "seu-refresh-token"
  }'
```

### 2. Documento Muito Grande (413)

```bash
# Reduzir qualidade ou dividir documento
echo "Erro: Documento excede 7.5MB. Considere:"
echo "- Reduzir qualidade do PDF"
echo "- Dividir em múltiplos documentos"
echo "- Usar compressão"
```

### 3. Tipo de Arquivo Não Suportado (415)

```bash
# Converter para formato suportado
echo "Tipos suportados: PDF, JPEG, PNG, GIF"
echo "Converta seu arquivo para um dos formatos suportados"
```

---

## Próximos Passos

Após dominar o fluxo básico, explore:

1. **[Documentação completa da API de Documentos](./documents.md)**
2. **[Documentação completa da API de Envelopes](./clicksign-envelopes.md)**
3. **Webhooks** para notificações em tempo real
4. **Integração com sistemas de notificação**
5. **Monitoramento de performance** e métricas

---

## Suporte

Para dúvidas ou problemas:

- Consulte os logs usando o correlation ID
- Verifique a documentação específica de cada endpoint
- Entre em contato com a equipe de desenvolvimento

**Dica:** Sempre teste primeiro no ambiente de sandbox antes de usar em produção!