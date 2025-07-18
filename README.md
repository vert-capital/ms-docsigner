# Go Clean Architecture Template

Este é um template de projeto Go que implementa os princípios da Clean Architecture, oferecendo uma base sólida para o desenvolvimento de APIs RESTful. O template inclui funcionalidades essenciais como autenticação JWT, integração com banco de dados PostgreSQL, mensageria com Kafka, e uma estrutura de testes robusta.

## Arquitetura

O projeto segue os princípios da **Clean Architecture** com separação clara de responsabilidades:

- **Entidades (entity/)**: Contém as regras de negócio fundamentais e estruturas de dados
- **Casos de Uso (usecase/)**: Implementa a lógica de negócio específica da aplicação
- **Infraestrutura (infrastructure/)**: Implementações de repositórios e integrações externas
- **Interface (api/)**: Camada de apresentação com handlers HTTP e middlewares
- **Injeção de Dependências**: Configurada no `main.go` para baixo acoplamento

## Stack Tecnológico

- **Linguagem**: Go 1.21+
- **Framework Web**: Gin
- **Banco de Dados**: PostgreSQL com GORM
- **Mensageria**: Apache Kafka
- **Autenticação**: JWT (JSON Web Tokens)
- **Criptografia**: bcrypt para senhas
- **Documentação**: Swagger/OpenAPI
- **Testes**: Go testing + Testify + GoConvey
- **Containerização**: Docker e Docker Compose
- **Monitoramento**: Elastic APM

## Pré-requisitos

Antes de começar, certifique-se de ter instalado:

