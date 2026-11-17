//go:build wireinject
// +build wireinject

package di

import (
	"github.com/google/wire"
	"github.com/rafapasa/openerp-jud/config"
	"github.com/rafapasa/openerp-jud/internal/mcp"
)

// InitializeApp monta o grafo de dependências da aplicação e devolve o servidor MCP pronto.
// A configuração é carregada pelo main.go e injetada como parâmetro.
func InitializeApp(cfg *config.Config) (mcp.Server, error) {
	wire.Build(ProviderSet)
	return nil, nil
}
