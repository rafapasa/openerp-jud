package client

import (
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/etoolstec/gokit/converter"
	"github.com/rafapasa/openerp-jud/internal/model"
	"github.com/rafapasa/openerp-jud/internal/observability/logger"
	"go.uber.org/zap"
)

type ProJudiParser struct {
}

var _ HTMLParser = (*ProJudiParser)(nil)

// NewEProcParser instancia um novo parser para o eProc TJSC.
func NewProjudiParser() *ProJudiParser {
	return &ProJudiParser{}
}
func (p *ProJudiParser) ParseProjudiDocumentos(body []byte, urlBase string) ([]model.Documento, error) {
	documentos := []model.Documento{}
	selection, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		logger.LogError("Erro ao parsear HTML", zap.Error(err))
		return nil, err
	}
	selection.Find("table#documentos_table > tbody > tr").Each(func(i int, tr *goquery.Selection) {
		documento := model.Documento{}
		tds := tr.Find("td")
		if tds.Length() < 3 {
			logger.LogError("Erro ao parsear tabela de documentos", zap.Any("HTML", tr.Text()))
			return
		}
		tds.Eq(2).Find("a").Each(func(j int, a *goquery.Selection) {
			href, _ := a.Attr("href")
			u, _ := url.Parse(href)
			id := u.Query().Get("id_documento")
			if id == "" {
				logger.LogError("Falha ao localizar ID", zap.String("URL", href), zap.Any("ERROR", err))
			}
			documento.ID = id
			documento.Descricao = strings.TrimSpace(tds.Eq(1).Text())
			documento.Nome = strings.TrimSpace(a.Text())
			documentos = append(documentos, documento)
		})
	})
	return documentos, nil
}

func (p *ProJudiParser) ParseProjudiProcessos(body []byte, urlBase string) ([]model.Processo, error) {

	return make([]model.Processo, 0), nil
}

func (p *ProJudiParser) ParseCapaProcesso(html string) (*model.Processo, error) {
	doc, err := GetHTMLDoc(html)
	if err != nil {
		logger.LogError("Erro ao parsear HTML", zap.Error(err))
		return nil, err
	}
	processo := model.Processo{Tribunal: model.TribunalTJPR}
	divConteudo := doc.Find("div#conteudoProcesso").First()
	processo.NumeroCNJ = strings.TrimSpace(divConteudo.Find("h2#numeroProcesso").First().Text())
	processo.Classe = strings.TrimSpace(divConteudo.Find("div.form-group span#classeProcesso").First().Text())
	processo.Comarca = strings.TrimSpace(divConteudo.Find("div.form-group span#comarcaProcesso").First().Text())
	processo.Vara = strings.TrimSpace(divConteudo.Find("div.form-group span#varaProcesso").First().Text())
	processo.Assunto = strings.TrimSpace(divConteudo.Find("div.form-group span#assuntoProcesso").First().Text())
	processo.Situacao = strings.TrimSpace(divConteudo.Find("div.form-group span#situacaoProcesso").First().Text())
	v, err := converter.ParseFloat(strings.TrimSpace(divConteudo.Find("div.form-group span#valorCausaProcesso").First().Text()))
	if err != nil {
		logger.LogWarn("Erro ao parsear valor da causa", zap.String("HTML", html), zap.Any("ERROR", err))
		v = 0.0
	}
	processo.ValorCausa = v

	divConteudo.Find("table#tabelaPartes > tbody > tr").Each(func(i int, s *goquery.Selection) {

		tds := s.Find("td")
		if tds.Length() < 2 {
			logger.LogError("Erro ao parsear tabela de partes", zap.String("HTML", s.Text()), zap.Any("ERROR", err))
			return
		}
		var partes model.ParteProcesso
		partes.Polo = strings.TrimSpace(tds.Eq(0).Text())
		partes.Nome = strings.TrimSpace(tds.Eq(1).Text())
		processo.Partes = append(processo.Partes, partes)
	})

	return &processo, nil
}
func (p *ProJudiParser) ParseDocumentosProcesso(html string) ([]model.Documento, error) {
	doc, err := GetHTMLDoc(html)
	if err != nil {
		logger.LogError("Erro ao parsear HTML", zap.Error(err))
		return nil, err
	}
	documentos := make([]model.Documento, 0)
	doc.Find("table#tabelaEventos > tbody > tr").Each(func(i int, s *goquery.Selection) {
		tds := s.Find("td")
		if tds.Length() < 3 {
			logger.LogError("Erro ao parsear tabela de documentos", zap.Any("HTML", s.Text()))
			return
		}

		id := tds.Eq(2).Find("a").AttrOr("data-id", "")
		if id == "" {
			href, _ := tds.Eq(2).Find("a").Attr("href")
			u, err := url.Parse(href)
			if err != nil {
				logger.LogError("Falha ao parsear url", zap.String("URL", href), zap.Any("ERROR", err))
				return
			}
			id = u.Query().Get("id")
			if id == "" {
				logger.LogError("Falha ao localizar ID", zap.String("URL", href), zap.Any("ERROR", err))
				return
			}
		}
		var documento model.Documento
		documento.ID = id
		documento.Descricao = strings.TrimSpace(tds.Eq(1).Text())
		documento.Nome = strings.TrimSpace(tds.Eq(2).Text())
		documentos = append(documentos, documento)
	})
	return documentos, nil
}

func (p *ProJudiParser) ParseTabelaProcessos(html string) ([]model.Processo, error) {
	processos := []model.Processo{}
	selection, err := goquery.NewDocumentFromReader(strings.NewReader(string(html)))
	if err != nil {
		logger.LogError("Erro ao parsear HTML", zap.Error(err))
		return nil, err
	}
	selection.Find("table#tabelaProcessos > tbody > tr").Each(func(i int, tr *goquery.Selection) {
		processo := model.Processo{}
		processo.Tribunal = model.TribunalTJPR
		tds := tr.Find("td")
		if tds.Length() < 4 {
			logger.LogError("Erro ao parsear tabela de processos",
				zap.Int("tds", tds.Length()),
				zap.String("html", tr.Text()),
			)
			return
		}
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
