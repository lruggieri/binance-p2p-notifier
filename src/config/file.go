package config

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"sync"
)

type FileManager struct {
	lock *sync.RWMutex

	configFilePath string
}

func (m *FileManager) GetConfig() Config {
	// supposing config can be read because this FileManager could be created
	fileBytes, _ := os.ReadFile(m.configFilePath)

	var cfg Config
	// supposing config is valid, since we could read during this Manager's creation
	_ = json.Unmarshal(fileBytes, &cfg)

	return cfg
}

func (m *FileManager) SaveConfig(cfg Config) {
	m.lock.Lock()
	defer m.lock.Unlock()

	cfgBytes, _ := json.MarshalIndent(cfg, "", "  ")
	_ = os.WriteFile(m.configFilePath, cfgBytes, 0644)
}

func NewFileManager(_ context.Context, configFilePath string) (Manager, error) {
	// check that we can interact with the file
	f, err := os.OpenFile(configFilePath, os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	m := &FileManager{
		lock:           &sync.RWMutex{},
		configFilePath: configFilePath,
	}

	fileBytes, err := io.ReadAll(f)
	var cfg Config
	if err = json.Unmarshal(fileBytes, &cfg); err != nil {
		// in case we start with an invalid config, create a default valid one
		m.SaveConfig(cfg)
	}

	return m, nil
}
