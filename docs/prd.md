# ms-docsigner Brownfield Enhancement PRD

## Intro Project Analysis and Context

### Existing Project Overview

#### Analysis Source
- IDE-based fresh analysis, documented in `docs/project-analysis.md`.

#### Current Project State
O `ms-docsigner` é um serviço Go que serve como um template robusto para novas aplicações, seguindo os princípios da Clean Architecture. Ele expõe uma API REST para gerenciamento de usuários, integra-se com Kafka para processamento assíncrono e utiliza PostgreSQL como banco de dados.

### Available Documentation Analysis

#### Available Documentation
- [x] Tech Stack Documentation
- [x] Source Tree/Architecture
- [x] Coding Standards (parcialmente inferido da análise)
- [x] API Documentation (via Swaggo)
- [ ] External API Documentation
- [ ] UX/UI Guidelines
- [ ] Technical Debt Documentation

*Nota: A análise completa do projeto está disponível em `docs/project-analysis.md`.*

### Enhancement Scope Definition

#### Enhancement Type
- [x] New Feature Addition (O próprio template é a "feature")
- [x] Technology Stack Upgrade (Padronização da stack para novos projetos)

#### Enhancement Description
Um template base para microsserviços Go que expõem APIs REST, se comunicam via Kafka e seguem a Clean Architecture, usado para subir projetos rápidos e padronizados.

#### Impact Assessment
- [x] Minimal Impact (O template é autocontido e serve como ponto de partida)

### Goals and Background Context

#### Goals
- Acelerar o tempo de desenvolvimento de novos microsserviços.
- Garantir a aplicação consistente da Clean Architecture em todos os projetos.
- Reduzir o tempo de configuração inicial (boilerplate).
- Padronizar logging, monitoramento e tratamento de erros.
- Facilitar o onboarding de novos desenvolvedores.
- Deixar o BMAD Method pré-configurado para acelerar o ciclo de vida de desenvolvimento.

#### Background Context
Este projeto foi criado para servir como uma fundação sólida e padronizada para o desenvolvimento de novos microsserviços em Go na organização. Ao encapsular as melhores práticas de Clean Architecture, configuração de ambiente, e integração com serviços essenciais como Kafka e PostgreSQL, ele resolve o problema de inconsistência entre projetos e o tempo gasto em configurações repetitivas. O objetivo é que qualquer nova equipe de desenvolvimento possa iniciar um novo serviço com uma base produtiva e de alta qualidade desde o primeiro dia, já integrada com o fluxo de trabalho do BMAD-Method.

### Change Log

| Change | Date | Version | Description | Author |
| :--- | :--- | :--- | :--- | :--- |
| Criação do PRD | 2025-07-18 | 1.0 | Versão inicial do PRD para o template base. | @pm (John) |

## Requisitos

Esta seção detalha os requisitos que o template deve cumprir para garantir que qualquer novo projeto construído a partir dele seja robusto, escalável e fácil de manter. As explicações são direcionadas a um desenvolvedor que está utilizando o template.

### Requisitos Funcionais (FR)

-   **FR1: API RESTful Pronta para Uso**
    -   **O que significa:** Ao iniciar um novo projeto, você já terá um servidor web (Gin) funcionando. Ele virá com um endpoint `/health` (essencial para health checks em ambientes como Docker e Kubernetes) e um exemplo completo de CRUD para a entidade `User`. Você verá na prática como uma rota é definida, como o handler processa a requisição e como ele chama a camada de lógica de negócio.
    -   **Por que é importante:** Isso economiza o tempo de escrever todo o código repetitivo de configuração do servidor e te dá um exemplo claro e funcional de como adicionar novos endpoints para as suas próprias funcionalidades.

-   **FR2: Integração com Kafka Funcional**
    -   **O que significa:** O template já inclui o código necessário para se conectar a um broker Kafka, enviar e receber mensagens. Haverá um exemplo de um "producer" (que envia uma mensagem quando um usuário é criado, por exemplo) e um "consumer" (que processa essa mensagem).
    -   **Por que é importante:** Remove a complexidade inicial de configurar clientes Kafka. Você pode simplesmente copiar e adaptar os exemplos para integrar seus próprios eventos, sem precisar se aprofundar na configuração da biblioteca.

-   **FR3: Camada de Persistência com GORM e Postgres**
    -   **O que significa:** A conexão com o banco de dados já está configurada. O projeto inclui uma implementação completa de um "repositório" para a entidade `User`, mostrando como realizar operações de Criar, Ler, Atualizar e Deletar (CRUD) usando GORM.
    -   **Por que é importante:** Fornece um padrão claro para todas as interações com o banco de dados. Você pode ver exatamente como criar um novo repositório para as suas entidades de negócio, seguindo o exemplo existente.

