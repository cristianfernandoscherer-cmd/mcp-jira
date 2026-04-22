# MCP Jira

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.19+-00ADD8?logo=go&logoColor=white)](https://golang.org)

Servidor MCP para integração com Jira desenvolvido em Go. Fornece acesso a comentários de refinamento técnico de tickets.

## Descrição

Este projeto implementa um servidor MCP simplificado que acessa comentários técnicos de tickets do Jira, ideal para integração com ferramentas compatíveis com MCP.

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

1. Acesse seu perfil do Jira
2. Vá para **Account Settings** → **Security**
3. Clique em **Create API token**
4. Gere um novo token e salve-o

## Uso

### Executar o servidor

```bash
./bin/mcp
```

## Tool Disponível

### `get_issue_with_comments`

Retorna o conteúdo de comentários de refinamento técnico de um ticket.

**Parâmetros:**
- `issue_key` (string): Chave do ticket (ex: KAN-4, PROJ-123)

**Resposta:**
```json
{
  "refinamento": "Conteúdo do comentário de refinamento técnico"
}
```

**Exemplo de uso:**
```json
{
  "method": "tools/call",
  "params": {
    "name": "get_issue_with_comments",
    "arguments": {
      "issue_key": "KAN-4"
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
│   ├── handlers/        # Handlers MCP
│   └── utils/           # Utilitários
├── pkg/
│   ├── jira/            # Cliente Jira
│   └── mcp/             # Servidor MCP
├── bin/                 # Binários
├── .env.example         # Exemplo de configuração
├── Makefile             # Tarefas de build
├── go.mod               # Módulo Go
└── go.sum               # Dependências
```

## Licença

MIT License - veja o arquivo [LICENSE](LICENSE).