package config

import (
	"fmt"
	"log"
	"sync"

	"os"

	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Server    ServerConfig              `yaml:"server"`
	API       APIConfig                 `yaml:"api"`
	Providers map[string]ProviderConfig `yaml:"providers"`
}

type ServerConfig struct {
	Port  int  `yaml:"port"`
	Debug bool `yaml:"debug"`
}

type APIConfig struct {
	Key string `yaml:"key"`
}

type ProviderConfig struct {
	Enabled   bool   `yaml:"enabled"`
	Token     string `yaml:"token,omitempty"`
	CloudName string `yaml:"cloud_name,omitempty"`
	APIKey    string `yaml:"api_key,omitempty"`
	APISecret string `yaml:"api_secret,omitempty"`
}

var (
	cfg     *Config
	cfgLock sync.RWMutex
)

func Load(configPath string) error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	var newCfg Config
	if err := yaml.Unmarshal(data, &newCfg); err != nil {
		return fmt.Errorf("解析配置文件失败: %w", err)
	}

	cfgLock.Lock()
	cfg = &newCfg
	cfgLock.Unlock()

	log.Println("配置文件加载成功")
	return nil
}

func WatchConfig(configPath string) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	go func() {
		defer watcher.Close()
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("配置文件发生变化，重新加载...")
					if err := Load(configPath); err != nil {
						log.Printf("重新加载配置文件失败: %v", err)
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Printf("配置文件监听错误: %v", err)
			}
		}
	}()

	return watcher.Add(configPath)
}

func Get() *Config {
	cfgLock.RLock()
	defer cfgLock.RUnlock()
	return cfg
}

func GetProvider(name string) (ProviderConfig, bool) {
	cfgLock.RLock()
	defer cfgLock.RUnlock()

	if cfg == nil {
		return ProviderConfig{}, false
	}

	provider, exists := cfg.Providers[name]
	return provider, exists && provider.Enabled
}
