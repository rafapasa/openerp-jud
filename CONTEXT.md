# CONTEXTO DO PROJETO: INTEGRAÇÃO JUDICIÁRIA (eProc TJSC & Projudi TJPR) VIA MCP EM GOLANG

> **Arquivo de Contexto para Agente / IDE**
> **Repositório:** `openerp-jud`
> **Módulo Go:** `github.com/rafapasa/openerp-jud`
> **Ambiente Atual:** Oracle Linux 9 (`el9.aarch64` - ARM64 / OCI)
> **Última atualização:** refatoração arquitetural (DI + interfaces)

---

## 1. Visão Geral e Objetivo do Projeto

O objetivo deste projeto é fornecer a um **ERP de terceiros** a capacidade de **consultar processos judiciais e baixar documentos/peças processuais** associadas a um CPF ou número de processo nos sistemas:

1. **eProc** (Tribunal de Justiça de Santa Catarina - TJSC);
2. **Projudi** (Tribunal de Justiça do Paraná - TJPR).

A operação é realizada em nome de um advogado habilitado portador de um **Certificado Digital ICP-Brasil A1 (`.pfx`)**.
O escopo inicial é **apenas leitura** (consulta de capa, andamentos, partes e download de documentos/peças), sem peticionamento neste momento.

---

## 2. Decisões Arquiteturais

### 2.1 Por que Go + MCP Server + Playwright?

1. **Ausência de APIs Abertas nos Tribunais:**
   - O eProc e o Projudi **não** fornecem endpoints REST abertos para advogados privados.
   - O padrão oficial do CNJ (**MNI - Modelo Nacional de Interoperabilidade**) é baseado em Web Services SOAP/XML e restrito a órgãos conveniados (Ministério Público, PGE, Defensoria).
   - APIs comerciais de mercado (Infosimples, Judit.io, Escavador) cobram por consulta ou exigem mensalidades fixas, além de muitas vezes não fornecerem acesso a peças restritas.

2. **Por que Go (Golang)?**
   - Alta performance e concorrência nativa (goroutines) para scraping e chamadas simultâneas.
   - Baixíssimo consumo de memória RAM comparado a Python/Node.
   - Compilação em binário único, facilitando o deploy no servidor Linux.

3. **Por que MCP (Model Context Protocol)?**
   - O serviço funciona como um **MCP Server**.
   - Expõe *tools* tipadas (`consultar_processos_cpf`, `obter_detalhes_processo`, `baixar_peca`) para serem consumidas deterministicamente por LLMs ou pelo ERP.

4. **Tratamento do Certificado A1 (.pfx):**
   - O Playwright permite injetar o arquivo `.pfx` e sua senha diretamente nas opções de contexto do navegador (`ClientCertificates`), executando o handshake TLS (mTLS) de forma headless e **sem exibir janelas nativas de seleção de certificado do sistema operacional**.

### 2.2 Injeção de Dependência com Google Wire

O projeto usa **`github.com/google/wire`** para montar o grafo de dependências em tempo de compilação.

**Motivação:** eliminar a classe de bugs de *nil interface* — quando um construtor falha e retorna `nil`, esse `nil` propagado como interface concreta causa panic em runtime. Com o wire, a falha de qualquer provider **aborta a inicialização** com erro explícito.

**Regra:** sem browser não há MCP. Falha do Chromium é fatal.

### 2.3 Regra de Nomenclatura (obrigatória para código novo)

| Elemento | Convenção | Exemplo |
|---|---|---|
| Interface | `PascalCase` | `ContextManager`, `EProcClient` |
| Struct concreta | `camelCase` (mesmo nome da interface) | `contextManager`, `eProcClient` |
| Construtor | `New` + nome da interface | `NewContextManager`, `NewEProcClient` |

**Toda função `NewXXXXX` pública retorna uma interface.** As structs concretas são privadas ao pacote.

**Exceção documentada:** tipos de bibliotecas externas (ex.: `*zap.Logger`) não são encapsulados.

