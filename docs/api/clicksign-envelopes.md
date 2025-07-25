# API de Envelopes do Clicksign

Esta documentação detalha todos os endpoints relacionados ao gerenciamento de envelopes no microserviço ms-docsigner, incluindo criação, consulta, ativação e integração com a API do Clicksign.

## Endpoints Disponíveis

### Headers Obrigatórios para Todos os Endpoints
```
Authorization: Bearer <jwt_token>
Content-Type: application/json
```

## 1. Criar Envelope
`POST /api/v1/envelopes`

Cria um novo envelope no Clicksign com documentos associados e signatários.

### Parâmetros do Request

| Campo | Tipo | Obrigatório | Descrição |
|-------|------|-------------|-----------|
| `name` | string | Sim | Nome do envelope (3-255 caracteres) |
| `description` | string | Não | Descrição do envelope (máx. 1000 caracteres) |
| `documents_ids` | array[int] | Condicional | IDs dos documentos existentes (mínimo 1) |
| `documents` | array[object] | Condicional | Documentos base64 para criação simultânea |
| `signatory_emails` | array[string] | Condicional | E-mails dos signatários (compatibilidade) |
| `signatories` | array[object] | Condicional | Signatários completos com dados detalhados ⭐ **NOVO** |
| `requirements` | array[object] | Não | Requirements de assinatura para aplicar automaticamente ⭐ **NOVO** |
| `message` | string | Não | Mensagem personalizada para signatários |
| `deadline_at` | string | Não | Prazo para assinatura (ISO 8601) |
| `remind_interval` | integer | Não | Intervalo de lembrete em dias (padrão: 3) |
| `auto_close` | boolean | Não | Fechar automaticamente após todas as assinaturas |

**⚠️ IMPORTANTE:** 
- Use **OU** `documents_ids` **OU** `documents`, nunca ambos na mesma requisição
- Use **OU** `signatory_emails` **OU** `signatories`, nunca ambos na mesma requisição

#### Objeto `documents` (para criação com base64)

| Campo | Tipo | Obrigatório | Descrição |
|-------|------|-------------|-----------|
| `name` | string | Sim | Nome do documento (3-255 caracteres) |
| `file_content_base64` | string | Sim | Conteúdo do arquivo em base64 (RAW - sem prefixos) |
| `description` | string | Não | Descrição do documento |

**⚠️ FORMATO BASE64:** O campo `file_content_base64` deve conter apenas o conteúdo base64 RAW, **SEM** prefixos como `data:application/pdf;base64,`. O sistema adiciona automaticamente o prefixo correto baseado no tipo MIME detectado.

#### Objeto `signatories` (para criação com dados completos) ⭐ **NOVA FUNCIONALIDADE**

| Campo | Tipo | Obrigatório | Descrição |
|-------|------|-------------|-----------|
| `name` | string | Sim | Nome completo do signatário (2-255 caracteres) |
| `email` | string | Sim | E-mail do signatário (formato válido) |
| `birthday` | string | Não | Data de nascimento (formato YYYY-MM-DD) |
| `phone_number` | string | Não | Telefone com código do país (ex: +5511999999999) |
| `has_documentation` | boolean | Não | Se possui documentação (padrão: false) |
| `refusable` | boolean | Não | Se pode recusar a assinatura (padrão: true) |
| `group` | integer | Não | Grupo de assinatura para ordem específica |
| `communicate_events` | object | Não | Configurações de notificação do signatário |

#### Objeto `requirements` (para criação com requirements automáticos) ⭐ **NOVA FUNCIONALIDADE**

| Campo | Tipo | Obrigatório | Descrição |
|-------|------|-------------|-----------|
| `action` | string | Sim | Tipo de ação: "sign", "agree", "provide_evidence" |
| `role` | string | Sim | Papel do signatário: "sign" |
| `auth` | string | Condicional | Método de autenticação: "email", "icp_brasil" (obrigatório para provide_evidence) |
| `document_id` | string | Não | ID do documento no Clicksign |
| `signer_id` | string | Não | ID do signatário no Clicksign |

