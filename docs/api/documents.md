# API de Documentos

Esta documentação detalha todos os endpoints relacionados ao gerenciamento de documentos no microserviço ms-docsigner.

## Endpoints Disponíveis

### 1. Criar Documento
`POST /api/v1/documents`

Cria um novo documento no sistema usando file_path (caminho de arquivo) ou file_content_base64 (conteúdo em base64).

#### Headers Obrigatórios
```
Authorization: Bearer <jwt_token>
Content-Type: application/json
```

#### Parâmetros do Request

| Campo | Tipo | Obrigatório | Descrição |
|-------|------|-------------|-----------|
| `name` | string | Sim | Nome do documento (3-255 caracteres) |
| `file_path` | string | Não* | Caminho absoluto do arquivo |
| `file_content_base64` | string | Não* | Conteúdo do arquivo em base64 (máx. 7.5MB) |
| `file_size` | integer | Condicional | Tamanho em bytes (obrigatório com file_path) |
| `mime_type` | string | Condicional | Tipo MIME (obrigatório com file_path) |
| `description` | string | Não | Descrição do documento (máx. 1000 caracteres) |

**Nota**: É obrigatório fornecer `file_path` OU `file_content_base64`, não ambos.

#### Tipos MIME Suportados
- `application/pdf`
- `image/jpeg`
- `image/png`
- `image/gif`

#### Exemplos de Request

##### Exemplo 1: Upload via Base64
```json
{
  "name": "Contrato de Prestação de Serviços",
  "file_content_base64": "JVBERi0xLjQKMyAwIG9iago8PAovVHlwZSAvQ2F0YWxvZwovT3V0bGluZXMgMiAwIFIKL1BhZ2VzIDQgMCBSCj4+CmVuZG9iago0IDAgb2JqCjw8Ci9UeXBlIC9QYWdlcwo...",
  "description": "Documento para assinatura digital"
}
```

##### Exemplo 2: Upload via File Path
```json
{
  "name": "NDA - Acordo de Confidencialidade",
  "file_path": "/uploads/nda_template.pdf",
  "file_size": 2048576,
  "mime_type": "application/pdf",
  "description": "Template de NDA para novos funcionários"
}
```

#### Response de Sucesso (201)
```json
{
  "id": 1,
  "name": "Contrato de Prestação de Serviços",
  "file_path": "/tmp/temp_12345.pdf",
  "file_size": 1048576,
  "mime_type": "application/pdf",
  "status": "draft",
  "clicksign_key": "",
  "description": "Documento para assinatura digital",
  "created_at": "2025-07-19T10:00:00Z",
  "updated_at": "2025-07-19T10:00:00Z"
}
```

#### Códigos de Erro
- `400` - Dados inválidos ou arquivo muito grande
- `401` - Token JWT ausente ou inválido
- `413` - Arquivo excede 7.5MB
- `415` - Tipo de arquivo não suportado
- `500` - Erro interno do servidor

---

### 2. Buscar Documento por ID
`GET /api/v1/documents/{id}`

Retorna um documento específico pelo ID.

#### Headers Obrigatórios
```
Authorization: Bearer <jwt_token>
```

#### Parâmetros da URL
- `id` (integer): ID do documento

