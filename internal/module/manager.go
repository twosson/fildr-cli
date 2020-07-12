package module

import (
	"fildr-cli/internal/log"
	"github.com/pkg/errors"
)

type ManagerInterface interface {
	Modules() []Module
	Register(module Module) error
	Unload()
}

type Manager struct {
	logger            log.Logger
	registeredModules []Module
	loadedModules     []Module
}

var _ ManagerInterface = (*Manager)(nil)

func NewManager(logger log.Logger) (*Manager, error) {
	manager := &Manager{
		logger: logger,
	}
	return manager, nil
}

func (m *Manager) Modules() []Module {
	return m.loadedModules
}

func (m *Manager) Register(module Module) error {
	m.registeredModules = append(m.registeredModules, module)

	if err := module.Start(); err != nil {
		return errors.Wrapf(err, "%s module failed to start", module.Name())
	}

	m.loadedModules = append(m.loadedModules, module)

	return nil
}

func (m *Manager) Unload() {
	for _, module := range m.loadedModules {
		module.Stop()
	}
}
