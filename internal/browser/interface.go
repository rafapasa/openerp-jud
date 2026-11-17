package browser

// ContextManager gerencia o ciclo de vida do navegador Chromium e cria
// contextos de navegação isolados (cookies, cache e certificados mTLS).
//
// É a única porta de entrada do pacote browser. Toda a dependência de
// github.com/mxschmitt/playwright-go fica encapsulada nas implementações
// concretas deste pacote, jamais vazando para os consumidores.
type ContextManager interface {
	// NewContext cria um novo contexto de navegação isolado, associando os
	// certificados A1 configurados às origens informadas.
	NewContext(origins ...string) (BrowserContext, error)

	// Close encerra a instância do navegador e desaloca os recursos do driver.
	Close() error
}

// BrowserContext representa um contexto de navegação isolado.
// Cada contexto possui cookies, cache e sessão TLS próprios.
type BrowserContext interface {
	// NewPage abre uma nova página (aba) dentro do contexto.
	NewPage() (Page, error)

	// Close encerra o contexto e libera os recursos associados.
	Close() error
}

// Page representa uma página (aba) do navegador.
//
// A interface cresce conforme a necessidade real de cada driver de tribunal.
// Operações de download (ExpectDownload/Download) serão adicionadas quando o
// fluxo de captura de peças for implementado.
type Page interface {
	// Goto navega para a URL informada e aguarda o carregamento.
	Goto(url string) error

	// Content retorna o HTML completo da página renderizada.
	Content() (string, error)

	// Close fecha a página.
	Close() error
}
