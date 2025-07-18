# Documentação da API de Envelopes do Clicksign

Este documento detalha o processo de criação, monitoramento e consulta de envelopes no Clicksign através da API do microserviço `ms-docsigner`.

## 1. Criação de Envelopes

Para criar um novo envelope no Clicksign, utilize o endpoint `POST /api/v1/envelopes`.

### Exemplo de Payload de Criação

```json
{
  "name": "Contrato de Prestação de Serviços",
  "locale": "pt-BR",
  "auto_close": true,
  "remind_interval": 3,
  "deadline_at": "2025-10-15T23:59:59Z",
  "default_subject": "Solicitação de assinatura do contrato"
}
```

### Exemplos de Uso da API

#### Exemplo 1: Criação de Envelope Simples

```bash
curl -X POST https://sandbox.clicksign.com/api/v3/envelopes \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Contrato de Prestação de Serviços",
    "locale": "pt-BR",
    "auto_close": true,
    "remind_interval": 3
  }'
```

#### Exemplo 2: Criação com Prazo Definido

```bash
curl -X POST https://sandbox.clicksign.com/api/v3/envelopes \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Contrato Urgente",
    "locale": "pt-BR",
    "auto_close": true,
    "remind_interval": 1,
    "deadline_at": "2025-08-15T23:59:59Z",
    "default_subject": "Assinatura urgente necessária"
  }'
```

#### Exemplo 3: Ativação de Envelope (Draft -> Running)

Após a criação de um envelope no status `draft`, ele pode ser ativado para assinatura utilizando o endpoint `PATCH /api/v3/envelopes/{envelope_id}`.

```bash
curl -X PATCH https://sandbox.clicksign.com/api/v3/envelopes/ENVELOPE_ID \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "status": "running"
  }'
```

## 2. Monitoramento e Consultas

### Consultar Status de Envelope

Para consultar os detalhes de um envelope específico, utilize o endpoint `GET /api/v1/envelopes/{id}`.

**Request:**

```bash
GET /api/v1/envelopes/123
```

**Response:**

```json
{
  "id": 123,
  "name": "Contrato de Prestação de Serviços - Cliente ABC",
  "status": "running",
  "clicksign_key": "12345678-1234-1234-1234-123456789012",
  "documents_ids": [1],
  "signatory_emails": ["empresa@exemplo.com", "cliente@abc.com"],
  "message": "Favor assinar o contrato de prestação de serviços conforme acordado.",
  "deadline_at": "2025-08-15T23:59:59Z",
  "created_at": "2025-07-18T10:00:00Z",
  "updated_at": "2025-07-18T10:15:00Z"
}
```

### Listar Envelopes Ativos

Para listar envelopes com base em filtros (ex: status), utilize o endpoint `GET /api/v1/envelopes`.

**Request:**

```bash
GET /api/v1/envelopes?status=running
```

**Response:**

```json
{
  "envelopes": [
    {
      "id": 123,
      "name": "Contrato de Prestação de Serviços - Cliente ABC",
      "status": "running",
      "created_at": "2025-07-18T10:00:00Z"
    },
    {
      "id": 124,
      "name": "NDA - Novos Funcionários Julho 2025",
      "status": "running",
      "created_at": "2025-07-18T11:00:00Z"
    }
  ],
  "total": 2
}
```

## 3. Tratamento de Erros e Cenários Especiais

### Erro de Validação

Quando o payload da requisição contém dados inválidos, a API retornará um erro de validação.

**Request com dados inválidos:**

```bash
POST /api/v1/envelopes
Content-Type: application/json

{
  "name": "",
  "documents_ids": [],
  "signatory_emails": ["email-invalido"]
}
```

**Response:**

```json
{
  "error": "Validation failed",
  "details": [
    {
      "field": "name",
      "message": "Name is required and must be at least 3 characters"
    },
    {
      "field": "documents_ids",
      "message": "At least one document is required"
    },
    {
      "field": "signatory_emails",
      "message": "Invalid email format: email-invalido"
    }
  ]
}
```

### Erro de Integração com Clicksign

Se houver problemas na comunicação com a API do Clicksign (ex: indisponibilidade), a API retornará um erro de serviço externo.

**Response quando API do Clicksign está indisponível:**

```json
{
  "error": "External service temporarily unavailable",
  "message": "Unable to connect to Clicksign API. Please try again later.",
  "retry_after": 300,
  "correlation_id": "abc123-def456-ghi789"
}
```

## 4. Casos de Uso Práticos do Microserviço

### Caso de Uso 1: Contrato de Prestação de Serviços

**Cenário:** Uma empresa precisa enviar um contrato de prestação de serviços para assinatura do cliente.

**Fluxo:**

1. Upload do documento PDF do contrato
2. Criação do envelope com informações do contrato
3. Adição dos signatários (empresa e cliente)
4. Ativação do envelope para assinatura
5. Monitoramento do status de assinatura

**Exemplo de Request:**

