package config

import (
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// Config armazena as configurações de execução do MCP Server e Playwright.
type Config struct {
	// Modo de transporte MCP: "stdio" ou "sse"
	Transport string
	// Porta para servidor HTTP/SSE (quando Transport == "sse")
	Port string

	// Certificado Digital ICP-Brasil A1 (.pfx / .p12)
	CertPath     string
	CertPassword string

	// Configurações do Navegador Headless
	Headless       bool
	BrowserTimeout time.Duration

	// Diretório para armazenamento temporário de downloads de peças
	DownloadsDir string
}

// LoadConfig carrega as configurações a partir de variáveis de ambiente com defaults seguros.
func LoadConfig() *Config {
	transport := getEnv("MCP_TRANSPORT", "stdio")
	port := getEnv("PORT", "8080")
	certPath := getEnv("CERT_PATH", filepath.Join("certs", "advogado.pfx"))
	certPassword := getEnv("CERT_PASSWORD", "")
	downloadsDir := getEnv("DOWNLOADS_DIR", filepath.Join("tmp", "downloads"))

	headless := true
	if v := os.Getenv("BROWSER_HEADLESS"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			headless = b
		}
	}

	timeoutSec := 45
	if v := os.Getenv("BROWSER_TIMEOUT_SEC"); v != "" {
		if s, err := strconv.Atoi(v); err == nil && s > 0 {
			timeoutSec = s
		}
	}

	return &Config{
		Transport:      transport,
		Port:           port,
		CertPath:       certPath,
		CertPassword:   certPassword,
		Headless:       headless,
		BrowserTimeout: time.Duration(timeoutSec) * time.Second,
		DownloadsDir:   downloadsDir,
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
