// Command capture_html grava HTML real do eProc (consulta pública) em testdata/.
//
// Limitações conhecidas (2026-11):
//   - IP de datacenter costuma cair no WAF F5 ("Verificação de segurança").
//   - POST de consulta exige Cloudflare Turnstile; sem token não há grid de processos.
//
// Uso:
//
//	go run ./cmd/capture_html -out testdata/eproc/public
//	go run ./cmd/capture_html -cpf 03624841980 -out testdata/eproc/public
//	go run ./cmd/capture_html -processo 5000000-00.2020.8.24.0000 -out testdata/eproc/public
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rafapasa/openerp-jud/config"
	"github.com/rafapasa/openerp-jud/internal/browser"
	"github.com/rafapasa/openerp-jud/internal/helper"
)

const (
	publicSearchURL = "https://eprocwebcon.tjsc.jus.br/consulta1g/externo_controlador.php?acao=processo_consulta_publica"
	publicOrigin    = "https://eprocwebcon.tjsc.jus.br"
)

func main() {
	outDir := flag.String("out", "testdata/eproc/public", "diretório de saída")
	cpf := flag.String("cpf", "", "CPF para preencher no formulário (somente dígitos ou formatado)")
	processo := flag.String("processo", "", "número CNJ para preencher")
	headless := flag.Bool("headless", true, "rodar Chromium headless")
	flag.Parse()

	if err := helper.EnsureDir(*outDir); err != nil {
		fail(err)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		fail(err)
	}
	cfg.Headless = *headless

	cm, err := browser.NewContextManager(cfg)
	if err != nil {
		fail(fmt.Errorf("browser: %w", err))
	}
	defer cm.Close()

	bctx, err := cm.NewContext(publicOrigin)
	if err != nil {
		fail(fmt.Errorf("contexto: %w", err))
	}
	defer bctx.Close()

	page, err := bctx.NewPage()
	if err != nil {
		fail(fmt.Errorf("página: %w", err))
	}
	defer page.Close()

	fmt.Println(">> navegando para consulta pública eProc")
	if err := page.Goto(publicSearchURL); err != nil {
		fail(fmt.Errorf("goto: %w", err))
	}

	html, err := page.Content()
	if err != nil {
		fail(err)
	}

	stamp := time.Now().Format("20060102_150405")
	formPath := filepath.Join(*outDir, fmt.Sprintf("browser_form_%s.html", stamp))
	if err := os.WriteFile(formPath, []byte(html), 0o644); err != nil {
		fail(err)
	}
	fmt.Println("   salvo:", formPath)

	if isWAF(html) {
		fmt.Println("!! WAF/F5 detectado — captura de lista/capa impossível neste IP/sessão.")
		fmt.Println("   Use rede não-datacenter ou resolva o desafio manualmente (headless=false).")
		os.Exit(2)
	}

	if strings.TrimSpace(*cpf) == "" && strings.TrimSpace(*processo) == "" {
		fmt.Println("OK formulário capturado. Passe -cpf ou -processo para tentar a consulta.")
		return
	}

	// A interface Page atual só expõe Goto/Content/Close.
	// Preenchimento de formulário exige evolução da porta browser.Page
	// (Fill/Click) ou uso temporário do playwright neste cmd.
	// Por ora gravamos o formulário e orientamos o próximo passo.
	_ = context.Background()
	fmt.Println("!! Preenchimento automático (CPF/processo) ainda depende de Page.Fill/Click na porta browser.")
	fmt.Println("   Dados pedidos — cpf=", *cpf, " processo=", *processo)
	fmt.Println("   Próximo incremento: estender browser.Page e completar o fluxo Turnstile + submit.")
}

func isWAF(html string) bool {
	l := strings.ToLower(html)
	return strings.Contains(l, "verificação de segurança") ||
		strings.Contains(l, "verificacao de seguranca") ||
		strings.Contains(l, "support id")
}

func fail(err error) {
	fmt.Fprintf(os.Stderr, "capture_html: %v\n", err)
	os.Exit(1)
}
