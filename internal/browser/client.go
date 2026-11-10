package browser

import (
	"fmt"
	"os"
	"sync"

	"github.com/mxschmitt/playwright-go"
	"github.com/rafapasa/openerp-jud/config"
)

// Manager gerencia o ciclo de vida do Playwright e do processo Chromium.
type Manager struct {
	cfg     *config.Config
	pw      *playwright.Playwright
	browser playwright.Browser
	mu      sync.Mutex
}

var _ ContextManager = (*Manager)(nil)

// NewManager inicializa o Playwright e sobe o navegador Chromium com flags de alta performance e stealth.
func NewManager(cfg *config.Config) (*Manager, error) {
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

	return &Manager{
		cfg:     cfg,
		pw:      pw,
		browser: browser,
	}, nil
}

// NewContext cria um BrowserContext isolado configurado com Stealth, headers e certificados A1.
func (m *Manager) NewContext(origins ...string) (playwright.BrowserContext, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	opts := playwright.BrowserNewContextOptions{
		UserAgent: playwright.String(DefaultUserAgent),
		Viewport: &playwright.Size{
			Width:  1366,
			Height: 768,
		},
		Locale:     playwright.String("pt-BR"),
		TimezoneId: playwright.String("America/Sao_Paulo"),
		AcceptDownloads: playwright.Bool(true),
		IgnoreHttpsErrors: playwright.Bool(true),
		ExtraHttpHeaders: map[string]string{
			"Accept-Language": "pt-BR,pt;q=0.9,en-US;q=0.8,en;q=0.7",
			"sec-ch-ua":       `"Chromium";v="128", "Not;A=Brand";v="24", "Google Chrome";v="128"`,
			"sec-ch-ua-mobile": "?0",
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

	ctx, err := m.browser.NewContext(opts)
	if err != nil {
		return nil, fmt.Errorf("falha ao criar browser context: %w", err)
	}

	// Injetar script de evasão em todas as páginas abertas neste contexto
	if err := ctx.AddInitScript(playwright.Script{
		Content: playwright.String(StealthInitScript),
	}); err != nil {
		ctx.Close()
		return nil, fmt.Errorf("falha ao injetar script stealth: %w", err)
	}

	return ctx, nil
}

// Close encerra a instância do navegador e do driver.
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var firstErr error
	if m.browser != nil {
		if err := m.browser.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	if m.pw != nil {
		if err := m.pw.Stop(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}
