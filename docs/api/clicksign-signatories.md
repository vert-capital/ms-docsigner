# API de Signatários do Clicksign

Esta documentação detalha todos os endpoints relacionados ao gerenciamento de signatários no microserviço ms-docsigner, incluindo criação, consulta, atualização, exclusão e integração com a API do Clicksign.

## Endpoints Disponíveis

### Headers Obrigatórios para Todos os Endpoints
```
Authorization: Bearer <jwt_token>
Content-Type: application/json
```

## 1. Criar Signatário
`POST /api/v1/envelopes/{envelope_id}/signatories`

Cria um novo signatário para um envelope específico.

### Parâmetros da URL
- `envelope_id` (integer): ID do envelope

### Parâmetros do Request

| Campo | Tipo | Obrigatório | Descrição |
|-------|------|-------------|-----------|
| `name` | string | Sim | Nome completo do signatário (2-255 caracteres) |
| `email` | string | Sim | E-mail do signatário (formato válido) |
| `birthday` | string | Não | Data de nascimento (formato YYYY-MM-DD) |
| `phone_number` | string | Não | Telefone com código do país (ex: +5511999999999) |
| `has_documentation` | boolean | Não | Se possui documentação (padrão: false) |
| `refusable` | boolean | Não | Se pode recusar a assinatura (padrão: true) |
| `group` | integer | Não | Grupo de assinatura para ordem específica (padrão: 1) |
| `communicate_events` | object | Não | Configurações de notificação do signatário |

#### Objeto `communicate_events`

| Campo | Tipo | Obrigatório | Descrição |
|-------|------|-------------|-----------|
| `document_signed` | string | Não | Notificação quando documento é assinado: "email", "sms", "none" (padrão: "email") |
| `signature_request` | string | Não | Notificação de solicitação de assinatura: "email", "sms", "none" (padrão: "email") |
| `signature_reminder` | string | Não | Notificação de lembrete: "email", "sms", "none" (padrão: "email") |

### Exemplo de Request

```json
{
  "name": "João Silva Santos",
  "email": "joao.silva@empresa.com",
  "birthday": "1985-03-15",
  "phone_number": "+5511987654321",
  "has_documentation": true,
  "refusable": false,
  "group": 1,
  "communicate_events": {
    "document_signed": "email",
    "signature_request": "email",
    "signature_reminder": "sms"
  }
}
```

### Response de Sucesso (201)

```json
{
  "id": 45,
  "name": "João Silva Santos",
  "email": "joao.silva@empresa.com",
  "envelope_id": 123,
  "birthday": "1985-03-15",
  "phone_number": "+5511987654321",
  "has_documentation": true,
  "refusable": false,
  "group": 1,
  "communicate_events": {
    "document_signed": "email",
    "signature_request": "email",
    "signature_reminder": "sms"
  },
  "created_at": "2025-07-25T10:00:00Z",
  "updated_at": "2025-07-25T10:00:00Z"
}
```

### Códigos de Erro
- `400` - ID do envelope inválido ou dados de validação incorretos
- `404` - Envelope não encontrado
- `401` - Não autorizado
- `500` - Erro interno

---

## 2. Listar Signatários do Envelope
`GET /api/v1/envelopes/{envelope_id}/signatories`

Retorna lista de todos os signatários de um envelope específico.

### Parâmetros da URL
- `envelope_id` (integer): ID do envelope