### Exemplos de Request

#### Opção A: Criação com Documentos Existentes (IDs)

```json
{
  "name": "Contrato de Prestação de Serviços - Cliente ABC",
  "description": "Contrato de desenvolvimento de software",
  "documents_ids": [1, 2],
  "signatories": [
    {
      "name": "Empresa Prestadora",
      "email": "empresa@exemplo.com",
      "has_documentation": false,
      "refusable": false
    },
    {
      "name": "Cliente ABC Ltda",
      "email": "cliente@abc.com",
      "phone_number": "+5511987654321",
      "has_documentation": true,
      "refusable": true
    }
  ],
  "message": "Favor assinar o contrato conforme acordado.",
  "deadline_at": "2025-08-15T23:59:59Z",
  "remind_interval": 3,
  "auto_close": true,
  "requirements": [
    {
      "action": "sign",
      "role": "sign",
      "auth": "email"
    },
    {
      "action": "provide_evidence",
      "role": "sign",
      "auth": "icp_brasil"
    }
  ]
}
```

#### Opção B: Criação com Documentos Base64 ⭐ **NOVA FUNCIONALIDADE**

```json
{
  "name": "Contrato de Prestação de Serviços - Cliente ABC",
  "description": "Contrato de desenvolvimento de software",
  "documents": [
    {
      "name": "contrato-principal.pdf",
      "description": "Contrato principal de prestação de serviços",
      "file_content_base64": "JVBERi0xLjQKMSAwIG9iag0KPDwNCi9UeXBlIC9DYXRhbG9nDQovUGFnZXMgMiAwIFINCj4+DQplbmRvYmoNCjIgMCBvYmoNCjw8DQovVHlwZSAvUGFnZXMNCi9LaWRzIFs..."
    },
    {
      "name": "anexo-tecnico.pdf",
      "description": "Anexo técnico com especificações",
      "file_content_base64": "JVBERi0xLjQKMyAwIG9iag0KPDwNCi9UeXBlIC9DYXRhbG9nDQovUGFnZXMgNCAwIFINCj4+DQplbmRvYmoNCjQgMCBvYmoNCjw8DQovVHlwZSAvUGFnZXMNCi9LaWRzIFs..."
    }
  ],
  "signatories": [
    {
      "name": "Empresa Prestadora",
      "email": "empresa@exemplo.com",
      "has_documentation": false,
      "refusable": false
    },
    {
      "name": "Cliente ABC Ltda", 
      "email": "cliente@abc.com",
      "phone_number": "+5511987654321",
      "has_documentation": true,
      "refusable": true
    }
  ],
  "message": "Favor assinar o contrato conforme acordado.",
  "deadline_at": "2025-08-15T23:59:59Z",
  "remind_interval": 3,
  "auto_close": true
}
```

### Response de Sucesso (201)

A resposta é idêntica para ambos os métodos de criação. Quando usando `documents` base64, os documentos são criados automaticamente e seus IDs aparecem em `documents_ids`:

```json
{
  "id": 123,
  "name": "Contrato de Prestação de Serviços - Cliente ABC",
  "description": "Contrato de desenvolvimento de software",
  "status": "draft",
  "clicksign_key": "12345678-1234-1234-1234-123456789012",
  "documents_ids": [45, 46],
  "signatories": [
    {
      "name": "Empresa Prestadora",
      "email": "empresa@exemplo.com",
      "has_documentation": false,
      "refusable": false
    },
    {
      "name": "Cliente ABC Ltda", 
      "email": "cliente@abc.com",
      "phone_number": "+5511987654321",
      "has_documentation": true,
      "refusable": true
    }
  ],
  "message": "Favor assinar o contrato conforme acordado.",
  "deadline_at": "2025-08-15T23:59:59Z",
  "remind_interval": 3,
  "auto_close": true,
  "created_at": "2025-07-19T10:00:00Z",
  "updated_at": "2025-07-19T10:00:00Z"
}
```

