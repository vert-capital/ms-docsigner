# Guia Completo: BMAD-Method para Projetos Existentes

Este guia fornece um passo a passo detalhado para usar o BMAD-Method em projetos já em andamento (brownfield), incluindo todos os agentes disponíveis, comandos e fluxo de trabalho recomendado.

## 📋 Índice

1. [Agentes Disponíveis](#agentes-disponíveis)
2. [Pré-requisitos e Instalação](#pré-requisitos-e-instalação)
3. [Configuração Inicial](#configuração-inicial)
4. [Fluxo de Trabalho Passo a Passo](#fluxo-de-trabalho-passo-a-passo)
5. [Comandos por Cenário](#comandos-por-cenário)
6. [Estrutura de Arquivos](#estrutura-de-arquivos)
7. [Troubleshooting](#troubleshooting)

## 🤖 Agentes Disponíveis

### **@analyst** - Analista de Projeto
**Descrição:** Especialista em análise de projetos existentes, documentação e pesquisa de mercado.

**Quando usar:**
- Documentar a estrutura atual do projeto
- Analisar padrões existentes no código
- Identificar inconsistências e áreas de melhoria
- Gerar documentação arquitetural focada

**Comandos principais:**
- `@analyst *document-project` - Documenta o estado atual do projeto
- `@analyst *brainstorm` - Brainstorming de melhorias

---

### **@pm** - Product Manager
**Descrição:** Gerente de produto responsável por requisitos, épicos e histórias de usuário.

**Quando usar:**
- Criar PRDs para novas funcionalidades
- Definir épicos e histórias para melhorias
- Planejar roadmap de desenvolvimento
- Gerenciar mudanças de escopo

**Comandos principais:**
- `@pm *create-doc brownfield-prd` - Cria PRD para projetos existentes
- `@pm *brownfield-create-epic` - Cria épico único para melhorias
- `@pm *correct-course` - Navega mudanças no projeto

---

### **@architect** - Arquiteto de Sistema
**Descrição:** Especialista em arquitetura de software, integração e design técnico.

**Quando usar:**
- Projetar arquitetura para novas funcionalidades
- Planejar estratégias de integração
- Definir padrões técnicos
- Avaliar impacto arquitetural

**Comandos principais:**
- `@architect *create-doc brownfield-architecture` - Arquitetura para brownfield
- `@architect *review-architecture` - Revisa arquitetura existente

---

### **@dev** - Desenvolvedor
**Descrição:** Desenvolvedor especializado em implementação de código e boas práticas.

**Quando usar:**
- Implementar funcionalidades específicas
- Refatorar código existente
- Resolver bugs e problemas técnicos
- Aplicar padrões de código

**Comandos principais:**
- `@dev implement [funcionalidade]` - Implementa funcionalidade específica
- `@dev refactor [componente]` - Refatora componente existente
- `@dev fix [bug]` - Corrige bug específico

---

### **@sm** - Story Manager
**Descrição:** Gerenciador de histórias responsável por criar e gerenciar user stories detalhadas.

**Quando usar:**
- Criar histórias de usuário detalhadas
- Quebrar épicos em histórias menores
- Definir critérios de aceitação
- Gerenciar backlog de desenvolvimento

**Comandos principais:**
- `@sm create-next-story` - Cria próxima história na sequência
- `@sm *create-doc user-story` - Cria documento de história específica

---

### **@po** - Product Owner
**Descrição:** Dono do produto responsável por validação, qualidade e alinhamento estratégico.

**Quando usar:**
- Validar planejamento e documentação
- Executar checklists de qualidade
- Aprovar mudanças de escopo
- Garantir alinhamento com objetivos

**Comandos principais:**
- `@po *execute-checklist po-master-checklist` - Executa checklist mestre
- `@po validate [documento]` - Valida documento específico

---

### **@qa** - Quality Assurance
**Descrição:** Especialista em qualidade, testes e validação de funcionalidades.

**Quando usar:**
- Definir estratégias de teste
- Criar planos de teste
- Validar implementações
- Garantir qualidade do código

**Comandos principais:**
- `@qa create-test-plan` - Cria plano de testes
- `@qa validate [funcionalidade]` - Valida funcionalidade específica

---

## 🔧 Pré-requisitos e Instalação

### Requisitos
- Node.js v20+
- Projeto existente com estrutura definida
- Gemini CLI instalado e configurado
- Conta Google para autenticação OAuth

### Instalação do Gemini CLI

**Passo 1: Instalar Gemini CLI**
```bash
# Instalar globalmente
npm install -g @google/gemini-cli

# OU executar diretamente sem instalação global
npx @google/gemini-cli
```

**Passo 2: Configurar Autenticação OAuth**
```bash
# Configurar autenticação OAuth (sem necessidade de API Key)
gemini -p "/auth"

# Uma janela do navegador será aberta para autenticação
# Siga as instruções na tela para fazer login com sua conta Google
```

**Passo 3: Verificar instalação**
```bash
gemini --version
```

### Instalação do BMAD-Method

**Passo 4: Instalar BMAD-Method**
```bash
npx bmad-method install
```

**Como executar:**
1. Abra o terminal na raiz do seu projeto
2. Execute o comando acima
3. Quando perguntado sobre o IDE, selecione **"Gemini CLI"**
4. Escolha **"Complete installation"** para projetos existentes
5. Confirme a detecção automática do projeto existente

**O que acontece:**
- Detecta automaticamente se é um projeto existente
- Oferece upgrade de V3 para V4 se necessário
- Cria estrutura de agentes no projeto
- Configura integração com Gemini CLI

---

## ⚙️ Configuração Inicial

### Configuração para Projetos Legados (V3)

**Arquivo: `core-config.yml`**
```yaml
prdVersion: v3
prdSharded: false
architectureVersion: v3
architectureSharded: false
devStoryLocation: docs/stories
```

### Configuração para Projetos Otimizados (V4)

**Arquivo: `core-config.yml`**
```yaml
prdVersion: v4
prdSharded: true
prdShardedLocation: docs/prd
architectureVersion: v4
architectureSharded: true
architectureShardedLocation: docs/architecture
devStoryLocation: .ai/stories
devLoadAlwaysFiles:
  - docs/architecture/tech-stack.md
  - docs/coding-guidelines.md
```

**Como configurar:**
1. Crie o arquivo `core-config.yml` na raiz do projeto
2. Escolha a configuração V3 (simples) ou V4 (avançada)
3. Ajuste os caminhos conforme sua estrutura de projeto

---

## 🚀 Fluxo de Trabalho Passo a Passo

### **Fase 1: Análise e Documentação (Obrigatória)**

#### Passo 1.1: Documentar Estado Atual
```bash
@analyst
*document-project
```

**Como executar:**
1. Abra o Gemini CLI digitando `gemini` no terminal
2. No prompt do Gemini CLI, digite: `@analyst`
3. Pressione Enter
4. Digite: `*document-project`
5. Pressione Enter

**Sintaxe alternativa no Gemini CLI:**
```bash
# Executar diretamente com prompt
gemini -p "@analyst *document-project"

# OU usar modo não-interativo
echo "@analyst *document-project" | gemini
```

**O que acontece:**
- Analisa estrutura de arquivos do projeto
- Identifica padrões de código existentes
- Documenta tecnologias utilizadas
- Gera relatório de estado atual
- Cria `docs/project-analysis.md`

**Tempo estimado:** 5-10 minutos

---

### **Fase 2: Planejamento de Melhorias**

#### Passo 2.1: Criar PRD para Brownfield
```bash
@pm
*create-doc brownfield-prd
```

**Como executar:**
1. No Gemini CLI, digite: `@pm`
2. Pressione Enter
3. Digite: `*create-doc brownfield-prd`
4. Responda às perguntas sobre:
   - Que melhorias você quer implementar
   - Objetivos da melhoria
   - Usuários impactados
   - Restrições técnicas

**Sintaxe alternativa:**
```bash
# Executar com prompt direto
gemini -p "@pm *create-doc brownfield-prd"
```

**O que acontece:**
- Cria PRD específico para projeto existente
- Analisa impacto nas funcionalidades atuais
- Define escopo das melhorias
- Gera `docs/prd.md` ou `docs/prd/` (se sharded)

**Tempo estimado:** 15-20 minutos

#### Passo 2.2: Criar Arquitetura para Brownfield
```bash
@architect
*create-doc brownfield-architecture
```

**Como executar:**
1. No Gemini CLI, digite: `@architect`
2. Digite: `*create-doc brownfield-architecture`
3. Forneça informações sobre:
   - Arquitetura atual do sistema
   - Pontos de integração necessários
   - Restrições técnicas existentes

**Sintaxe alternativa:**
```bash
# Executar com prompt direto
gemini -p "@architect *create-doc brownfield-architecture"
```

**O que acontece:**
- Projeta estratégia de integração
- Identifica riscos técnicos
- Define padrões de compatibilidade
- Cria diagramas de componentes
- Gera `docs/architecture.md`

**Tempo estimado:** 20-30 minutos

---

### **Fase 3: Validação e Aprovação**

#### Passo 3.1: Executar Checklist de Validação
```bash
@po
*execute-checklist po-master-checklist
```

**Como executar:**
1. No Gemini CLI, digite: `@po`
2. Digite: `*execute-checklist po-master-checklist`
3. Revise cada item do checklist apresentado
4. Confirme ou solicite ajustes

**Sintaxe alternativa:**
```bash
# Executar com prompt direto
gemini -p "@po *execute-checklist po-master-checklist"
```

**O que acontece:**
- Verifica compatibilidade com sistema existente
- Confirma que não há breaking changes
- Valida estratégias de mitigação de riscos
- Aprova ou rejeita o planejamento
- Gera relatório de validação

**Tempo estimado:** 10-15 minutos

---

### **Fase 4: Criação de Épicos e Histórias**

#### Passo 4.1: Criar Épico para Brownfield
```bash
@pm
*brownfield-create-epic
```

**Como executar:**
1. Digite: `@pm`
2. Digite: `*brownfield-create-epic`
3. Defina:
   - Nome do épico
   - Objetivo principal
   - Critérios de sucesso

**O que acontece:**
- Cria épico único e abrangente
- Foca em integração incremental
- Define sequência de implementação
- Inclui verificações de integridade

#### Passo 4.2: Criar Histórias Detalhadas
```bash
@sm
create-next-story
```

**Como executar:**
1. Digite: `@sm`
2. Digite: `create-next-story`
3. Repita para cada história necessária

**O que acontece:**
- Cria histórias de usuário detalhadas
- Define critérios de aceitação
- Inclui verificações de funcionalidades existentes
- Estabelece critérios de rollback

---

### **Fase 5: Implementação**

#### Passo 5.1: Implementar História Específica
```bash
@dev implement story [número]
```

**Exemplo:**
```bash
@dev implement story 1.1
```

**Como executar:**
1. Digite: `@dev`
2. Digite: `implement story 1.1` (substitua pelo número da história)
3. Forneça contexto adicional se necessário

**O que acontece:**
- Implementa código para a história específica
- Segue padrões existentes do projeto
- Inclui testes quando necessário
- Verifica compatibilidade com código existente

---

## 📁 Comandos por Cenário

### **Cenário 1: Adicionar Nova Funcionalidade**

**Sequência de comandos:**
```bash
# 1. Analisar projeto atual
@analyst
*document-project

# 2. Planejar nova funcionalidade
@pm
*create-doc brownfield-prd

# 3. Projetar arquitetura
@architect
*create-doc brownfield-architecture

# 4. Validar planejamento
@po
*execute-checklist po-master-checklist

# 5. Criar épico
@pm
*brownfield-create-epic

# 6. Criar histórias
@sm
create-next-story

# 7. Implementar
@dev implement story 1.1
```

### **Cenário 2: Refatorar Código Existente**

**Sequência de comandos:**
```bash
# 1. Documentar estado atual
@analyst
*document-project

# 2. Identificar áreas de refatoração
@architect
*review-architecture

# 3. Planejar refatoração
@pm
*create-doc brownfield-prd

# 4. Implementar refatoração
@dev refactor [componente]

# 5. Validar resultado
@qa validate [componente]
```

### **Cenário 3: Corrigir Bug Complexo**

**Sequência de comandos:**
```bash
# 1. Analisar problema
@analyst
*document-project

# 2. Identificar causa raiz
@dev analyze bug [descrição]

# 3. Planejar correção
@architect
*design-fix [bug]

# 4. Implementar correção
@dev fix [bug]

# 5. Testar correção
@qa validate [correção]
```

### **Cenário 4: Migração de Tecnologia**

**Sequência de comandos:**
```bash
# 1. Documentar estado atual
@analyst
*document-project

# 2. Planejar migração
@architect
*create-doc migration-architecture

# 3. Criar roadmap
@pm
*create-migration-roadmap

# 4. Validar plano
@po
*execute-checklist migration-checklist

# 5. Implementar por fases
@dev implement migration-phase [número]
```

---

## 📂 Estrutura de Arquivos Gerada

Após seguir o fluxo completo, sua estrutura de projeto terá:

```
projeto-existente/
├── docs/
│   ├── prd.md                     # Product Requirements Document
│   ├── architecture.md            # Documento de Arquitetura
│   ├── project-analysis.md        # Análise do projeto atual
│   └── stories/                   # Histórias de usuário
│       ├── epic-1/
│       │   ├── 1.1.story.md
│       │   ├── 1.2.story.md
│       │   └── ...
├── .ai/                           # Configurações BMAD (V4)
│   ├── agents/                    # Agentes configurados
│   └── templates/                 # Templates personalizados
├── core-config.yml                # Configuração principal
└── [estrutura existente do projeto]
```

---

## 🔧 Troubleshooting

### **Problema: Agente não responde**
**Solução:**
```bash
# Verificar status da instalação BMAD
npx bmad-method status

# Verificar se Gemini CLI está funcionando
gemini --version

# Testar conexão básica
gemini -p "Hello, test message"

# Reinstalar BMAD se necessário
npx bmad-method install
```

### **Problema: Comandos não reconhecidos**
**Solução:**
1. Verifique se está usando a sintaxe correta: `@agente` seguido de `comando`
2. Confirme que o Gemini CLI está configurado corretamente:
   ```bash
   # Verificar autenticação OAuth
   gemini -p "/auth"

   # Testar comando básico
   gemini -p "test"
   ```
3. Reinstale BMAD com: `npx bmad-method install`
4. Verifique se os arquivos de agente foram criados em `.gemini/`

### **Problema: "Command not found: gemini"**
**Solução:**
```bash
# Se instalou globalmente, verificar PATH
npm list -g @google/gemini-cli

# Usar npx como alternativa
npx @google/gemini-cli

# Reinstalar globalmente
npm install -g @google/gemini-cli
```

### **Problema: Erro de autenticação**
**Solução:**
```bash
# Reconfigurar autenticação OAuth
gemini -p "/auth"

# Verificar se a autenticação está funcionando
gemini -p "test de conexão"

# Para Google Cloud (se necessário para outros serviços)
gcloud auth application-default login
export GOOGLE_CLOUD_PROJECT="YOUR_PROJECT_ID"
```

### **Problema: Documentos não são gerados**
**Solução:**
1. Verifique permissões de escrita na pasta `docs/`
2. Confirme configuração no `core-config.yml`
3. Execute: `@analyst *document-project` novamente

### **Problema: Conflitos com estrutura existente**
**Solução:**
1. Ajuste caminhos no `core-config.yml`
2. Use configuração V3 para projetos mais simples
3. Customize `devLoadAlwaysFiles` para incluir arquivos importantes

---

## 📝 Dicas Importantes

1. **Sempre comece com `@analyst *document-project`** - É fundamental entender o estado atual
2. **Use a validação do PO** - `@po *execute-checklist` evita problemas futuros
3. **Implemente incrementalmente** - Uma história por vez para minimizar riscos
4. **Mantenha documentação atualizada** - Re-execute análises após mudanças significativas
5. **Teste em ambiente isolado** - Sempre teste melhorias antes de aplicar em produção
6. **Use comandos diretos quando necessário** - `gemini -p "comando"` para execução rápida
7. **Aproveite o modo não-interativo** - `echo "comando" | gemini` para scripts
8. **Configure autenticação OAuth** - Use `/auth` para configurar autenticação
9. **Use sandboxing quando apropriado** - `gemini -s` para execução segura
10. **Monitore uso de tokens** - Use `/stats` para acompanhar consumo

## 🎛️ Comandos Úteis do Gemini CLI

### Comandos de Sistema
```bash
/help          # Exibir ajuda
/stats         # Mostrar estatísticas de uso
/about         # Informações da versão
/clear         # Limpar tela (Ctrl+L)
/quit          # Sair do CLI
```

### Comandos de Configuração
```bash
/theme         # Alterar tema visual
/auth          # Configurar autenticação
/editor        # Selecionar editor preferido
```

### Comandos de Contexto
```bash
/memory show   # Mostrar contexto atual
/memory add    # Adicionar ao contexto
/memory refresh # Recarregar contexto
```

### Comandos de Ferramentas
```bash
/tools         # Listar ferramentas disponíveis
/mcp           # Status dos servidores MCP
```

### Injeção de Arquivos
```bash
@arquivo.txt                    # Incluir arquivo específico
@pasta/                        # Incluir conteúdo da pasta
@pasta/ Analise este código    # Incluir pasta com prompt
```

### Execução de Shell
```bash
!ls -la                        # Executar comando shell
!git status                    # Verificar status do git
```

---

## 🎯 Próximos Passos

Após dominar este fluxo básico, explore:
- Expansion packs específicos para sua tecnologia
- Automação de CI/CD com BMAD
- Integração com outras ferramentas de desenvolvimento
- Customização de agentes para necessidades específicas

---

**Versão do documento:** 1.0
**Última atualização:** Janeiro 2025
**Compatível com:** BMAD-Method V4+