### Exemplo de Request
```bash
GET /api/v1/envelopes/123/signatories
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### Response de Sucesso (200)

```json
{
  "signatories": [
    {
      "id": 45,
      "name": "João Silva Santos",
      "email": "joao.silva@empresa.com",
      "envelope_id": 123,
      "birthday": "1985-03-15",
      "phone_number": "+5511987654321",
      "has_documentation": true,
      "refusable": false,
      "group": 1,
      "communicate_events": {
        "document_signed": "email",
        "signature_request": "email",
        "signature_reminder": "sms"
      },
      "created_at": "2025-07-25T10:00:00Z",
      "updated_at": "2025-07-25T10:00:00Z"
    },
    {
      "id": 46,
      "name": "Maria Santos Costa",
      "email": "maria.santos@cliente.com",
      "envelope_id": 123,
      "phone_number": "+5511912345678",
      "has_documentation": false,
      "refusable": true,
      "group": 2,
      "communicate_events": {
        "document_signed": "email",
        "signature_request": "email",
        "signature_reminder": "email"
      },
      "created_at": "2025-07-25T10:05:00Z",
      "updated_at": "2025-07-25T10:05:00Z"
    }
  ],
  "total": 2
}
```

### Códigos de Erro
- `400` - ID do envelope inválido
- `404` - Envelope não encontrado
- `401` - Não autorizado
- `500` - Erro interno

---

## 3. Buscar Signatário por ID
`GET /api/v1/signatories/{id}`

Retorna um signatário específico pelo ID.

### Parâmetros da URL
- `id` (integer): ID do signatário

### Exemplo de Request
```bash
GET /api/v1/signatories/45
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### Response de Sucesso (200)
```json
{
  "id": 45,
  "name": "João Silva Santos",
  "email": "joao.silva@empresa.com",
  "envelope_id": 123,
  "birthday": "1985-03-15",
  "phone_number": "+5511987654321",
  "has_documentation": true,
  "refusable": false,
  "group": 1,
  "communicate_events": {
    "document_signed": "email",
    "signature_request": "email",
    "signature_reminder": "sms"
  },
  "created_at": "2025-07-25T10:00:00Z",
  "updated_at": "2025-07-25T10:00:00Z"
}
```

### Códigos de Erro
- `400` - ID inválido
- `401` - Não autorizado
- `404` - Signatário não encontrado
- `500` - Erro interno

---

## 4. Atualizar Signatário
`PUT /api/v1/signatories/{id}`

Atualiza informações de um signatário existente.

### Parâmetros da URL
- `id` (integer): ID do signatário

### Parâmetros do Request

Todos os campos são opcionais. Apenas os campos fornecidos serão atualizados.

| Campo | Tipo | Obrigatório | Descrição |
|-------|------|-------------|-----------|
| `name` | string | Não | Nome completo do signatário (2-255 caracteres) |
| `email` | string | Não | E-mail do signatário (formato válido) |
| `envelope_id` | integer | Não | ID do envelope (para mover signatário) |
| `birthday` | string | Não | Data de nascimento (formato YYYY-MM-DD) |
| `phone_number` | string | Não | Telefone com código do país |
| `has_documentation` | boolean | Não | Se possui documentação |
| `refusable` | boolean | Não | Se pode recusar a assinatura |
| `group` | integer | Não | Grupo de assinatura |
| `communicate_events` | object | Não | Configurações de notificação |

### Exemplo de Request

```json
{
  "name": "João Silva Santos Junior",
  "phone_number": "+5511999887766",
  "group": 2,
  "communicate_events": {
    "document_signed": "sms",
    "signature_request": "email",
    "signature_reminder": "none"
  }
}
```

### Response de Sucesso (200)

```json
{
  "id": 45,
  "name": "João Silva Santos Junior",
  "email": "joao.silva@empresa.com",
  "envelope_id": 123,
  "birthday": "1985-03-15",
  "phone_number": "+5511999887766",
  "has_documentation": true,
  "refusable": false,
  "group": 2,
  "communicate_events": {
    "document_signed": "sms",
    "signature_request": "email",
    "signature_reminder": "none"
  },
  "created_at": "2025-07-25T10:00:00Z",
  "updated_at": "2025-07-25T10:30:00Z"
}
```

### Códigos de Erro
- `400` - ID inválido ou dados de validação incorretos
- `401` - Não autorizado
- `404` - Signatário não encontrado
- `500` - Erro interno

---

## 5. Remover Signatário
`DELETE /api/v1/signatories/{id}`

Remove um signatário do sistema.

### Parâmetros da URL
- `id` (integer): ID do signatário

### Exemplo de Request
```bash
DELETE /api/v1/signatories/45
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### Response de Sucesso (204)
```
No Content
```

### Códigos de Erro
- `400` - ID inválido
- `401` - Não autorizado
- `404` - Signatário não encontrado
- `500` - Erro interno

---

## 6. Enviar Signatários para Clicksign
`POST /api/v1/envelopes/{envelope_id}/send`

Envia todos os signatários de um envelope para o Clicksign para processamento.

### Parâmetros da URL
- `envelope_id` (integer): ID do envelope

### Pré-requisitos
- O envelope deve existir e ter uma chave Clicksign válida
- O envelope deve ter pelo menos um signatário

### Exemplo de Request
```bash
POST /api/v1/envelopes/123/send
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### Response de Sucesso (200)

