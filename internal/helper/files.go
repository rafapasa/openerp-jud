package helper

import (
	"fmt"
	"os"
	"path/filepath"
)

// EnsureDir garante que o diretório informado exista, criando-o (e seus pais) se necessário.
// Centraliza o uso de "os" para operações de sistema de arquivos.
func EnsureDir(path string) error {
	if path == "" {
		return fmt.Errorf("caminho de diretório vazio")
	}
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("falha ao criar diretório %s: %w", path, err)
	}
	return nil
}

// FileExists indica se o caminho informado existe e é um arquivo regular.
func FileExists(path string) bool {
	if path == "" {
		return false
	}
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// DirExists indica se o caminho informado existe e é um diretório.
func DirExists(path string) bool {
	if path == "" {
		return false
	}
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// JoinPath une segmentos de caminho de forma portável.
func JoinPath(elem ...string) string {
	return filepath.Join(elem...)
}
