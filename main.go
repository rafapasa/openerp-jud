package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/rafapasa/openerp-jud/config"
	"github.com/rafapasa/openerp-jud/internal/browser"
	"github.com/rafapasa/openerp-jud/internal/client"
	"github.com/rafapasa/openerp-jud/internal/mcp"
	"github.com/rafapasa/openerp-jud/internal/tribunal"
)

func main() {
	// 1. Carregar configurações
	cfg := config.LoadConfig()

	// 2. Garantir diretório de downloads temporários
	if err := os.MkdirAll(cfg.DownloadsDir, 0755); err != nil {
		log.Fatalf("Falha ao criar diretório de downloads: %v", err)
	}

	// 3. Inicializar o gerenciador de navegador Playwright/Chromium
	browserManager, err := browser.NewManager(cfg)
	if err != nil {
		log.Printf("AVISO: Falha ao inicializar o Chromium: %v", err)
		log.Printf("O servidor MCP continuará, mas as chamadas de automação retornarão erro até que as dependências do Playwright estejam disponíveis.")
	} else {
		defer browserManager.Close()
	}

	// 4. Inicializar drivers dos tribunais
	eprocClient := client.NewEProcClient(cfg, browserManager)
	projudiClient := client.NewProjudiClient(cfg, browserManager)

	// 5. Registrar os drivers no Registry
	registry := tribunal.NewRegistry()
	registry.Register(eprocClient)
	registry.Register(projudiClient)

	// 6. Criar Handler e Servidor MCP
	handler := mcp.NewToolHandler(cfg, registry)
	server := mcp.NewServer(cfg, handler)

	// 7. Configurar cancelamento gracioso via sinais do SO
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// 8. Iniciar o servidor MCP (Stdio ou SSE)
	log.Printf("Servidor MCP 'openerp-jud' pronto. Modo de transporte: %s", cfg.Transport)
	if err := server.Serve(ctx); err != nil {
		log.Fatalf("Erro na execução do servidor MCP: %v", err)
	}
}
