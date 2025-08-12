 # Resumo da Implementação - Funcionalidade de Webhooks

## Arquivos Criados/Modificados

### 1. Entidade
- **`src/entity/entity_webhook.go`** - Nova entidade para representar webhooks

### 2. DTOs
- **`src/api/handlers/dtos/webhook_dto.go`** - DTOs para requisições e respostas de webhook

### 3. Repository
- **`src/infrastructure/repository/repository_webhook.go`** - Repositório para operações de banco de dados

### 4. UseCase
- **`src/usecase/webhook/usecase_webhook_interface.go`** - Interface do usecase
- **`src/usecase/webhook/usecase_webhook_service.go`** - Implementação do usecase

### 5. Handlers
- **`src/api/handlers/handlers_webhook.go`** - Handlers HTTP para webhooks
- **`src/api/handlers/handlers_webhook_mount.go`** - Montagem das rotas

### 6. API Principal
- **`src/api/api.go`** - Adicionada chamada para montar handlers de webhook

### 7. Documentação
- **`docs/api/webhooks.md`** - Documentação da API
- **`docs/webhooks/README.md`** - Documentação técnica detalhada
- **`docs/database/migrations/001_create_webhooks_table.sql`** - Script de migração

## Funcionalidades Implementadas

### ✅ Recebimento de Webhooks
- Endpoint `POST /webhooks` para receber webhooks do Clicksign
- Validação de payload JSON
- Validação de campos obrigatórios
- Logging estruturado

### ✅ Processamento de Eventos
- **auto_close**: Atualiza status do envelope para "completed"
- **sign**: Salva dados do evento para histórico
- **signature_started**: Salva dados do evento para histórico
- **add_signer**: Salva dados do evento para histórico
- **upload**: Salva dados do evento para histórico

### ✅ Gerenciamento de Webhooks
- Listagem com filtros (`GET /webhooks`)
- Busca por ID (`GET /webhooks/{id}`)
- Busca por document key (`GET /webhooks/document/{document_key}`)
- Listagem de pendentes (`GET /webhooks/pending`)
- Listagem de falhados (`GET /webhooks/failed`)

### ✅ Reprocessamento
- Endpoint para reprocessar webhooks falhados (`POST /webhooks/{id}/retry`)
- Reset de status de "failed" para "pending"

### ✅ Exclusão
- Endpoint para deletar webhooks (`DELETE /webhooks/{id}`)

### ✅ Banco de Dados
- Tabela `webhooks` com índices otimizados
- Campos para rastreamento completo
- Suporte a diferentes status

## Estrutura da Tabela

```sql
CREATE TABLE webhooks (
    id SERIAL PRIMARY KEY,
    event_name VARCHAR(255) NOT NULL,
    event_data TEXT,
    document_key VARCHAR(255),
    account_key VARCHAR(255),
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    processed_at TIMESTAMP,
    error TEXT,
    raw_payload TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

## Endpoints Disponíveis

| Método | Endpoint                                   | Descrição                       |
| ------ | ------------------------------------------ | ------------------------------- |
| POST   | `/api/v1/webhooks`                         | Receber webhook do Clicksign    |
| GET    | `/api/v1/webhooks`                         | Listar webhooks com filtros     |
| GET    | `/api/v1/webhooks/pending`                 | Listar webhooks pendentes       |
| GET    | `/api/v1/webhooks/failed`                  | Listar webhooks que falharam    |
| GET    | `/api/v1/webhooks/document/{document_key}` | Listar webhooks de um documento |
| GET    | `/api/v1/webhooks/{id}`                    | Buscar webhook por ID           |
| POST   | `/api/v1/webhooks/{id}/retry`              | Reprocessar webhook             |
| DELETE | `/api/v1/webhooks/{id}`                    | Deletar webhook                 |

## Fluxo de Processamento

1. **Recebimento**: Clicksign envia webhook para `POST /api/v1/webhooks`
2. **Validação**: Handler valida payload e campos obrigatórios
3. **Persistência**: Webhook é salvo no banco com status "pending"
4. **Processamento**: UseCase processa evento específico
5. **Atualização**: Status é atualizado para "processed" ou "failed"
6. **Logging**: Todo o processo é logado

## Processamento Específico - auto_close

Para o evento `auto_close` (funcionalidade principal):

1. Verifica se o documento está com status "closed"
2. Busca o envelope correspondente pelo ClicksignKey
3. Atualiza o status do envelope para "completed"
4. Salva os dados raw do webhook no envelope
5. Marca o webhook como processado

## Configuração Necessária

### 1. Banco de Dados
Execute o script de migração:
```sql
-- docs/database/migrations/001_create_webhooks_table.sql
```

### 2. Clicksign
Configure a URL do webhook no painel administrativo:
```
https://seu-dominio.com/api/v1/webhooks
```

## Monitoramento

### Endpoints de Monitoramento
- `GET /api/v1/webhooks/pending` - Verificar webhooks pendentes
- `GET /api/v1/webhooks/failed` - Verificar webhooks que falharam
- `GET /api/v1/webhooks/document/{document_key}` - Verificar webhooks de um documento

### Logs
O sistema registra logs detalhados para cada webhook com informações como:
- Evento recebido
- Document key
- Account key
- Status do processamento
- Erros (se houver)

## Extensibilidade

A arquitetura permite fácil adição de novos tipos de eventos:

1. Adicionar método no DTO para identificar o evento
2. Adicionar case no switch do processamento
3. Implementar função específica de processamento

## Próximos Passos

### Melhorias Sugeridas

1. **Autenticação**: Implementar autenticação para os webhooks
2. **Rate Limiting**: Adicionar limitação de taxa
3. **Retry Automático**: Implementar retry automático para webhooks falhados
4. **Métricas**: Adicionar métricas de performance
5. **Notificações**: Implementar notificações para webhooks falhados
6. **Testes**: Adicionar testes unitários e de integração

### Funcionalidades Futuras

1. **Webhook Signing**: Implementar assinatura de webhooks para segurança
2. **Event Sourcing**: Implementar event sourcing para auditoria completa
3. **Webhook Templates**: Criar templates para diferentes tipos de eventos
4. **Dashboard**: Interface web para monitoramento de webhooks

## Conclusão

A implementação da funcionalidade de webhooks está completa e funcional, seguindo as melhores práticas de arquitetura limpa e padrões de projeto. O sistema está preparado para receber e processar eventos do Clicksign, com foco especial no evento `auto_close` conforme solicitado.

A estrutura é extensível e permite fácil adição de novos tipos de eventos no futuro, mantendo a organização e manutenibilidade do código. 