-   **FR4: Estrutura da Clean Architecture Pré-definida**
    -   **O que significa:** A estrutura de pastas (`entity`, `usecase`, `infrastructure`, etc.) força a separação de responsabilidades. A "Regra da Dependência" é aplicada: o código da lógica de negócio (`usecase`) não depende de detalhes de tecnologia (como Gin ou Postgres). O exemplo do `User` demonstra o fluxo completo, desde a chegada de uma requisição na API até a sua persistência no banco de dados, passando por todas as camadas.
    -   **Por que é importante:** Esta é a regra de ouro do template. Ela garante que, mesmo que você não seja um especialista em arquitetura, seu projeto crescerá de forma organizada, testável e fácil de manter.

### Requisitos Não Funcionais (NFR)

-   **NFR1: Sistema de Logging Estruturado**
    -   **O que significa:** Um logger (Logrus) já está configurado e disponível em toda a aplicação. Você pode registrar informações ou erros de forma padronizada (em formato JSON), o que é ideal para ferramentas de análise de logs como Datadog, Splunk ou ELK.
    -   **Por que é importante:** Padroniza a forma como os logs são escritos em todos os microsserviços, tornando a depuração e o monitoramento em produção muito mais simples e eficientes.

-   **NFR2: Configuração via Variáveis de Ambiente**
    -   **O que significa:** O projeto não possui dados sensíveis (como senhas de banco de dados) escritos diretamente no código. Todas as configurações são lidas de variáveis de ambiente, e o arquivo `.env.sample` mostra exatamente quais variáveis seu serviço precisa para funcionar.
    -   **Por que é importante:** É uma prática essencial de segurança e flexibilidade (12-Factor App). Permite que o mesmo código seja executado em diferentes ambientes (desenvolvimento, teste, produção) apenas mudando as configurações, sem alterar uma linha de código.

-   **NFR3: Containerização com Docker Pronta**
    -   **O que significa:** O projeto vem com um `Dockerfile` otimizado (usando multi-stage builds) que compila sua aplicação e cria uma imagem Docker leve e segura, pronta para ser executada.
    -   **Por que é importante:** Facilita enormemente a distribuição e o deploy da sua aplicação. Com um único comando (`docker build`), você tem um artefato padronizado que pode rodar em qualquer lugar.

-   **NFR4: Exemplos de Testes Unitários**
    -   **O que significa:** Você encontrará arquivos como `usecase_user_service_test.go` que mostram como escrever testes para a sua lógica de negócio. Os exemplos demonstram como "mockar" (simular) dependências externas, como o banco de dados, para que você possa testar sua lógica de forma isolada.
    -   **Por que é importante:** Fornece um ponto de partida claro para a escrita de testes, diminuindo a barreira para que você adicione testes para suas novas funcionalidades e garantindo a qualidade e a confiabilidade do seu código.

## Epic and Story Structure

Para este projeto, que visa a criação de um template reutilizável, o trabalho será estruturado como um único épico focado em preparar e documentar a base de código para futuras equipes.

### Epic 1: Preparação do Template Base para Reutilização

**Epic Goal**: Garantir que o template `ms-docsigner` seja limpo, bem documentado e facilmente reutilizável para acelerar o desenvolvimento de novos microsserviços Go.

--- 

### Story 1.1: Revisão e Limpeza do Código de Exemplo

**Como** um mantenedor do template,
**Eu quero** remover qualquer lógica de negócio específica do `ms-docsigner` que não seja genérica,
**Para que** o template contenha apenas código de exemplo claro e reutilizável.

#### Acceptance Criteria
1.  O código relacionado à entidade `User` deve ser revisado e mantido como o exemplo principal.
2.  Qualquer outra lógica de negócio específica que não sirva como um bom exemplo genérico deve ser removida.
3.  As configurações no `.env.sample` devem refletir apenas as variáveis necessárias para um serviço genérico.
4.  O código deve estar livre de comentários ou `TODOs` específicos do projeto `ms-docsigner`.

--- 

### Story 1.2: Finalização da Documentação do Template

**Como** um desenvolvedor que vai usar o template,
**Eu quero** um `README.md` claro que explique como usar o template, como configurar o ambiente e como criar um novo serviço a partir dele,
**Para que** eu possa começar a trabalhar rapidamente sem precisar de ajuda externa.

#### Acceptance Criteria
1.  O `README.md` principal do projeto deve ser atualizado para focar no uso do template.
2.  Deve haver uma seção "Getting Started" com um passo-a-passo claro.
3.  As instruções devem incluir como clonar o projeto, como configurar as variáveis de ambiente e como executar o serviço pela primeira vez.
4.  Deve haver uma breve explicação sobre a estrutura de pastas e como adicionar novas funcionalidades (novas entidades, casos de uso, etc.).