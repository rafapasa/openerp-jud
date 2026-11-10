# 📋 DEMANDAS E CRONOGRAMA: MCP SERVER JUDICIÁRIO (eProc TJSC & Projudi TJPR)

> **Projeto:** `openerp-jud`  
> **Responsabilidade:** Servidor MCP em Go + Automação Playwright Headless com Certificado Digital A1.  
> **Premissa:** O ERP de terceiros já existe e será o cliente que invocará as ferramentas (Tools) expostas por este servidor MCP.

---

## ⏱️ Cronograma Estimado (2 a 3 Semanas)

Considerando a infraestrutura já provisionada (Oracle Linux 9 ARM64, Go 1.22+ e dependências de sistema do Chromium configuradas):

| Fase | Escopo Principal | Duração Est. | Meta de Entrega |
| :--- | :--- | :--- | :--- |
| **Fase 1** | **Bypass Anti-bot (Stealth) + mTLS A1** | Dias 1 - 2 | Navegador acessa eProc TJSC autenticado via `.pfx` sem bloqueio de WAF/Cloudflare. |
| **Fase 2** | **Driver eProc TJSC (Scraper & Parser)** | Dias 3 - 5 | Consultas por CPF/Processo e download de peças funcionando programmaticamente. |
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

---

### ÉPICO 0: Segurança e Higiene de Repositório
*Garantir que segredos, certificados e credenciais nunca vazem.*

- [x] **DEM-001**: Atualizar `.gitignore` para proibir qualquer arquivo `.pfx`, `.pem`, `.crt`, `.key` e diretório `certs/`.
- [ ] **DEM-002**: Estruturar carregamento seguro de certificados A1 via variável de ambiente ou caminho de volume seguro (com senha passada via `ENV` ou secret manager).

---

### ÉPICO 1: Engine de Navegação & Evasão Anti-Bot (Stealth)
*Resolver o bloqueio do WAF/Cloudflare/F5 no Oracle Linux ARM64.*

- [ ] **DEM-101**: **Camada Stealth no Playwright Go**
  - Configurar launch options com `--disable-blink-features=AutomationControlled`.
  - Definir headers realistas de navegador desktop (`User-Agent`, `Accept-Language`, `sec-ch-ua`).
  - Injetar script (`page.AddInitScript`) para mascarar `navigator.webdriver` e emular plugins.
  - *Critério de Aceite:* Acessar o eproc TJSC sem cair na tela de "Verificação de segurança do portal institucional".
- [ ] **DEM-102**: **Injeção de Certificado A1 (.pfx) no Contexto TLS**
  - Configurar `BrowserNewContextOptions.ClientCertificates` com certificado de teste.
  - Testar handshake mútuo (mTLS) direto no endpoint seguro do tribunal.
  - *Critério de Aceite:* Navegador headless ser reconhecido como advogado autenticado sem abrir popup do SO.
- [ ] **DEM-103**: **Gerenciamento de Contextos Isolados (Browser Pool)**
  - Criar pacote `internal/browser` para reutilizar a instância do Chromium e abrir `BrowserContext` descartáveis por requisição.
  - *Critério de Aceite:* Baixo consumo de RAM e isolamento de cookies entre consultas simultâneas.

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

- [ ] **DEM-401**: **Setup do Servidor MCP e Transportes**
  - Inicializar servidor MCP usando `mark3labs/mcp-go`.
  - Suportar modo Stdio (CLI/local) e modo SSE (HTTP/remoto para o ERP consumir via rede).
- [ ] **DEM-402**: **Tool `consultar_processos_cpf`**
  - Parâmetros: `cpf` (string), `tribunal` ("TJSC" | "TJPR" | "AMBOS").
  - Retorno: Lista JSON padronizada de processos encontrados.
- [ ] **DEM-403**: **Tool `obter_detalhes_processo`**
  - Parâmetros: `numero_processo` (string, formato CNJ), `tribunal` ("TJSC" | "TJPR").
  - Retorno: JSON estruturado com capa, partes, valor da causa, juízo e últimas movimentações.
- [ ] **DEM-404**: **Tool `listar_pecas_processo`**
  - Parâmetros: `numero_processo` (string), `tribunal` ("TJSC" | "TJPR").
  - Retorno: Lista de documentos disponíveis com seus IDs, datas, descrições e tamanhos.
- [ ] **DEM-405**: **Tool `baixar_peca`**
  - Parâmetros: `tribunal`, `numero_processo`, `documento_id`.
  - Retorno: Caminho do arquivo temporário baixado, ou payload codificado para entrega direta ao ERP.

---

### ÉPICO 5: Resiliência, Concorrência e Operação
*Garantir estabilidade em produção no Oracle Linux ARM64.*

- [ ] **DEM-501**: **Controle de Timeout e Retentativas com Backoff**
  - Tratar lentidões e instabilidades frequentes dos portais do TJSC/TJPR.
- [ ] **DEM-502**: **Gestão de Memória e Ciclo de Vida do Chromium**
  - Prevenir vazamento de processos `chromium` zumbis ou acúmulo excessivo de memória RAM.
  - Implementar encerramento forçado de contextos órfãos.
- [ ] **DEM-503**: **Logs Estruturados e Auditoria**
  - Registro de auditoria de consultas (sem logar senhas de certificados).
  - Medição de tempo de resposta por tribunal.

---

## 🚀 Próxima Ação Imediata (Dia 1 / Passo 1)

1. **Atualizar `.gitignore`** imediatamente para proteger a pasta de certificados.
2. **Implementar a camada Stealth no `test_browser.go`** e validar se o eProc TJSC deixa de exibir a tela de verificação de segurança.