```bash
# 1. Criar documento no sistema
POST /api/v1/documents
Content-Type: application/json

{
  "name": "Contrato de Prestação de Serviços - Cliente ABC",
  "file_path": "/uploads/contrato_abc_2025.pdf",
  "file_size": 1048576,
  "mime_type": "application/pdf",
  "description": "Contrato de desenvolvimento de software"
}

# 2. Criar envelope no Clicksign
POST /api/v1/envelopes
Content-Type: application/json

{
  "name": "Contrato de Prestação de Serviços - Cliente ABC",
  "description": "Contrato de desenvolvimento de software para o cliente ABC",
  "documents_ids": [1],
  "signatory_emails": ["empresa@exemplo.com", "cliente@abc.com"],
  "message": "Favor assinar o contrato de prestação de serviços conforme acordado.",
  "deadline_at": "2025-08-15T23:59:59Z"
}
```

**Exemplo de Response:**

```json
{
  "id": 123,
  "name": "Contrato de Prestação de Serviços - Cliente ABC",
  "status": "draft",
  "clicksign_key": "12345678-1234-1234-1234-123456789012",
  "created_at": "2025-07-18T10:00:00Z",
  "updated_at": "2025-07-18T10:00:00Z"
}
```

### Caso de Uso 2: Acordo de Confidencialidade (NDA)

**Cenário:** RH precisa coletar assinatura de NDA de novos funcionários.

**Fluxo:**

1. Upload do template de NDA
2. Criação do envelope com prazo de 48 horas
3. Envio para múltiplos funcionários
4. Configuração de lembretes diários

**Exemplo de Request:**

```bash
POST /api/v1/envelopes
Content-Type: application/json

{
  "name": "NDA - Novos Funcionários Julho 2025",
  "description": "Acordo de confidencialidade para novos colaboradores",
  "documents_ids": [2],
  "signatory_emails": [
    "joao.silva@empresa.com",
    "maria.santos@empresa.com",
    "carlos.oliveira@empresa.com"
  ],
  "message": "Bem-vindo(a) à empresa! Por favor, assine o acordo de confidencialidade.",
  "deadline_at": "2025-07-20T17:00:00Z",
  "remind_interval": 1
}
```

### Caso de Uso 3: Termo de Consentimento Médico

**Cenário:** Clínica médica precisa coletar consentimento para procedimento.

**Fluxo:**

1. Upload do termo de consentimento
2. Criação de envelope urgente (24h)
3. Envio para paciente e responsável
4. Ativação imediata do envelope

**Exemplo de Request:**

```bash
POST /api/v1/envelopes
Content-Type: application/json

{
  "name": "Termo de Consentimento - Procedimento Cirúrgico",
  "description": "Consentimento para cirurgia do paciente João Silva",
  "documents_ids": [3],
  "signatory_emails": [
    "paciente@email.com",
    "responsavel@email.com"
  ],
  "message": "Termo de consentimento para procedimento cirúrgico agendado para amanhã.",
  "deadline_at": "2025-07-19T12:00:00Z",
  "remind_interval": 2,
  "auto_close": true
}
```

### Caso de Uso 4: Contrato de Locação Residencial

**Cenário:** Imobiliária precisa formalizar contrato de locação.

**Fluxo:**

1. Upload do contrato de locação
2. Criação de envelope com múltiplos signatários
3. Prazo de 7 dias para assinatura
4. Lembretes a cada 2 dias

**Exemplo de Request:**

```bash
POST /api/v1/envelopes
Content-Type: application/json

{
  "name": "Contrato de Locação - Apartamento Centro",
  "description": "Contrato de locação residencial - Rua das Flores, 123",
  "documents_ids": [4],
  "signatory_emails": [
    "proprietario@email.com",
    "inquilino@email.com",
    "fiador@email.com",
    "imobiliaria@email.com"
  ],
  "message": "Contrato de locação residencial para assinatura de todas as partes.",
  "deadline_at": "2025-07-25T23:59:59Z",
  "remind_interval": 2
}
```

### Caso de Uso 5: Acordo de Parceria Empresarial

**Cenário:** Duas empresas precisam formalizar uma parceria comercial.

**Fluxo:**

1. Upload do acordo de parceria
2. Criação de envelope com representantes legais
3. Prazo de 15 dias para análise e assinatura
4. Lembretes semanais

**Exemplo de Request:**

```bash
POST /api/v1/envelopes
Content-Type: application/json

{
  "name": "Acordo de Parceria Comercial - Empresa XYZ",
  "description": "Acordo de parceria para desenvolvimento conjunto de produtos",
  "documents_ids": [5],
  "signatory_emails": [
    "diretor@empresaA.com",
    "legal@empresaA.com",
    "ceo@empresaXYZ.com",
    "juridico@empresaXYZ.com"
  ],
  "message": "Acordo de parceria comercial entre as empresas para análise e assinatura.",
  "deadline_at": "2025-08-02T17:00:00Z",
  "remind_interval": 7
}
```
