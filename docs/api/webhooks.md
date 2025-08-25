# API de Webhooks

Esta documentação descreve a API de webhooks do sistema, que permite receber e processar eventos do Clicksign.

## Visão Geral

A API de webhooks é responsável por:

1. **Receber webhooks** do Clicksign com informações sobre eventos de documentos
2. **Processar eventos** específicos como fechamento automático de envelopes
3. **Armazenar histórico** de todos os webhooks recebidos
4. **Gerenciar status** dos webhooks (pendente, processado, falhou)
5. **Permitir reprocessamento** de webhooks que falharam

## Endpoints

### POST /api/v1/webhooks

Recebe um webhook do Clicksign.

**Payload de exemplo:**
```json
{
    "event": {
        "name": "auto_close",
        "data": null,
        "occurred_at": "2025-08-11T11:43:22.282-03:00"
    },
    "document": {
        "key": "4d210b08-54c0-4716-83a6-25d3b5f8f8ad",
        "account_key": "c34e9873-bb05-481c-9c54-144ca8da782c",
        "status": "closed",
        "auto_close": true
    }
}
```

**Resposta de sucesso (200):**
```json
{
    "id": 1,
    "event_name": "auto_close",
    "document_key": "4d210b08-54c0-4716-83a6-25d3b5f8f8ad",
    "account_key": "c34e9873-bb05-481c-9c54-144ca8da782c",
    "status": "processed",
    "processed_at": "2025-08-11T14:43:22.000Z",
    "created_at": "2025-08-11T14:43:22.000Z",
    "updated_at": "2025-08-11T14:43:22.000Z"
}
```

### GET /api/v1/webhooks

Lista webhooks com filtros opcionais.

**Parâmetros de query:**
- `event_name` (opcional): Nome do evento
- `document_key` (opcional): Chave do documento
- `account_key` (opcional): Chave da conta
- `status` (opcional): Status do webhook (pending, processed, failed)
- `page` (opcional): Página (padrão: 1)
- `limit` (opcional): Limite por página (padrão: 10)

**Exemplo de requisição:**
```
GET /api/v1/webhooks?event_name=auto_close&status=processed&page=1&limit=20
```

**Resposta:**
```json
{
    "webhooks": [
        {
            "id": 1,
            "event_name": "auto_close",
            "document_key": "4d210b08-54c0-4716-83a6-25d3b5f8f8ad",
            "account_key": "c34e9873-bb05-481c-9c54-144ca8da782c",
            "status": "processed",
            "processed_at": "2025-08-11T14:43:22.000Z",
            "created_at": "2025-08-11T14:43:22.000Z",
            "updated_at": "2025-08-11T14:43:22.000Z"
        }
    ],
    "total": 1
}
```

### GET /api/v1/webhooks/pending

Lista webhooks pendentes de processamento.

### GET /api/v1/webhooks/failed

Lista webhooks que falharam no processamento.

### GET /api/v1/webhooks/document/{document_key}

Lista webhooks de um documento específico.

### GET /api/v1/webhooks/{id}

Busca um webhook específico por ID.

### POST /api/v1/webhooks/{id}/retry

Tenta reprocessar um webhook que falhou.

**Resposta:**
```json
{
    "success": true,
    "message": "Webhook marcado para reprocessamento"
}
```

### DELETE /api/v1/webhooks/{id}

Remove um webhook do sistema.

**Resposta:**
```json
{
    "success": true,
    "message": "Webhook deletado com sucesso"
}
```

## Tipos de Eventos

### auto_close

Evento disparado quando um envelope é fechado automaticamente pelo Clicksign.

**Processamento:**
- Verifica se o documento está com status "closed"
- Busca o envelope correspondente pelo ClicksignKey
- Atualiza o status do envelope para "completed"
- Salva os dados raw do webhook no envelope

### sign

Evento disparado quando um documento é assinado.

### signature_started

Evento disparado quando o processo de assinatura é iniciado.

### add_signer

Evento disparado quando um signatário é adicionado ao documento.

### upload

Evento disparado quando um documento é enviado.

## Status dos Webhooks

- **pending**: Webhook recebido, aguardando processamento
- **processed**: Webhook processado com sucesso
- **failed**: Webhook falhou no processamento

## Configuração no Clicksign

Para receber webhooks do Clicksign, configure a URL do webhook no painel administrativo:

```
https://seu-dominio.com/api/v1/webhooks
```

## Tratamento de Erros

### Erro 400 - Bad Request
- JSON inválido
- Dados obrigatórios faltando
- ID inválido

### Erro 404 - Not Found
- Webhook não encontrado

### Erro 500 - Internal Server Error
- Erro no processamento do webhook
- Erro de banco de dados

## Logs

Todos os webhooks são logados com as seguintes informações:
- Evento recebido
- Document key
- Account key
- Status do processamento
- Erros (se houver)

## Monitoramento

Para monitorar o funcionamento dos webhooks:

1. **Verifique webhooks pendentes:**
   ```
   GET /api/v1/webhooks/pending
   ```

2. **Verifique webhooks que falharam:**
   ```
   GET /api/v1/webhooks/failed
   ```

3. **Reprocesse webhooks falhados:**
   ```
   POST /api/v1/webhooks/{id}/retry
   ```

## Exemplo de Integração

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

# Verificar webhooks processados
curl http://localhost:8080/api/v1/webhooks?status=processed

# Verificar webhooks de um documento específico
curl http://localhost:8080/api/v1/webhooks/document/4d210b08-54c0-4716-83a6-25d3b5f8f8ad
``` 