package client

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/rafapasa/openerp-jud/internal/observability/logger"
	"go.uber.org/zap"
)

func GetHTMLDoc(html string) (*goquery.Document, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		logger.LogError("Erro parse documento html", zap.Any("ERROR", err))
		return nil, err
	}
	return doc, nil
}
