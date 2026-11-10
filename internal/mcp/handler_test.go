package mcp_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/rafapasa/openerp-jud/config"
	"github.com/rafapasa/openerp-jud/internal/mcp"
	"github.com/rafapasa/openerp-jud/internal/model"
	"github.com/rafapasa/openerp-jud/internal/tribunal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockTribunalDriver implementa a interface tribunal.Driver para testes unitários.
type MockTribunalDriver struct {
	mock.Mock
}

var _ tribunal.Driver = (*MockTribunalDriver)(nil)

func (m *MockTribunalDriver) Tribunal() model.Tribunal {
	args := m.Called()
	return args.Get(0).(model.Tribunal)
}

func (m *MockTribunalDriver) ConsultarPorCPF(ctx context.Context, cpf string) ([]model.Processo, error) {
	args := m.Called(ctx, cpf)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Processo), args.Error(1)
}

func (m *MockTribunalDriver) ObterProcesso(ctx context.Context, numeroCNJ string) (*model.Processo, error) {
	args := m.Called(ctx, numeroCNJ)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Processo), args.Error(1)
}

func (m *MockTribunalDriver) BaixarPeca(ctx context.Context, numeroCNJ, docID, destino string) (string, error) {
	args := m.Called(ctx, numeroCNJ, docID, destino)
	return args.String(0), args.Error(1)
}

func setupTestEnvironment() (*config.Config, *tribunal.InMemRegistry) {
	cfg := &config.Config{
		DownloadsDir: "/tmp/openerp-jud-test",
	}
	reg := tribunal.NewRegistry()
	return cfg, reg
}

func TestHandleConsultarProcessosCPF_Sucesso_TribunalEspecifico(t *testing.T) {
	cfg, reg := setupTestEnvironment()

	mockTJSC := new(MockTribunalDriver)
	mockTJSC.On("Tribunal").Return(model.TribunalTJSC)

	esperados := []model.Processo{
		{
			NumeroCNJ: "5001234-56.2026.8.24.0001",
			Tribunal:  model.TribunalTJSC,
			Classe:    "Procedimento Comum Cível",
			Vara:      "1ª Vara Cível de Florianópolis",
			Situacao:  "Ativo",
		},
	}
	mockTJSC.On("ConsultarPorCPF", mock.Anything, "123.456.789-00").Return(esperados, nil)

	reg.Register(mockTJSC)
	handler := mcp.NewToolHandler(cfg, reg)

	// Act
	resultado, err := handler.HandleConsultarProcessosCPF(context.Background(), "123.456.789-00", model.TribunalTJSC)

	// Assert
	require.NoError(t, err)
	assert.Len(t, resultado, 1)
	assert.Equal(t, "5001234-56.2026.8.24.0001", resultado[0].NumeroCNJ)
	assert.Equal(t, model.TribunalTJSC, resultado[0].Tribunal)
	mockTJSC.AssertExpectations(t)
}

func TestHandleConsultarProcessosCPF_Federado_Todos(t *testing.T) {
	cfg, reg := setupTestEnvironment()

	mockTJSC := new(MockTribunalDriver)
	mockTJSC.On("Tribunal").Return(model.TribunalTJSC)
	mockTJSC.On("ConsultarPorCPF", mock.Anything, "12345678900").Return([]model.Processo{
		{NumeroCNJ: "5001111-11.2026.8.24.0001", Tribunal: model.TribunalTJSC},
	}, nil)

	mockTJPR := new(MockTribunalDriver)
	mockTJPR.On("Tribunal").Return(model.TribunalTJPR)
	mockTJPR.On("ConsultarPorCPF", mock.Anything, "12345678900").Return([]model.Processo{
		{NumeroCNJ: "0002222-22.2026.8.16.0001", Tribunal: model.TribunalTJPR},
	}, nil)

	reg.Register(mockTJSC)
	reg.Register(mockTJPR)
	handler := mcp.NewToolHandler(cfg, reg)

	// Act: Tribunal = TODOS
	resultado, err := handler.HandleConsultarProcessosCPF(context.Background(), "12345678900", model.TribunalTodos)

	// Assert: deve unificar processos de ambos os tribunais
	require.NoError(t, err)
	assert.Len(t, resultado, 2)

	numeros := []string{resultado[0].NumeroCNJ, resultado[1].NumeroCNJ}
	assert.Contains(t, numeros, "5001111-11.2026.8.24.0001")
	assert.Contains(t, numeros, "0002222-22.2026.8.16.0001")

	mockTJSC.AssertExpectations(t)
	mockTJPR.AssertExpectations(t)
}

func TestHandleConsultarProcessosCPF_Erro_CPF_Vazio(t *testing.T) {
	cfg, reg := setupTestEnvironment()
	handler := mcp.NewToolHandler(cfg, reg)

	// Act
	resultado, err := handler.HandleConsultarProcessosCPF(context.Background(), "", model.TribunalTJSC)

	// Assert
	require.Error(t, err)
	assert.Nil(t, resultado)
	assert.Contains(t, err.Error(), "cpf obrigatório")
}

