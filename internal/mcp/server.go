package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
	"github.com/rafapasa/openerp-jud/config"
	"github.com/rafapasa/openerp-jud/internal/model"
)

// MCPServer encapsula o servidor MCP e registra as ferramentas disponíveis para o ERP.
type MCPServer struct {
	cfg     *config.Config
	server  *mcpserver.MCPServer
	handler ToolHandler
}

var _ Server = (*MCPServer)(nil)

// NewServer cria uma nova instância do servidor MCP.
func NewServer(cfg *config.Config, handler ToolHandler) *MCPServer {
	s := mcpserver.NewMCPServer(
		"openerp-jud",
		"1.0.0",
		mcpserver.WithToolCapabilities(true),
	)

	return &MCPServer{
		cfg:     cfg,
		server:  s,
		handler: handler,
	}
}

// RegisterTools registra todas as ferramentas judiciais disponíveis.
func (s *MCPServer) RegisterTools() error {
	// 1. Tool: consultar_processos_cpf
	s.server.AddTool(
		mcp.NewTool(
			"consultar_processos_cpf",
			mcp.WithDescription("Consulta processos judiciais associados ao CPF da parte nos tribunais atendidos (TJSC e/ou TJPR)."),
			mcp.WithString("cpf", mcp.Required(), mcp.Description("CPF do cliente/parte no formato 000.000.000-00 ou numérico")),
			mcp.WithString("tribunal", mcp.Description("Tribunal a consultar: 'TJSC', 'TJPR' ou 'TODOS' (padrão: TODOS)"), mcp.Enum("TJSC", "TJPR", "TODOS")),
		),
		s.handleConsultarCPF,
	)

	// 2. Tool: obter_detalhes_processo
	s.server.AddTool(
		mcp.NewTool(
			"obter_detalhes_processo",
			mcp.WithDescription("Obtém dados detalhados da capa, partes e movimentações de um processo pelo número CNJ."),
			mcp.WithString("numero_processo", mcp.Required(), mcp.Description("Número único CNJ do processo (ex: 0000000-00.0000.8.24.0000)")),
			mcp.WithString("tribunal", mcp.Required(), mcp.Description("Tribunal do processo: 'TJSC' ou 'TJPR'"), mcp.Enum("TJSC", "TJPR")),
		),
		s.handleObterDetalhes,
	)

	// 3. Tool: baixar_peca
	s.server.AddTool(
		mcp.NewTool(
			"baixar_peca",
			mcp.WithDescription("Realiza o download de uma peça/documento judicial em PDF e retorna o caminho local do arquivo."),
			mcp.WithString("tribunal", mcp.Required(), mcp.Description("Tribunal: 'TJSC' ou 'TJPR'"), mcp.Enum("TJSC", "TJPR")),
			mcp.WithString("numero_processo", mcp.Required(), mcp.Description("Número do processo")),
			mcp.WithString("documento_id", mcp.Required(), mcp.Description("ID ou código identificador do documento/anexo")),
		),
		s.handleBaixarPeca,
	)

	return nil
}

func (s *MCPServer) handleConsultarCPF(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cpf, err := request.RequireString("cpf")
	if err != nil || cpf == "" {
		return mcp.NewToolResultError("argumento 'cpf' é obrigatório"), nil
	}

	tribStr := request.GetString("tribunal", string(model.TribunalTodos))
	trib := model.Tribunal(tribStr)

	processos, err := s.handler.HandleConsultarProcessosCPF(ctx, cpf, trib)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("erro ao consultar processos: %v", err)), nil
	}

	bytes, err := json.MarshalIndent(processos, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("falha ao serializar resultado: %v", err)), nil
	}

	return mcp.NewToolResultText(string(bytes)), nil
}

func (s *MCPServer) handleObterDetalhes(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	numero, err := request.RequireString("numero_processo")
	if err != nil || numero == "" {
		return mcp.NewToolResultError("argumento 'numero_processo' é obrigatório"), nil
	}

	tribStr, err := request.RequireString("tribunal")
	if err != nil || tribStr == "" {
		return mcp.NewToolResultError("argumento 'tribunal' é obrigatório ('TJSC' ou 'TJPR')"), nil
	}

	proc, err := s.handler.HandleObterDetalhesProcesso(ctx, numero, model.Tribunal(tribStr))
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("erro ao obter detalhes do processo: %v", err)), nil
	}

	bytes, err := json.MarshalIndent(proc, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("falha ao serializar resultado: %v", err)), nil
	}

	return mcp.NewToolResultText(string(bytes)), nil
}

func (s *MCPServer) handleBaixarPeca(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tribStr, err := request.RequireString("tribunal")
	if err != nil || tribStr == "" {
		return mcp.NewToolResultError("argumento 'tribunal' é obrigatório"), nil
	}

	numero, err := request.RequireString("numero_processo")
	if err != nil || numero == "" {
		return mcp.NewToolResultError("argumento 'numero_processo' é obrigatório"), nil
	}

	docID, err := request.RequireString("documento_id")
	if err != nil || docID == "" {
		return mcp.NewToolResultError("argumento 'documento_id' é obrigatório"), nil
	}

	caminhoArquivo, err := s.handler.HandleBaixarPeca(ctx, model.Tribunal(tribStr), numero, docID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("erro ao baixar peça: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf(`{"status": "sucesso", "caminho_arquivo": "%s"}`, caminhoArquivo)), nil
}

// Serve inicia o servidor MCP no modo configurado (Stdio ou SSE/HTTP).
func (s *MCPServer) Serve(ctx context.Context) error {
	if err := s.RegisterTools(); err != nil {
		return fmt.Errorf("falha ao registrar tools: %w", err)
	}

	if s.cfg.Transport == "sse" {
		addr := ":" + s.cfg.Port
		log.Printf("[MCP] Iniciando servidor MCP via SSE na porta %s...", addr)
		sse := mcpserver.NewSSEServer(s.server)
		return sse.Start(addr)
	}

	log.Printf("[MCP] Iniciando servidor MCP via Stdio...")
	return mcpserver.ServeStdio(s.server)
}

