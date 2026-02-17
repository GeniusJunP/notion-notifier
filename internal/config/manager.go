package config

import (
	"sync"
)

type ValidationError struct {
	Err error
}

func (e ValidationError) Error() string {
	if e.Err == nil {
		return "validation failed"
	}
	return e.Err.Error()
}

func (e ValidationError) Unwrap() error {
	return e.Err
}

type Manager struct {
	mu      sync.RWMutex
	cfg     Config
	env     Env
	cfgPath string
	envPath string
}

func NewManager(cfgPath, envPath string) (*Manager, error) {
	cfg, err := LoadConfig(cfgPath)
	if err != nil {
		return nil, err
	}
	env, err := LoadEnv(envPath)
	if err != nil {
		return nil, err
	}
	env = ApplyEnvOverrides(env)
	return &Manager{cfg: cfg, env: env, cfgPath: cfgPath, envPath: envPath}, nil
}

func (m *Manager) Snapshot() (Config, Env) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.cfg, m.env
}

func (m *Manager) Config() Config {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.cfg
}

func (m *Manager) Env() Env {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.env
}

func (m *Manager) UpdateConfig(cfg Config) (Config, error) {
	cfg = NormalizeConfig(cfg)
	if err := ValidateConfig(cfg); err != nil {
		return Config{}, ValidationError{Err: err}
	}
	if err := WriteConfig(m.cfgPath, cfg); err != nil {
		return Config{}, err
	}
	m.mu.Lock()
	m.cfg = cfg
	m.mu.Unlock()
	return cfg, nil
}

func (m *Manager) Reload() error {
	cfg, err := LoadConfig(m.cfgPath)
	if err != nil {
		return err
	}
	env, err := LoadEnv(m.envPath)
	if err != nil {
		return err
	}
	env = ApplyEnvOverrides(env)
	m.mu.Lock()
	m.cfg = cfg
	m.env = env
	m.mu.Unlock()
	return nil
}
