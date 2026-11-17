package client_test

import (
	"testing"

	"github.com/rafapasa/openerp-jud/internal/client"
	"github.com/rafapasa/openerp-jud/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================
// FIXTURES — eProc TJSC
//
// Baseadas na estrutura REAL confirmada em:
//   testdata/eproc/CÓDIGO FONTE - EPROC.html  (sessão autenticada, v9.23.5-3.0.8)
//
// Padrões reais observados:
//   - Container de tela: div#divInfraAreaTelaD
//   - Tabelas: table.infraTable
//   - Cabeçalho: thead > tr > th.infraTh
//   - Linhas: tr.infraTrClara / tr.infraTrEscura
//   - Células: td com atributo valign="top"
//
// ATENÇÃO: os DADOS (números de processo, nomes, valores) são fictícios.
// A ESTRUTURA é real. Quando houver captura da grid/capa/eventos reais,
// estes fixtures devem ser substituídos pelos arquivos capturados.
// ============================================================

// Fixture 1: Tabela de resultados de busca por CPF no eProc.
// Estrutura baseada em table.infraTable com th.infraTh e tr.infraTrClara.
const htmlTabelaEprocMock = `
<!DOCTYPE html>
<html lang="pt-br">
<body class="bootstrap-styles instancia-1g">
<div id="divInfraAreaGlobal">
  <div id="divInfraAreaTelaD" class="container-fluid">
    <table width="99%" class="infraTable" summary="Tabela de Processos">
      <thead>
        <tr>
          <th class="infraTh">Processo</th>
          <th class="infraTh">Classe</th>
          <th class="infraTh">Órgão Julgador</th>
          <th class="infraTh">Situação</th>
        </tr>
      </thead>
      <tbody>
        <tr class="infraTrClara">
          <td valign="top">
            <a href="controlador.php?acao=processo_selecionar&num_processo=50012345620268240001&hash=abc123">
              5001234-56.2026.8.24.0001
            </a>
          </td>
          <td valign="top">Procedimento Comum Cível</td>
          <td valign="top">1ª Vara Cível da Comarca de Florianópolis</td>
          <td valign="top">Em andamento</td>
        </tr>
        <tr class="infraTrEscura">
          <td valign="top">
            <a href="controlador.php?acao=processo_selecionar&num_processo=50098761220258240038&hash=def456">
              5009876-12.2025.8.24.0038
            </a>
          </td>
          <td valign="top">Execução de Título Extrajudicial</td>
          <td valign="top">2ª Vara Cível da Comarca de Joinville</td>
          <td valign="top">Baixado</td>
        </tr>
      </tbody>
    </table>
  </div>
</div>
</body>
</html>
`

// Fixture 1b: Tabela com uma linha malformada (apenas 2 td) entre linhas válidas.
const htmlTabelaComLinhaMalformadaMock = `
<!DOCTYPE html>
<html lang="pt-br">
<body>
<div id="divInfraAreaTelaD">
  <table width="99%" class="infraTable" summary="Tabela de Processos">
    <thead>
      <tr>
        <th class="infraTh">Processo</th>
        <th class="infraTh">Classe</th>
        <th class="infraTh">Órgão Julgador</th>
        <th class="infraTh">Situação</th>
      </tr>
    </thead>
    <tbody>
      <tr class="infraTrClara">
        <td valign="top"><a href="#">5001234-56.2026.8.24.0001</a></td>
        <td valign="top">Procedimento Comum Cível</td>
        <td valign="top">1ª Vara Cível de Florianópolis</td>
        <td valign="top">Em andamento</td>
      </tr>
      <tr class="infraTrClara">
        <td valign="top">Linha quebrada</td>
        <td valign="top">sem td suficiente</td>
      </tr>
      <tr class="infraTrEscura">
        <td valign="top"><a href="#">5009876-12.2025.8.24.0038</a></td>
        <td valign="top">Execução de Título Extrajudicial</td>
        <td valign="top">2ª Vara Cível de Joinville</td>
        <td valign="top">Baixado</td>
      </tr>
    </tbody>
  </table>
</div>
</body>
</html>
`

// Fixture 2: Mensagem de busca sem resultados no eProc.
const htmlTabelaVaziaMock = `
<!DOCTYPE html>
<html lang="pt-br">
<body>
<div id="divInfraAreaTelaD">
  <div class="infraMensagemAlerta">
    Nenhum registro encontrado para os parâmetros informados.
  </div>
</div>
</body>
</html>
`

// Fixture 3: Capa do processo no eProc TJSC.
// Estrutura baseada em table.infraTable com td.infraTdLabel + td de valor.
const htmlCapaProcessoMock = `
<!DOCTYPE html>
<html lang="pt-br">
<body>
<div id="divInfraAreaTelaD" class="container-fluid">
  <fieldset id="fldCapa" class="infraFieldset">
    <legend class="infraLegendObrigatorio">Capa do Processo</legend>
    <table id="tblCapa" class="infraTable" summary="Tabela de Capa do Processo">
      <tr class="infraTrClara">
        <td class="infraTdLabel" valign="top">Processo:</td>
        <td id="txtNumProcesso" valign="top">5001234-56.2026.8.24.0001</td>
        <td class="infraTdLabel" valign="top">Data de Autuação:</td>
        <td id="txtAutuacao" valign="top">15/02/2026 14:30:00</td>
      </tr>
      <tr class="infraTrEscura">
        <td class="infraTdLabel" valign="top">Classe da Ação:</td>
        <td id="txtClasse" valign="top">Procedimento Comum Cível</td>
        <td class="infraTdLabel" valign="top">Situação:</td>
        <td id="txtSituacao" valign="top">Ativo</td>
      </tr>
      <tr class="infraTrClara">
        <td class="infraTdLabel" valign="top">Órgão Julgador:</td>
        <td id="txtOrgaoJulgador" valign="top">1ª Vara Cível</td>
        <td class="infraTdLabel" valign="top">Comarca:</td>
        <td id="txtComarca" valign="top">Florianópolis</td>
      </tr>
      <tr class="infraTrEscura">
        <td class="infraTdLabel" valign="top">Assunto:</td>
        <td id="txtAssunto" valign="top">Indenização por Dano Material</td>
        <td class="infraTdLabel" valign="top">Valor da Causa:</td>
        <td id="txtValorCausa" valign="top">R$ 45.000,00</td>
      </tr>
    </table>
  </fieldset>

  <fieldset id="fldPartes" class="infraFieldset">
    <legend class="infraLegendObrigatorio">Partes e Representantes</legend>
    <table id="tblPartes" class="infraTable" summary="Tabela de Partes">
      <thead>
        <tr>
          <th class="infraTh">Polo</th>
          <th class="infraTh">Nome</th>
        </tr>
      </thead>
      <tbody>
        <tr class="infraTrClara">
          <td valign="top">Ativo</td>
          <td valign="top">João da Silva</td>
        </tr>
        <tr class="infraTrEscura">
          <td valign="top">Passivo</td>
          <td valign="top">Seguradora Alfa S.A.</td>
        </tr>
      </tbody>
    </table>
  </fieldset>
</div>
</body>
</html>
`

// Fixture 3b: Capa com múltiplos blocos de dados dentro de #divInfraAreaTelaD.
// O parser DEVE ler apenas o fieldset da capa (#fldCapa) e ignorar blocos extras.
const htmlCapaProcessoDivsAninhadosMock = `
<!DOCTYPE html>
<html lang="pt-br">
<body>
<div id="divInfraAreaTelaD">
  <fieldset id="fldCapa" class="infraFieldset">
    <table id="tblCapa" class="infraTable">
      <tr class="infraTrClara">
        <td class="infraTdLabel" valign="top">Processo:</td>
        <td id="txtNumProcesso" valign="top">5001234-56.2026.8.24.0001</td>
      </tr>
      <tr class="infraTrEscura">
        <td class="infraTdLabel" valign="top">Classe da Ação:</td>
        <td id="txtClasse" valign="top">Procedimento Comum Cível</td>
      </tr>
      <tr class="infraTrClara">
        <td class="infraTdLabel" valign="top">Valor da Causa:</td>
        <td id="txtValorCausa" valign="top">R$ 45.000,00</td>
      </tr>
    </table>
  </fieldset>

  <fieldset id="fldOutroBloco" class="infraFieldset">
    <table class="infraTable">
      <tr class="infraTrClara">
        <td class="infraTdLabel" valign="top">Classe da Ação:</td>
        <td id="txtClasse" valign="top">CLASSE ERRADA QUE NAO DEVE SER LIDA</td>
      </tr>
      <tr class="infraTrEscura">
        <td class="infraTdLabel" valign="top">Valor da Causa:</td>
        <td id="txtValorCausa" valign="top">R$ 999.999,99</td>
      </tr>
    </table>
  </fieldset>
</div>
</body>
</html>
`

// Fixture 3c: Capa sem a célula de valor da causa.
const htmlCapaProcessoSemValorMock = `
<!DOCTYPE html>
<html lang="pt-br">
<body>
<div id="divInfraAreaTelaD">
  <fieldset id="fldCapa" class="infraFieldset">
    <table id="tblCapa" class="infraTable">
      <tr class="infraTrClara">
        <td class="infraTdLabel" valign="top">Processo:</td>
        <td id="txtNumProcesso" valign="top">5001234-56.2026.8.24.0001</td>
      </tr>
      <tr class="infraTrEscura">
        <td class="infraTdLabel" valign="top">Classe da Ação:</td>
        <td id="txtClasse" valign="top">Procedimento Comum Cível</td>
      </tr>
    </table>
  </fieldset>
</div>
</body>
</html>
`

// Fixture 4: Tabela de eventos/documentos do eProc.
// Estrutura baseada em table.infraTable com links de documento.
const htmlEventosDocumentosMock = `
<!DOCTYPE html>
<html lang="pt-br">
<body>
<div id="divInfraAreaTelaD">
  <fieldset id="fldEventos" class="infraFieldset">
    <legend class="infraLegendObrigatorio">Eventos</legend>
    <table id="tblEventos" class="infraTable" summary="Tabela de Eventos">
      <thead>
        <tr>
          <th class="infraTh">Evento</th>
          <th class="infraTh">Descrição</th>
          <th class="infraTh">Documentos</th>
        </tr>
      </thead>
      <tbody>
        <tr class="infraTrClara">
          <td valign="top">1</td>
          <td valign="top">Petição Inicial</td>
          <td valign="top">
            <a class="infraLinkDocumento" href="controlador.php?acao=acessar_documento&doc=doc101&hash=abc123def">
              PETICAO1.pdf
            </a>
          </td>
        </tr>
        <tr class="infraTrEscura">
          <td valign="top">2</td>
          <td valign="top">Despacho</td>
          <td valign="top">
            <a class="infraLinkDocumento" href="controlador.php?acao=acessar_documento&doc=doc102&hash=xyz987uvw">
              DESP1.pdf
            </a>
          </td>
        </tr>
      </tbody>
    </table>
  </fieldset>
</div>
</body>
</html>
`

// Fixture 4b: Documentos com data-id explícito (quando o eProc renderiza o atributo).
const htmlEventosDocumentosComDataIDMock = `
<!DOCTYPE html>
<html lang="pt-br">
<body>
<div id="divInfraAreaTelaD">
  <table id="tblEventos" class="infraTable">
    <thead>
      <tr>
        <th class="infraTh">Evento</th>
        <th class="infraTh">Descrição</th>
        <th class="infraTh">Documentos</th>
      </tr>
    </thead>
    <tbody>
      <tr class="infraTrClara">
        <td valign="top">1</td>
        <td valign="top">Petição Inicial</td>
        <td valign="top">
          <a class="infraLinkDocumento" data-id="doc201" href="controlador.php?acao=acessar_documento&doc=doc201">
            PETICAO1.pdf
          </a>
        </td>
      </tr>
    </tbody>
  </table>
</div>
</body>
</html>
`

// Fixture 4c: Documento sem ID identificável (link sem doc= e sem data-id).
const htmlEventosDocumentoSemIDMock = `
<!DOCTYPE html>
<html lang="pt-br">
<body>
<div id="divInfraAreaTelaD">
  <table id="tblEventos" class="infraTable">
    <tbody>
      <tr class="infraTrClara">
        <td valign="top">1</td>
        <td valign="top">Petição Inicial</td>
        <td valign="top">
          <a class="infraLinkDocumento" href="controlador.php?acao=outra_acao">
            PETICAO1.pdf
          </a>
        </td>
      </tr>
    </tbody>
  </table>
</div>
</body>
</html>
`

// Fixture 4d: Tabela de eventos vazia.
const htmlEventosVazioMock = `
<!DOCTYPE html>
<html lang="pt-br">
<body>
<div id="divInfraAreaTelaD">
  <table id="tblEventos" class="infraTable">
    <thead><tr><th class="infraTh">Evento</th><th class="infraTh">Descrição</th><th class="infraTh">Documentos</th></tr></thead>
    <tbody></tbody>
  </table>
</div>
</body>
</html>
`

// Fixture 5: HTML inválido (tag não fechada e malformada).
const htmlInvalidoMock = `<html><body><div><table id="tblEventos"><tr><td>`

// --- TESTES UNITÁRIOS ---

func TestEProcParser_ParseTabelaProcessos_Sucesso(t *testing.T) {
	parser := client.NewEProcParser()

	processos, err := parser.ParseTabelaProcessos(htmlTabelaEprocMock)

	require.NoError(t, err)
	require.Len(t, processos, 2, "deve extrair exatamente 2 processos da tabela HTML")

	p1 := processos[0]
	assert.Equal(t, "5001234-56.2026.8.24.0001", p1.NumeroCNJ)
	assert.Equal(t, "Procedimento Comum Cível", p1.Classe)
	assert.Equal(t, "1ª Vara Cível da Comarca de Florianópolis", p1.Vara)
	assert.Equal(t, "Em andamento", p1.Situacao)
	assert.Equal(t, model.TribunalTJSC, p1.Tribunal)

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

	require.Len(t, proc.Partes, 2)
	assert.Equal(t, "João da Silva", proc.Partes[0].Nome)
	assert.Equal(t, "Ativo", proc.Partes[0].Polo)
	assert.Equal(t, "Seguradora Alfa S.A.", proc.Partes[1].Nome)
	assert.Equal(t, "Passivo", proc.Partes[1].Polo)
}

func TestEProcParser_ParseCapaProcesso_BlocosAninhados(t *testing.T) {
	parser := client.NewEProcParser()

	proc, err := parser.ParseCapaProcesso(htmlCapaProcessoDivsAninhadosMock)

	require.NoError(t, err)
	require.NotNil(t, proc)

	// Deve ler SOMENTE de #fldCapa, ignorando #fldOutroBloco
	assert.Equal(t, "Procedimento Comum Cível", proc.Classe,
		"deve ignorar o bloco extra e ler apenas #fldCapa")
	assert.Equal(t, 45000.0, proc.ValorCausa,
		"deve ignorar o valor do bloco extra e ler apenas #fldCapa")
}

func TestEProcParser_ParseCapaProcesso_ValorCausaAusente(t *testing.T) {
	parser := client.NewEProcParser()

	proc, err := parser.ParseCapaProcesso(htmlCapaProcessoSemValorMock)

	require.NoError(t, err)
	require.NotNil(t, proc)
	assert.Equal(t, 0.0, proc.ValorCausa,
		"deve retornar 0 quando a célula de valor da causa não existir (após log de warning)")
	assert.Equal(t, "5001234-56.2026.8.24.0001", proc.NumeroCNJ)
	assert.Equal(t, "Procedimento Comum Cível", proc.Classe)
}

func TestEProcParser_ParseDocumentosProcesso_Sucesso(t *testing.T) {
	parser := client.NewEProcParser()

	docs, err := parser.ParseDocumentosProcesso(htmlEventosDocumentosMock)

	require.NoError(t, err)
	require.Len(t, docs, 2, "deve extrair 2 documentos da árvore de eventos")

	assert.Equal(t, "doc101", docs[0].ID, "deve extrair o parâmetro doc= do href")
	assert.Equal(t, "PETICAO1.pdf", docs[0].Nome)
	assert.Equal(t, "Petição Inicial", docs[0].Descricao)

	assert.Equal(t, "doc102", docs[1].ID)
	assert.Equal(t, "DESP1.pdf", docs[1].Nome)
	assert.Equal(t, "Despacho", docs[1].Descricao)
}

func TestEProcParser_ParseDocumentosProcesso_DataID(t *testing.T) {
	parser := client.NewEProcParser()

	docs, err := parser.ParseDocumentosProcesso(htmlEventosDocumentosComDataIDMock)

	require.NoError(t, err)
	require.Len(t, docs, 1)
	assert.Equal(t, "doc201", docs[0].ID, "data-id tem precedência sobre o href")
	assert.Equal(t, "PETICAO1.pdf", docs[0].Nome)
}

func TestEProcParser_ParseDocumentosProcesso_SemID(t *testing.T) {
	parser := client.NewEProcParser()

	docs, err := parser.ParseDocumentosProcesso(htmlEventosDocumentoSemIDMock)

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

	assert.NotPanics(t, func() {
		_, _ = parser.ParseTabelaProcessos(htmlInvalidoMock)
	}, "parser não deve entrar em panic com HTML malformado")
}
