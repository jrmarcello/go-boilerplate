# Decisão de Arquitetura: Estratégia de Configuração

## Contexto

A aplicação precisa ser configurável em múltiplos ambientes: **Desenvolvimento Local**, **Infraestrutura Docker** e **Produção (Kubernetes)**. Precisamos de uma estratégia unificada que evite duplicidade e mantenha a conformidade com o 12-Factor App.

## Decisão

Adotamos **Viper** como biblioteca de configuração com prioridade para **Variáveis de Ambiente**, centralizando a configuração local em um único arquivo `.env`.

### Hierarquia de Prioridade

| Prioridade | Fonte | Uso |
| ---------- | ----- | --- |
| 🥇 Alta | Variáveis de Ambiente | Kubernetes (ConfigMaps/Secrets), Docker Compose |
| 🥈 Média | Arquivo `.env` | Desenvolvimento local |
| 🥉 Baixa | Defaults no Código | Fallback seguro (`localhost`) |

## Justificativa

1. **Single Source of Truth (Local)**: O arquivo `.env` na raiz é consumido simultaneamente pelo Docker Compose, Go Application (Viper) e Makefile.
2. **Transparência em Produção**: O K8s injeta configurações via Env Vars, que têm precedência máxima.
3. **Simplicidade (DX)**: O desenvolvedor precisa apenas criar um arquivo `.env`.

## Consequências

- **Positivas**:
  - Eliminamos arquivos duplicados (`docker/.env`, `config.yaml`).
  - `make dev` e `make docker-up` funcionam em harmonia.
  - Comportamento determinístico em produção.

- **Negativas**:
  - Dependência da biblioteca Viper.
  - Necessidade de documentar a hierarquia de configuração.

## Implementação

### Configuração do Viper

```go
// config/config.go
func Load() (*Config, error) {
    v := viper.New()

    // 1. Defaults
    setDefaults(v)

    // 2. Arquivo .env (opcional)
    v.SetConfigFile(".env")
    v.SetConfigType("env")
    _ = v.ReadInConfig()

    // 3. Variáveis de Ambiente (precedência máxima)
    v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
    v.AutomaticEnv()

    var cfg Config
    return &cfg, v.Unmarshal(&cfg)
}
```

### Mapeamento de Variáveis

| Struct Field | Env Var | Default |
| ------------ | ------- | ------- |
| `Server.Port` | `SERVER_PORT` | `8080` |
| `DB.DSN` | `DB_DSN` | `postgres://...` |
| `Redis.Enabled` | `REDIS_ENABLED` | `false` |