### 2.4 Isolamento de Dependências Externas

O pacote `github.com/mxschmitt/playwright-go` é importado **exclusivamente** por `internal/browser/client.go`.

Os consumidores (`internal/client`, `internal/mcp`, `main.go`) trabalham apenas com as interfaces próprias do pacote `browser`:

- `browser.ContextManager` — gerencia o Chromium e cria contextos
- `browser.BrowserContext` — contexto de navegação isolado
- `browser.Page` — página/aba

**Objetivo:** trocar o motor de automação (Playwright → outro) exige mudanças apenas em `internal/browser`.

### 2.5 Centralização de `os`

Toda operação de sistema de arquivos passa por `internal/helper/files.go` (`EnsureDir`, `FileExists`, `DirExists`, `JoinPath`). O import de `os` fica concentrado nesse pacote.

### 2.6 Logger

O logger **não participa do grafo do wire**.

- `logger.NewLogger(cfg)` cria a instância e a armazena em variável privada do pacote.
- `logger.GetLogger()` retorna a instância ou cria fallback seguro.
- `logger.FromContext(ctx)` retorna o logger do contexto → global → `zap.NewNop()` (**nunca nil**).
- Adapters limpos em `internal/observability/logger/context.go`: `Debug`, `Info`, `Warn`, `Error`, `Fatal` — todos recebem `ctx`.

O `main.go` inicializa o logger **antes** do wire e usa `zapLogger` diretamente (código mínimo e legível).

---

## 3. Stack Tecnológica

| Componente | Versão / Pacote |
|---|---|
| Linguagem | Go 1.27+ |
| Framework MCP | `github.com/mark3labs/mcp-go` v1.5.0 |
| Automação de Navegador | `github.com/mxschmitt/playwright-go` v0.6401.0 |
| Injeção de Dependência | `github.com/google/wire` v0.7.0 |
| Parse HTML | `github.com/PuerkitoBio/goquery` v1.13.0 |
| Logging | `go.uber.org/zap` v1.28.0 |
| Testes | `github.com/stretchr/testify` v1.11.1 |
| Toolkit interno | `github.com/etoolstec/gokit` v1.0.12 (`apperror`, `converter`) |
| Ambiente | Oracle Linux 9 / ARM64 (aarch64) |

---

## 4. Estrutura de Código

```text
openerp-jud/
├── CONTEXT.md                    # Este arquivo
├── DEMANDAS.md                   # Backlog e cronograma
├── go.mod
├── go.sum
├── main.go                       # Entrypoint: config → logger → wire → Serve
├── config/
│   └── config.go                 # LoadConfig() (*Config, error) + validações
├── certs/                        # Certificados A1 (IGNORADO NO GIT)
│   └── advogado.pfx
└── internal/
    ├── browser/                  # ÚNICO pacote que importa playwright
    │   ├── interface.go          # ContextManager, BrowserContext, Page
    │   ├── client.go             # contextManager, browserContext, page
    │   └── stealth.go            # StealthInitScript, DefaultUserAgent
    │
    ├── client/                   # Drivers de automação dos tribunais
    │   ├── interface.go          # EProcClient, ProjudiClient, HTMLParser
    │   ├── eproc.go              # eProcClient (driver TJSC)
    │   ├── projudi.go            # projudiClient (driver TJPR)
    │   ├── html.go               # GetHTMLDoc (parse HTML → goquery)
    │   ├── eproc_parser.go       # EProcParser (CONGELADO — ver §7)
    │   └── projudi_parser.go     # ProJudiParser (CONGELADO — ver §7)
    │
    ├── di/                       # Injeção de dependência
    │   ├── provider.go           # ProviderSet (sets de providers)
    │   ├── wire.go               # InitializeApp (build tag wireinject)
    │   └── wire_gen.go           # Grafo gerado pelo wire
    │
    ├── helper/
    │   └── files.go              # EnsureDir, FileExists, DirExists, JoinPath
    │
    ├── mcp/                      # Servidor MCP
    │   ├── interface.go          # Server, ToolHandler
    │   ├── handler.go            # defaultToolHandler (delega ao Registry)
    │   └── server.go             # mcpServer (tools + transporte Stdio/SSE)
    │
    ├── model/
    │   └── types.go              # Processo, Documento, ParteProcesso, Movimentacao
    │
    ├── observability/logger/
    │   ├── logger.go             # NewLogger, GetLogger, LogInfo/Warn/Error
    │   └── context.go            # Adapters com ctx + trace_id/span_id
    │
    └── tribunal/
        ├── interface.go          # Driver (contrato genérico de tribunal)
        └── registry.go           # Registry, inMemRegistry, NewRegistry(eproc, projudi)
```