**📝 Nota:** No exemplo acima, os documentos base64 foram automaticamente criados com IDs `45` e `46`, e já estão sincronizados no Clicksign com seus respectivos `clicksign_key`.

### Códigos de Erro

#### Erros Gerais
- `401` - Token JWT ausente ou inválido
- `500` - Erro interno ou falha na integração Clicksign

#### Erros de Validação (400)
- **Dados inválidos:** Campos obrigatórios ausentes ou formato incorreto
- **Documentos não encontrados:** IDs de documentos em `documents_ids` não existem
- **Conflito de parâmetros:** Uso simultâneo de `documents_ids` e `documents`
- **Base64 inválido:** Conteúdo `file_content_base64` malformado
- **Arquivo muito grande:** Documento excede 7.5MB após decodificação
- **Tipo não suportado:** MIME type do documento não é PDF, JPEG, PNG ou GIF

#### Exemplos de Respostas de Erro

**Conflito de parâmetros:**
```json
{
  "error": "Validation failed",
  "message": "não é possível fornecer documents_ids e documents ao mesmo tempo"
}
```

**Base64 inválido:**
```json
{
  "error": "Invalid request",
  "message": "failed to process base64 content for document 'contrato.pdf': invalid base64 encoding"
}
```

**Tipo de arquivo não suportado:**
```json
{
  "error": "Invalid request", 
  "message": "unsupported file type for document 'documento.docx': file type not supported. Supported types: PDF, JPEG, PNG, GIF"
}
```

---

## 2. Buscar Envelope por ID
`GET /api/v1/envelopes/{id}`

Retorna um envelope específico pelo ID.

### Parâmetros da URL
- `id` (integer): ID do envelope

