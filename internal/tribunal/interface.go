package tribunal

import (
	"context"

	"github.com/rafapasa/openerp-jud/internal/model"
)

// Service define o contrato de alto nível para operações nos tribunais (eProc TJSC, Projudi TJPR).
type Driver interface {
	Tribunal() model.Tribunal
	ConsultarPorCPF(ctx context.Context, cpf string) ([]model.Processo, error)
	ObterProcesso(ctx context.Context, numeroCNJ string) (*model.Processo, error)
	BaixarPeca(ctx context.Context, numeroCNJ, docID, destino string) (string, error)
}

// Registry gerencia a resolução de drivers por tribunal.
type Registry interface {
	// Register adiciona um driver ao registro.
	Register(driver Driver)

	// Get recupera o driver correspondente ao tribunal solicitado.
	Get(tribunal model.Tribunal) (Driver, error)

	// List retorna todos os drivers registrados.
	List() []Driver
}
