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

// ProJudiParser implementa a interface HTMLParser para o Projudi TJPR.
type ProJudiParser struct {
}

var _ HTMLParser = (*ProJudiParser)(nil)

// NewProjudiParser instancia um novo parser para o Projudi TJPR.
func NewProjudiParser() *ProJudiParser {
	return &ProJudiParser{}
}

// ParseTabelaProcessos extrai os processos a partir da tabela de resultados de busca por CPF do Projudi.
func (p *ProJudiParser) ParseTabelaProcessos(html string) ([]model.Processo, error) {
	doc, err := GetHTMLDoc(html)
	if err != nil {
		return nil, apperror.NewInternalError("Erro ao parsear documento html", err)
	}

	processos := make([]model.Processo, 0)

	doc.Find("table#tabelaProcessos > tbody > tr").Each(func(i int, tr *goquery.Selection) {
		tds := tr.Find("td")
		if tds.Length() < 4 {
			logger.LogWarn("Linha da tabela de processos ignorada por falta de colunas",
				zap.Int("tds", tds.Length()),
				zap.String("linha", tr.Text()),
			)
			return
		}

		var processo model.Processo
		processo.Tribunal = model.TribunalTJPR
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

// ParseCapaProcesso extrai os metadados principais da capa de um processo no Projudi.
func (p *ProJudiParser) ParseCapaProcesso(html string) (*model.Processo, error) {
	doc, err := GetHTMLDoc(html)
	if err != nil {
		return nil, apperror.NewInternalError("Erro ao parsear documento html", err)
	}

	processo := model.Processo{Tribunal: model.TribunalTJPR}

	divConteudo := doc.Find("div#conteudoProcesso").First()
	processo.NumeroCNJ = strings.TrimSpace(divConteudo.Find("h2#numeroProcesso").First().Text())
	processo.Classe = strings.TrimSpace(divConteudo.Find("div.form-group span#classeProcesso").First().Text())
	processo.Comarca = strings.TrimSpace(divConteudo.Find("div.form-group span#comarcaProcesso").First().Text())
	processo.Vara = strings.TrimSpace(divConteudo.Find("div.form-group span#varaProcesso").First().Text())
	processo.Assunto = strings.TrimSpace(divConteudo.Find("div.form-group span#assuntoProcesso").First().Text())
	processo.Situacao = strings.TrimSpace(divConteudo.Find("div.form-group span#situacaoProcesso").First().Text())

	strValorCausa := strings.TrimSpace(divConteudo.Find("div.form-group span#valorCausaProcesso").First().Text())
	valorCausa, err := converter.ParseFloat(strValorCausa)
	if err != nil {
		logger.LogWarn("Falha ao converter valor da causa",
			zap.String("valor", strValorCausa),
			zap.Any("ERROR", err),
		)
		valorCausa = 0
	}
	processo.ValorCausa = valorCausa

	divConteudo.Find("table#tabelaPartes > tbody > tr").Each(func(i int, s *goquery.Selection) {
		tds := s.Find("td")
		if tds.Length() < 2 {
			logger.LogWarn("Linha da tabela de partes ignorada por falta de colunas",
				zap.Int("tds", tds.Length()),
				zap.String("linha", s.Text()),
			)
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
func (p *ProJudiParser) ParseDocumentosProcesso(html string) ([]model.Documento, error) {
	doc, err := GetHTMLDoc(html)
	if err != nil {
		return nil, apperror.NewInternalError("Erro ao parsear documento html", err)
	}

	documentos := make([]model.Documento, 0)

	doc.Find("table#tabelaEventos > tbody > tr").Each(func(i int, s *goquery.Selection) {
		tds := s.Find("td")
		if tds.Length() < 3 {
			logger.LogWarn("Linha da tabela de eventos ignorada por falta de colunas",
				zap.Int("tds", tds.Length()),
				zap.String("linha", s.Text()),
			)
			return
		}

		descricao := strings.TrimSpace(tds.Eq(1).Text())

		tds.Eq(2).Find("a").Each(func(j int, a *goquery.Selection) {
			id := strings.TrimSpace(a.AttrOr("data-id", ""))
			if id == "" {
				href, _ := a.Attr("href")
				u, err := url.Parse(href)
				if err != nil {
					logger.LogError("Falha ao parsear url do documento",
						zap.String("URL", href),
						zap.Any("ERROR", err),
					)
					return
				}
				id = strings.TrimSpace(u.Query().Get("id"))
			}

			if id == "" {
				logger.LogWarn("Documento descartado por não possuir ID identificável",
					zap.String("descricao", descricao),
					zap.String("nome", strings.TrimSpace(a.Text())),
				)
				return
			}

			var documento model.Documento
			documento.ID = id
			documento.Descricao = descricao
			documento.Nome = strings.TrimSpace(a.Text())

			documentos = append(documentos, documento)
		})
	})

	return documentos, nil
}
