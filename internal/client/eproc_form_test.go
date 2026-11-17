package client_test

import (
	"testing"

	"github.com/rafapasa/openerp-jud/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================
// FIXTURES — Formulário de consulta do eProc TJSC
//
// ESTE É O ÚNICO FIXTURE 100% REAL DO PROJETO.
//
// Fonte: testdata/eproc/public/form_consulta_publica_fragment.html
// Capturado em: 2026-11-13
// URL: https://eprocwebcon.tjsc.jus.br/consulta1g/externo_controlador.php?acao=processo_consulta_publica
// Charset original: iso-8859-1
//
// Observação: o POST exige Cloudflare Turnstile (sitekey 0x4AAAAAAA9zgl8AcFbokEy4).
// ============================================================

const htmlFormConsultaEprocReal = `
<form
        id="frmProcessoLista"
        method="post"
        onsubmit="return OnSubmitForm();"
        action="externo_controlador.php?acao=processo_consulta_publica&acao_origem=principal&acao_retorno=processo_consulta_publica"
    >
    <input type="hidden" id="hdnInfraTipoPagina" name="hdnInfraTipoPagina" value="1"/>
<div id="divInfraBarraComandosSuperior" class="infraBarraComandos">
<button type="submit" accesskey="C" name="sbmNovo" id="sbmNovo" value="Consultar" class="infraButton btn btn-primary"><span class="infraTeclaAtalho">C</span>onsultar</button>
</div>
<div id="divInfraAreaDados" class="infraAreaDados" style="height:auto;">

        <fieldset class="infraFieldset">
            <legend>Critérios da consulta processual</legend>
            <div class="row align-items-end">
                <div class="col-12 col-md-4 form-group">
                    <label id="lblNumProcesso" for="txtNumProcesso" accesskey="s" class="infraLabelObrigatorio">
                        Nº Proce<span class="infraTeclaAtalho">s</span>so:
                    </label>
                    <input type="text" id="txtNumProcesso" name="txtNumProcesso" class="infraText form-control"
                        onkeypress="return infraMascara(this,event,'#######-##.####.#.##.####');"
                        value="" maxlength="25" />
                </div>
            </div>

            <div class="row align-items-end">
                <div class="col-12 col-md-7">
                    <div class="row align-items-center">
                        <div class="col form-group mb-0 mb-md-3">
                            <label id="lblNumChave" for="txtNumChave" accesskey="C" class="infraLabelObrigatorio">
                                <span class="infraTeclaAtalho">C</span>have Processo:
                            </label>
                            <input type="text" id="txtNumChave" name="txtNumChave" class="infraText form-control" value="" />
                        </div>
                    </div>
                </div>
            </div>

            <div class="row align-items-end">
                <div class="col-12 col-md-4 form-group">
                    <label id="lblNumChaveDocumento" for="txtNumChaveDocumento" class="infraLabelObrigatorio">Chave Documento:</label>
                    <input type="text" id="txtNumChaveDocumento" name="txtNumChaveDocumento" class="infraText form-control" value="" />
                </div>
            </div>

            <div class="row align-items-end">
                <div class="col-12 col-md-6 form-group">
                    <label id="lblStrParte" for="txtStrParte" class="infraLabelObrigatorio">Nome da Parte:</label>
                    <input type="text" id="txtStrParte" name="txtStrParte" class="infraText form-control" value="" />
                    <input type="hidden" name="chkFonetica" value="N" />
                    <input type="checkbox" id="chkFonetica" name="chkFonetica" value="S" />
                    <label for="chkFonetica">Pesquisa fonética</label>
                </div>
            </div>

            <div class="row align-items-end">
                <div class="col-12 col-md-4 form-group">
                    <input type="radio" id="rdoPessoaFisica" name="rdoTipo" value="CPF" checked="checked" />
                    <label for="rdoPessoaFisica">CPF</label>
                    <input type="radio" id="rdoPessoaJuridica" name="rdoTipo" value="CNPJ" />
                    <label for="rdoPessoaJuridica">CNPJ</label>
                    <label id="lblCpfCnpj" for="txtCpfCnpj" class="infraLabelObrigatorio">CPF/CNPJ:</label>
                    <input type="text" id="txtCpfCnpj" name="txtCpfCnpj" class="infraText form-control" value="" />
                </div>
            </div>

            <div class="row align-items-end">
                <div class="col-12 col-md-4 form-group">
                    <label id="lblStrOAB" for="txtStrOAB" accesskey="O" class="infraLabelObrigatorio">
                        <span class="infraTeclaAtalho">O</span>AB:
                    </label>
                    <input type="text" id="txtStrOAB" name="txtStrOAB" class="infraText form-control" value="" />
                </div>
            </div>
        </fieldset>

<div id="divInfraCaptcha" class="infraAreaDados">
  <!-- Cloudflare Turnstile — obrigatório para POST de consulta -->
  <div class="cf-turnstile" data-sitekey="0x4AAAAAAA9zgl8AcFbokEy4" data-callback="callbackCloudflare"></div>
  <input type="hidden" id="hdnInfraCaptcha" name="hdnInfraCaptcha" value="0" />
</div>
<script src="https://challenges.cloudflare.com/turnstile/v0/api.js" async defer></script>
</div>
</form>
`

// Fixture: formulário sem captcha (para validar o caminho sem Turnstile).
const htmlFormConsultaSemCaptcha = `
<form id="frmProcessoLista" method="post" action="externo_controlador.php?acao=processo_consulta_publica">
    <div id="divInfraAreaDados" class="infraAreaDados">
        <fieldset class="infraFieldset">
            <input type="text" id="txtNumProcesso" name="txtNumProcesso" class="infraText form-control" value="" />
            <input type="text" id="txtStrParte" name="txtStrParte" class="infraText form-control" value="" />
            <input type="radio" id="rdoPessoaFisica" name="rdoTipo" value="CPF" checked="checked" />
            <input type="radio" id="rdoPessoaJuridica" name="rdoTipo" value="CNPJ" />
            <input type="text" id="txtCpfCnpj" name="txtCpfCnpj" class="infraText form-control" value="" />
            <input type="text" id="txtStrOAB" name="txtStrOAB" class="infraText form-control" value="" />
        </fieldset>
        <button type="submit" name="sbmNovo" id="sbmNovo" value="Consultar" class="infraButton">Consultar</button>
    </div>
</form>
`

// Fixture: HTML sem formulário de consulta.
const htmlFormConsultaAusente = `
<!DOCTYPE html>
<html lang="pt-br">
<body>
<div id="divInfraAreaTelaD">
    <p>Nenhum formulário nesta página.</p>
</div>
</body>
</html>
`

// ============================================================
// TESTES — ParseFormularioConsulta (eProc)
// ============================================================

func TestEProcParser_ParseFormularioConsulta_Real(t *testing.T) {
	parser := client.NewEProcParser()

	form, err := parser.ParseFormularioConsulta(htmlFormConsultaEprocReal)

	require.NoError(t, err)
	require.NotNil(t, form, "deve extrair o formulário de consulta do HTML real")

	// Metadados do <form>
	assert.Equal(t, "post", form.Method)
	assert.Contains(t, form.Action, "acao=processo_consulta_publica")

	// Seletores dos campos — todos confirmados no HTML real
	assert.Equal(t, "#txtNumProcesso", form.CampoNumeroProcesso)
	assert.Equal(t, "#txtStrParte", form.CampoNomeParte)
	assert.Equal(t, "#txtCpfCnpj", form.CampoDocumento)
	assert.Equal(t, "#txtStrOAB", form.CampoOAB)

	// Radios de tipo de documento
	assert.Equal(t, "#rdoPessoaFisica", form.RadioPessoaFisica)
	assert.Equal(t, "#rdoPessoaJuridica", form.RadioPessoaJuridica)

	// Botão de submissão
	assert.Equal(t, "#sbmNovo", form.BotaoSubmit)

	// Captcha — o formulário real exige Cloudflare Turnstile
	assert.True(t, form.PossuiCaptcha, "o formulário real do eProc exige captcha")
	assert.Equal(t, "0x4AAAAAAA9zgl8AcFbokEy4", form.CaptchaSiteKey)
}

func TestEProcParser_ParseFormularioConsulta_SemCaptcha(t *testing.T) {
	parser := client.NewEProcParser()

	form, err := parser.ParseFormularioConsulta(htmlFormConsultaSemCaptcha)

	require.NoError(t, err)
	require.NotNil(t, form)

	assert.False(t, form.PossuiCaptcha, "não deve detectar captcha quando ausente")
	assert.Empty(t, form.CaptchaSiteKey)

	assert.Equal(t, "#txtNumProcesso", form.CampoNumeroProcesso)
	assert.Equal(t, "#txtCpfCnpj", form.CampoDocumento)
	assert.Equal(t, "#sbmNovo", form.BotaoSubmit)
}

func TestEProcParser_ParseFormularioConsulta_Ausente(t *testing.T) {
	parser := client.NewEProcParser()

	form, err := parser.ParseFormularioConsulta(htmlFormConsultaAusente)

	require.NoError(t, err)
	assert.Nil(t, form, "deve retornar nil quando não houver formulário de consulta")
}

func TestEProcParser_ParseFormularioConsulta_HTMLInvalido(t *testing.T) {
	parser := client.NewEProcParser()

	assert.NotPanics(t, func() {
		_, _ = parser.ParseFormularioConsulta(htmlInvalidoMock)
	}, "parser não deve entrar em panic com HTML malformado")
}
