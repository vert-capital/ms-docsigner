# Instruções para I.A - Projeto MS-DocSigner

## Configurações Gerais

- **Idioma**: Sempre responda em português do Brasil
- **Contexto**: Este é um microserviço Go para assinatura de documentos com integração Clicksign

## Regras de Desenvolvimento

### ⚠️ IMPORTANTE - Gerenciamento de Containers

- **NUNCA faça rebuild do container** - O projeto tem hot reload configurado
- **NUNCA execute comandos como `docker build` ou `docker-compose build`**
- Após alterações no código, use apenas `make log` para verificar se as mudanças foram aplicadas

### Comandos Essenciais do Projeto

```bash
# Iniciar o ambiente de desenvolvimento (se não estiver rodando)
make up

# Verificar status dos containers
make status

# Visualizar logs em tempo real (após alterações de código)
make log

# Parar todos os containers (se necessário)
make down
```

### Estrutura do Projeto

- **Linguagem**: Go
- **Arquitetura**: Clean Architecture
- **Banco de dados**: PostgreSQL
- **Containerização**: Docker com hot reload habilitado

### Fluxo de Desenvolvimento

1. Faça alterações no código Go em `src/`
2. O hot reload detectará automaticamente as mudanças
3. Use `make log` para verificar se a aplicação foi recarregada

### Observações para I.A

- Sempre considere a arquitetura limpa do projeto ao sugerir alterações
- Mantenha a separação de responsabilidades (entity, usecase, infrastructure)
- Quando criar novos endpoints, seguir o padrão existente em `src/api/handlers/`
- Para testes, usar os mocks já disponíveis em `src/mocks/`
