# API de Documentos

Esta documentação detalha todos os endpoints relacionados ao gerenciamento de documentos no microserviço ms-docsigner, incluindo upload via base64, consulta, atualização e exclusão de documentos.

## Endpoints Disponíveis

### Headers Obrigatórios para Todos os Endpoints
```
Authorization: Bearer <jwt_token>
Content-Type: application/json
```

## 1. Criar Documento
`POST /api/v1/documents`

Cria um novo documento usando file_path (caminho absoluto) ou conteúdo base64.

### Parâmetros do Request

| Campo | Tipo | Obrigatório | Descrição |
|-------|------|-------------|-----------|
| `name` | string | Sim | Nome do documento (3-255 caracteres) |
| `file_path` | string | Condicional | Caminho absoluto do arquivo (usar OU file_content_base64) |
| `file_content_base64` | string | Condicional | Conteúdo do arquivo em base64 (usar OU file_path). Máximo 7.5MB após decodificação |
| `file_size` | integer | Condicional | Tamanho do arquivo em bytes (obrigatório com file_path, opcional com base64) |
| `mime_type` | string | Condicional | Tipo MIME (obrigatório com file_path, opcional com base64) |
| `description` | string | Não | Descrição do documento (máx. 1000 caracteres) |

### Tipos MIME Suportados
- `application/pdf` - Documentos PDF
- `image/jpeg` - Imagens JPEG
- `image/png` - Imagens PNG
- `image/gif` - Imagens GIF

### Exemplo 1: Upload via Base64

```json
{
  "name": "Contrato de Prestação de Serviços",
  "file_content_base64": "JVBERi0xLjQKMSAwIG9iag0KPDwNCi9UeXBlIC9DYXRhbG9nDQovUGFnZXMgMiAwIFINCj4+DQplbmRvYmoNCjIgMCBvYmoNCjw8DQovVHlwZSAvUGFnZXMNCi9LaWRzIFs...",
  "description": "Documento para assinatura digital"
}
```

### Exemplo 2: Upload via File Path

```json
{
  "name": "Contrato de Prestação de Serviços",
  "file_path": "/uploads/documents/contrato_cliente_abc.pdf",
  "file_size": 2048576,
  "mime_type": "application/pdf",
  "description": "Documento para assinatura digital"
}
```

### Response de Sucesso (201)

```json
{
  "id": 1,
  "name": "Contrato de Prestação de Serviços",
  "file_path": "/tmp/processed_document_1627123456.pdf",
  "file_size": 2048576,
  "mime_type": "application/pdf",
  "status": "draft",
  "clicksign_key": "",
  "description": "Documento para assinatura digital",
  "created_at": "2025-07-19T10:00:00Z",
  "updated_at": "2025-07-19T10:00:00Z"
}
```

### Códigos de Erro
- `400` - Dados inválidos, arquivo muito grande (>7.5MB) ou tipo não suportado
- `401` - Token JWT ausente ou inválido
- `413` - Arquivo muito grande (> 7.5MB)
- `415` - Tipo de arquivo não suportado
- `500` - Erro interno do servidor

### Validações Específicas

#### Upload via Base64:
- Conteúdo base64 é automaticamente validado
- MIME type é detectado automaticamente se não fornecido
- Tamanho é calculado automaticamente após decodificação
- Limite máximo: 7.5MB após decodificação

#### Upload via File Path:
- `file_size` e `mime_type` são obrigatórios
- Arquivo deve existir no caminho especificado
- Validação de tipo MIME é obrigatória

---

## 2. Buscar Documento por ID
`GET /api/v1/documents/{id}`

Retorna um documento específico pelo ID.

### Parâmetros da URL
- `id` (integer): ID do documento

