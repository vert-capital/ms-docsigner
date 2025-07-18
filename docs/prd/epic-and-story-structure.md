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