### Exemplo de Request
```bash
GET /api/v1/envelopes/123
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### Response de Sucesso (200)
```json
{
  "id": 123,
  "name": "Contrato de Prestação de Serviços - Cliente ABC",
  "description": "Contrato de desenvolvimento de software",
  "status": "running",
  "clicksign_key": "12345678-1234-1234-1234-123456789012",
  "documents_ids": [1, 2],
  "signatories": [
    {
      "name": "Empresa Prestadora",
      "email": "empresa@exemplo.com",
      "has_documentation": false,
      "refusable": false
    },
    {
      "name": "Cliente ABC Ltda", 
      "email": "cliente@abc.com",
      "phone_number": "+5511987654321",
      "has_documentation": true,
      "refusable": true
    }
  ],
  "message": "Favor assinar o contrato conforme acordado.",
  "deadline_at": "2025-08-15T23:59:59Z",
  "remind_interval": 3,
  "auto_close": true,
  "created_at": "2025-07-19T10:00:00Z",
  "updated_at": "2025-07-19T10:15:00Z"
}
```

### Códigos de Erro
- `400` - ID inválido
- `401` - Não autorizado
- `404` - Envelope não encontrado
- `500` - Erro interno

---

## 3. Listar Envelopes
`GET /api/v1/envelopes`

Retorna uma lista de envelopes com filtros opcionais.

### Parâmetros de Query (opcionais)

| Parâmetro | Tipo | Descrição |
|-----------|------|-----------|
| `search` | string | Buscar por nome do envelope |
| `status` | string | Filtrar por status (draft, running, closed) |
| `clicksign_key` | string | Filtrar por chave do Clicksign |

### Exemplos de Request

#### Listar todos os envelopes
```bash
GET /api/v1/envelopes
```

#### Filtrar por status
```bash
GET /api/v1/envelopes?status=running
```

#### Buscar por nome
```bash
GET /api/v1/envelopes?search=contrato
```

### Response de Sucesso (200)
```json
{
  "envelopes": [
    {
      "id": 123,
      "name": "Contrato de Prestação de Serviços - Cliente ABC",
      "description": "Contrato de desenvolvimento de software",
      "status": "running",
      "clicksign_key": "12345678-1234-1234-1234-123456789012",
      "documents_ids": [1, 2],
      "signatories": [
    {
      "name": "Empresa Prestadora",
      "email": "empresa@exemplo.com",
      "has_documentation": false,
      "refusable": false
    },
    {
      "name": "Cliente ABC Ltda", 
      "email": "cliente@abc.com",
      "phone_number": "+5511987654321",
      "has_documentation": true,
      "refusable": true
    }
  ],
      "message": "Favor assinar o contrato conforme acordado.",
      "deadline_at": "2025-08-15T23:59:59Z",
      "remind_interval": 3,
      "auto_close": true,
      "created_at": "2025-07-19T10:00:00Z",
      "updated_at": "2025-07-19T10:15:00Z"
    }
  ],
  "total": 1
}
```

---

## 4. Ativar Envelope
`POST /api/v1/envelopes/{id}/activate`

Ativa um envelope para iniciar o processo de assinatura. Move o envelope do status `draft` para `running`.

### Parâmetros da URL
- `id` (integer): ID do envelope

### Exemplo de Request
```bash
POST /api/v1/envelopes/123/activate
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### Response de Sucesso (200)
```json
{
  "id": 123,
  "name": "Contrato de Prestação de Serviços - Cliente ABC",
  "description": "Contrato de desenvolvimento de software",
  "status": "running",
  "clicksign_key": "12345678-1234-1234-1234-123456789012",
  "documents_ids": [1, 2],
  "signatories": [
    {
      "name": "Empresa Prestadora",
      "email": "empresa@exemplo.com",
      "has_documentation": false,
      "refusable": false
    },
    {
      "name": "Cliente ABC Ltda", 
      "email": "cliente@abc.com",
      "phone_number": "+5511987654321",
      "has_documentation": true,
      "refusable": true
    }
  ],
  "message": "Favor assinar o contrato conforme acordado.",
  "deadline_at": "2025-08-15T23:59:59Z",
  "remind_interval": 3,
  "auto_close": true,
  "created_at": "2025-07-19T10:00:00Z",
  "updated_at": "2025-07-19T10:20:00Z"
}
```

### Códigos de Erro
- `400` - ID inválido ou envelope já ativado
- `401` - Não autorizado
- `404` - Envelope não encontrado
- `500` - Erro interno ou falha na ativação

---

## 🆕 Gerenciamento de Requirements (Requisitos de Assinatura)

### 1. Criar Requirement para Envelope
`POST /api/v1/envelopes/{envelope_id}/requirements`

Adiciona um novo requirement (requisito de assinatura) a um envelope existente.

#### Parâmetros da URL
- `envelope_id` (integer): ID do envelope

#### Parâmetros do Request
```json
{
  "action": "sign",
  "role": "sign", 
  "auth": "email",
  "document_id": "doc_123",
  "signer_id": "signer_456"
}
```

| Campo | Tipo | Obrigatório | Descrição |
|-------|------|-------------|-----------|
| `action` | string | Sim | Tipo de ação: "sign", "agree", "provide_evidence" |
| `role` | string | Sim | Papel do signatário: "sign" |
| `auth` | string | Condicional | Método de autenticação: "email", "icp_brasil" (obrigatório para provide_evidence) |
| `document_id` | string | Não | ID do documento no Clicksign |
| `signer_id` | string | Não | ID do signatário no Clicksign |

#### Response de Sucesso (201)
```json
{
  "id": 1,
  "envelope_id": 123,
  "clicksign_key": "req_12345678-1234-1234-1234-123456789012",
  "action": "sign",
  "role": "sign",
  "auth": "email",
  "document_id": "doc_123",
  "signer_id": "signer_456",
  "status": "pending",
  "created_at": "2025-07-20T10:00:00Z",
  "updated_at": "2025-07-20T10:00:00Z"
}
```

