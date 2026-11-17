package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Config armazena as configurações de execução do MCP Server e Playwright.
type Config struct {
	Environment string `mapstructure:"ENVIRONMENT"`
	TimeZone    string `mapstructure:"TIME_ZONE"`

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

	// ============================================
	// LOGGING
	// ============================================
	LogLevel    string `mapstructure:"LOG_LEVEL"`
	LogEncoding string `mapstructure:"LOG_ENCODING"`
	LogOutput   string `mapstructure:"LOG_OUTPUT"`
	LogFile     string `mapstructure:"LOG_FILE"`
}

// LoadConfig carrega as configurações a partir de variáveis de ambiente com defaults seguros.
func LoadConfig() (*Config, error) {
	transport := getEnv("MCP_TRANSPORT", "stdio")
	if transport != "stdio" && transport != "sse" {
		return nil, fmt.Errorf("MCP_TRANSPORT inválido: %q (esperado 'stdio' ou 'sse')", transport)
	}

	port := getEnv("PORT", "8080")
	if transport == "sse" {
		if _, err := strconv.Atoi(port); err != nil {
			return nil, fmt.Errorf("PORT inválido: %q", port)
		}
	}

	certPath := getEnv("CERT_PATH", filepath.Join("certs", "advogado.pfx"))
	certPassword := getEnv("CERT_PASSWORD", "")
	downloadsDir := getEnv("DOWNLOADS_DIR", filepath.Join("tmp", "downloads"))

	headless := true
	if v := os.Getenv("BROWSER_HEADLESS"); v != "" {
		b, err := strconv.ParseBool(v)
		if err != nil {
			return nil, fmt.Errorf("BROWSER_HEADLESS inválido: %q", v)
		}
		headless = b
	}

	timeoutSec := 45
	if v := os.Getenv("BROWSER_TIMEOUT_SEC"); v != "" {
		s, err := strconv.Atoi(v)
		if err != nil || s <= 0 {
			return nil, fmt.Errorf("BROWSER_TIMEOUT_SEC inválido: %q", v)
		}
		timeoutSec = s
	}

	return &Config{
		Environment:    getEnv("ENVIRONMENT", "development"),
		TimeZone:       getEnv("TIME_ZONE", "America/Sao_Paulo"),
		Transport:      transport,
		Port:           port,
		CertPath:       certPath,
		CertPassword:   certPassword,
		Headless:       headless,
		BrowserTimeout: time.Duration(timeoutSec) * time.Second,
		DownloadsDir:   downloadsDir,
		LogLevel:       getEnv("LOG_LEVEL", "info"),
		LogEncoding:    getEnv("LOG_ENCODING", "json"),
		LogOutput:      getEnv("LOG_OUTPUT", "stdout"),
		LogFile:        getEnv("LOG_FILE", filepath.Join("logs", "mcp-server.log")),
	}, nil
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

// IsProduction indica se o ambiente é de produção.
func (c *Config) IsProduction() bool {
	env := strings.ToLower(strings.TrimSpace(c.Environment))
	if env == "" {
		env = strings.ToLower(strings.TrimSpace(os.Getenv("ENVIRONMENT")))
	}
	return env == "production" || env == "prod" || env == "prd"
}