- [Go 1.21+](https://golang.org/dl/)
- [Docker](https://www.docker.com/get-started)
- [Docker Compose](https://docs.docker.com/compose/install/)
- [Make](https://www.gnu.org/software/make/) (opcional, mas recomendado)

## Getting Started

### 1. Clonando o Projeto

```bash
git clone <seu-repositorio>
cd template_golang
```

### 2. Configuração do Ambiente

Copie o arquivo de exemplo de variáveis de ambiente:

```bash
make init
```

Ou manualmente:

```bash
cp src/.env.sample src/.env
```

### 3. Configuração das Variáveis de Ambiente

Edite o arquivo `src/.env` com suas configurações:

```env
# Configuração de Logging
LOG_LEVEL=DEBUG
GIN_MODE=debug
GORM_LOG_LEVEL=DEBUG

# Banco de Dados
POSTGRES_DB=gotemplate
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_HOST=db
POSTGRES_PORT=5432

# Kafka
KAFKA_BOOTSTRAP_SERVER=kafka:9092
KAFKA_CLIENT_ID=gotemplate
KAFKA_GROUP_ID=gotemplate

# JWT
JWT_SECRET_KEY=your-jwt-secret-key-here

# Email (opcional)
EMAIL_HOST=mail
EMAIL_PORT=1025
EMAIL_FROM=noreply@example.com

# Admin padrão
DEFAULT_ADMIN_MAIL=admin@example.com
DEFAULT_ADMIN_PASSWORD=admin123
```

### 4. Primeira Execução

Construa e execute os containers:

```bash
make up
```

A aplicação estará disponível em:
- **API**: http://localhost:8080
- **Swagger**: http://localhost:8080/swagger/index.html

### 5. Comandos Úteis

```bash
# Ver logs da aplicação
make log

# Ver logs de todos os serviços
make logs

# Parar os serviços
make stop

# Executar testes
make test

# Executar testes com cobertura
make coverage

# Executar shell no container
make sh app

# Gerar documentação Swagger
make swagger
```

## Estrutura do Projeto

```
template_golang/
├── src/                          # Código fonte da aplicação
│   ├── api/                      # Camada de apresentação
│   │   ├── handlers/            # Handlers HTTP
│   │   └── middleware/          # Middlewares
│   ├── entity/                  # Entidades de domínio
│   ├── usecase/                 # Casos de uso (lógica de negócio)
│   │   └── user/               # Exemplo: casos de uso de usuário
│   ├── infrastructure/          # Infraestrutura
│   │   ├── postgres/           # Configuração do banco
│   │   └── repository/         # Implementação de repositórios
│   ├── kafka/                   # Integração com Kafka
│   ├── pkg/                     # Pacotes utilitários
│   │   ├── auth/               # Autenticação e autorização
│   │   ├── logger/             # Sistema de logs
│   │   └── utils/              # Utilitários gerais
│   ├── mocks/                   # Mocks para testes
│   └── main.go                  # Ponto de entrada da aplicação
├── docs/                        # Documentação do projeto
├── docker-compose.yml           # Configuração do Docker Compose
├── Makefile                     # Comandos automatizados
└── README.md                    # Este arquivo
```

## Testando a Aplicação

### Executar todos os testes

```bash
make test
```

### Executar testes com cobertura

```bash
make coverage
```

### Executar testes em modo watch

```bash
make test-watch
```

### Interface web para testes (GoConvey)

```bash
make test-watch-web
```

## Adicionando Novas Funcionalidades

### Criando uma Nova Entidade

1. **Defina a entidade** em `src/entity/`:

```go
// src/entity/entity_produto.go
package entity

import "time"

type Produto struct {
    ID        uint      `json:"id" gorm:"primaryKey"`
    Nome      string    `json:"nome" gorm:"size:100;not null"`
    Preco     float64   `json:"preco" gorm:"not null"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

2. **Crie os testes da entidade** em `src/entity/entity_produto_test.go`

### Criando um Novo Caso de Uso

1. **Defina a interface** em `src/usecase/produto/`:

```go
// src/usecase/produto/usecase_produto_interface.go
package produto

import "app/entity"

type ProdutoUseCaseInterface interface {
    CreateProduto(produto *entity.Produto) (*entity.Produto, error)
    GetProduto(id uint) (*entity.Produto, error)
    // ... outros métodos
}
```

2. **Implemente o serviço** em `src/usecase/produto/usecase_produto_service.go`

3. **Crie os testes** em `src/usecase/produto/usecase_produto_service_test.go`

### Criando um Novo Repositório

1. **Implemente o repositório** em `src/infrastructure/repository/`:

```go
// src/infrastructure/repository/repository_produto.go
package repository

import (
    "app/entity"
    "gorm.io/gorm"
)

type ProdutoRepository struct {
    db *gorm.DB
}

func NewProdutoRepository(db *gorm.DB) *ProdutoRepository {
    return &ProdutoRepository{db: db}
}

func (r *ProdutoRepository) Create(produto *entity.Produto) (*entity.Produto, error) {
    if err := r.db.Create(produto).Error; err != nil {
        return nil, err
    }
    return produto, nil
}
```

### Criando Novos Handlers

1. **Implemente os handlers** em `src/api/handlers/`:

```go
// src/api/handlers/handlers_produto.go
package handlers

import (
    "net/http"
    "github.com/gin-gonic/gin"
)

func (h *Handlers) CreateProduto(c *gin.Context) {
    // Implementação do handler
}
```

2. **Registre as rotas** em `src/api/api.go`

### Exemplo Completo: Gerador CRUD

O template inclui um gerador automático de CRUD. Para usá-lo:

```bash
# Instalar o gerador
make install_generator

# Gerar novo CRUD
make generator_crud
```

## Gerenciamento de Dependências

### Adicionar nova dependência

```bash
make dep_install github.com/nova/dependencia
```

### Atualizar dependências

```bash
make auto_install
```

### Limpar módulos

```bash
make mod_tidy
```

## Configuração para Produção

### Variáveis de Ambiente para Produção

```env
# Logs mínimos para produção
LOG_LEVEL=WARN
GIN_MODE=release
GORM_LOG_LEVEL=ERROR

# Configurações de segurança
JWT_SECRET_KEY=sua-chave-secreta-super-forte
ISRELEASE=true

# Configurações de banco específicas
POSTGRES_HOST=seu-host-producao
POSTGRES_DB=seu-banco-producao
# ... outras configurações
```

### Build para Produção

```bash
docker-compose -f docker-compose.yml build --no-cache
```

## Contribuindo

1. Faça um fork do projeto
2. Crie uma branch para sua feature (`git checkout -b feature/nova-feature`)
3. Commit suas mudanças (`git commit -am 'Adiciona nova feature'`)
4. Push para a branch (`git push origin feature/nova-feature`)
5. Crie um Pull Request

## Licença

Este projeto está sob a licença MIT. Veja o arquivo `LICENSE` para mais detalhes.

## Suporte

Se você encontrar problemas ou tiver dúvidas:

1. Verifique a documentação em `docs/`
2. Consulte os logs com `make log`
3. Verifique os testes com `make test`
4. Abra uma issue no repositório

## Recursos Adicionais

- [Documentação da Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Go Best Practices](https://golang.org/doc/effective_go.html)
- [Gin Framework](https://gin-gonic.com/)
- [GORM Documentation](https://gorm.io/)
- [Swagger/OpenAPI](https://swagger.io/)