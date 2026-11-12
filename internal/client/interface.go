package client

import "github.com/rafapasa/openerp-jud/internal/model"

// HTMLParser define as operações de extração e conversão de dados do HTML do eProc.
// Permite testes unitários de parser com snapshots HTML puros, sem necessitar do navegador.
type HTMLParser interface {
	// ParseTabelaProcessos extrai a lista de processos a partir da tabela de resultados de busca por CPF.
	ParseTabelaProcessos(html string) ([]model.Processo, error)

	// ParseCapaProcesso extrai os metadados da capa (partes, vara, valor da causa, situação) do processo.
	ParseCapaProcesso(html string) (*model.Processo, error)

	// ParseDocumentosProcesso extrai a lista de eventos e peças disponíveis para download.
	ParseDocumentosProcesso(html string) ([]model.Documento, error)
}
