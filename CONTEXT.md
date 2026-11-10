# CONTEXTO DO PROJETO: INTEGRAÇÃO JUDICIÁRIA (eProc TJSC & Projudi TJPR) VIA MCP EM GOLANG

> **Arquivo de Contexto para Agente / IDE Antigravity**  
> **Repositório:** `openerp-jud`  
> **Ambiente Atual:** Oracle Linux 9 (`el9.aarch64` - ARM64 / OCI)  
> **Data:** Novembro de 2026  

---

## 1. Visão Geral e Objetivo do Projeto

O objetivo deste projeto é fornecer a um **ERP de terceiros** a capacidade de **consultar processos judiciais e baixar documentos/peças processuais** associadas a um CPF ou número de processo nos sistemas:
1. **eproc** (Tribunal de Justiça de Santa Catarina - TJSC);
2. **Projudi** (Tribunal de Justiça do Paraná - TJPR).

A operação é realizada em nome de um advogado habilitado portador de um **Certificado Digital ICP-Brasil A1 (`.pfx`)**.  
O escopo inicial é **apenas leitura** (consulta de capa, andamentos, partes e download de documentos/peças), sem peticionamento neste momento.

---

## 2. Decisões Arquiteturais

### Por que Go (Golang) + MCP Server + Playwright?
1. **Ausência de APIs Abertas nos Tribunais:**
   - O eproc e o Projudi **não** fornecem endpoints REST abertos para advogados privados.
   - O padrão oficial do CNJ (**MNI - Modelo Nacional de Interoperabilidade**) é baseado em Web Services SOAP/XML e restrito a órgãos conveniados (Ministério Público, PGE, Defensoria).
   - APIs comerciais de mercado (Infosimples, Judit.io, Escavador) cobram por consulta ou exigem mensalidades fixas (R$ 1.000+ na Judit ou franquia de R$ 100 na Infosimples), além de muitas vezes não fornecerem acesso a peças restritas.
2. **Por que Go (Golang)?**
   - Alta performance e concorrência nativa (goroutines) para scraping e chamadas simultâneas.
   - Baixíssimo consumo de memória RAM comparado a Python/Node.
   - Compilação em binário único, facilitando o deploy no servidor Linux.
3. **Por que MCP (Model Context Protocol)?**
   - O serviço funciona como um **MCP Server**.
   - Expõe *tools* tipadas (`consultar_eproc_cpf`, `consultar_projudi_processo`, `baixar_peca`) para serem consumidas deterministicamente por LLMs ou pelo ERP.
4. **Tratamento do Certificado A1 (.pfx):**
   - O Playwright permite injetar o arquivo `.pfx` e sua senha diretamente nas opções de contexto do navegador (`ClientCertificates`), executando o handshake TLS (mTLS) de forma headless e **sem exibir janelas nativas de seleção de certificado do sistema operacional**.

---

## 3. Stack Tecnológica e Dependências

* **Linguagem:** Go 1.22+ (ou superior)
* **Framework MCP:** `github.com/mark3labs/mcp-go` (v1.5.0+)
* **Automação de Navegador:** `github.com/mxschmitt/playwright-go` (Chromium headless)
* **Ambiente de Execução:** Oracle Linux 9 / ARM64 (aarch64)

---

## 4. Histórico de Problemas e Correções no Ambiente Linux ARM64

Durante a configuração do ambiente no servidor Linux da Oracle Cloud, os seguintes obstáculos foram superados:

1. **Incompatibilidade de Caminho de Módulo no Go:**
   - *Problema:* `go get github.com/playwright-community/playwright-go` falhava informando que o módulo declarava seu caminho como `github.com/mxschmitt/playwright-go`.
   - *Solução:* Utilizar o caminho canônico do módulo:
     ```bash
     go get github.com/mxschmitt/playwright-go
     ```
2. **Falha de Instalação com `--with-deps` (falta do `apt-get`):**
   - *Problema:* O comando padrão tentava executar `apt-get`, mas o Oracle Linux usa `dnf`.
   - *Solução:* Baixar o Chromium sem a flag de dependências:
     ```bash
     go run github.com/mxschmitt/playwright-go/cmd/playwright@latest install chromium
     ```
3. **Biblioteca Compartilhada Faltante no Oracle Linux 9 (`libatspi`):**
   - *Problema:* `No match for argument: libatspi` no `dnf`.
   - *Solução:* No RHEL/Oracle Linux 9, a biblioteca é fornecida pelo pacote:
     ```bash
     sudo dnf install -y at-spi2-core
     ```
4. **Smoke Test do Chromium:**
   - O Chromium ARM64 foi inicializado com sucesso em modo headless e acessou o eproc do TJSC.

---

## 5. Ponto Atual de Parada e Próximo Desafio

### Situação Atual:
Ao acessar a URL pública de consulta do eproc (`https://eproc1g.tjsc.jus.br/...`), o título da página retornou:
> `Tribunal de Justiça de Santa Catarina - Verificação de segurança do portal institucional`

### Diagnóstico:
O WAF/Defesa Anti-bot do TJSC (F5 BIG-IP / Cloudflare) interceptou a requisição porque:
1. O IP de origem pertence à faixa de data center (Oracle Cloud OCI);
2. O Chromium padrão do Playwright expõe `navigator.webdriver = true` e não continha headers/user-agent reais.

### Próxima Ação Imediata:
Validar a camada de **Stealth / Evasão** no Playwright em Go:
- Iniciar o Chromium com `--disable-blink-features=AutomationControlled`;
- Definir `UserAgent` desktop realista e resolução de tela (`Viewport`);
- Injetar script (`page.AddInitScript`) para ocultar `navigator.webdriver`;
- Testar o handshake TLS fornecendo o certificado `.pfx` no `ClientCertificates`.

---

## 6. Estrutura de Código Sugerida para o Repositório

```text
openerp-jud/
├── CONTEXT.md               # Este arquivo de contexto
├── go.mod
├── go.sum
├── main.go                  # Ponto de entrada do Servidor MCP (Stdio ou SSE)
├── certs/                   # Diretório de certificados (IGNORADO NO GIT)
│   ├── .gitignore
│   └── advogado.pfx
├── internal/
│   ├── browser/
│   │   └── client.go        # Inicialização do Playwright, Chromium e Stealth
│   ├── eproc/
│   │   ├── auth.go          # Handshake mTLS e sessão no eproc TJSC
│   │   ├── search.go        # Consulta por CPF / Processo
│   │   └── download.go      # Captura de links e download de PDFs
│   └── projudi/
│       ├── auth.go          # Sessão e login no Projudi TJPR
│       ├── search.go        # Consulta no Projudi
│       └── download.go      # Download de peças no Projudi
└── test_browser.go          # Script de validação e testes rápidos
```

---

## 7. Instruções para o Próximo Passo na IDE

1. **Não commitar certificados no Git:** Garantir que a pasta `certs/` e arquivos `.pfx`/`.pem` estejam no `.gitignore`.
2. **Executar o script com Stealth:** Rodar a versão atualizada do `test_browser.go` com flags anti-detecção e salvar a captura de tela `tela_eproc.png` para inspecionar se a página de verificação de segurança foi superada.
3. **Plugar o Certificado A1:** Colocar o `.pfx` de teste no diretório `certs/` e adicionar a URL de origem `https://eproc1g.tjsc.jus.br` no `BrowserNewContextOptions.ClientCertificates`.
4. **Construir as ferramentas do MCP:** Estruturar o `main.go` com `mcp-go` para expor a tool `consultar_eproc_cpf`.