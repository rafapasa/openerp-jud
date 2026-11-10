package browser

import (
	"github.com/mxschmitt/playwright-go"
)

// ContextManager define o contrato para criação e gerenciamento do ciclo de vida
// de instâncias isoladas do navegador Chromium com suporte a certificados A1 e stealth.
type ContextManager interface {
	// NewContext cria um novo contexto de navegação isolado com cookies e certificados mTLS.
	NewContext(origins ...string) (playwright.BrowserContext, error)

	// Close encerra a instância do navegador e desaloca recursos.
	Close() error
}

