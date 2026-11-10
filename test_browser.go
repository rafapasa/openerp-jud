package main

import (
	"fmt"
	"log"

	"github.com/mxschmitt/playwright-go"
)

func main() {
	fmt.Println("1. Inicializando Playwright driver...")
	pw, err := playwright.Run()
	if err != nil {
		log.Fatalf("Erro ao iniciar Playwright: %v", err)
	}
	defer pw.Stop()

	fmt.Println("2. Abrindo Chromium headless...")
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
	})
	if err != nil {
		log.Fatalf("Erro ao iniciar Chromium: %v", err)
	}
	defer browser.Close()

	page, err := browser.NewPage()
	if err != nil {
		log.Fatalf("Erro ao abrir aba: %v", err)
	}

	fmt.Println("3. Navegando para o eproc do TJSC...")
	if _, err = page.Goto("https://eproc1g.tjsc.jus.br/eproc/externo_controlador.php?acao=processo_consulta_publica"); err != nil {
		log.Fatalf("Erro ao acessar eproc: %v", err)
	}

	title, err := page.Title()
	if err != nil {
		log.Fatalf("Erro ao ler título da página: %v", err)
	}

	fmt.Printf("✅ Sucesso! Título da página carregada: %s\n", title)
}