### Exemplo de Request
```bash
GET /api/v1/documents/1
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### Response de Sucesso (200)
```json
{
  "id": 1,
  "name": "Contrato de Prestação de Serviços",
  "file_path": "/tmp/processed_document_1627123456.pdf",
  "file_size": 2048576,
  "mime_type": "application/pdf",
  "status": "ready",
  "clicksign_key": "12345678-1234-1234-1234-123456789012",
  "description": "Documento para assinatura digital",
  "created_at": "2025-07-19T10:00:00Z",
  "updated_at": "2025-07-19T10:15:00Z"
}
```

### Códigos de Erro
- `400` - ID inválido
- `401` - Não autorizado
- `404` - Documento não encontrado
- `500` - Erro interno

---

## 3. Listar Documentos
`GET /api/v1/documents`

Retorna uma lista de documentos com filtros opcionais.

### Parâmetros de Query (opcionais)

| Parâmetro | Tipo | Descrição |
|-----------|------|-----------|
| `search` | string | Buscar por nome do documento |
| `status` | string | Filtrar por status (draft, ready, processing, sent) |
| `clicksign_key` | string | Filtrar por chave do Clicksign |

### Exemplos de Request

#### Listar todos os documentos
```bash
GET /api/v1/documents
```

#### Filtrar por status
```bash
GET /api/v1/documents?status=ready
```

#### Buscar por nome
```bash
GET /api/v1/documents?search=contrato
```

### Response de Sucesso (200)
```json
{
  "documents": [
    {
      "id": 1,
      "name": "Contrato de Prestação de Serviços",
      "file_path": "/tmp/processed_document_1627123456.pdf",
      "file_size": 2048576,
      "mime_type": "application/pdf",
      "status": "ready",
      "clicksign_key": "12345678-1234-1234-1234-123456789012",
      "description": "Documento para assinatura digital",
      "created_at": "2025-07-19T10:00:00Z",
      "updated_at": "2025-07-19T10:15:00Z"
    },
    {
      "id": 2,
      "name": "NDA - Acordo de Confidencialidade",
      "file_path": "/tmp/processed_document_1627123789.pdf",
      "file_size": 1536000,
      "mime_type": "application/pdf",
      "status": "draft",
      "clicksign_key": "",
      "description": "Acordo de confidencialidade padrão",
      "created_at": "2025-07-19T11:00:00Z",
      "updated_at": "2025-07-19T11:00:00Z"
    }
  ],
  "total": 2
}
```

---

## 4. Atualizar Documento
`PUT /api/v1/documents/{id}`

Atualiza um documento existente. Apenas campos fornecidos serão atualizados.

### Parâmetros da URL
- `id` (integer): ID do documento

### Parâmetros do Request (todos opcionais)

| Campo | Tipo | Descrição |
|-------|------|-----------|
| `name` | string | Novo nome do documento (3-255 caracteres) |
| `description` | string | Nova descrição (máx. 1000 caracteres) |
| `status` | string | Novo status (draft, ready, processing, sent) |

### Exemplo de Request
```bash
PUT /api/v1/documents/1
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
Content-Type: application/json

{
  "name": "Contrato de Prestação de Serviços - Atualizado",
  "description": "Documento atualizado com nova versão",
  "status": "ready"
}
```

### Response de Sucesso (200)
```json
{
  "id": 1,
  "name": "Contrato de Prestação de Serviços - Atualizado",
  "file_path": "/tmp/processed_document_1627123456.pdf",
  "file_size": 2048576,
  "mime_type": "application/pdf",
  "status": "ready",
  "clicksign_key": "12345678-1234-1234-1234-123456789012",
  "description": "Documento atualizado com nova versão",
  "created_at": "2025-07-19T10:00:00Z",
  "updated_at": "2025-07-19T10:30:00Z"
}
```

### Códigos de Erro
- `400` - Dados inválidos ou transição de status inválida
- `401` - Não autorizado
- `404` - Documento não encontrado
- `500` - Erro interno

---

## 5. Deletar Documento
`DELETE /api/v1/documents/{id}`

Remove um documento do sistema.

### Parâmetros da URL
- `id` (integer): ID do documento

### Exemplo de Request
```bash
DELETE /api/v1/documents/1
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### Response de Sucesso (200)
```json
{
  "message": "Documento deletado com sucesso"
}
```

