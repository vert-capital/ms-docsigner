Para manter a consistência e a qualidade, todos os projetos derivados deste template devem seguir os seguintes padrões:

- **Estrutura de Arquivos:** A estrutura de pastas baseada na Clean Architecture deve ser mantida. Novas funcionalidades devem ser adicionadas criando novos arquivos dentro das pastas existentes (ex: `usecase/new_entity/`) ou, se necessário, novas pastas que sigam a mesma lógica de separação.
- **Injeção de Dependência:** Todas as dependências devem ser inicializadas no `main.go` e injetadas nos construtores dos componentes. Nenhum componente deve criar suas próprias dependências.
- **Tratamento de Erros:** Os erros devem ser tratados na camada onde ocorrem. Erros de camadas mais internas devem ser propagados para as camadas externas, onde serão logados e convertidos em uma resposta apropriada (ex: um erro de banco de dados no repositório se torna um erro 500 na API).
- **Configuração:** Toda a configuração deve ser lida de variáveis de ambiente, conforme definido no `config/` e exemplificado no `.env.sample`.
