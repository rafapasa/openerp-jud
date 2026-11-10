package client

import (
	"context"
	"fmt"

	"github.com/rafapasa/openerp-jud/config"
	"github.com/rafapasa/openerp-jud/internal/browser"
	"github.com/rafapasa/openerp-jud/internal/model"
	"github.com/rafapasa/openerp-jud/internal/tribunal"
)

const (
	TJSCBaseURL      = "https://eproc1g.tjsc.jus.br"
	TJSCPublicSearch = "https://eproc1g.tjsc.jus.br/eproc/externo_controlador.php?acao=processo_consulta_publica"
)

// Client implementa o driver de automação do eProc TJSC.
type EProcClient struct {
	cfg            *config.Config
	browserManager browser.ContextManager
}

var _ Service = (*EProcClient)(nil)
var _ tribunal.Driver = (*EProcClient)(nil)

// NewClient cria uma nova instância do driver do eProc TJSC.
func NewEProcClient(cfg *config.Config, bm browser.ContextManager) *EProcClient {
	return &EProcClient{
		cfg:            cfg,
		browserManager: bm,
	}
}

// Tribunal retorna o identificador do tribunal atendido.
func (c *EProcClient) Tribunal() model.Tribunal {
	return model.TribunalTJSC
}

// ConsultarPorCPF consulta processos vinculados ao CPF no eProc TJSC.
func (c *EProcClient) ConsultarPorCPF(ctx context.Context, cpf string) ([]model.Processo, error) {
	bCtx, err := c.browserManager.NewContext(TJSCBaseURL)
	if err != nil {
		return nil, fmt.Errorf("eproc tjsc: erro ao criar contexto: %w", err)
	}
	defer bCtx.Close()

	page, err := bCtx.NewPage()
	if err != nil {
		return nil, fmt.Errorf("eproc tjsc: erro ao criar página: %w", err)
	}
	defer page.Close()

	// Navegação inicial
	if _, err := page.Goto(TJSCPublicSearch); err != nil {
		return nil, fmt.Errorf("eproc tjsc: falha ao acessar portal: %w", err)
	}

	// TODO (Fase 2): Preencher formulário de consulta por CPF, aguardar grid e extrair dados
	return []model.Processo{}, nil
}

// ObterProcesso busca detalhes completos de um processo pelo número único CNJ.
func (c *EProcClient) ObterProcesso(ctx context.Context, numeroCNJ string) (*model.Processo, error) {
	bCtx, err := c.browserManager.NewContext(TJSCBaseURL)
	if err != nil {
		return nil, fmt.Errorf("eproc tjsc: erro ao criar contexto: %w", err)
	}
	defer bCtx.Close()

	page, err := bCtx.NewPage()
	if err != nil {
		return nil, fmt.Errorf("eproc tjsc: erro ao criar página: %w", err)
	}
	defer page.Close()

	// TODO (Fase 2): Navegar para tela do processo e parsear dados de capa e movimentações
	return &model.Processo{
		NumeroCNJ: numeroCNJ,
		Tribunal:  model.TribunalTJSC,
	}, nil
}

// BaixarPeca realiza o download de uma peça/documento específico em PDF.
func (c *EProcClient) BaixarPeca(ctx context.Context, numeroCNJ, docID, destino string) (string, error) {
	bCtx, err := c.browserManager.NewContext(TJSCBaseURL)
	if err != nil {
		return "", fmt.Errorf("eproc tjsc: erro ao criar contexto: %w", err)
	}
	defer bCtx.Close()

	// TODO (Fase 2): Navegar, clicar no documento e salvar PDF no diretório de destino
	return "", nil
}