#### Todos os signatários enviados com sucesso:
```json
{
  "signatories": [
    {
      "id": 45,
      "name": "João Silva Santos",
      "email": "joao.silva@empresa.com",
      "envelope_id": 123,
      "created_at": "2025-07-25T10:00:00Z",
      "updated_at": "2025-07-25T10:00:00Z"
    },
    {
      "id": 46,
      "name": "Maria Santos Costa",
      "email": "maria.santos@cliente.com",
      "envelope_id": 123,
      "created_at": "2025-07-25T10:05:00Z",
      "updated_at": "2025-07-25T10:05:00Z"
    }
  ],
  "total": 2,
  "successful_sends": 2,
  "message": "All signatories sent to Clicksign successfully"
}
```

#### Envio parcial com erros:
```json
{
  "signatories": [
    {
      "id": 45,
      "name": "João Silva Santos",
      "email": "joao.silva@empresa.com",
      "envelope_id": 123,
      "created_at": "2025-07-25T10:00:00Z",
      "updated_at": "2025-07-25T10:00:00Z"
    },
    {
      "id": 46,
      "name": "Maria Santos Costa",
      "email": "maria.santos@cliente.com",
      "envelope_id": 123,
      "created_at": "2025-07-25T10:05:00Z",
      "updated_at": "2025-07-25T10:05:00Z"
    }
  ],
  "total": 2,
  "successful_sends": 1,
  "failed_sends": 1,
  "errors": [
    "Failed to send signatory Maria Santos Costa (maria.santos@cliente.com) to Clicksign: invalid email domain"
  ]
}
```

### Códigos de Erro
- `400` - ID do envelope inválido, envelope não possui chave Clicksign ou não possui signatários
- `401` - Não autorizado
- `404` - Envelope não encontrado
- `500` - Erro interno

---

## Regras de Validação

### Nome (name)
- **Obrigatório** na criação
- Mínimo 2 caracteres, máximo 255 caracteres
- Deve conter pelo menos nome e sobrenome

### Email
- **Obrigatório** na criação
- Deve ser um formato de email válido
- Exemplo: `usuario@dominio.com`

### Data de Nascimento (birthday)
- **Opcional**
- Formato obrigatório: `YYYY-MM-DD`
- Deve ser uma data válida
- Exemplo: `1985-03-15`

### Telefone (phone_number)
- **Opcional**
- Formato obrigatório: internacional com `+` seguido de 8-15 dígitos
- Exemplo: `+5511987654321`

### Grupo (group)
- **Opcional** (padrão: 1)
- Deve ser um número inteiro positivo
- Utilizado para definir ordem de assinatura

### Configurações de Comunicação (communicate_events)
- **Opcional**
- Valores aceitos: `"email"`, `"sms"`, `"none"`
- Padrões:
  - `document_signed`: `"email"`
  - `signature_request`: `"email"`
  - `signature_reminder`: `"email"`

---

## Exemplos de Uso da API

### Exemplo 1: Criação de Signatário Básico

```bash
curl -X POST https://api.ms-docsigner.com/api/v1/envelopes/123/signatories \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Carlos Alberto Silva",
    "email": "carlos.silva@empresa.com"
  }'
```

### Exemplo 2: Criação de Signatário Completo

```bash
curl -X POST https://api.ms-docsigner.com/api/v1/envelopes/123/signatories \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Ana Paula Santos",
    "email": "ana.paula@cliente.com",
    "birthday": "1990-07-20",
    "phone_number": "+5511912345678",
    "has_documentation": true,
    "refusable": false,
    "group": 2,
    "communicate_events": {
      "document_signed": "sms",
      "signature_request": "email",
      "signature_reminder": "email"
    }
  }'
```

### Exemplo 3: Atualização Parcial de Signatário

```bash
curl -X PUT https://api.ms-docsigner.com/api/v1/signatories/45 \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "phone_number": "+5511999888777",
    "group": 3
  }'
```

### Exemplo 4: Envio de Signatários para Clicksign

```bash
curl -X POST https://api.ms-docsigner.com/api/v1/envelopes/123/send \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json"
```

---

## Fluxo de Trabalho Recomendado

### 1. Criação de Envelope com Signatários

1. **Criar envelope** usando `/api/v1/envelopes`
2. **Adicionar signatários** usando `/api/v1/envelopes/{id}/signatories`
3. **Enviar signatários para Clicksign** usando `/api/v1/envelopes/{id}/send`
4. **Ativar envelope** usando `/api/v1/envelopes/{id}/activate`

