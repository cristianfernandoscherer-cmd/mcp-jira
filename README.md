# MCP Jira

Servidor MCP para integração com Jira. Retorna título, descrição e comentário de refinamento técnico de tickets.

## Descrição

Este projeto implementa um servidor MCP que busca dados de uma issue no Jira e retorna:

- **titulo** — summary da issue
- **descricao** — corpo da descrição em JSON (formato ADF do Jira)
- **refinamento** — comentário completo que contém a frase "refinamento técnico", ou mensagem de fallback

## Requisitos

- Go 1.19+
- Conta no Jira com API access
- Variáveis de ambiente configuradas

## Instalação

1. Clone o repositório:
```bash
git clone <repository-url>
cd mcp-jira
```

2. Build o projeto:
```bash
make build
```

3. Configure as variáveis de ambiente:
```bash
cp .env.example .env
# Edite o arquivo .env com suas credenciais do Jira
```

## Configuração

### Variáveis de Ambiente

```env
JIRA_BASE_URL=https://seu-jira.atlassian.net
JIRA_EMAIL=seu-email@empresa.com
JIRA_API_TOKEN=seu-token-de-api
MCP_MODE=stdio
DEBUG=false
```

### Obter Token de API Jira

Existem diferentes formas de gerar tokens para a API do Jira, dependendo do tipo de autenticação necessária:

- **Account API Tokens**: Para autenticação pessoal (usuário individual)
- **Service Account Tokens**: Para integrações de aplicativos (scopes específicos)

> **Importante:** A escolha do tipo de token pode alterar a URL base (`JIRA_BASE_URL`) necessária para a API.

#### Tipos de URL Base

Dependendo do método de autenticação, a `JIRA_BASE_URL` terá formatos diferentes:

**Classic (Account API Token):**
```env
JIRA_BASE_URL=https://<seu-dominio>.atlassian.net
```
Exemplo:
```env
JIRA_BASE_URL=https://cristianfernandoscherer.atlassian.net
```

**Granular (Service Account Token com tenant UUID):**
```env
JIRA_BASE_URL=https://api.atlassian.com/ex/jira/<tenant-uuid>
```
Exemplo:
```env
JIRA_BASE_URL=https://api.atlassian.com/ex/jira/ad6204c2-c64b-4998-a087-3027c78133ec
```

Para mais detalhes sobre os diferentes tipos de tokens e seus escopos, consulte: [How to map API token scopes with API](https://community.atlassian.com/forums/Jira-questions/How-to-map-API-token-scopes-with-API/qaq-p/3019748)

#### Passos para gerar um token pessoal:

1. Acesse seu perfil do Jira
2. Vá para **Account Settings** → **Security**
3. Clique em **Create API token**
4. Gere um novo token e salve-o

#### Descobrir Tenant UUID

Para obter o UUID do seu tenant Jira, acesse:
```
https://<seu-dominio>.atlassian.net/_edge/tenant_info
```

Exemplo:
```
https://cristianfernandoscherer.atlassian.net/_edge/tenant_info
```

## Uso

### Executar o servidor

```bash
./run-mcp.sh
```

O script carrega o `.env` do diretório do projeto (se `JIRA_API_TOKEN` não estiver definido) e executa `./bin/mcp`.

Alternativa direta:

```bash
./bin/mcp
```

## Tool Disponível

### `get_task_with_comments`

Retorna título, descrição e o comentário de refinamento técnico de um ticket.

**Parâmetros:**
- `task_key` (string): Chave do ticket (ex: LED-52940, KAN-4)

**Resposta:**
```json
{
  "titulo": "Título da issue",
  "descricao": "{ ... JSON ADF da descrição ... }",
  "refinamento": "{ ... JSON do comentário completo ... }"
}
```

- `descricao` retorna o ADF original do Jira em JSON formatado; pode vir vazio
- `refinamento` retorna o comentário inteiro (autor, data, body) quando contém "refinamento técnico"
- Se nenhum comentário corresponder, `refinamento` retorna: `Nenhum comentário de refinamento técnico encontrado`

**Exemplo de uso:**
```json
{
  "method": "tools/call",
  "params": {
    "name": "get_task_with_comments",
    "arguments": {
      "task_key": "LED-52940"
    }
  }
}
```

## Comandos Make

- `make build` - Compilar o projeto
- `make run` - Executar o servidor
- `make clean` - Remover binários

## Estrutura

```
.
├── cmd/
│   └── main.go          # Entry point
├── internal/
│   ├── config/          # Configuração
│   └── handlers/        # Handlers MCP
├── pkg/
│   ├── jira/            # Cliente Jira
│   └── mcp/             # Servidor MCP
├── bin/                 # Binários
├── run-mcp.sh           # Script de execução
├── .env.example         # Exemplo de configuração
├── Makefile             # Tarefas de build
├── go.mod               # Módulo Go
└── go.sum               # Dependências
```
