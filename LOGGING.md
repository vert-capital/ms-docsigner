# Configuração de Logging

Este projeto possui configuração flexível de logging através de variáveis de ambiente, permitindo diferentes níveis de verbosidade para desenvolvimento, teste e produção.

## Variáveis de Ambiente

### LOG_LEVEL
Controla o nível geral de logs da aplicação.

**Valores aceitos:**
- `DEBUG`: Logs muito detalhados (desenvolvimento)
- `INFO`: Logs informativos (desenvolvimento/staging)
- `WARN`: Apenas warnings e erros (produção)
- `ERROR`: Apenas erros críticos
- `SILENT`: Nenhum log

**Padrão:** `INFO`

### GIN_MODE
Controla o modo do framework Gin.

**Valores aceitos:**
- `debug`: Modo desenvolvimento (logs verbosos do Gin)
- `release`: Modo produção (logs mínimos do Gin)
- `test`: Modo teste

**Padrão:** `release`

### GORM_LOG_LEVEL
Controla o nível de logs do GORM (ORM/banco de dados).

**Valores aceitos:**
- `DEBUG`: Mostra todas as queries SQL com detalhes
- `INFO`: Mostra queries SQL básicas
- `WARN`: Mostra apenas queries lentas e warnings
- `ERROR`: Mostra apenas erros de banco
- `SILENT`: Nenhum log de banco

**Padrão:** `WARN`

## Configurações Recomendadas por Ambiente

### Desenvolvimento
```bash
LOG_LEVEL=DEBUG
GIN_MODE=debug
GORM_LOG_LEVEL=INFO
```

### Staging/Homologação
```bash
LOG_LEVEL=INFO
GIN_MODE=release
GORM_LOG_LEVEL=WARN
```

### Produção
```bash
LOG_LEVEL=WARN
GIN_MODE=release
GORM_LOG_LEVEL=ERROR
```

### Teste
```bash
LOG_LEVEL=SILENT
GIN_MODE=test
GORM_LOG_LEVEL=SILENT
```

## Comportamentos

### Logs do Gin
- Em modo `debug`: Mostra todos os requests HTTP com detalhes
- Em modo `release`: Logs mínimos, apenas se `LOG_LEVEL=INFO` ou superior
- Recovery middleware sempre ativo para capturar panics

### Logs do GORM
- `DEBUG`: Mostra todas as queries SQL com parâmetros
- `INFO`: Mostra queries SQL básicas
- `WARN`: Mostra apenas queries que demoram mais que 1 segundo
- `ERROR`: Mostra apenas erros de banco
- `SILENT`: Nenhum log de banco

### Vantagens

1. **Redução de logs em produção**: Evita logs excessivos que podem impactar performance
2. **Flexibilidade**: Cada ambiente pode ter configuração específica
3. **Debug simplificado**: Fácil ativar logs verbosos para debugging
4. **Conformidade**: Logs estruturados para auditoria e monitoramento

## Exemplo de Uso

Para ativar logs verbosos temporariamente em produção:

```bash
# Definir variáveis de ambiente
export LOG_LEVEL=DEBUG
export GORM_LOG_LEVEL=INFO

# Reiniciar aplicação
```

Para voltar aos logs de produção:

```bash
export LOG_LEVEL=WARN
export GORM_LOG_LEVEL=ERROR
```

## Monitoramento

Com essas configurações, você pode:

1. **Produção**: Manter logs mínimos para performance
2. **Debug**: Ativar logs detalhados quando necessário
3. **Auditoria**: Configurar níveis apropriados para compliance
4. **Performance**: Reduzir overhead de I/O em produção
