// internal/observability/logger/logger.go
package logger

import (
	"os"
	"path/filepath"

	"github.com/rafapasa/openerp-jud/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	instance *zap.Logger
	level    zapcore.Level
)

// LoggerConfig contém as configurações do logger
type LoggerConfig struct {
	Level      string `mapstructure:"LOG_LEVEL"`
	Encoding   string `mapstructure:"LOG_ENCODING"`
	OutputPath string `mapstructure:"LOG_OUTPUT_PATH"`
	LogFile    string `mapstructure:"LOG_FILE"`
}

// DefaultConfig retorna a configuração padrão
func DefaultConfig() LoggerConfig {
	return LoggerConfig{
		Level:      "info",
		Encoding:   "json",
		OutputPath: "stdout",
		LogFile:    "logs/mcp-server.log",
	}
}

// NewLogger cria uma nova instância do logger
func NewLogger(cfg *config.Config) (*zap.Logger, func(), error) {
	// ============================================
	// 1. CONFIGURAÇÃO
	// ============================================
	logCfg := LoggerConfig{
		Level:      cfg.LogLevel,
		Encoding:   cfg.LogEncoding,
		OutputPath: cfg.LogOutput,
		LogFile:    cfg.LogFile,
	}

	// Valores padrão
	if logCfg.Level == "" {
		logCfg.Level = "info"
	}
	if logCfg.Encoding == "" {
		logCfg.Encoding = "json"
	}
	if logCfg.OutputPath == "" {
		logCfg.OutputPath = "stdout"
	}
	if logCfg.LogFile == "" {
		logCfg.LogFile = "logs/mcp-server.log"
	}

	// ============================================
	// 2. NÍVEL DE LOG
	// ============================================
	lvl, err := zapcore.ParseLevel(logCfg.Level)
	if err != nil {
		lvl = zapcore.InfoLevel
	}
	level = lvl

	// ============================================
	// 3. ENCODER (FORMATO)
	// ============================================
	isDev := !cfg.IsProduction()

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Configuração para desenvolvimento (console colorido)
	if isDev || logCfg.Encoding == "console" {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("15:04:05.000")
		encoderConfig.ConsoleSeparator = " | "
	} else {
		// Produção (JSON)
		encoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	var encoder zapcore.Encoder
	if isDev || logCfg.Encoding == "console" {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	// ============================================
	// 4. SAÍDA (TERMINAL E/OU ARQUIVO)
	// ============================================
	// Lógica clara:
	//   - LOG_OUTPUT=stdout  → apenas terminal
	//   - LOG_OUTPUT=file    → apenas arquivo
	//   - LOG_OUTPUT=both    → ambos (padrão)
	//   - LOG_OUTPUT vazio   → ambos (padrão)

	var writers []zapcore.WriteSyncer

	switch logCfg.OutputPath {
	case "stdout":
		// Apenas terminal
		writers = append(writers, zapcore.AddSync(os.Stdout))

	case "file":
		// Apenas arquivo
		fileWriter, err := openLogFile(logCfg.LogFile)
		if err != nil {
			// Fallback para stdout se não conseguir abrir o arquivo
			writers = append(writers, zapcore.AddSync(os.Stdout))
		} else {
			writers = append(writers, fileWriter)
		}

	default: // "both" ou vazio
		// Terminal + arquivo
		writers = append(writers, zapcore.AddSync(os.Stdout))

		fileWriter, err := openLogFile(logCfg.LogFile)
		if err == nil {
			writers = append(writers, fileWriter)
		}
	}

	writeSyncer := zapcore.NewMultiWriteSyncer(writers...)
	core := zapcore.NewCore(encoder, writeSyncer, level)

	// ============================================
	// 5. CRIA O LOGGER
	// ============================================
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	cleanup := func() { _ = logger.Sync() }
	instance = logger
	return logger, cleanup, nil
}

// openLogFile abre (ou cria) o arquivo de log
func openLogFile(path string) (zapcore.WriteSyncer, error) {
	// Cria o diretório se não existir
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}

	// Abre o arquivo
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	return zapcore.AddSync(file), nil
}

// GetLogger retorna a instância global do logger
func GetLogger() *zap.Logger {
	if instance == nil {
		// Fallback: cria logger básico se não inicializado
		cfg := config.Config{}
		logger, _, _ := NewLogger(&cfg)
		return logger
	}
	return instance
}

// Sync sincroniza os logs pendentes
func Sync() error {
	if instance != nil {
		return instance.Sync()
	}
	return nil
}

// SetLevel altera o nível de log em runtime
func SetLevel(lvl zapcore.Level) {
	level = lvl
}

// GetLevel retorna o nível atual
func GetLevel() zapcore.Level {
	return level
}

// LogInfo loga em nível INFO
func LogInfo(msg string, fields ...zap.Field) {
	GetLogger().Info(msg, fields...)
}

func LogError(msg string, fields ...zap.Field) {
	GetLogger().Error(msg, fields...)
}

func LogDebug(msg string, fields ...zap.Field) {
	GetLogger().Debug(msg, fields...)
}

func LogWarn(msg string, fields ...zap.Field) {
	GetLogger().Warn(msg, fields...)
}
