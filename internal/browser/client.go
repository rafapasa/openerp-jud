package browser

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/mxschmitt/playwright-go"
	"github.com/rafapasa/openerp-jud/config"
)

// contextManager é a implementação concreta de ContextManager.
// Encapsula toda a dependência de playwright-go neste pacote.
type contextManager struct {
	cfg            *config.Config
	pw             *playwright.Playwright
	browser        playwright.Browser
	defaultTimeout time.Duration
	mu             sync.Mutex
}

var _ ContextManager = (*contextManager)(nil)

// NewContextManager inicializa o Playwright e sobe o navegador Chromium
// com flags de alta performance e stealth.
func NewContextManager(cfg *config.Config) (ContextManager, error) {
	pw, err := playwright.Run()
	if err != nil {
		return nil, fmt.Errorf("falha ao iniciar playwright: %w", err)
	}

	// Argumentos do Chromium para Oracle Linux ARM64, modo headless e anti-detecção
	args := []string{
		"--no-sandbox",
		"--disable-setuid-sandbox",
		"--disable-dev-shm-usage",
		"--disable-blink-features=AutomationControlled",
		"--disable-infobars",
		"--window-size=1366,768",
		"--no-first-run",
		"--no-default-browser-check",
	}

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(cfg.Headless),
		Args:     args,
	})
	if err != nil {
		pw.Stop()
		return nil, fmt.Errorf("falha ao lançar chromium: %w", err)
	}

	return &contextManager{
		cfg:            cfg,
		pw:             pw,
		browser:        browser,
		defaultTimeout: cfg.BrowserTimeout,
	}, nil
}

// NewContext cria um BrowserContext isolado configurado com Stealth, headers e certificados A1.
func (m *contextManager) NewContext(origins ...string) (BrowserContext, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	opts := playwright.BrowserNewContextOptions{
		UserAgent: playwright.String(DefaultUserAgent),
		Viewport: &playwright.Size{
			Width:  1366,
			Height: 768,
		},
		Locale:            playwright.String("pt-BR"),
		TimezoneId:        playwright.String("America/Sao_Paulo"),
		AcceptDownloads:   playwright.Bool(true),
		IgnoreHttpsErrors: playwright.Bool(true),
		ExtraHttpHeaders: map[string]string{
			"Accept-Language":    "pt-BR,pt;q=0.9,en-US;q=0.8,en;q=0.7",
			"sec-ch-ua":          `"Chromium";v="128", "Not;A=Brand";v="24", "Google Chrome";v="128"`,
			"sec-ch-ua-mobile":   "?0",
			"sec-ch-ua-platform": `"Linux"`,
		},
	}

	// Injetar certificado ICP-Brasil A1 caso o arquivo exista
	if m.cfg.CertPath != "" {
		if _, err := os.Stat(m.cfg.CertPath); err == nil {
			var clientCerts []playwright.ClientCertificate
			for _, origin := range origins {
				cert := playwright.ClientCertificate{
					Origin:  origin,
					PfxPath: &m.cfg.CertPath,
				}
				if m.cfg.CertPassword != "" {
					cert.Passphrase = &m.cfg.CertPassword
				}
				clientCerts = append(clientCerts, cert)
			}
			if len(clientCerts) > 0 {
				opts.ClientCertificates = clientCerts
			}
		}
	}

	pwCtx, err := m.browser.NewContext(opts)
	if err != nil {
		return nil, fmt.Errorf("falha ao criar browser context: %w", err)
	}

	// Injetar script de evasão em todas as páginas abertas neste contexto
	if err := pwCtx.AddInitScript(playwright.Script{
		Content: playwright.String(StealthInitScript),
	}); err != nil {
		pwCtx.Close()
		return nil, fmt.Errorf("falha ao injetar script stealth: %w", err)
	}

	return &browserContext{
		ctx:            pwCtx,
		defaultTimeout: m.defaultTimeout,
	}, nil
}

// Close encerra a instância do navegador e do driver.
func (m *contextManager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var firstErr error
	if m.browser != nil {
		if err := m.browser.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
		m.browser = nil
	}
	if m.pw != nil {
		if err := m.pw.Stop(); err != nil && firstErr == nil {
			firstErr = err
		}
		m.pw = nil
	}
	return firstErr
}

// browserContext é a implementação concreta de BrowserContext.
type browserContext struct {
	ctx            playwright.BrowserContext
	defaultTimeout time.Duration
}

var _ BrowserContext = (*browserContext)(nil)

// NewPage abre uma nova página dentro do contexto e aplica o timeout padrão configurado.
func (c *browserContext) NewPage() (Page, error) {
	pwPage, err := c.ctx.NewPage()
	if err != nil {
		return nil, fmt.Errorf("falha ao criar página: %w", err)
	}

	// Aplica BrowserTimeout (config) a todas as operações da página (goto, click, wait, etc.).
	if c.defaultTimeout > 0 {
		ms := float64(c.defaultTimeout / time.Millisecond)
		pwPage.SetDefaultTimeout(ms)
		pwPage.SetDefaultNavigationTimeout(ms)
	}

	return &page{p: pwPage}, nil
}

// Close encerra o contexto de navegação.
func (c *browserContext) Close() error {
	return c.ctx.Close()
}

// page é a implementação concreta de Page.
type page struct {
	p playwright.Page
}

var _ Page = (*page)(nil)

// Goto navega para a URL informada e aguarda o carregamento.
func (p *page) Goto(url string) error {
	if _, err := p.p.Goto(url); err != nil {
		return fmt.Errorf("falha ao navegar para %s: %w", url, err)
	}
	return nil
}

// Content retorna o HTML completo da página renderizada.
func (p *page) Content() (string, error) {
	html, err := p.p.Content()
	if err != nil {
		return "", fmt.Errorf("falha ao obter conteúdo da página: %w", err)
	}
	return html, nil
}

// Close fecha a página.
func (p *page) Close() error {
	return p.p.Close()
}
