package tribunal

import (
	"fmt"
	"sync"

	"github.com/rafapasa/openerp-jud/internal/model"
)

// InMemRegistry implementa a interface Registry armazenando os drivers em memória de forma thread-safe.
type InMemRegistry struct {
	mu      sync.RWMutex
	drivers map[model.Tribunal]Driver
}

var _ Registry = (*InMemRegistry)(nil)

// NewRegistry cria uma nova instância do registro de drivers judiciais.
func NewRegistry() *InMemRegistry {
	return &InMemRegistry{
		drivers: make(map[model.Tribunal]Driver),
	}
}

// Register adiciona um driver ao mapa do registro.
func (r *InMemRegistry) Register(driver Driver) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.drivers[driver.Tribunal()] = driver
}

// Get recupera o driver solicitado.
func (r *InMemRegistry) Get(tribunal model.Tribunal) (Driver, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	driver, ok := r.drivers[tribunal]
	if !ok {
		return nil, fmt.Errorf("driver para o tribunal '%s' não encontrado ou não registrado", tribunal)
	}
	return driver, nil
}

// List retorna a lista de todos os drivers registrados.
func (r *InMemRegistry) List() []Driver {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list := make([]Driver, 0, len(r.drivers))
	for _, d := range r.drivers {
		list = append(list, d)
	}
	return list
}

