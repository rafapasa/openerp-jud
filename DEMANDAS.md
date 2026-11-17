# 📋 DEMANDAS E CRONOGRAMA: MCP SERVER JUDICIÁRIO (eProc TJSC & Projudi TJPR)

> **Projeto:** `openerp-jud`
> **Responsabilidade:** Servidor MCP em Go + Automação Playwright Headless com Certificado Digital A1.
> **Premissa:** O ERP de terceiros já existe e será o cliente que invocará as ferramentas (Tools) expostas por este servidor MCP.
> **Última atualização:** após refatoração arquitetural (DI + interfaces + isolamento do Playwright)

---

## ⏱️ Cronograma Estimado (2 a 3 Semanas)

| Fase | Escopo Principal | Duração Est. | Meta de Entrega |
| :--- | :--- | :--- | :--- |
| **Fase 1** | **Bypass Anti-bot (Stealth) + mTLS A1** | Dias 1 - 2 | Navegador acessa eProc TJSC autenticado via `.pfx` sem bloqueio de WAF/Cloudflare. |
| **Fase 2** | **Driver eProc TJSC (Scraper & Parser)** | Dias 3 - 5 | Consultas por CPF/Processo e download de peças funcionando programaticamente. |
| **Fase 3** | **Driver Projudi TJPR (Scraper & Parser)** | Dias 6 - 8 | Fluxo equivalente ao eProc implementado e validado no TJPR. |
| **Fase 4** | **MCP Server (`mcp-go`) & Tooling** | Dias 9 - 11 | Exposição das ferramentas MCP (Stdio / SSE), tipagem JSON Schema e pool de navegadores. |
| **Fase 5** | **Robustez, Concorrência & Integração ERP** | Dias 12 - 14 | Tratamento de quedas, timeouts, rotação de sessões, limpeza de disco e doc de consumo. |

---

## 📊 Matriz de Demandas / Backlog de Desenvolvimento

### Legenda de Status:
- `[ ]` Não Iniciado
- `[/]` Em Andamento
- `[x]` Concluído
- `[!]` Impedimento / Bloco de Risco
- `[~]` Congelado (aguardando dependência externa)

---

### ÉPICO 0: Segurança e Higiene de Repositório
*Garantir que segredos, certificados e credenciais nunca vazem.*

- [x] **DEM-001**: Atualizar `.gitignore` para proibir qualquer arquivo `.pfx`, `.pem`, `.crt`, `.key` e diretório `certs/`.
- [x] **DEM-002**: Carregamento seguro de certificados A1 via variável de ambiente (`CERT_PATH` / `CERT_PASSWORD`), com validação em `config.LoadConfig()`.

---

### ÉPICO 1: Engine de Navegação & Evasão Anti-Bot (Stealth)
*Resolver o bloqueio do WAF/Cloudflare/F5 no Oracle Linux ARM64.*

- [x] **DEM-101**: **Camada Stealth no Playwright Go**
  - `--disable-blink-features=AutomationControlled` configurado.
  - Headers realistas (`User-Agent`, `Accept-Language`, `sec-ch-ua`) configurados.
  - Script `AddInitScript` mascarando `navigator.webdriver`, `window.chrome`, plugins e idiomas.
  - *Pendente:* validar em produção que a tela de "Verificação de segurança" foi superada.

- [x] **DEM-102**: **Injeção de Certificado A1 (.pfx) no Contexto TLS**
  - `ClientCertificates` configurado em `internal/browser/client.go`.
  - Certificado associado por origem (`origins ...string`).
  - *Pendente:* validar handshake mTLS real contra o endpoint do tribunal.

- [x] **DEM-103**: **Gerenciamento de Contextos Isolados**
  - Pacote `internal/browser` reutiliza a instância do Chromium e abre `BrowserContext` descartáveis por requisição.
  - `ContextManager` / `BrowserContext` / `Page` isolam o `playwright-go` do resto da aplicação.

---

### ÉPICO 2: Driver TJSC (eProc)
*Operações de leitura e extração de dados no eProc 1G/2G do TJSC.*

- [ ] **DEM-201**: **Autenticação e Manutenção de Sessão no eProc**
  - Implementar login automatizado via certificado A1.
  - Identificar tokens de sessão / cookies necessários para manter a navegação ativa.

