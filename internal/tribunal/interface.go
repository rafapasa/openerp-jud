package tribunal

import (
	"context"

	"github.com/rafapasa/openerp-jud/internal/model"
)

// EProcClient define o contrato do driver de automação do eProc TJSC.
// Implementa tribunal.Driver e é consumido pelo Registry e pelo MCP Server.
type EProcClient interface {
	Driver
}

// ProjudiClient define o contrato do driver de automação do Projudi TJPR.
// Implementa tribunal.Driver e é consumido pelo Registry e pelo MCP Server.
type ProjudiClient interface {
	Driver
}

// Driver define o contrato de alto nível para operações nos tribunais (eProc TJSC, Projudi TJPR).
type Driver interface {
	Tribunal() model.Tribunal
	ConsultarPorCPF(ctx context.Context, cpf string) ([]model.Processo, error)
	ObterProcesso(ctx context.Context, numeroCNJ string) (*model.Processo, error)
	BaixarPeca(ctx context.Context, numeroCNJ, docID, destino string) (string, error)
}
