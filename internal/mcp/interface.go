package mcp

import (
	"context"

	"github.com/rafapasa/openerp-jud/internal/model"
)

// Server define o ciclo de vida do servidor MCP.
type Server interface {
	// RegisterTools registra todas as ferramentas disponíveis no MCP Server.
	RegisterTools() error

	// Serve inicializa a escuta de requisições no transporte configurado (Stdio ou SSE/HTTP).
	Serve(ctx context.Context) error
}

// ToolHandler define os contratos de execução das ferramentas (Tools) expostas pelo MCP Server para o ERP.
// Ao ser desacoplada em interface, permite testar os manipuladores de tools e a serialização JSON
// de forma 100% isolada com mocks rápidos.
type ToolHandler interface {
	// HandleConsultarProcessosCPF processa a busca de processos por CPF no(s) tribunal(is) especificado(s).
	HandleConsultarProcessosCPF(ctx context.Context, cpf string, tribunal model.Tribunal) ([]model.Processo, error)

	// HandleObterDetalhesProcesso processa a recuperação detalhada de capa e movimentações de um processo.
	HandleObterDetalhesProcesso(ctx context.Context, numeroProcesso string, tribunal model.Tribunal) (*model.Processo, error)

	// HandleBaixarPeca processa o download da peça/documento judicial para entrega ao ERP.
	HandleBaixarPeca(ctx context.Context, tribunal model.Tribunal, numeroProcesso, docID string) (string, error)
}

