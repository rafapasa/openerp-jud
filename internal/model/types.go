package model

import "time"

// Tribunal representa os sistemas judiciais suportados.
type Tribunal string

const (
	TribunalTJSC  Tribunal = "TJSC"  // eProc 1G / 2G TJSC
	TribunalTJPR  Tribunal = "TJPR"  // Projudi TJPR
	TribunalTodos Tribunal = "TODOS" // Consulta federada em ambos
)

// ParteProcesso representa uma parte no processo (autor, réu, advogado, etc).
type ParteProcesso struct {
	Nome      string   `json:"nome"`
	Documento string   `json:"documento,omitempty"` // CPF/CNPJ se disponível
	Polo      string   `json:"polo"`                // Ativo, Passivo, Terceiro
	Advogados []string `json:"advogados,omitempty"`
}

// Documento representa uma peça/anexo processual.
type Documento struct {
	ID        string    `json:"id"`
	Nome      string    `json:"nome"`
	Descricao string    `json:"descricao,omitempty"`
	Data      time.Time `json:"data,omitempty"`
	Tipo      string    `json:"tipo,omitempty"`
	Tamanho   int64     `json:"tamanho_bytes,omitempty"`
	URL       string    `json:"url,omitempty"`
	FilePath  string    `json:"file_path,omitempty"` // Caminho local se já baixado
}

// Movimentacao representa um evento ou andamento processual.
type Movimentacao struct {
	Numero     int         `json:"numero,omitempty"`
	Data       time.Time   `json:"data"`
	Descricao  string      `json:"descricao"`
	Complemento string     `json:"complemento,omitempty"`
	Documentos []Documento `json:"documentos,omitempty"`
}

// Processo representa os metadados e andamento de um processo judicial.
type Processo struct {
	NumeroCNJ          string         `json:"numero_cnj"`
	NumeroAntigo       string         `json:"numero_antigo,omitempty"`
	Tribunal           Tribunal       `json:"tribunal"`
	Comarca            string         `json:"comarca,omitempty"`
	Vara               string         `json:"vara,omitempty"`
	Classe             string         `json:"classe,omitempty"`
	Assunto            string         `json:"assunto,omitempty"`
	DataDistribuicao   *time.Time     `json:"data_distribuicao,omitempty"`
	ValorCausa         float64        `json:"valor_causa,omitempty"`
	Situacao           string         `json:"situacao,omitempty"` // Ativo, Suspenso, Baixado
	SegredoJustica     bool           `json:"segredo_justica"`
	Partes             []ParteProcesso `json:"partes,omitempty"`
	Movimentacoes      []Movimentacao `json:"movimentacoes,omitempty"`
	TotalMovimentacoes int            `json:"total_movimentacoes,omitempty"`
}

// ConsultaFiltro parâmetros de busca no tribunal.
type ConsultaFiltro struct {
	CPF            string   `json:"cpf,omitempty"`
	CNPJ           string   `json:"cnpj,omitempty"`
	NumeroProcesso string   `json:"numero_processo,omitempty"`
	Tribunal       Tribunal `json:"tribunal"`
}

// FormularioConsulta descreve os seletores e metadados do formulário de consulta
// de um portal de tribunal, permitindo que o driver preencha os campos sem
// depender de seletores hardcoded.
type FormularioConsulta struct {
	// Action é a URL de destino do <form>.
	Action string `json:"action,omitempty"`
	// Method é o método HTTP do <form> (GET ou POST).
	Method string `json:"method,omitempty"`

	// CampoNumeroProcesso é o seletor do input de número do processo.
	CampoNumeroProcesso string `json:"campo_numero_processo,omitempty"`
	// CampoNomeParte é o seletor do input de nome da parte.
	CampoNomeParte string `json:"campo_nome_parte,omitempty"`
	// CampoDocumento é o seletor do input de CPF/CNPJ.
	CampoDocumento string `json:"campo_documento,omitempty"`
	// CampoOAB é o seletor do input de OAB.
	CampoOAB string `json:"campo_oab,omitempty"`

	// RadioPessoaFisica é o seletor do radio de CPF.
	RadioPessoaFisica string `json:"radio_pessoa_fisica,omitempty"`
	// RadioPessoaJuridica é o seletor do radio de CNPJ.
	RadioPessoaJuridica string `json:"radio_pessoa_juridica,omitempty"`

	// BotaoSubmit é o seletor do botão de submissão.
	BotaoSubmit string `json:"botao_submit,omitempty"`

	// PossuiCaptcha indica se o formulário exige resolução de captcha.
	PossuiCaptcha bool `json:"possui_captcha"`
	// CaptchaSiteKey é a chave do site do captcha, quando presente.
	CaptchaSiteKey string `json:"captcha_site_key,omitempty"`
}