### 2. Listar Requirements do Envelope
`GET /api/v1/envelopes/{envelope_id}/requirements`

Lista todos os requirements de um envelope específico.

#### Response de Sucesso (200)
```json
{
  "requirements": [
    {
      "id": 1,
      "envelope_id": 123,
      "clicksign_key": "req_12345678-1234-1234-1234-123456789012",
      "action": "sign",
      "role": "sign",
      "auth": "email",
      "document_id": "doc_123",
      "signer_id": "signer_456",
      "status": "pending",
      "created_at": "2025-07-20T10:00:00Z",
      "updated_at": "2025-07-20T10:00:00Z"
    }
  ],
  "total": 1
}
```

### 3. Obter Requirement Específico
`GET /api/v1/requirements/{requirement_id}`

Retorna detalhes de um requirement específico.

### 4. Atualizar Requirement
`PUT /api/v1/requirements/{requirement_id}`

Atualiza informações de um requirement existente.

#### Parâmetros do Request
```json
{
  "action": "provide_evidence",
  "auth": "icp_brasil",
  "status": "completed"
}
```

### 5. Remover Requirement
`DELETE /api/v1/requirements/{requirement_id}`

Remove um requirement do envelope.

#### Response de Sucesso (200)
```json
{
  "message": "Requirement deletado com sucesso"
}
```

### Criação de Envelope com Requirements Automáticos ⭐ **NOVA FUNCIONALIDADE**

Agora você pode criar envelopes e requirements simultaneamente usando o campo `requirements` no payload de criação de envelope:

```json
{
  "name": "Contrato com Requirements Específicos",
  "description": "Contrato com requisitos de assinatura personalizados",
  "documents_ids": [1, 2],
  "signatories": [
    {
      "name": "Cliente ABC",
      "email": "cliente@abc.com",
      "refusable": true
    }
  ],
  "requirements": [
    {
      "action": "sign",
      "role": "sign",
      "auth": "email",
      "document_id": "doc_123",
      "signer_id": "signer_456"
    },
    {
      "action": "provide_evidence",
      "role": "sign", 
      "auth": "icp_brasil",
      "document_id": "doc_124",
      "signer_id": "signer_456"
    }
  ],
  "deadline_at": "2025-08-15T23:59:59Z"
}
```

### Tipos de Requirements

#### Actions Disponíveis
- **`agree`**: Requisito de concordância/aceite
- **`sign`**: Requisito de assinatura digital
- **`provide_evidence`**: Requisito de fornecimento de evidência (requer autenticação)

#### Métodos de Autenticação
- **`email`**: Autenticação via email (padrão)
- **`icp_brasil`**: Certificado digital ICP-Brasil

#### Validações de Negócio
- Action `provide_evidence` **obrigatoriamente** requer campo `auth`
- Role atualmente suporta apenas `sign`
- `document_id` e `signer_id` devem referenciar recursos válidos no Clicksign
- Envelope deve existir antes de criar requirements

---

## 🆕 Gerenciamento de Signatários

### 1. Criar Signatário
`POST /api/v1/envelopes/{envelope_id}/signatories`

Adiciona um novo signatário a um envelope existente.

#### Parâmetros da URL
- `envelope_id` (integer): ID do envelope

#### Parâmetros do Request
```json
{
  "name": "João Silva",
  "email": "joao.silva@email.com",
  "phone_number": "+5511987654321",
  "birthday": "1985-03-15",
  "has_documentation": true,
  "refusable": false,
  "group": 1
}
```

### 2. Listar Signatários do Envelope
`GET /api/v1/envelopes/{envelope_id}/signatories`

Lista todos os signatários de um envelope específico.

### 3. Obter Signatário Específico
`GET /api/v1/signatories/{id}`

Retorna detalhes de um signatário específico.

### 4. Atualizar Signatário
`PUT /api/v1/signatories/{id}`

Atualiza informações de um signatário existente.

### 5. Remover Signatário
`DELETE /api/v1/signatories/{id}`