func TestHandleConsultarProcessosCPF_TribunalNaoRegistrado(t *testing.T) {
	cfg, reg := setupTestEnvironment()
	handler := mcp.NewToolHandler(cfg, reg)

	// Act
	resultado, err := handler.HandleConsultarProcessosCPF(context.Background(), "12345678900", model.Tribunal("TRF4"))

	// Assert
	require.Error(t, err)
	assert.Nil(t, resultado)
	assert.Contains(t, err.Error(), "não encontrado")
}

func TestHandleObterDetalhesProcesso_Sucesso(t *testing.T) {
	cfg, reg := setupTestEnvironment()

	mockTJSC := new(MockTribunalDriver)
	mockTJSC.On("Tribunal").Return(model.TribunalTJSC)

	agora := time.Now()
	detalheEsperado := &model.Processo{
		NumeroCNJ:        "5001234-56.2026.8.24.0001",
		Tribunal:         model.TribunalTJSC,
		Comarca:          "Joinville",
		Vara:             "2ª Vara da Fazenda Pública",
		Classe:           "Mandado de Segurança",
		DataDistribuicao: &agora,
		ValorCausa:       150000.0,
		Partes: []model.ParteProcesso{
			{Nome: "Empresa XPTO Ltda", Polo: "Ativo"},
			{Nome: "Estado de Santa Catarina", Polo: "Passivo"},
		},
		Movimentacoes: []model.Movimentacao{
			{Numero: 1, Descricao: "Distribuído por sorteio"},
		},
	}

	mockTJSC.On("ObterProcesso", mock.Anything, "5001234-56.2026.8.24.0001").Return(detalheEsperado, nil)
	reg.Register(mockTJSC)

	handler := mcp.NewToolHandler(cfg, reg)

	// Act
	proc, err := handler.HandleObterDetalhesProcesso(context.Background(), "5001234-56.2026.8.24.0001", model.TribunalTJSC)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, proc)
	assert.Equal(t, "5001234-56.2026.8.24.0001", proc.NumeroCNJ)
	assert.Equal(t, "Joinville", proc.Comarca)
	assert.Equal(t, 150000.0, proc.ValorCausa)
	assert.Len(t, proc.Partes, 2)
	assert.Len(t, proc.Movimentacoes, 1)
	mockTJSC.AssertExpectations(t)
}

func TestHandleObterDetalhesProcesso_Erro_NumeroVazio(t *testing.T) {
	cfg, reg := setupTestEnvironment()
	handler := mcp.NewToolHandler(cfg, reg)

	proc, err := handler.HandleObterDetalhesProcesso(context.Background(), "", model.TribunalTJSC)

	require.Error(t, err)
	assert.Nil(t, proc)
	assert.Contains(t, err.Error(), "numero do processo obrigatório")
}

func TestHandleBaixarPeca_Sucesso(t *testing.T) {
	cfg, reg := setupTestEnvironment()

	mockTJPR := new(MockTribunalDriver)
	mockTJPR.On("Tribunal").Return(model.TribunalTJPR)

	caminhoEsperado := "/tmp/openerp-jud-test/0002222-22.2026.8.16.0001_doc123.pdf"
	mockTJPR.On("BaixarPeca", mock.Anything, "0002222-22.2026.8.16.0001", "doc123", cfg.DownloadsDir).
		Return(caminhoEsperado, nil)

	reg.Register(mockTJPR)
	handler := mcp.NewToolHandler(cfg, reg)

	// Act
	caminho, err := handler.HandleBaixarPeca(context.Background(), model.TribunalTJPR, "0002222-22.2026.8.16.0001", "doc123")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, caminhoEsperado, caminho)
	mockTJPR.AssertExpectations(t)
}

func TestHandleBaixarPeca_Erro_DocIDVazio(t *testing.T) {
	cfg, reg := setupTestEnvironment()
	handler := mcp.NewToolHandler(cfg, reg)

	caminho, err := handler.HandleBaixarPeca(context.Background(), model.TribunalTJPR, "0002222-22.2026.8.16.0001", "")

	require.Error(t, err)
	assert.Empty(t, caminho)
	assert.Contains(t, err.Error(), "id do documento obrigatório")
}

func TestHandleBaixarPeca_Erro_Driver(t *testing.T) {
	cfg, reg := setupTestEnvironment()

	mockTJSC := new(MockTribunalDriver)
	mockTJSC.On("Tribunal").Return(model.TribunalTJSC)
	mockTJSC.On("BaixarPeca", mock.Anything, "5001234-56.2026.8.24.0001", "doc999", cfg.DownloadsDir).
		Return("", errors.New("documento em segredo de justiça"))

	reg.Register(mockTJSC)
	handler := mcp.NewToolHandler(cfg, reg)

	caminho, err := handler.HandleBaixarPeca(context.Background(), model.TribunalTJSC, "5001234-56.2026.8.24.0001", "doc999")

	require.Error(t, err)
	assert.Empty(t, caminho)
	assert.Contains(t, err.Error(), "documento em segredo de justiça")
	mockTJSC.AssertExpectations(t)
}

