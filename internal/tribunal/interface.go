package tribunal

import (
	"context"

	"github.com/rafapasa/openerp-jud/internal/model"
)

// Driver define o contrato padronizado para operações de consulta e download em qualquer tribunal judicial.
type Driver interface {
	// Tribunal retorna a identificação do tribunal atendido por este driver (ex: TJSC, TJPR).
	Tribunal() model.Tribunal

	// ConsultarPorCPF busca processos vinculados ao CPF informado.
	ConsultarPorCPF(ctx context.Context, cpf string) ([]model.Processo, error)

	// ObterProcesso obtém a capa, dados gerais e histórico de um processo pelo número único CNJ.
	ObterProcesso(ctx context.Context, numeroCNJ string) (*model.Processo, error)

	// BaixarPeca realiza o download de uma peça processual específica e retorna o caminho do arquivo baixado.
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

