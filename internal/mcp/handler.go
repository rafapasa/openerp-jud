package mcp

import (
	"context"
	"fmt"
	"sync"

	"github.com/rafapasa/openerp-jud/config"
	"github.com/rafapasa/openerp-jud/internal/model"
	"github.com/rafapasa/openerp-jud/internal/tribunal"
)

// DefaultToolHandler implementa a interface ToolHandler delegando as operações
// para os drivers judiciais apropriados via Registry.
type DefaultToolHandler struct {
	cfg      *config.Config
	registry tribunal.Registry
}

var _ ToolHandler = (*DefaultToolHandler)(nil)

// NewToolHandler cria uma nova instância de DefaultToolHandler.
func NewToolHandler(cfg *config.Config, registry tribunal.Registry) *DefaultToolHandler {
	return &DefaultToolHandler{
		cfg:      cfg,
		registry: registry,
	}
}

// HandleConsultarProcessosCPF consulta processos por CPF no(s) tribunal(is) especificado(s).
func (h *DefaultToolHandler) HandleConsultarProcessosCPF(ctx context.Context, cpf string, trib model.Tribunal) ([]model.Processo, error) {
	if cpf == "" {
		return nil, fmt.Errorf("cpf obrigatório para consulta")
	}

	if trib == model.TribunalTodos || trib == "" {
		// Consulta concorrente em todos os drivers registrados
		drivers := h.registry.List()
		if len(drivers) == 0 {
			return nil, fmt.Errorf("nenhum driver de tribunal registrado")
		}

		type result struct {
			processos []model.Processo
			err       error
		}

		resCh := make(chan result, len(drivers))
		var wg sync.WaitGroup

		for _, d := range drivers {
			wg.Add(1)
			go func(drv tribunal.Driver) {
				defer wg.Done()
				procs, err := drv.ConsultarPorCPF(ctx, cpf)
				resCh <- result{processos: procs, err: err}
			}(d)
		}

		wg.Wait()
		close(resCh)

		var todosProcessos []model.Processo
		var ultErr error
		for res := range resCh {
			if res.err != nil {
				ultErr = res.err
				continue
			}
			todosProcessos = append(todosProcessos, res.processos...)
		}

		if len(todosProcessos) == 0 && ultErr != nil {
			return nil, fmt.Errorf("falha ao consultar tribunais: %w", ultErr)
		}

		return todosProcessos, nil
	}

	driver, err := h.registry.Get(trib)
	if err != nil {
		return nil, err
	}

	return driver.ConsultarPorCPF(ctx, cpf)
}

// HandleObterDetalhesProcesso recupera capa e andamentos de um processo pelo número CNJ.
func (h *DefaultToolHandler) HandleObterDetalhesProcesso(ctx context.Context, numeroProcesso string, trib model.Tribunal) (*model.Processo, error) {
	if numeroProcesso == "" {
		return nil, fmt.Errorf("numero do processo obrigatório")
	}

	driver, err := h.registry.Get(trib)
	if err != nil {
		return nil, err
	}

	return driver.ObterProcesso(ctx, numeroProcesso)
}

// HandleBaixarPeca gerencia o download de uma peça judicial.
func (h *DefaultToolHandler) HandleBaixarPeca(ctx context.Context, trib model.Tribunal, numeroProcesso, docID string) (string, error) {
	if docID == "" {
		return "", fmt.Errorf("id do documento obrigatório")
	}

	driver, err := h.registry.Get(trib)
	if err != nil {
		return "", err
	}

	return driver.BaixarPeca(ctx, numeroProcesso, docID, h.cfg.DownloadsDir)
}