### 2. Gerenciamento de Signatários

1. **Listar signatários** de um envelope
2. **Atualizar informações** de signatários se necessário
3. **Remover signatários** que não devem mais participar
4. **Reenviar signatários** para Clicksign após modificações

### 3. Casos de Uso Práticos

#### Contrato Empresarial com Múltiplos Signatários
```json
// Signatário 1: Representante da empresa (obrigatório)
{
  "name": "Diretor Comercial - Empresa ABC",
  "email": "diretor@empresa.com",
  "has_documentation": false,
  "refusable": false,
  "group": 1
}

// Signatário 2: Cliente (pode recusar)
{
  "name": "Cliente XYZ Ltda",
  "email": "contrato@cliente.com",
  "phone_number": "+5511987654321",
  "has_documentation": true,
  "refusable": true,
  "group": 2
}

// Signatário 3: Testemunha (opcional)
{
  "name": "João Silva - Testemunha",
  "email": "joao.silva@testemunha.com",
  "refusable": true,
  "group": 3,
  "communicate_events": {
    "document_signed": "none",
    "signature_request": "email",
    "signature_reminder": "sms"
  }
}
```

---

## Tratamento de Erros

### Erros de Validação (400)

#### Nome inválido:
```json
{
  "error": "Validation failed",
  "message": "Name is required and must be at least 2 characters"
}
```

#### Email inválido:
```json
{
  "error": "Validation failed", 
  "message": "invalid email format: email-invalido"
}
```

#### Data de nascimento inválida:
```json
{
  "error": "Validation failed",
  "message": "birthday must be in YYYY-MM-DD format, got: 15/03/1985"
}
```

#### Telefone inválido:
```json
{
  "error": "Validation failed",
  "message": "phone number must be in international format (+xxxxxxxx), got: 11987654321"
}
```

#### Grupo inválido:
```json
{
  "error": "Validation failed",
  "message": "group must be a positive integer, got: -1"
}
```

### Erros de Negócio (400/404)

#### Envelope não encontrado:
```json
{
  "error": "Envelope not found",
  "message": "The specified envelope does not exist"
}
```

#### Envelope não está pronto para envio:
```json
{
  "error": "Envelope not ready",
  "message": "Envelope must be created in Clicksign before sending signatories"
}
```

#### Nenhum signatário para enviar:
```json
{
  "error": "No signatories",
  "message": "Envelope must have at least one signatory before sending"
}
```

---

## Integração com Clicksign

### Mapeamento de Dados

O sistema mapeia automaticamente os dados dos signatários para o formato esperado pela API do Clicksign:

#### Dados Obrigatórios
- `name` → Nome completo do signatário
- `email` → Email para notificações e acesso

#### Dados Opcionais
- `birthday` → Data de nascimento para verificação
- `phone_number` → Telefone para notificações SMS
- `has_documentation` → Indica se possui documentos para verificação
- `refusable` → Se pode recusar a assinatura
- `group` → Ordem de assinatura

#### Configurações de Comunicação
- `communicate_events.document_signed` → Como notificar quando documento é assinado
- `communicate_events.signature_request` → Como notificar solicitação de assinatura  
- `communicate_events.signature_reminder` → Como notificar lembretes

### Sincronização

- **Criação local**: Signatários são criados primeiro no banco local
- **Envio para Clicksign**: Utilizar endpoint `/send` para sincronizar
- **Atualizações**: Modificações locais não são sincronizadas automaticamente
- **Exclusão**: Remover localmente não remove do Clicksign

---

## Notas Importantes

1. **Ordem de Criação**: Recomenda-se criar signatários antes de ativar o envelope
2. **Validação de Email**: Emails são validados tanto localmente quanto pelo Clicksign
3. **Grupos de Assinatura**: Utilize grupos para controlar a ordem de assinatura
4. **Notificações**: Configure adequadamente as preferências de comunicação
5. **Documentação**: Campo `has_documentation` afeta o processo de verificação no Clicksign
6. **Recusa**: Signatários com `refusable: false` não podem recusar a assinatura
7. **Telefone Internacional**: Sempre utilize formato internacional para números de telefone
8. **Sincronização**: Após modificações, use o endpoint `/send` para sincronizar com Clicksign