---

## 5. Grafo de Dependências (Wire)

```text
config.Config  (parâmetro de InitializeApp)
      │
      └─> browser.NewContextManager ──> browser.ContextManager
                │
                ├─> client.NewEProcClient ────> client.EProcClient ────┐
                │                                                       │
                └─> client.NewProjudiClient ──> client.ProjudiClient ──┤
                                                                        │
                                    tribunal.NewRegistry(eproc, projudi)│
                                                │                       │
                                                └─> tribunal.Registry ◄─┘
                                                          │
                                                          └─> mcp.NewToolHandler ──> mcp.ToolHandler
                                                                      │
                                                                      └─> mcp.NewServer ──> mcp.Server
```

**Regenerar o grafo:**

```bash
go run github.com/google/wire/cmd/wire ./internal/di
```

---

## 6. Fluxo de Inicialização (`main.go`)

```text
1. config.LoadConfig()          → (*Config, error)   — falha = log.Fatalf
2. logger.NewLogger(cfg)        → (*zap.Logger, cleanup, error)
3. helper.EnsureDir(DownloadsDir)
4. di.InitializeApp(cfg)        → (mcp.Server, error) — falha = Fatal
5. app.Serve(ctx)               → Stdio ou SSE
```

---

## 7. Estado Atual dos Parsers (CONGELADOS)

Os arquivos `eproc_parser.go` e `projudi_parser.go` estão **deliberadamente congelados**.

**Motivo:** os fixtures HTML usados nos testes são **estruturais/inventados**, não capturados dos portais reais. Qualquer ajuste de seletor agora seria baseado em suposição e teria de ser refeito.

**Decisão:** aguardar a captura de snapshots HTML reais do eProc TJSC e do Projudi TJPR. Somente então os parsers serão revisados — e possivelmente reescritos.

**Consequência:** os testes `eproc_parser_test.go` e `projudi_parser_test.go` podem não refletir o comportamento real e não devem ser tratados como especificação definitiva.

---

## 8. Histórico de Problemas e Correções no Ambiente Linux ARM64

1. **Incompatibilidade de Caminho de Módulo no Go**
   - *Problema:* `go get github.com/playwright-community/playwright-go` falhava — o módulo declara seu caminho como `github.com/mxschmitt/playwright-go`.
   - *Solução:* usar o caminho canônico:
     ```bash
     go get github.com/mxschmitt/playwright-go
     ```

2. **Falha de Instalação com `--with-deps` (falta do `apt-get`)**
   - *Problema:* o comando padrão tentava executar `apt-get`, mas o Oracle Linux usa `dnf`.
   - *Solução:* baixar o Chromium sem a flag de dependências:
     ```bash
     go run github.com/mxschmitt/playwright-go/cmd/playwright@latest install chromium
     ```

3. **Biblioteca Compartilhada Faltante (`libatspi`)**
   - *Problema:* `No match for argument: libatspi` no `dnf`.
   - *Solução:* no RHEL/Oracle Linux 9 a biblioteca vem do pacote:
     ```bash
     sudo dnf install -y at-spi2-core
     ```

