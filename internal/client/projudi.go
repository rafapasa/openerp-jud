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
	TJPRBaseURL = "https://projudi.tjpr.jus.br"
	TJPRLogin   = "https://projudi.tjpr.jus.br/projudi/usuario/loginCertificado.do"
)

// Client implementa o driver de automação do Projudi TJPR.
type ProjudiClient struct {
	cfg            *config.Config
	browserManager browser.ContextManager
}

var _ Service = (*ProjudiClient)(nil)
var _ tribunal.Driver = (*ProjudiClient)(nil)

// NewClient cria uma nova instância do driver do Projudi TJPR.
func NewProjudiClient(cfg *config.Config, bm browser.ContextManager) *ProjudiClient {
	return &ProjudiClient{
		cfg:            cfg,
		browserManager: bm,
	}
}

// Tribunal retorna o identificador do tribunal atendido.
func (c *ProjudiClient) Tribunal() model.Tribunal {
	return model.TribunalTJPR
}

// ConsultarPorCPF consulta processos vinculados ao CPF no Projudi TJPR.
func (c *ProjudiClient) ConsultarPorCPF(ctx context.Context, cpf string) ([]model.Processo, error) {
	bCtx, err := c.browserManager.NewContext(TJPRBaseURL)
	if err != nil {
		return nil, fmt.Errorf("projudi tjpr: erro ao criar contexto: %w", err)
	}
	defer bCtx.Close()

	page, err := bCtx.NewPage()
	if err != nil {
		return nil, fmt.Errorf("projudi tjpr: erro ao criar página: %w", err)
	}
	defer page.Close()

	// TODO (Fase 3): Autenticar e buscar por CPF
	return []model.Processo{}, nil
}

// ObterProcesso busca detalhes completos de um processo pelo número único CNJ.
func (c *ProjudiClient) ObterProcesso(ctx context.Context, numeroCNJ string) (*model.Processo, error) {
	bCtx, err := c.browserManager.NewContext(TJPRBaseURL)
	if err != nil {
		return nil, fmt.Errorf("projudi tjpr: erro ao criar contexto: %w", err)
	}
	defer bCtx.Close()

	// TODO (Fase 3): Consultar processo no Projudi
	return &model.Processo{
		NumeroCNJ: numeroCNJ,
		Tribunal:  model.TribunalTJPR,
	}, nil
}

// BaixarPeca realiza o download de uma peça/documento específico do Projudi.
func (c *ProjudiClient) BaixarPeca(ctx context.Context, numeroCNJ, docID, destino string) (string, error) {
	bCtx, err := c.browserManager.NewContext(TJPRBaseURL)
	if err != nil {
		return "", fmt.Errorf("projudi tjpr: erro ao criar contexto: %w", err)
	}
	defer bCtx.Close()

	// TODO (Fase 3): Localizar anexo e baixar
	return "", nil
}