- [ ] **DEM-202**: **Consulta de Processos por CPF do Cliente**
  - Navegar para tela de consulta por parte/documento.
  - Preencher CPF, submeter e aguardar renderização da grid de processos.
  - Parsear tabela de resultados: Número do processo, Vara, Classe, Situação, Polo (Ativo/Passivo).

- [ ] **DEM-203**: **Consulta de Detalhes por Número Único (CNJ)**
  - Acessar processo específico e extrair capa completa, partes, advogados e histórico de movimentações.

- [ ] **DEM-204**: **Download de Peças Processuais (PDFs)**
  - Localizar tabela de eventos/documentos anexos.
  - Acionar evento de download pelo Playwright interceptando o stream de resposta.
  - Salvar arquivo temporário ou converter para Base64 / Stream para retorno ao ERP.

> **Nota:** os métodos `ConsultarPorCPF`, `ObterProcesso` e `BaixarPeca` de `eProcClient` existem como esqueleto (criam contexto, abrem página, navegam) mas contêm `TODO (Fase 2)` no ponto de extração.

---

### ÉPICO 3: Driver TJPR (Projudi)
*Operações no sistema Projudi do Tribunal de Justiça do Paraná.*

- [ ] **DEM-301**: **Autenticação mTLS no Projudi**
  - Mapear fluxo de entrada com certificado ICP-Brasil no Projudi TJPR.
  - Tratar particularidades de frames/iframes legados do Projudi.

- [ ] **DEM-302**: **Consulta por CPF / CNPJ no Projudi**
  - Submeter formulário de busca de processos vinculados à parte.
  - Tratar paginação e extrair listagem padronizada.

- [ ] **DEM-303**: **Consulta por Número do Processo no Projudi**
  - Extrair dados de capa, status, movimentações e prazos pendentes.

- [ ] **DEM-304**: **Download de Peças / Movimentos no Projudi**
  - Extrair PDFs anexos às movimentações do Projudi.

---

### ÉPICO 4: Camada de Exposição MCP Server (`mcp-go`)
*Interface padrão que o ERP consumirá determinística e estruturadamente.*

- [x] **DEM-401**: **Setup do Servidor MCP e Transportes**
  - Servidor MCP inicializado com `mark3labs/mcp-go`.
  - Suporte a modo Stdio (CLI/local) e modo SSE (HTTP/remoto).
  - `mcp.Server` exposto como interface; `mcpServer` privado.

- [x] **DEM-402**: **Tool `consultar_processos_cpf`**
  - Parâmetros: `cpf` (string), `tribunal` ("TJSC" | "TJPR" | "TODOS").
  - Retorno: Lista JSON padronizada de processos encontrados.
  - Consulta federada concorrente quando `tribunal = TODOS`.

- [x] **DEM-403**: **Tool `obter_detalhes_processo`**
  - Parâmetros: `numero_processo` (string, formato CNJ), `tribunal` ("TJSC" | "TJPR").
  - Retorno: JSON estruturado com capa, partes, valor da causa, juízo e últimas movimentações.

- [x] **DEM-404**: **Tool `baixar_peca`**
  - Parâmetros: `tribunal`, `numero_processo`, `documento_id`.
  - Retorno: Caminho do arquivo temporário baixado.

> **Nota:** as tools estão registradas e funcionais na camada MCP. O retorno real depende da implementação dos drivers (Épicos 2 e 3).

---

### ÉPICO 5: Resiliência, Concorrência e Operação
*Garantir estabilidade em produção no Oracle Linux ARM64.*

- [ ] **DEM-501**: **Controle de Timeout e Retentativas com Backoff**
  - Tratar lentidões e instabilidades frequentes dos portais do TJSC/TJPR.
  - `cfg.BrowserTimeout` já existe em `config.Config` — falta aplicar via `page.SetDefaultTimeout`.

- [ ] **DEM-502**: **Gestão de Memória e Ciclo de Vida do Chromium**
  - Prevenir vazamento de processos `chromium` zumbis ou acúmulo excessivo de memória RAM.
  - Implementar encerramento forçado de contextos órfãos.

- [ ] **DEM-503**: **Logs Estruturados e Auditoria**
  - Registro de auditoria de consultas (sem logar senhas de certificados).
  - Medição de tempo de resposta por tribunal.
  - Adapters `logger.Debug/Info/Warn/Error/Fatal(ctx, ...)` já disponíveis em `internal/observability/logger/context.go`.