Remove um signatário do envelope.

### 6. Enviar Signatários para Clicksign
`POST /api/v1/envelopes/{envelope_id}/send`

Envia todos os signatários do envelope para o Clicksign após a criação.

---

## Estados do Envelope

| Estado | Descrição |
|--------|-----------|
| `draft` | Envelope criado, aguardando ativação |
| `running` | Envelope ativo, processo de assinatura em andamento |
| `closed` | Envelope finalizado (todas as assinaturas coletadas) |

---

## 🆕 Funcionalidade: Criação de Envelope com Documentos Base64

### Visão Geral

A nova funcionalidade permite criar envelopes e documentos simultaneamente em uma única requisição, eliminando a necessidade de múltiplas chamadas à API. Ideal para:

- **Aplicações frontend** que lidam com upload de arquivos
- **Integrações rápidas** que precisam de menos latência
- **Fluxos simplificados** onde você tem o documento em memória

### Comparação: Fluxo Antigo vs. Novo

#### Fluxo Antigo (3 requisições)
```
1. POST /api/v1/documents (criar documento)
2. POST /api/v1/envelopes (criar envelope com document_id)
3. POST /api/v1/envelopes/{id}/activate (ativar envelope)
```

#### Fluxo Novo (2 requisições) ⭐
```
1. POST /api/v1/envelopes (criar envelope + documentos)
2. POST /api/v1/envelopes/{id}/activate (ativar envelope)
```

### Vantagens do Novo Fluxo

1. **Performance**: Reduz latência e número de requisições HTTP
2. **Atomicidade**: Operação é totalmente atômica - ou tudo é criado ou nada é criado
3. **Simplicidade**: Menos código e menos gerenciamento de estado
4. **Consistência**: Documentos são automaticamente associados ao envelope
5. **Integração Clicksign**: Documentos são criados diretamente no Clicksign
6. **Auto-conversão**: Sistema converte automaticamente base64 RAW para data URI com MIME type correto

### Limitações e Considerações

- **Tamanho do payload**: Requisições maiores devido ao base64
- **Timeout**: Pode levar mais tempo para processar múltiplos documentos
- **Rollback**: Em caso de erro, toda a operação é desfeita
- **Limite de documentos**: Recomendado máximo 5 documentos por envelope

---

## Integração com Documentos Base64

### ⭐ Fluxo Novo (Recomendado): Documento Base64 + Envelope em Uma Operação

```bash
curl -X POST https://api.ms-docsigner.com/api/v1/envelopes \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Envelope - Contrato Cliente ABC",
    "description": "Contrato de prestação de serviços",
    "documents": [
      {
        "name": "contrato-abc.pdf",
        "file_content_base64": "JVBERi0xLjQKM...",
        "description": "Contrato de prestação de serviços"
      }
    ],
    "signatories": [
    {
      "name": "Empresa Prestadora",
      "email": "empresa@exemplo.com",
      "has_documentation": false,
      "refusable": false
    },
    {
      "name": "Cliente ABC Ltda", 
      "email": "cliente@abc.com",
      "phone_number": "+5511987654321",
      "has_documentation": true,
      "refusable": true
    }
  ],
    "deadline_at": "2025-08-15T23:59:59Z"
  }'
```

### Fluxo Antigo (Ainda Suportado): Documento Base64 → Envelope → Ativação

1. **Criar documento via base64**
```bash
curl -X POST https://api.ms-docsigner.com/api/v1/documents \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Contrato Cliente ABC",
    "file_content_base64": "JVBERi0xLjQKM...",
    "description": "Contrato de prestação de serviços"
  }'
```

