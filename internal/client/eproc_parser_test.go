package client_test

import (
	"testing"

	"github.com/rafapasa/openerp-jud/internal/client"
	"github.com/rafapasa/openerp-jud/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Fixture 1: Tabela de resultados de busca por CPF típica do eProc TJSC
const htmlTabelaEprocMock = `
<!DOCTYPE html>
<html>
<body>
<table id="conteudoTabela" class="infraTable" summary="Tabela de Processos">
    <thead>
        <tr class="infraTh">
            <th>Processo</th>
            <th>Classe</th>
            <th>Órgão Julgador / Vara</th>
            <th>Situação</th>
        </tr>
    </thead>
    <tbody>
        <tr class="infraTrClara">
            <td>
                <a href="/eproc/externo_controlador.php?acao=processo_selecionar&num_processo=50012345620268240001">
                    5001234-56.2026.8.24.0001
                </a>
            </td>
            <td>Procedimento Comum Cível</td>
            <td>1ª Vara Cível da Comarca de Florianópolis</td>
            <td>Em andamento</td>
        </tr>
        <tr class="infraTrEscura">
            <td>
                <a href="/eproc/externo_controlador.php?acao=processo_selecionar&num_processo=50098761220258240038">
                    5009876-12.2025.8.24.0038
                </a>
            </td>
            <td>Execução de Título Extrajudicial</td>
            <td>2ª Vara Cível da Comarca de Joinville</td>
            <td>Baixado</td>
        </tr>
    </tbody>
</table>
</body>
</html>
`

// Fixture 1b: Tabela com uma linha malformada (apenas 2 td) entre linhas válidas
const htmlTabelaComLinhaMalformadaMock = `
<!DOCTYPE html>
<html>
<body>
<table id="conteudoTabela" class="infraTable">
    <tbody>
        <tr>
            <td><a href="#">5001234-56.2026.8.24.0001</a></td>
            <td>Procedimento Comum Cível</td>
            <td>1ª Vara Cível de Florianópolis</td>
            <td>Em andamento</td>
        </tr>
        <tr>
            <td>Linha quebrada</td>
            <td>sem td suficiente</td>
        </tr>
        <tr>
            <td><a href="#">5009876-12.2025.8.24.0038</a></td>
            <td>Execução de Título Extrajudicial</td>
            <td>2ª Vara Cível de Joinville</td>
            <td>Baixado</td>
        </tr>
    </tbody>
</table>
</body>
</html>
`

// Fixture 2: Mensagem de busca sem resultados no eProc
const htmlTabelaVaziaMock = `
<!DOCTYPE html>
<html>
<body>
<div class="infraMensagemAlerta">
    Nenhum registro encontrado para os parâmetros informados.
</div>
</body>
</html>
`

// Fixture 3: Capa do processo no eProc TJSC
const htmlCapaProcessoMock = `
<!DOCTYPE html>
<html>
<body>
    <div id="divConteudo">
    <span id="txtNumProcesso">5001234-56.2026.8.24.0001</span>
    <div class="infraAreaDados">
        <label>Classe:</label>
        <span id="txtClasse">Procedimento Comum Cível</span>

        <label>Comarca:</label>
        <span id="txtComarca">Florianópolis</span>

        <label>Órgão Julgador:</label>
        <span id="txtOrgaoJulgador">1ª Vara Cível</span>

        <label>Assunto:</label>
        <span id="txtAssunto">Indenização por Dano Material</span>

        <label>Valor da Causa:</label>
        <span id="txtValorCausa">R$ 45.000,00</span>

        <label>Situação:</label>
        <span id="txtSituacao">Ativo</span>
    </div>

    <!-- Tabela de Partes -->
    <table id="tblPartes" class="infraTable">
        <tr>
            <td class="polo">Ativo</td>
            <td class="nome">João da Silva</td>
        </tr>
        <tr>
            <td class="polo">Passivo</td>
            <td class="nome">Seguradora Alfa S.A.</td>
        </tr>
    </table>
</div>
</body>
</html>
`

// Fixture 3b: Capa com múltiplos <div> dentro de #divConteudo.
// O parser DEVE extrair dados de div.infraAreaDados e IGNORAR os outros divs.
// Este fixture falharia com a implementação antiga (Find("div") genérico).
const htmlCapaProcessoDivsAninhadosMock = `
<!DOCTYPE html>
<html>
<body>
    <div id="divConteudo">
        <span id="txtNumProcesso">5001234-56.2026.8.24.0001</span>

        <div class="infraAreaDados">
            <span id="txtClasse">Procedimento Comum Cível</span>
            <span id="txtComarca">Florianópolis</span>
            <span id="txtOrgaoJulgador">1ª Vara Cível</span>
            <span id="txtAssunto">Indenização por Dano Material</span>
            <span id="txtValorCausa">R$ 45.000,00</span>
            <span id="txtSituacao">Ativo</span>
        </div>

        <div class="outra-area-que-nao-e-a-capa">
            <span id="txtClasse">CLASSE ERRADA QUE NAO DEVE SER LIDA</span>
            <span id="txtValorCausa">R$ 999.999,99</span>
        </div>
    </div>
</body>
</html>
`

// Fixture 3c: Capa sem o span de valor da causa
const htmlCapaProcessoSemValorMock = `
<!DOCTYPE html>
<html>
<body>
    <div id="divConteudo">
        <span id="txtNumProcesso">5001234-56.2026.8.24.0001</span>
        <div class="infraAreaDados">
            <span id="txtClasse">Procedimento Comum Cível</span>
            <span id="txtComarca">Florianópolis</span>
            <span id="txtOrgaoJulgador">1ª Vara Cível</span>
            <span id="txtAssunto">Indenização por Dano Material</span>
            <span id="txtSituacao">Ativo</span>
        </div>
    </div>
</body>
</html>
`

// Fixture 4: Tabela de eventos/documentos do eProc
const htmlEventosDocumentosMock = `
<!DOCTYPE html>
<html>
<body>
<table id="tblEventos" class="infraTable">
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
                <a href="/eproc/externo_controlador.php?acao=documento_download&id_documento=doc101" data-id="doc101">
                    PETICAO1.pdf
                </a>
            </td>
        </tr>
        <tr>
            <td>2</td>
            <td>Despacho</td>
            <td>
                <a href="/eproc/externo_controlador.php?acao=documento_download&id_documento=doc102" data-id="doc102">
                    DESP1.pdf
                </a>
            </td>
        </tr>
    </tbody>
</table>
</body>
</html>
`

// Fixture 4b: Documentos SEM data-id, apenas com href contendo id_documento
const htmlEventosDocumentosSemDataIDMock = `
<!DOCTYPE html>
<html>
<body>
<table id="tblEventos" class="infraTable">
    <tbody>
        <tr>
            <td>1</td>
            <td>Petição Inicial</td>
            <td>
                <a href="/eproc/externo_controlador.php?acao=documento_download&id_documento=doc201">
                    PETICAO1.pdf
                </a>
            </td>
        </tr>
        <tr>
            <td>2</td>
            <td>Despacho</td>
            <td>
                <a href="/eproc/externo_controlador.php?acao=documento_download&id_documento=doc202">
                    DESP1.pdf
                </a>
            </td>
        </tr>
    </tbody>
</table>
</body>
</html>
`

// Fixture 4c: Documento sem data-id E sem id_documento no href (link vazio)
const htmlEventosDocumentoSemIDMock = `
<!DOCTYPE html>
<html>
<body>
<table id="tblEventos" class="infraTable">
    <tbody>
        <tr>
            <td>1</td>
            <td>Petição Inicial</td>
            <td>
                <a href="/eproc/externo_controlador.php?acao=outra_acao">
                    PETICAO1.pdf
                </a>
            </td>
        </tr>
    </tbody>
</table>
</body>
</html>
`

// Fixture 4d: Tabela de eventos vazia (sem tbody tr)
const htmlEventosVazioMock = `
<!DOCTYPE html>
<html>
<body>
<table id="tblEventos" class="infraTable">
    <thead><tr><th>Evento</th><th>Descrição</th><th>Documentos</th></tr></thead>
    <tbody></tbody>
</table>
</body>
</html>
`

// Fixture 5: HTML inválido (tag não fechada e malformada)
const htmlInvalidoMock = `<html><body><div><table id="conteudoTabela"><tr><td>`

// --- TESTES UNITÁRIOS ---

func TestEProcParser_ParseTabelaProcessos_Sucesso(t *testing.T) {
	parser := client.NewEProcParser()

	processos, err := parser.ParseTabelaProcessos(htmlTabelaEprocMock)

	require.NoError(t, err)
	require.Len(t, processos, 2, "deve extrair exatamente 2 processos da tabela HTML")

	// Validação do Processo 1
	p1 := processos[0]
	assert.Equal(t, "5001234-56.2026.8.24.0001", p1.NumeroCNJ)
	assert.Equal(t, "Procedimento Comum Cível", p1.Classe)
	assert.Equal(t, "1ª Vara Cível da Comarca de Florianópolis", p1.Vara)
	assert.Equal(t, "Em andamento", p1.Situacao)
	assert.Equal(t, model.TribunalTJSC, p1.Tribunal)

	// Validação do Processo 2
	p2 := processos[1]
	assert.Equal(t, "5009876-12.2025.8.24.0038", p2.NumeroCNJ)
	assert.Equal(t, "Execução de Título Extrajudicial", p2.Classe)
	assert.Equal(t, "2ª Vara Cível da Comarca de Joinville", p2.Vara)
	assert.Equal(t, "Baixado", p2.Situacao)
	assert.Equal(t, model.TribunalTJSC, p2.Tribunal)
}

func TestEProcParser_ParseTabelaProcessos_Vazia(t *testing.T) {
	parser := client.NewEProcParser()

	processos, err := parser.ParseTabelaProcessos(htmlTabelaVaziaMock)

	require.NoError(t, err)
	assert.Empty(t, processos, "deve retornar lista vazia quando não houver processos")
	assert.NotNil(t, processos, "deve retornar slice não-nil (make(..., 0)) para serialização JSON consistente")
}

func TestEProcParser_ParseTabelaProcessos_LinhaMalformada(t *testing.T) {
	parser := client.NewEProcParser()

	processos, err := parser.ParseTabelaProcessos(htmlTabelaComLinhaMalformadaMock)

	require.NoError(t, err)
	require.Len(t, processos, 2, "linha com menos de 4 td deve ser ignorada")

	assert.Equal(t, "5001234-56.2026.8.24.0001", processos[0].NumeroCNJ)
	assert.Equal(t, "5009876-12.2025.8.24.0038", processos[1].NumeroCNJ)
}

func TestEProcParser_ParseCapaProcesso_Sucesso(t *testing.T) {
	parser := client.NewEProcParser()

	proc, err := parser.ParseCapaProcesso(htmlCapaProcessoMock)

	require.NoError(t, err)
	require.NotNil(t, proc, "deve retornar os dados de capa preenchidos")

	assert.Equal(t, "5001234-56.2026.8.24.0001", proc.NumeroCNJ)
	assert.Equal(t, "Procedimento Comum Cível", proc.Classe)
	assert.Equal(t, "Florianópolis", proc.Comarca)
	assert.Equal(t, "1ª Vara Cível", proc.Vara)
	assert.Equal(t, "Indenização por Dano Material", proc.Assunto)
	assert.Equal(t, 45000.0, proc.ValorCausa)
	assert.Equal(t, "Ativo", proc.Situacao)
	assert.Equal(t, model.TribunalTJSC, proc.Tribunal)

	// Partes
	require.Len(t, proc.Partes, 2)
	assert.Equal(t, "João da Silva", proc.Partes[0].Nome)
	assert.Equal(t, "Ativo", proc.Partes[0].Polo)
	assert.Equal(t, "Seguradora Alfa S.A.", proc.Partes[1].Nome)
	assert.Equal(t, "Passivo", proc.Partes[1].Polo)
}

func TestEProcParser_ParseCapaProcesso_DivsAninhados(t *testing.T) {
	parser := client.NewEProcParser()

	proc, err := parser.ParseCapaProcesso(htmlCapaProcessoDivsAninhadosMock)

	require.NoError(t, err)
	require.NotNil(t, proc)

	// Deve ler SOMENTE de div.infraAreaDados, ignorando o div "outra-area-que-nao-e-a-capa"
	assert.Equal(t, "Procedimento Comum Cível", proc.Classe,
		"deve ignorar o div extra e ler apenas div.infraAreaDados")
	assert.Equal(t, 45000.0, proc.ValorCausa,
		"deve ignorar o valor do div extra e ler apenas div.infraAreaDados")
}

func TestEProcParser_ParseCapaProcesso_ValorCausaAusente(t *testing.T) {
	parser := client.NewEProcParser()

	proc, err := parser.ParseCapaProcesso(htmlCapaProcessoSemValorMock)

	require.NoError(t, err)
	require.NotNil(t, proc)
	assert.Equal(t, 0.0, proc.ValorCausa,
		"deve retornar 0 quando o span de valor da causa não existir (após log de warning)")
	assert.Equal(t, "5001234-56.2026.8.24.0001", proc.NumeroCNJ)
	assert.Equal(t, "Procedimento Comum Cível", proc.Classe)
}

func TestEProcParser_ParseDocumentosProcesso_Sucesso(t *testing.T) {
	parser := client.NewEProcParser()

	docs, err := parser.ParseDocumentosProcesso(htmlEventosDocumentosMock)

	require.NoError(t, err)
	require.Len(t, docs, 2, "deve extrair 2 documentos da árvore de eventos")

	assert.Equal(t, "doc101", docs[0].ID)
	assert.Equal(t, "PETICAO1.pdf", docs[0].Nome)
	assert.Equal(t, "Petição Inicial", docs[0].Descricao)

	assert.Equal(t, "doc102", docs[1].ID)
	assert.Equal(t, "DESP1.pdf", docs[1].Nome)
	assert.Equal(t, "Despacho", docs[1].Descricao)
}

func TestEProcParser_ParseDocumentosProcesso_FallbackHref(t *testing.T) {
	parser := client.NewEProcParser()

	docs, err := parser.ParseDocumentosProcesso(htmlEventosDocumentosSemDataIDMock)

	require.NoError(t, err)
	require.Len(t, docs, 2, "deve extrair 2 documentos usando fallback de href")

	assert.Equal(t, "doc201", docs[0].ID, "deve extrair id_documento do href quando data-id estiver ausente")
	assert.Equal(t, "PETICAO1.pdf", docs[0].Nome)
	assert.Equal(t, "doc202", docs[1].ID)
	assert.Equal(t, "DESP1.pdf", docs[1].Nome)
}

func TestEProcParser_ParseDocumentosProcesso_SemID(t *testing.T) {
	parser := client.NewEProcParser()

	docs, err := parser.ParseDocumentosProcesso(htmlEventosDocumentoSemIDMock)

	// Comportamento desejado após correção do bug residual:
	// documento sem ID identificável NÃO deve ser adicionado à lista.
	require.NoError(t, err)
	assert.Empty(t, docs, "documento sem ID identificável deve ser descartado, não incluído com ID vazio")
}

func TestEProcParser_ParseDocumentosProcesso_Vazio(t *testing.T) {
	parser := client.NewEProcParser()

	docs, err := parser.ParseDocumentosProcesso(htmlEventosVazioMock)

	require.NoError(t, err)
	assert.Empty(t, docs)
}

func TestEProcParser_ParseTabelaProcessos_HTMLInvalido(t *testing.T) {
	parser := client.NewEProcParser()

	// goquery/html são tolerantes a HTML malformado em geral; este caso
	// valida que, mesmo com entrada quebrada, o parser NÃO entra em panic
	// e retorna ou lista vazia ou erro — nunca panic.
	assert.NotPanics(t, func() {
		_, _ = parser.ParseTabelaProcessos(htmlInvalidoMock)
	}, "parser não deve entrar em panic com HTML malformado")
}
