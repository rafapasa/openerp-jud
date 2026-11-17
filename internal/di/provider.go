package di

import (
	"github.com/google/wire"
	"github.com/rafapasa/openerp-jud/internal/browser"
	"github.com/rafapasa/openerp-jud/internal/client"
	"github.com/rafapasa/openerp-jud/internal/mcp"
	"github.com/rafapasa/openerp-jud/internal/tribunal"
)

// ============================================
// SETS DE PROVIDERS
// ============================================

// providerSetBrowser provê o gerenciador de contextos do Chromium.
var providerSetBrowser = wire.NewSet(
	browser.NewContextManager,
)

// providerSetClient provê os drivers de automação dos tribunais.
var providerSetClient = wire.NewSet(
	client.NewEProcClient,
	client.NewProjudiClient,
)

// providerSetTribunal provê o registro de drivers judiciais.
var providerSetTribunal = wire.NewSet(
	tribunal.NewRegistry,
)

// providerSetHandlers provê o handler de tools e o servidor MCP.
var providerSetHandlers = wire.NewSet(
	mcp.NewToolHandler,
	mcp.NewServer,
)

// ============================================
// SET AGREGADOR
// ============================================

// ProviderSet agrega todos os providers da aplicação.
var ProviderSet = wire.NewSet(
	providerSetBrowser,
	providerSetClient,
	providerSetTribunal,
	providerSetHandlers,
)
