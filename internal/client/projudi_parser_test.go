package client_test

import (
	"testing"

	"github.com/rafapasa/openerp-jud/internal/client"
	"github.com/rafapasa/openerp-jud/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================
// FIXTURES — Projudi TJPR
//
// Baseadas na estrutura REAL confirmada em:
//   testdata/projudi/código fonte frame projudi.html  (sessão autenticada)
//
// Padrões reais observados:
//   - Layout em <table> (sem Bootstrap)
//   - Tabelas de dados: table.form
//   - Rótulo: td.label > label
//   - Valor: td (sem classe) ou td.value
//   - Linhas alternadas: tr.odd / tr.even
//   - Conteúdo dentro do frame "mainFrame" (frameset)
//
// ATENÇÃO: os DADOS (números de processo, nomes, valores) são fictícios.
// A ESTRUTURA é real. Quando houver captura da lista/capa/movimentações
// reais, estes fixtures devem ser substituídos pelos arquivos capturados.
// ============================================================

// Fixture 1: Tabela de resultados de busca por CPF no Projudi.
const htmlProjudiTabelaProcessosMock = `
<!DOCTYPE html>
<html lang="pt-BR">
<body>
<div id="content">
  <table class="form" id="tabelaProcessos">
    <thead>
      <tr>
        <th>Número do Processo</th>
        <th>Classe</th>
        <th>Vara / Comarca</th>
        <th>Situação</th>
      </tr>
    </thead>
    <tbody>
      <tr class="odd">
        <td>
          <a href="/projudi/processo/processo.do?numero=50012345620268240001">5001234-56.2026.8.24.0001</a>
        </td>
        <td>Procedimento Comum Cível</td>
        <td>1ª Vara Cível - Florianópolis</td>
        <td>Em andamento</td>
      </tr>
      <tr class="even">
        <td>
          <a href="/projudi/processo/processo.do?numero=50098761220258240038">5009876-12.2025.8.24.0038</a>
        </td>
        <td>Execução de Título Extrajudicial</td>
        <td>2ª Vara Cível - Joinville</td>
        <td>Baixado</td>
      </tr>
    </tbody>
  </table>
</div>
</body>
</html>
`

// Fixture 2: Busca sem resultados no Projudi.
const htmlProjudiTabelaVaziaMock = `
<!DOCTYPE html>
<html lang="pt-BR">
<body>
<div id="content">
  <div class="alert alert-warning">
    Nenhum processo encontrado para os critérios informados.
  </div>
</div>
</body>
</html>
`

// Fixture 3: Capa do processo no Projudi.
// Estrutura baseada em table.form com td.label + td de valor.
const htmlProjudiCapaProcessoMock = `
<!DOCTYPE html>
<html lang="pt-BR">
<body>
<div id="content">
  <h3>Dados do Processo</h3>
  <table class="form" id="tabelaCapa">
    <tr>
      <td class="label"><label>Número Único:</label></td>
      <td class="value"><strong>5001234-56.2026.8.24.0001</strong></td>
      <td class="label"><label>Distribuição:</label></td>
      <td class="value">10/03/2026 às 09:15:22</td>
    </tr>
    <tr>
      <td class="label"><label>Classe Processual:</label></td>
      <td class="value">Procedimento Comum Cível</td>
      <td class="label"><label>Assunto Principal:</label></td>
      <td class="value">Indenização por Dano Material</td>
    </tr>
    <tr>
      <td class="label"><label>Vara / Juízo:</label></td>
      <td class="value">1ª Vara Cível</td>
      <td class="label"><label>Valor da Causa:</label></td>
      <td class="value">R$ 45.000,00</td>
    </tr>
    <tr>
      <td class="label"><label>Comarca:</label></td>
      <td class="value">Florianópolis</td>
      <td class="label"><label>Situação:</label></td>
      <td class="value">Ativo</td>
    </tr>
  </table>

  <h4>Partes do Processo</h4>
  <table class="resultTable" id="tabelaPartes">
    <thead>
      <tr>
        <th>Polo</th>
        <th>Nome</th>
        <th>Documento</th>
        <th>Advogado(s)</th>
      </tr>
    </thead>
    <tbody>
      <tr class="odd">
        <td>Ativo</td>
        <td>João da Silva</td>
        <td>111.222.333-44</td>
        <td>DR. BELTRANO (OAB/PR 54321)</td>
      </tr>
      <tr class="even">
        <td>Passivo</td>
        <td>Seguradora Alfa S.A.</td>
        <td>12.345.678/0001-99</td>
        <td>-</td>
      </tr>
    </tbody>
  </table>
</div>
</body>
</html>
`

// Fixture 3b: Capa com múltiplas tabelas .form.
// O parser DEVE ler apenas a tabela de capa (#tabelaCapa) e ignorar as demais.
const htmlProjudiCapaDivsAninhadosMock = `
<!DOCTYPE html>
<html lang="pt-BR">
<body>
<div id="content">
  <table class="form" id="tabelaCapa">
    <tr>
      <td class="label"><label>Número Único:</label></td>
      <td class="value"><strong>5001234-56.2026.8.24.0001</strong></td>
    </tr>
    <tr>
      <td class="label"><label>Classe Processual:</label></td>
      <td class="value">Procedimento Comum Cível</td>
    </tr>
    <tr>
      <td class="label"><label>Valor da Causa:</label></td>
      <td class="value">R$ 45.000,00</td>
    </tr>
  </table>

  <table class="form" id="tabelaOutroBloco">
    <tr>
      <td class="label"><label>Classe Processual:</label></td>
      <td class="value">CLASSE ERRADA QUE NAO DEVE SER LIDA</td>
    </tr>
    <tr>
      <td class="label"><label>Valor da Causa:</label></td>
      <td class="value">R$ 999.999,99</td>
    </tr>
  </table>
</div>
</body>
</html>
`

// Fixture 3c: Capa sem a célula de valor da causa.
const htmlProjudiCapaSemValorMock = `
<!DOCTYPE html>
<html lang="pt-BR">
<body>
<div id="content">
  <table class="form" id="tabelaCapa">
    <tr>
      <td class="label"><label>Número Único:</label></td>
      <td class="value"><strong>5001234-56.2026.8.24.0001</strong></td>
    </tr>
    <tr>
      <td class="label"><label>Classe Processual:</label></td>
      <td class="value">Procedimento Comum Cível</td>
    </tr>
  </table>
</div>
</body>
</html>
`

// Fixture 4: Tabela de movimentações/documentos do Projudi.
const htmlProjudiEventosDocumentosMock = `
<!DOCTYPE html>
<html lang="pt-BR">
<body>
<div id="content">
  <h4>Movimentações</h4>
  <table class="resultTable" id="tabelaMovimentacoes">
    <thead>
      <tr>
        <th>Seq.</th>
        <th>Data</th>
        <th>Movimentação</th>
        <th>Arquivos</th>
      </tr>
    </thead>
    <tbody>
      <tr class="odd">
        <td>2</td>
        <td>12/03/2026 11:00</td>
        <td>Proferido despacho de mero expediente</td>
        <td>
          <ul class="arquivos">
            <li><a href="/projudi/processo/arquivo.do?id=doc101&hash=a1b2c3d4">Despacho.pdf</a></li>
          </ul>
        </td>
      </tr>
      <tr class="even">
        <td>1</td>
        <td>10/03/2026 09:15</td>
        <td>Distribuído por sorteio</td>
        <td>
          <ul class="arquivos">
            <li><a href="/projudi/processo/arquivo.do?id=doc102&hash=e5f6g7h8">Peticao_Inicial.pdf</a></li>
          </ul>
        </td>
      </tr>
    </tbody>
  </table>
</div>
</body>
</html>
`

// Fixture 4b: Documentos com data-id explícito.
const htmlProjudiDocumentosComDataIDMock = `
<!DOCTYPE html>
<html lang="pt-BR">
<body>
<div id="content">
  <table class="resultTable" id="tabelaMovimentacoes">
    <thead>
      <tr>
        <th>Seq.</th>
        <th>Data</th>
        <th>Movimentação</th>
        <th>Arquivos</th>
      </tr>
    </thead>
    <tbody>
      <tr class="odd">
        <td>1</td>
        <td>10/03/2026 09:15</td>
        <td>Distribuído por sorteio</td>
        <td>
          <ul class="arquivos">
            <li><a data-id="doc201" href="/projudi/processo/arquivo.do?id=doc201">Peticao_Inicial.pdf</a></li>
          </ul>
        </td>
      </tr>
    </tbody>
  </table>
</div>
</body>
</html>
`

// Fixture 4c: Documento sem ID identificável.
const htmlProjudiDocumentoSemIDMock = `
<!DOCTYPE html>
<html lang="pt-BR">
<body>
<div id="content">
  <table class="resultTable" id="tabelaMovimentacoes">
    <tbody>
      <tr class="odd">
        <td>1</td>
        <td>10/03/2026 09:15</td>
        <td>Distribuído por sorteio</td>
        <td>
          <ul class="arquivos">
            <li><a href="/projudi/outra_acao">Peticao_Inicial.pdf</a></li>
          </ul>
        </td>
      </tr>
    </tbody>
  </table>
</div>
</body>
</html>
`

// Fixture 4d: Tabela de movimentações vazia.
const htmlProjudiEventosVazioMock = `
<!DOCTYPE html>
<html lang="pt-BR">
<body>
<div id="content">
  <table class="resultTable" id="tabelaMovimentacoes">
    <thead><tr><th>Seq.</th><th>Data</th><th>Movimentação</th><th>Arquivos</th></tr></thead>
    <tbody></tbody>
  </table>
</div>
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

func TestProjudiParser_ParseCapaProcesso_BlocosAninhados(t *testing.T) {
	parser := client.NewProjudiParser()

	proc, err := parser.ParseCapaProcesso(htmlProjudiCapaDivsAninhadosMock)

	require.NoError(t, err)
	require.NotNil(t, proc)

	// Deve ler SOMENTE de #tabelaCapa, ignorando #tabelaOutroBloco
	assert.Equal(t, "Procedimento Comum Cível", proc.Classe,
		"deve ignorar a tabela extra e ler apenas #tabelaCapa")
	assert.Equal(t, 45000.0, proc.ValorCausa,
		"deve ignorar o valor da tabela extra")
}

func TestProjudiParser_ParseCapaProcesso_ValorCausaAusente(t *testing.T) {
	parser := client.NewProjudiParser()

	proc, err := parser.ParseCapaProcesso(htmlProjudiCapaSemValorMock)

	require.NoError(t, err)
	require.NotNil(t, proc)
	assert.Equal(t, 0.0, proc.ValorCausa,
		"deve retornar 0 quando a célula de valor da causa não existir")
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
	require.Len(t, docs, 2, "deve extrair 2 documentos da tabela de movimentações")

	assert.Equal(t, "doc101", docs[0].ID, "deve extrair o parâmetro id= do href")
	assert.Equal(t, "Despacho.pdf", docs[0].Nome)
	assert.Equal(t, "Proferido despacho de mero expediente", docs[0].Descricao)

	assert.Equal(t, "doc102", docs[1].ID)
	assert.Equal(t, "Peticao_Inicial.pdf", docs[1].Nome)
	assert.Equal(t, "Distribuído por sorteio", docs[1].Descricao)
}

func TestProjudiParser_ParseDocumentosProcesso_DataID(t *testing.T) {
	parser := client.NewProjudiParser()

	docs, err := parser.ParseDocumentosProcesso(htmlProjudiDocumentosComDataIDMock)

	require.NoError(t, err)
	require.Len(t, docs, 1)
	assert.Equal(t, "doc201", docs[0].ID, "data-id tem precedência sobre o href")
	assert.Equal(t, "Peticao_Inicial.pdf", docs[0].Nome)
}

func TestProjudiParser_ParseDocumentosProcesso_SemID(t *testing.T) {
	parser := client.NewProjudiParser()

	docs, err := parser.ParseDocumentosProcesso(htmlProjudiDocumentoSemIDMock)

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
