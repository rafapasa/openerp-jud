package client_test

import (
	"testing"

	"github.com/rafapasa/openerp-jud/internal/client"
	"github.com/rafapasa/openerp-jud/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================
// FIXTURES — ProJudi TJSC (mock estrutural)
// Ajustar seletores conforme HTML real da implantação alvo.
// ============================================================

// Fixture 1: Tabela de resultados de busca por CPF no ProJudi.
// ProJudi costuma usar tabelas com classe "table" ou "tabelaLista".
const htmlProjudiTabelaProcessosMock = `
<!DOCTYPE html>
<html>
<body>
<table id="tabelaProcessos" class="table table-striped">
    <thead>
        <tr>
            <th>Número do Processo</th>
            <th>Classe</th>
            <th>Vara / Comarca</th>
            <th>Situação</th>
        </tr>
    </thead>
    <tbody>
        <tr>
            <td>
                <a href="/projudi/processo.do?numero=50012345620268240001">
                    5001234-56.2026.8.24.0001
                </a>
            </td>
            <td>Procedimento Comum Cível</td>
            <td>1ª Vara Cível - Florianópolis</td>
            <td>Em andamento</td>
        </tr>
        <tr>
            <td>
                <a href="/projudi/processo.do?numero=50098761220258240038">
                    5009876-12.2025.8.24.0038
                </a>
            </td>
            <td>Execução de Título Extrajudicial</td>
            <td>2ª Vara Cível - Joinville</td>
            <td>Baixado</td>
        </tr>
    </tbody>
</table>
</body>
</html>
`

// Fixture 2: Busca sem resultados no ProJudi.
const htmlProjudiTabelaVaziaMock = `
<!DOCTYPE html>
<html>
<body>
<div class="alert alert-warning">
    Nenhum processo encontrado para os critérios informados.
</div>
</body>
</html>
`

// Fixture 3: Capa do processo no ProJudi.
// ProJudi frequentemente usa <div class="form-group"> com <label> + <span>.
const htmlProjudiCapaProcessoMock = `
<!DOCTYPE html>
<html>
<body>
<div id="conteudoProcesso">
    <h2 id="numeroProcesso">5001234-56.2026.8.24.0001</h2>

    <div class="form-group">
        <label>Classe:</label>
        <span id="classeProcesso">Procedimento Comum Cível</span>
    </div>
    <div class="form-group">
        <label>Comarca:</label>
        <span id="comarcaProcesso">Florianópolis</span>
    </div>
    <div class="form-group">
        <label>Vara:</label>
        <span id="varaProcesso">1ª Vara Cível</span>
    </div>
    <div class="form-group">
        <label>Assunto:</label>
        <span id="assuntoProcesso">Indenização por Dano Material</span>
    </div>
    <div class="form-group">
        <label>Valor da Causa:</label>
        <span id="valorCausaProcesso">R$ 45.000,00</span>
    </div>
    <div class="form-group">
        <label>Situação:</label>
        <span id="situacaoProcesso">Ativo</span>
    </div>

    <table id="tabelaPartes" class="table">
        <thead><tr><th>Polo</th><th>Nome</th></tr></thead>
        <tbody>
            <tr>
                <td>Ativo</td>
                <td>João da Silva</td>
            </tr>
            <tr>
                <td>Passivo</td>
                <td>Seguradora Alfa S.A.</td>
            </tr>
        </tbody>
    </table>
</div>
</body>
</html>
`

// Fixture 3b: Capa com múltiplos divs (regressão do Bug 1 do eProc).
const htmlProjudiCapaDivsAninhadosMock = `
<!DOCTYPE html>
<html>
<body>
<div id="conteudoProcesso">
    <h2 id="numeroProcesso">5001234-56.2026.8.24.0001</h2>

    <div class="form-group">
        <label>Classe:</label>
        <span id="classeProcesso">Procedimento Comum Cível</span>
    </div>
    <div class="form-group">
        <label>Valor da Causa:</label>
        <span id="valorCausaProcesso">R$ 45.000,00</span>
    </div>

    <div class="sidebar-extra">
        <span id="classeProcesso">CLASSE ERRADA QUE NAO DEVE SER LIDA</span>
        <span id="valorCausaProcesso">R$ 999.999,99</span>
    </div>
</div>
</body>
</html>
`

// Fixture 3c: Capa sem valor da causa.
const htmlProjudiCapaSemValorMock = `
<!DOCTYPE html>
<html>
<body>
<div id="conteudoProcesso">
    <h2 id="numeroProcesso">5001234-56.2026.8.24.0001</h2>
    <div class="form-group">
        <label>Classe:</label>
        <span id="classeProcesso">Procedimento Comum Cível</span>
    </div>
</div>
</body>
</html>
`

// Fixture 4: Tabela de eventos/documentos do ProJudi.
const htmlProjudiEventosDocumentosMock = `
<!DOCTYPE html>
<html>
<body>
<table id="tabelaEventos" class="table">
    <thead>
        <tr>
            <th>Evento</th>
            <th>Descrição</th>
            <th>Documentos</th>
        </tr>
    </thead>
    <tbody>
        <tr>
            <td>1</td>
            <td>Petição Inicial</td>
            <td>
                <a href="/projudi/documento.do?id=doc101" data-id="doc101">
                    PETICAO1.pdf
                </a>
            </td>
        </tr>
        <tr>
            <td>2</td>
            <td>Despacho</td>
            <td>
                <a href="/projudi/documento.do?id=doc102" data-id="doc102">
                    DESP1.pdf
                </a>
            </td>
        </tr>
    </tbody>
</table>
</body>
</html>
`

// Fixture 4b: Documentos sem data-id, apenas href com id.
const htmlProjudiDocumentosSemDataIDMock = `
<!DOCTYPE html>
<html>
<body>
<table id="tabelaEventos" class="table">
    <tbody>
        <tr>
            <td>1</td>
            <td>Petição Inicial</td>
            <td>
                <a href="/projudi/documento.do?id=doc201">PETICAO1.pdf</a>
            </td>
        </tr>
    </tbody>
</table>
</body>
</html>
`

// Fixture 4c: Documento sem ID identificável.
const htmlProjudiDocumentoSemIDMock = `
<!DOCTYPE html>
<html>
<body>
<table id="tabelaEventos" class="table">
    <tbody>
        <tr>
            <td>1</td>
            <td>Petição Inicial</td>
            <td>
                <a href="/projudi/outra_acao">PETICAO1.pdf</a>
            </td>
        </tr>
    </tbody>
</table>
</body>
</html>
`

// Fixture 4d: Tabela de eventos vazia.
const htmlProjudiEventosVazioMock = `
<!DOCTYPE html>
<html>
<body>
<table id="tabelaEventos" class="table">
    <thead><tr><th>Evento</th><th>Descrição</th><th>Documentos</th></tr></thead>
    <tbody></tbody>
</table>
</body>
</html>
`

// Fixture 5: HTML malformado (não deve causar panic).
const htmlProjudiInvalidoMock = `<html><body><div><table id="tabelaProcessos"><tr><td>`

// ============================================================
// TESTES — ParseTabelaProcessos
// ============================================================

func TestProjudiParser_ParseTabelaProcessos_Sucesso(t *testing.T) {
	parser := client.NewProjudiParser()

	processos, err := parser.ParseTabelaProcessos(htmlProjudiTabelaProcessosMock)

	require.NoError(t, err)
	require.Len(t, processos, 2, "deve extrair exatamente 2 processos da tabela HTML")

	p1 := processos[0]
	assert.Equal(t, "5001234-56.2026.8.24.0001", p1.NumeroCNJ)
	assert.Equal(t, "Procedimento Comum Cível", p1.Classe)
	assert.Equal(t, "1ª Vara Cível - Florianópolis", p1.Vara)
	assert.Equal(t, "Em andamento", p1.Situacao)
	assert.Equal(t, model.TribunalTJPR, p1.Tribunal)

	p2 := processos[1]
	assert.Equal(t, "5009876-12.2025.8.24.0038", p2.NumeroCNJ)
	assert.Equal(t, "Execução de Título Extrajudicial", p2.Classe)
	assert.Equal(t, "2ª Vara Cível - Joinville", p2.Vara)
	assert.Equal(t, "Baixado", p2.Situacao)
	assert.Equal(t, model.TribunalTJPR, p2.Tribunal)
}

func TestProjudiParser_ParseTabelaProcessos_Vazia(t *testing.T) {
	parser := client.NewProjudiParser()

	processos, err := parser.ParseTabelaProcessos(htmlProjudiTabelaVaziaMock)

	require.NoError(t, err)
	assert.Empty(t, processos, "deve retornar lista vazia quando não houver processos")
	assert.NotNil(t, processos, "deve retornar slice não-nil para serialização JSON consistente")
}

func TestProjudiParser_ParseTabelaProcessos_HTMLInvalido(t *testing.T) {
	parser := client.NewProjudiParser()

	assert.NotPanics(t, func() {
		_, _ = parser.ParseTabelaProcessos(htmlProjudiInvalidoMock)
	}, "parser não deve entrar em panic com HTML malformado")
}

// ============================================================
// TESTES — ParseCapaProcesso
// ============================================================

func TestProjudiParser_ParseCapaProcesso_Sucesso(t *testing.T) {
	parser := client.NewProjudiParser()

	proc, err := parser.ParseCapaProcesso(htmlProjudiCapaProcessoMock)

	require.NoError(t, err)
	require.NotNil(t, proc)

	assert.Equal(t, "5001234-56.2026.8.24.0001", proc.NumeroCNJ)
	assert.Equal(t, "Procedimento Comum Cível", proc.Classe)
	assert.Equal(t, "Florianópolis", proc.Comarca)
	assert.Equal(t, "1ª Vara Cível", proc.Vara)
	assert.Equal(t, "Indenização por Dano Material", proc.Assunto)
	assert.Equal(t, 45000.0, proc.ValorCausa)
	assert.Equal(t, "Ativo", proc.Situacao)
	assert.Equal(t, model.TribunalTJPR, proc.Tribunal)

	require.Len(t, proc.Partes, 2)
	assert.Equal(t, "João da Silva", proc.Partes[0].Nome)
	assert.Equal(t, "Ativo", proc.Partes[0].Polo)
	assert.Equal(t, "Seguradora Alfa S.A.", proc.Partes[1].Nome)
	assert.Equal(t, "Passivo", proc.Partes[1].Polo)
}

func TestProjudiParser_ParseCapaProcesso_DivsAninhados(t *testing.T) {
	parser := client.NewProjudiParser()

	proc, err := parser.ParseCapaProcesso(htmlProjudiCapaDivsAninhadosMock)

	require.NoError(t, err)
	require.NotNil(t, proc)

	// Deve ler SOMENTE de div.form-group, ignorando div.sidebar-extra
	assert.Equal(t, "Procedimento Comum Cível", proc.Classe,
		"deve ignorar o div extra e ler apenas os form-groups da capa")
	assert.Equal(t, 45000.0, proc.ValorCausa,
		"deve ignorar o valor do div extra")
}

func TestProjudiParser_ParseCapaProcesso_ValorCausaAusente(t *testing.T) {
	parser := client.NewProjudiParser()

	proc, err := parser.ParseCapaProcesso(htmlProjudiCapaSemValorMock)

	require.NoError(t, err)
	require.NotNil(t, proc)
	assert.Equal(t, 0.0, proc.ValorCausa,
		"deve retornar 0 quando o span de valor da causa não existir")
	assert.Equal(t, "5001234-56.2026.8.24.0001", proc.NumeroCNJ)
	assert.Equal(t, "Procedimento Comum Cível", proc.Classe)
}

// ============================================================
// TESTES — ParseDocumentosProcesso
// ============================================================

func TestProjudiParser_ParseDocumentosProcesso_Sucesso(t *testing.T) {
	parser := client.NewProjudiParser()

	docs, err := parser.ParseDocumentosProcesso(htmlProjudiEventosDocumentosMock)

	require.NoError(t, err)
	require.Len(t, docs, 2, "deve extrair 2 documentos da árvore de eventos")

	assert.Equal(t, "doc101", docs[0].ID)
	assert.Equal(t, "PETICAO1.pdf", docs[0].Nome)
	assert.Equal(t, "Petição Inicial", docs[0].Descricao)

	assert.Equal(t, "doc102", docs[1].ID)
	assert.Equal(t, "DESP1.pdf", docs[1].Nome)
	assert.Equal(t, "Despacho", docs[1].Descricao)
}

func TestProjudiParser_ParseDocumentosProcesso_FallbackHref(t *testing.T) {
	parser := client.NewProjudiParser()

	docs, err := parser.ParseDocumentosProcesso(htmlProjudiDocumentosSemDataIDMock)

	require.NoError(t, err)
	require.Len(t, docs, 1, "deve extrair documento usando fallback de href")
	assert.Equal(t, "doc201", docs[0].ID,
		"deve extrair id do href quando data-id estiver ausente")
	assert.Equal(t, "PETICAO1.pdf", docs[0].Nome)
}

func TestProjudiParser_ParseDocumentosProcesso_SemID(t *testing.T) {
	parser := client.NewProjudiParser()

	docs, err := parser.ParseDocumentosProcesso(htmlProjudiDocumentoSemIDMock)

	// Mesmo contrato do eProc: documento sem ID identificável deve ser descartado.
	require.NoError(t, err)
	assert.Empty(t, docs,
		"documento sem ID identificável deve ser descartado, não incluído com ID vazio")
}

func TestProjudiParser_ParseDocumentosProcesso_Vazio(t *testing.T) {
	parser := client.NewProjudiParser()

	docs, err := parser.ParseDocumentosProcesso(htmlProjudiEventosVazioMock)

	require.NoError(t, err)
	assert.Empty(t, docs)
}
