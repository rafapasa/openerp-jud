package client

import (
	"github.com/rafapasa/openerp-jud/internal/model"
	"github.com/rafapasa/openerp-jud/internal/tribunal"
)

// EProcClient define o contrato do driver de automação do eProc TJSC.
// Implementa tribunal.Driver e é consumido pelo Registry e pelo MCP Server.
type EProcClient interface {
	tribunal.Driver
}

// ProjudiClient define o contrato do driver de automação do Projudi TJPR.
// Implementa tribunal.Driver e é consumido pelo Registry e pelo MCP Server.
type ProjudiClient interface {
	tribunal.Driver
}

// HTMLParser define as operações de extração e conversão de dados do HTML dos tribunais.
// Permite testes unitários de parser com snapshots HTML puros, sem necessitar do navegador.
type HTMLParser interface {
	// ParseTabelaProcessos extrai a lista de processos a partir da tabela de resultados de busca por CPF.
	ParseTabelaProcessos(html string) ([]model.Processo, error)

	// ParseCapaProcesso extrai os metadados da capa (partes, vara, valor da causa, situação) do processo.
	ParseCapaProcesso(html string) (*model.Processo, error)

	// ParseDocumentosProcesso extrai a lista de eventos e peças disponíveis para download.
	ParseDocumentosProcesso(html string) ([]model.Documento, error)
}

// FormParser define as operações de extração dos seletores de formulários de consulta.
//
// Existe porque o driver precisa preencher formulários cujos seletores variam entre
// os portais (eProc público, eProc autenticado, Projudi). Centralizar a extração no
// parser permite testar os seletores contra HTML real capturado, sem subir navegador.
type FormParser interface {
	// ParseFormularioConsulta extrai os seletores e metadados do formulário de consulta.
	ParseFormularioConsulta(html string) (*model.FormularioConsulta, error)
}
