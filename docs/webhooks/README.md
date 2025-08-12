# Funcionalidade de Webhooks

## Visão Geral

A funcionalidade de webhooks permite que o sistema receba e processe eventos em tempo real do Clicksign, mantendo a sincronização entre os dois sistemas.

## Arquitetura

### Componentes

1. **Entity (Entidade)**
   - `EntityWebhook`: Representa um webhook no sistema

2. **Repository (Repositório)**
   - `RepositoryWebhook`: Gerencia operações de banco de dados para webhooks

3. **UseCase (Casos de Uso)**
   - `UsecaseWebhookService`: Implementa a lógica de negócio para processamento de webhooks

4. **Handler (Controlador)**
   - `WebhookHandler`: Gerencia as requisições HTTP relacionadas a webhooks

5. **DTOs (Data Transfer Objects)**
   - `WebhookRequestDTO`: Estrutura para receber webhooks do Clicksign
   - `WebhookResponseDTO`: Estrutura para respostas da API

## Eventos Suportados

### 1. auto_close
- **Descrição**: Disparado quando um envelope é fechado automaticamente
- **Processamento**: 
  - Atualiza o status do envelope para "completed"
  - Salva dados raw do webhook
- **Prioridade**: Alta (funcionalidade principal)

### 2. sign
- **Descrição**: Disparado quando um documento é assinado
- **Processamento**: Salva dados do evento para histórico

### 3. signature_started
- **Descrição**: Disparado quando o processo de assinatura é iniciado
- **Processamento**: Salva dados do evento para histórico

### 4. add_signer
- **Descrição**: Disparado quando um signatário é adicionado
- **Processamento**: Salva dados do evento para histórico

### 5. upload
- **Descrição**: Disparado quando um documento é enviado
- **Processamento**: Salva dados do evento para histórico

## Fluxo de Processamento

```
1. Clicksign envia webhook → POST /webhooks
2. Handler valida payload
3. UseCase processa evento específico
4. Repository salva no banco
5. Status é atualizado (pending → processed/failed)
```

## Configuração

### 1. Banco de Dados

Execute a migração para criar a tabela de webhooks:

```sql
-- Ver arquivo: docs/database/migrations/001_create_webhooks_table.sql
```

### 2. Clicksign

Configure a URL do webhook no painel administrativo do Clicksign:

```
https://seu-dominio.com/api/v1/webhooks
```

### 3. Variáveis de Ambiente

Certifique-se de que as seguintes variáveis estão configuradas:

```env
# Configurações do banco de dados
DB_HOST=localhost
DB_PORT=5432
DB_NAME=docsigner
DB_USER=postgres
DB_PASSWORD=password

# Configurações de log
LOG_LEVEL=INFO
```

## Monitoramento

### Endpoints de Monitoramento

1. **Webhooks Pendentes**
   ```
   GET /api/v1/webhooks/pending
   ```

2. **Webhooks que Falharam**
   ```
   GET /api/v1/webhooks/failed
   ```

3. **Webhooks por Documento**
   ```
   GET /api/v1/webhooks/document/{document_key}
   ```

### Logs

O sistema registra logs detalhados para cada webhook:

```json
{
  "level": "info",
  "msg": "Processing webhook",
  "event_name": "auto_close",
  "document_key": "4d210b08-54c0-4716-83a6-25d3b5f8f8ad",
  "account_key": "c34e9873-bb05-481c-9c54-144ca8da782c"
}
```

## Tratamento de Erros

### Cenários de Erro

1. **JSON Inválido**
   - Status: 400 Bad Request
   - Ação: Retorna erro de validação

2. **Dados Obrigatórios Faltando**
   - Status: 400 Bad Request
   - Ação: Retorna erro específico

3. **Erro de Processamento**
   - Status: 500 Internal Server Error
   - Ação: Webhook é marcado como "failed"

4. **Envelope Não Encontrado**
   - Status: 200 OK (não é erro)
   - Ação: Webhook é processado mas envelope não é atualizado

### Reprocessamento

Webhooks que falharam podem ser reprocessados:

```
POST /api/v1/webhooks/{id}/retry
```

## Testes

### Teste Manual

```bash
# Enviar webhook de teste
curl -X POST http://localhost:8080/api/v1/webhooks \
  -H "Content-Type: application/json" \
  -d '{
    "event": {
        "name": "auto_close",
        "occurred_at": "2025-08-11T11:43:22.282-03:00"
    },
    "document": {
        "key": "4d210b08-54c0-4716-83a6-25d3b5f8f8ad",
        "account_key": "c34e9873-bb05-481c-9c54-144ca8da782c",
        "status": "closed"
    }
}'
```

### Teste Automatizado

Os testes estão localizados em:
- `src/api/handlers/handlers_webhook_test.go`

## Extensibilidade

### Adicionando Novos Eventos

1. **Adicione o evento no DTO:**
   ```go
   func (w *WebhookRequestDTO) IsNewEvent() bool {
       return w.Event.Name == "new_event"
   }
   ```

2. **Adicione o processamento no UseCase:**
   ```go
   case "new_event":
       return u.ProcessNewEvent(webhookDTO, webhook)
   ```

3. **Implemente a função de processamento:**
   ```go
   func (u *UsecaseWebhookService) ProcessNewEvent(webhookDTO *dtos.WebhookRequestDTO, webhook *entity.EntityWebhook) error {
       // Lógica específica do evento
       return nil
   }
   ```

## Performance

### Otimizações Implementadas

1. **Índices no Banco**
   - `event_name`: Para filtrar por tipo de evento
   - `document_key`: Para buscar webhooks de um documento
   - `status`: Para filtrar por status
   - `created_at`: Para ordenação cronológica

2. **Processamento Assíncrono**
   - Webhooks são salvos imediatamente
   - Processamento específico é feito em background

3. **Logging Estruturado**
   - Logs em JSON para fácil parsing
   - Níveis de log configuráveis

## Segurança

### Validações Implementadas

1. **Validação de Payload**
   - Verificação de campos obrigatórios
   - Validação de formato JSON

2. **Sanitização de Dados**
   - Escape de caracteres especiais
   - Validação de tipos de dados

3. **Rate Limiting**
   - Implementar conforme necessário

## Troubleshooting

### Problemas Comuns

1. **Webhook não está sendo processado**
   - Verifique logs do sistema
   - Confirme se a URL está correta no Clicksign
   - Verifique se o banco está acessível

2. **Envelope não está sendo atualizado**
   - Verifique se o ClicksignKey está correto
   - Confirme se o envelope existe no sistema

3. **Erros de validação**
   - Verifique o formato do payload
   - Confirme se todos os campos obrigatórios estão presentes

### Comandos Úteis

```bash
# Verificar webhooks pendentes
curl http://localhost:8080/api/v1/webhooks/pending

# Verificar webhooks que falharam
curl http://localhost:8080/api/v1/webhooks/failed

# Reprocessar webhook falhado
curl -X POST http://localhost:8080/api/v1/webhooks/1/retry
``` 