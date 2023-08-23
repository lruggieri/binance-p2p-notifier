package config

type Manager interface {
	GetConfig() Config
	SaveConfig(cfg Config)
}
