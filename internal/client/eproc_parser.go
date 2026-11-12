package client

import (
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/etoolstec/gokit/apperror"
	"github.com/etoolstec/gokit/converter"
	"github.com/rafapasa/openerp-jud/internal/model"
	"github.com/rafapasa/openerp-jud/internal/observability/logger"
	"go.uber.org/zap"
)

// EProcParser implementa a interface HTMLParser para o eProc TJSC.
type EProcParser struct {
}

var _ HTMLParser = (*EProcParser)(nil)

// NewEProcParser instancia um novo parser para o eProc TJSC.
func NewEProcParser() *EProcParser {
	return &EProcParser{}
}

// ParseTabelaProcessos extrai os processos a partir da tabela de resultados de consulta pública/CPF do eProc.
// (Stub vazio para o ciclo TDD: implemente a lógica para fazer o teste passar!)
func (p *EProcParser) ParseTabelaProcessos(html string) ([]model.Processo, error) {
	doc, err := GetHTMLDoc(html)
	if err != nil {
		return nil, apperror.NewInternalError("Erro ao parsear documento html", err)
	}

	processos := make([]model.Processo, 0)

	doc.Find("table#conteudoTabela tbody tr").Each(func(i int, s *goquery.Selection) {
		tds := s.Find("td")
		if tds.Length() < 4 {
			return
		}
		var processo model.Processo
		processo.Tribunal = model.TribunalTJSC
		processo.NumeroCNJ = strings.TrimSpace(tds.Eq(0).Find("a").Text())
		if processo.NumeroCNJ == "" {
			processo.NumeroCNJ = strings.TrimSpace(tds.Eq(0).Text())
		}
		processo.Classe = strings.TrimSpace(tds.Eq(1).Text())
		processo.Vara = strings.TrimSpace(tds.Eq(2).Text())
		processo.Situacao = strings.TrimSpace(tds.Eq(3).Text())

		processos = append(processos, processo)

	})
	return processos, nil
}

// ParseCapaProcesso extrai os metadados principais da capa de um processo no eProc.
// (Stub vazio para o ciclo TDD)
func (p *EProcParser) ParseCapaProcesso(html string) (*model.Processo, error) {
	doc, err := GetHTMLDoc(html)
	if err != nil {
		return nil, apperror.NewInternalError("Erro ao parsear documento html", nil)
	}
	var processo model.Processo
	processo.Tribunal = model.TribunalTJSC
	divConteudo := doc.Find("div#divConteudo")
	processo.NumeroCNJ = strings.TrimSpace(divConteudo.Find("span#txtNumProcesso").Text())
	divDados := divConteudo.Find("div.infraAreaDados")
	processo.Classe = strings.TrimSpace(divDados.Find("span#txtClasse").Text())
	processo.Comarca = strings.TrimSpace(divDados.Find("span#txtComarca").Text())
	processo.Vara = strings.TrimSpace(divDados.Find("span#txtOrgaoJulgador").Text())
	processo.Assunto = strings.TrimSpace(divDados.Find("span#txtAssunto").Text())
	strValue := strings.TrimSpace(divDados.Find("span#txtValorCausa").Text())
	v, err := converter.ParseFloat(strValue)
	if err != nil {
		logger.LogWarn("Falha ParseFloat",
			zap.String("Valor", strValue),
			zap.Any("ERROR", err),
		)
		v = 0
	}
	processo.ValorCausa = v
	processo.Situacao = strings.TrimSpace(divDados.Find("span#txtSituacao").Text())

	divConteudo.Find("table#tblPartes tbody tr").Each(func(i int, s *goquery.Selection) {
		tds := s.Find("td")
		if tds.Length() < 2 {
			return
		}
		var parte model.ParteProcesso
		parte.Polo = strings.TrimSpace(tds.Eq(0).Text())
		parte.Nome = strings.TrimSpace(tds.Eq(1).Text())

		processo.Partes = append(processo.Partes, parte)
	})

	return &processo, nil
}

// ParseDocumentosProcesso extrai a lista de eventos e peças disponíveis na árvore do processo.
// (Stub vazio para o ciclo TDD)
func (p *EProcParser) ParseDocumentosProcesso(html string) ([]model.Documento, error) {
	// TODO: Seu código entra aqui!
	doc, err := GetHTMLDoc(html)
	if err != nil {
		return nil, apperror.NewInternalError("Erro ao parsear documento html", nil)
	}

	documentos := make([]model.Documento, 0)
	doc.Find("table#tblEventos tbody tr").Each(func(i int, s *goquery.Selection) {
		tds := s.Find("td")
		if tds.Length() < 3 {
			return
		}
		var documento model.Documento

		id := tds.Eq(2).Find("a").AttrOr("data-id", "")
		if id == "" {
			href, _ := tds.Eq(2).Find("a").Attr("href")
			u, err := url.Parse(href)
			if err != nil {
				logger.LogError("Falha ao parsear url", zap.String("URL", href), zap.Any("ERROR", err))
				return
			}
			id = u.Query().Get("id_documento")
			if id == "" {
				logger.LogError("Falha ao localizar ID", zap.String("URL", href), zap.Any("ERROR", err))
				return
			}
		}
		documento.ID = id
		documento.Descricao = strings.TrimSpace(tds.Eq(1).Text())
		documento.Nome = strings.TrimSpace(tds.Eq(2).Find("a").Text())

		documentos = append(documentos, documento)

	})
	return documentos, nil
}
