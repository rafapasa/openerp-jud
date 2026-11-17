package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"github.com/rafapasa/openerp-jud/config"
	"github.com/rafapasa/openerp-jud/internal/di"
	"github.com/rafapasa/openerp-jud/internal/helper"
	"github.com/rafapasa/openerp-jud/internal/observability/logger"
)

func main() {
	// 1. Carregar configurações
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Erro ao carregar config: %v", err)
	}

	// 2. Inicializar logger
	zapLogger, cleanup, err := logger.NewLogger(cfg)
	if err != nil {
		log.Fatalf("Erro logger: %v", err)
	}
	defer cleanup()

	zapLogger.Info("Iniciando MCP Universal",
		zap.String("env", cfg.Environment),
		zap.String("log_level", cfg.LogLevel),
		zap.Duration("browser_timeout", cfg.BrowserTimeout),
	)

	// 3. Garantir diretório de downloads temporários
	if err := helper.EnsureDir(cfg.DownloadsDir); err != nil {
		zapLogger.Fatal("Falha ao criar diretório de downloads", zap.Error(err))
	}

	// 4. Contexto cancelável por SIGINT/SIGTERM — base do shutdown gracioso
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 5. WIRE — servidor MCP chega com todas as dependências injetadas
	//    (ContextManager sob posse do Server para Close ordenado)
	app, err := di.InitializeApp(cfg)
	if err != nil {
		zapLogger.Fatal("Erro ao inicializar app via Wire", zap.Error(err))
	}
	defer func() {
		if cerr := app.Close(); cerr != nil {
			zapLogger.Error("Erro ao encerrar recursos do navegador", zap.Error(cerr))
		} else {
			zapLogger.Info("Recursos do navegador encerrados")
		}
	}()

	// 6. Serve em goroutine para permitir reação ao sinal de cancelamento
	serveErr := make(chan error, 1)
	go func() {
		zapLogger.Info("Servidor MCP 'openerp-jud' pronto", zap.String("transport", cfg.Transport))
		serveErr <- app.Serve(ctx)
	}()

	select {
	case err := <-serveErr:
		if err != nil {
			zapLogger.Fatal("Erro na execução do servidor MCP", zap.Error(err))
		}
	case <-ctx.Done():
		zapLogger.Info("Sinal de encerramento recebido, iniciando shutdown")
	}
}