2. **Criar envelope com documento**
```bash
curl -X POST https://api.ms-docsigner.com/api/v1/envelopes \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Envelope - Contrato Cliente ABC",
    "documents_ids": [1],
    "signatories": [
    {
      "name": "Empresa Prestadora",
      "email": "empresa@exemplo.com",
      "has_documentation": false,
      "refusable": false
    },
    {
      "name": "Cliente ABC Ltda", 
      "email": "cliente@abc.com",
      "phone_number": "+5511987654321",
      "has_documentation": true,
      "refusable": true
    }
  ],
    "deadline_at": "2025-08-15T23:59:59Z"
  }'
```

### Ativação do Envelope (Para Ambos os Fluxos)

```bash
curl -X POST https://api.ms-docsigner.com/api/v1/envelopes/123/activate \
  -H "Authorization: Bearer <token>"
```

### 🔧 Detalhes Técnicos da Conversão Base64

**O que acontece internamente:**

1. **Recebimento**: Sistema recebe base64 RAW (ex: `JVBERi0xLjQKM...`)
2. **Detecção MIME**: Sistema detecta tipo do arquivo automaticamente
3. **Conversão**: Adiciona prefixo data URI (ex: `data:application/pdf;base64,JVBERi0xLjQKM...`)
4. **Envio**: Envia para Clicksign no formato esperado

**Tipos MIME Suportados:**
- `application/pdf` → PDFs
- `image/jpeg` → Imagens JPEG/JPG
- `image/png` → Imagens PNG
- `image/gif` → Imagens GIF

---

## Exemplos de Uso da API

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
  "signatories": [
    {
      "name": "Empresa Prestadora",
      "email": "empresa@exemplo.com",
      "has_documentation": false,
      "refusable": false
    },
    {
      "name": "Cliente ABC Ltda", 
      "email": "cliente@abc.com",
      "phone_number": "+5511987654321",
      "has_documentation": true,
      "refusable": true
    }
  ],
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

#### Método Tradicional (Fluxo Antigo)

**Fluxo:**
1. Upload do documento PDF do contrato
2. Criação do envelope com informações do contrato
3. Adição dos signatários (empresa e cliente)
4. Ativação do envelope para assinatura
5. Monitoramento do status de assinatura

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
  "signatories": [
    {
      "name": "Empresa Prestadora",
      "email": "empresa@exemplo.com",
      "has_documentation": false,
      "refusable": false
    },
    {
      "name": "Cliente ABC Ltda", 
      "email": "cliente@abc.com",
      "phone_number": "+5511987654321",
      "has_documentation": true,
      "refusable": true
    }
  ],
  "message": "Favor assinar o contrato de prestação de serviços conforme acordado.",
  "deadline_at": "2025-08-15T23:59:59Z"
}
```

#### Método Simplificado ⭐ **NOVA FUNCIONALIDADE**

```bash
# Criar envelope + documento em uma única operação
POST /api/v1/envelopes
Content-Type: application/json

{
  "name": "Contrato de Prestação de Serviços - Cliente ABC",
  "description": "Contrato de desenvolvimento de software para o cliente ABC",
  "documents": [
    {
      "name": "contrato-abc-2025.pdf",
      "description": "Contrato de desenvolvimento de software",
      "file_content_base64": "JVBERi0xLjQKMSAwIG9iag0KPDwNCi9UeXBlIC9DYXRhbG9nDQovUGFnZXMgMiAwIFINCj4+DQplbmRvYmoNCjIgMCBvYmoNCjw8DQovVHlwZSAvUGFnZXMNCi9LaWRzIFs..."
    }
  ],
  "signatories": [
    {
      "name": "Empresa Prestadora",
      "email": "empresa@exemplo.com",
      "has_documentation": false,
      "refusable": false
    },
    {
      "name": "Cliente ABC Ltda", 
      "email": "cliente@abc.com",
      "phone_number": "+5511987654321",
      "has_documentation": true,
      "refusable": true
    }
  ],
  "message": "Favor assinar o contrato de prestação de serviços conforme acordado.",
  "deadline_at": "2025-08-15T23:59:59Z"
}
```

**Vantagem:** Reduz de 2 para 1 requisição, processo mais rápido e atômico.

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
