package tribunal

import (
	"fmt"
	"sync"

	"github.com/rafapasa/openerp-jud/internal/model"
)

// Registry gerencia a resolução de drivers por tribunal.
type Registry interface {
	// Register adiciona um driver ao registro.
	Register(driver Driver)

	// Get recupera o driver correspondente ao tribunal solicitado.
	Get(tribunal model.Tribunal) (Driver, error)

	// List retorna todos os drivers registrados.
	List() []Driver
}

// inMemRegistry implementa a interface Registry armazenando os drivers em memória de forma thread-safe.
type inMemRegistry struct {
	mu      sync.RWMutex
	drivers map[model.Tribunal]Driver
}

var _ Registry = (*inMemRegistry)(nil)

// NewRegistry cria o registro de drivers judiciais já populado com os
// drivers suportados pela aplicação.
func NewRegistry(eproc Driver, projudi Driver) Registry {
	r := &inMemRegistry{
		drivers: make(map[model.Tribunal]Driver),
	}

	r.Register(eproc)
	r.Register(projudi)

	return r
}

// Register adiciona um driver ao mapa do registro.
func (r *inMemRegistry) Register(driver Driver) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.drivers[driver.Tribunal()] = driver
}

// Get recupera o driver solicitado.
func (r *inMemRegistry) Get(tribunal model.Tribunal) (Driver, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	driver, ok := r.drivers[tribunal]
	if !ok {
		return nil, fmt.Errorf("driver para o tribunal '%s' não encontrado ou não registrado", tribunal)
	}
	return driver, nil
}

// List retorna a lista de todos os drivers registrados.
func (r *inMemRegistry) List() []Driver {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list := make([]Driver, 0, len(r.drivers))
	for _, d := range r.drivers {
		list = append(list, d)
	}
	return list
}