#### Exemplo de Request
```bash
GET /api/v1/documents/1
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

#### Response de Sucesso (200)
```json
{
  "id": 1,
  "name": "Contrato de Prestação de Serviços",
  "file_path": "/uploads/contrato_2025.pdf",
  "file_size": 1048576,
  "mime_type": "application/pdf",
  "status": "ready",
  "clicksign_key": "abc123-def456-ghi789",
  "description": "Documento para assinatura digital",
  "created_at": "2025-07-19T10:00:00Z",
  "updated_at": "2025-07-19T10:15:00Z"
}
```

#### Códigos de Erro
- `400` - ID inválido
- `401` - Não autorizado
- `404` - Documento não encontrado
- `500` - Erro interno

---

### 3. Listar Documentos
`GET /api/v1/documents`

Retorna uma lista de documentos com filtros opcionais.

#### Headers Obrigatórios
```
Authorization: Bearer <jwt_token>
```

#### Parâmetros de Query (opcionais)

| Parâmetro | Tipo | Descrição |
|-----------|------|-----------|
| `search` | string | Buscar por nome do documento |
| `status` | string | Filtrar por status (draft, ready, processing, sent) |
| `clicksign_key` | string | Filtrar por chave do Clicksign |

#### Exemplos de Request

##### Listar todos os documentos
```bash
GET /api/v1/documents
```

##### Filtrar por status
```bash
GET /api/v1/documents?status=ready
```

##### Buscar por nome
```bash
GET /api/v1/documents?search=contrato
```

##### Múltiplos filtros
```bash
GET /api/v1/documents?status=ready&search=nda
```

#### Response de Sucesso (200)
```json
{
  "documents": [
    {
      "id": 1,
      "name": "Contrato de Prestação de Serviços",
      "file_path": "/uploads/contrato_2025.pdf",
      "file_size": 1048576,
      "mime_type": "application/pdf",
      "status": "ready",
      "clicksign_key": "abc123-def456-ghi789",
      "description": "Documento para assinatura digital",
      "created_at": "2025-07-19T10:00:00Z",
      "updated_at": "2025-07-19T10:15:00Z"
    },
    {
      "id": 2,
      "name": "NDA - Acordo de Confidencialidade",
      "file_path": "/uploads/nda_template.pdf",
      "file_size": 512000,
      "mime_type": "application/pdf",
      "status": "draft",
      "clicksign_key": "",
      "description": "Template de NDA",
      "created_at": "2025-07-19T11:00:00Z",
      "updated_at": "2025-07-19T11:00:00Z"
    }
  ],
  "total": 2
}
```

---

### 4. Atualizar Documento
`PUT /api/v1/documents/{id}`

Atualiza um documento existente.

#### Headers Obrigatórios
```
Authorization: Bearer <jwt_token>
Content-Type: application/json
```

#### Parâmetros da URL
- `id` (integer): ID do documento

#### Parâmetros do Request (todos opcionais)

| Campo | Tipo | Descrição |
|-------|------|-----------|
| `name` | string | Novo nome do documento (3-255 caracteres) |
| `description` | string | Nova descrição (máx. 1000 caracteres) |
| `status` | string | Novo status (draft, ready, processing, sent) |

#### Exemplo de Request
```json
{
  "name": "Contrato de Prestação de Serviços - Atualizado",
  "description": "Documento atualizado com nova versão",
  "status": "ready"
}
```

#### Response de Sucesso (200)
```json
{
  "id": 1,
  "name": "Contrato de Prestação de Serviços - Atualizado",
  "file_path": "/uploads/contrato_2025.pdf",
  "file_size": 1048576,
  "mime_type": "application/pdf",
  "status": "ready",
  "clicksign_key": "abc123-def456-ghi789",
  "description": "Documento atualizado com nova versão",
  "created_at": "2025-07-19T10:00:00Z",
  "updated_at": "2025-07-19T12:30:00Z"
}
```

#### Códigos de Erro
- `400` - Dados inválidos ou transição de status inválida
- `401` - Não autorizado
- `404` - Documento não encontrado
- `500` - Erro interno

---

### 5. Deletar Documento
`DELETE /api/v1/documents/{id}`

Remove um documento do sistema.

#### Headers Obrigatórios
```
Authorization: Bearer <jwt_token>
```

#### Parâmetros da URL
- `id` (integer): ID do documento

#### Exemplo de Request
```bash
DELETE /api/v1/documents/1
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

#### Response de Sucesso (200)
```json
{
  "message": "Documento deletado com sucesso"
}
```

#### Códigos de Erro
- `400` - ID inválido
- `401` - Não autorizado
- `404` - Documento não encontrado
- `500` - Erro interno

---

## Estados do Documento

| Estado | Descrição |
|--------|-----------|
| `draft` | Documento criado, ainda não processado |
| `ready` | Documento processado e pronto para uso |
| `processing` | Documento sendo processado |
| `sent` | Documento enviado para assinatura |

## Validações e Limitações

### Tamanho de Arquivo
- **Máximo**: 7.5MB após decodificação do base64
- **Verificação**: Automática durante o upload

### Tipos de Arquivo
- **PDF**: `application/pdf`
- **JPEG**: `image/jpeg`
- **PNG**: `image/png`
- **GIF**: `image/gif`

### Validações de Campo
- **Nome**: Obrigatório, 3-255 caracteres
- **Descrição**: Opcional, máximo 1000 caracteres
- **Base64**: Detecção automática de MIME type e tamanho

## Exemplos de Curl

### Criar documento via base64
```bash
curl -X POST https://api.ms-docsigner.com/api/v1/documents \
  -H "Authorization: Bearer <seu-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Meu Contrato",
    "file_content_base64": "JVBERi0xLjQKM...",
    "description": "Contrato de prestação de serviços"
  }'
```

### Buscar documento por ID
```bash
curl -X GET https://api.ms-docsigner.com/api/v1/documents/1 \
  -H "Authorization: Bearer <seu-token>"
```

### Listar documentos com filtro
```bash
curl -X GET "https://api.ms-docsigner.com/api/v1/documents?status=ready&search=contrato" \
  -H "Authorization: Bearer <seu-token>"
```

### Atualizar documento
```bash
curl -X PUT https://api.ms-docsigner.com/api/v1/documents/1 \
  -H "Authorization: Bearer <seu-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Contrato Atualizado",
    "status": "ready"
  }'
```

### Deletar documento
```bash
curl -X DELETE https://api.ms-docsigner.com/api/v1/documents/1 \
  -H "Authorization: Bearer <seu-token>"
```

## Troubleshooting Comum

### Erro 413 - Arquivo muito grande
**Problema**: Upload de arquivo excede 7.5MB
**Solução**: Comprimir o arquivo ou dividir em partes menores

### Erro 415 - Tipo não suportado
**Problema**: Formato de arquivo não aceito
**Solução**: Converter para PDF, JPEG, PNG ou GIF

### Erro 400 - Base64 inválido
**Problema**: Conteúdo base64 corrompido
**Solução**: Verificar codificação e integridade do base64

### Erro 401 - Token inválido
**Problema**: JWT expirado ou malformado
**Solução**: Renovar token de autenticação