### Códigos de Erro
- `400` - ID inválido
- `401` - Não autorizado
- `404` - Documento não encontrado
- `500` - Erro interno

---

## Estados do Documento

| Estado | Descrição |
|--------|-----------|
| `draft` | Documento criado, aguardando processamento |
| `ready` | Documento processado e pronto para uso em envelopes |
| `processing` | Documento sendo processado no Clicksign |
| `sent` | Documento enviado para assinatura |

---

## Exemplos de Uso Prático

### Exemplo 1: Upload de Contrato via Base64

**Cenário:** Sistema frontend precisa fazer upload de um PDF diretamente do browser.

```bash
curl -X POST https://api.ms-docsigner.com/api/v1/documents \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Contrato Cliente ABC",
    "file_content_base64": "JVBERi0xLjQKMSAwIG9iag0KPDwNCi9UeXBlIC9DYXRhbG9nDQovUGFnZXMgMiAwIFINCj4+DQplbmRvYmoNCjIgMCBvYmoNCjw8DQovVHlwZSAvUGFnZXMNCi9LaWRzIFs...",
    "description": "Contrato de prestação de serviços para cliente ABC"
  }'
```

### Exemplo 2: Upload de Documento via File Path

**Cenário:** Sistema backend com acesso ao filesystem.

```bash
curl -X POST https://api.ms-docsigner.com/api/v1/documents \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "NDA Funcionários",
    "file_path": "/uploads/2025/nda_template.pdf",
    "file_size": 1536000,
    "mime_type": "application/pdf",
    "description": "Template de NDA para novos funcionários"
  }'
```

### Exemplo 3: Atualização de Status

**Cenário:** Marcar documento como pronto para uso.

```bash
curl -X PUT https://api.ms-docsigner.com/api/v1/documents/1 \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "status": "ready"
  }'
```

### Exemplo 4: Busca por Documentos Prontos

**Cenário:** Listar documentos disponíveis para criação de envelopes.

```bash
curl -X GET "https://api.ms-docsigner.com/api/v1/documents?status=ready" \
  -H "Authorization: Bearer <token>"
```

---

## Integração com Envelopes

Após criar e preparar documentos, eles podem ser utilizados na criação de envelopes:

```bash
# 1. Criar documento
POST /api/v1/documents
{
  "name": "Contrato Cliente XYZ",
  "file_content_base64": "JVBERi0xLjQK..."
}

# 2. Usar documento em envelope
POST /api/v1/envelopes
{
  "name": "Envelope - Contrato Cliente XYZ",
  "documents_ids": [1],
  "signatory_emails": ["cliente@xyz.com", "empresa@example.com"]
}
```

---

## Tratamento de Erros

### Erro de Validação Base64

**Request com base64 inválido:**
```json
{
  "name": "Documento Teste",
  "file_content_base64": "invalid-base64-content"
}
```

**Response (400):**
```json
{
  "error": "Invalid base64",
  "message": "Base64 content is not valid or cannot be decoded"
}
```

### Erro de Arquivo Muito Grande

**Response (413):**
```json
{
  "error": "File too large",
  "message": "File size exceeds maximum limit of 7.5MB"
}
```

### Erro de Tipo de Arquivo Não Suportado

**Response (415):**
```json
{
  "error": "Unsupported file type",
  "message": "File type 'text/plain' is not supported. Supported types: PDF, JPEG, PNG, GIF"
}
```

---

## Monitoramento e Logs

Todas as operações incluem:
- `X-Correlation-ID` header para rastreabilidade
- Logs estruturados com contexto
- Métricas de performance
- Limpeza automática de arquivos temporários

### Header de Correlação

```bash
curl -X POST /api/v1/documents \
  -H "Authorization: Bearer <token>" \
  -H "X-Correlation-ID: custom-trace-id-123" \
  -H "Content-Type: application/json" \
  -d '{ ... }'
```

Se não fornecido, um ID de correlação será gerado automaticamente.