4. **Smoke Test do Chromium**
   - O Chromium ARM64 inicializou com sucesso em modo headless e acessou o eProc do TJSC.

---

## 9. Desafio Pendente: WAF / Anti-bot do TJSC

### Situação
Ao acessar a URL pública de consulta do eProc (`https://eproc1g.tjsc.jus.br/...`), o título retornado foi:

> `Tribunal de Justiça de Santa Catarina - Verificação de segurança do portal institucional`

### Diagnóstico
O WAF/Defesa Anti-bot do TJSC (F5 BIG-IP / Cloudflare) interceptou a requisição porque:
1. O IP de origem pertence à faixa de data center (Oracle Cloud OCI);
2. O Chromium padrão do Playwright expõe `navigator.webdriver = true` e não continha headers/user-agent realistas.

### Camada Stealth já implementada
`internal/browser/stealth.go` injeta, via `AddInitScript`, um script que:
- Oculta `navigator.webdriver`;
- Simula `window.chrome`;
- Simula plugins válidos;
- Define idiomas `pt-BR`;
- Faz mock da API de permissões.

`internal/browser/client.go` configura:
- `--disable-blink-features=AutomationControlled`;
- `UserAgent` desktop realista (`DefaultUserAgent`);
- `Viewport` 1366x768;
- Headers `sec-ch-ua`, `Accept-Language`;
- `Locale` pt-BR e `TimezoneId` America/Sao_Paulo.

### Próxima Ação
Validar se a camada Stealth supera a tela de verificação de segurança do TJSC e, em caso positivo, capturar o HTML real das telas de consulta para descongelar os parsers (§7).

---

## 10. Segurança e Higiene de Repositório

1. **Nunca commitar certificados.** A pasta `certs/` e arquivos `.pfx`, `.pem`, `.crt`, `.key` devem estar no `.gitignore`.
2. **Senha do certificado** deve vir de variável de ambiente (`CERT_PASSWORD`) ou secret manager — nunca do código.
3. **Logs de auditoria** não devem registrar senhas de certificado.

---

## 11. Variáveis de Ambiente

| Variável | Default | Descrição |
|---|---|---|
| `ENVIRONMENT` | `development` | `development` / `production` |
| `TIME_ZONE` | `America/Sao_Paulo` | Timezone da aplicação |
| `MCP_TRANSPORT` | `stdio` | `stdio` ou `sse` |
| `PORT` | `8080` | Porta do servidor SSE |
| `CERT_PATH` | `certs/advogado.pfx` | Caminho do certificado A1 |
| `CERT_PASSWORD` | *(vazio)* | Senha do certificado A1 |
| `BROWSER_HEADLESS` | `true` | Modo headless do Chromium |
| `BROWSER_TIMEOUT_SEC` | `45` | Timeout padrão do navegador (segundos) |
| `DOWNLOADS_DIR` | `tmp/downloads` | Diretório de peças baixadas |
| `LOG_LEVEL` | `info` | `debug` / `info` / `warn` / `error` |
| `LOG_ENCODING` | `json` | `json` ou `console` |
| `LOG_OUTPUT` | `stdout` | `stdout` / `file` / `both` |
| `LOG_FILE` | `logs/mcp-server.log` | Caminho do arquivo de log |

---

## 12. Regras de Código (herdadas do BASE.JSON)

1. Erros **sempre** via `apperror` (`NewNotFoundError`, `NewBadRequestError`, `NewInternalError`, `NewValidationError`, `NewForbiddenError`). Nunca `errors.New`.
2. Conversões DTO ↔ Model **sempre** via `gokit/mapper`.
3. Respostas HTTP **sempre** via `gokit/response`.
4. **Não** criar DTOs novos sem necessidade — usar os existentes.
5. Verificar nomes antes de declarar funções, structs e variáveis globais, para evitar redeclaração.
6. Ao usar funções/structs de outros arquivos, confirmar que existem e usar os nomes corretos.