---

### ÉPICO 6: Arquitetura e Injeção de Dependência
*Refatoração estrutural concluída — base para todos os épicos.*

- [x] **DEM-601**: **Isolamento do Playwright**
  - `github.com/mxschmitt/playwright-go` importado **apenas** por `internal/browser/client.go`.
  - Interfaces próprias: `ContextManager`, `BrowserContext`, `Page`.

- [x] **DEM-602**: **Regra de Nomenclatura (interface pública / struct privada)**
  - Aplicada em `browser`, `client`, `mcp`, `tribunal`.
  - Construtores `NewXXXXX` retornam interfaces.

- [x] **DEM-603**: **Injeção de Dependência com Google Wire**
  - `internal/di/provider.go` — `ProviderSet` agregando os sets.
  - `internal/di/wire.go` — `InitializeApp(cfg)`.
  - `internal/di/wire_gen.go` — grafo gerado.
  - Falha de provider **aborta** a inicialização (elimina nil interface panic).

- [x] **DEM-604**: **Centralização de `os`**
  - `internal/helper/files.go` — `EnsureDir`, `FileExists`, `DirExists`, `JoinPath`.

- [x] **DEM-605**: **Configuração com validação**
  - `config.LoadConfig() (*Config, error)` valida `MCP_TRANSPORT`, `PORT`, `BROWSER_HEADLESS`, `BROWSER_TIMEOUT_SEC`.
  - Campos de logging (`LogLevel`, `LogEncoding`, `LogOutput`, `LogFile`) agora são preenchidos.

- [x] **DEM-606**: **Registry com injeção de drivers**
  - `tribunal.NewRegistry(eproc, projudi)` recebe os drivers e registra internamente.
  - `inMemRegistry` privado; `Registry` exposto como interface.

---

### ÉPICO 7: Parsers HTML (CONGELADOS)
*Bloqueados até a captura de HTML real dos portais.*

- [~] **DEM-701**: **Parser eProc TJSC**
  - `EProcParser` implementa `ParseTabelaProcessos`, `ParseCapaProcesso`, `ParseDocumentosProcesso`.
  - **Congelado:** fixtures dos testes são estruturais/inventados, não capturados do portal real.
  - **Desbloqueio:** capturar snapshot HTML real das telas de consulta, capa e eventos.

- [~] **DEM-702**: **Parser Projudi TJPR**
  - `ProJudiParser` implementa os mesmos três métodos.
  - **Congelado:** mesma razão do DEM-701.

- [ ] **DEM-703**: **Captura de Snapshots HTML Reais**
  - Usar o navegador com Stealth para salvar o HTML das telas:
    - eProc: consulta por CPF, capa do processo, árvore de eventos.
    - Projudi: consulta por CPF, capa do processo, tabela de documentos.
  - **Pré-requisito:** DEM-101 e DEM-102 validados em produção.

---

## 🚧 Pendências Técnicas Conhecidas

| # | Item | Ação necessária |
|---|---|---|
| 1 | `internal/mcp/handler_test.go` não compila | Usa `tribunal.InMemRegistry` (agora privado) e `NewRegistry()` sem argumentos. Reescrever passando dois mocks. |
| 2 | `go.mod` com `google/wire` como `// indirect` | Rodar `go mod tidy`. |
| 3 | `wire_gen.go` escrito manualmente | Regenerar com `go run github.com/google/wire/cmd/wire ./internal/di`. |
| 4 | Parsers congelados | Ver Épico 7. |
| 5 | `page.SetDefaultTimeout` não aplicado | Usar `cfg.BrowserTimeout` nos drivers (DEM-501). |

---

## 🚀 Próxima Ação Imediata

1. **Validar a camada Stealth em produção** — confirmar que o eProc TJSC não exibe mais a tela de verificação de segurança.
2. **Capturar snapshots HTML reais** (DEM-703) — desbloqueia o Épico 7.
3. **Corrigir `handler_test.go`** — restabelecer `go build ./...` e `go vet ./...` limpos.
4. **Implementar `ConsultarPorCPF` do eProc** (DEM-202) — primeiro fluxo ponta a ponta.
