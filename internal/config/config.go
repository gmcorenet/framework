package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Config struct {
	data map[string]interface{}
}

var (
	instance *Config
	once     sync.Once
)

func GetInstance() *Config {
	once.Do(func() {
		instance = &Config{data: make(map[string]interface{})}
	})
	return instance
}

func (c *Config) Load(path string) *Config {
	info, err := os.Stat(path)
	if err != nil {
		return c
	}

	if info.IsDir() {
		entries, _ := os.ReadDir(path)
		for _, entry := range entries {
			if !entry.IsDir() {
				resolved := filepath.Join(path, entry.Name())
				if !strings.HasPrefix(filepath.Clean(resolved), filepath.Clean(path)+string(os.PathSeparator)) && filepath.Clean(resolved) != filepath.Clean(path) {
					continue
				}
				c.loadFile(resolved)
			}
		}
	} else {
		c.loadFile(path)
	}

	return c
}

func (c *Config) loadFile(path string) {
	ext := filepath.Ext(path)
	key := filepath.Base(path[:len(path)-len(ext)])

	switch ext {
	case ".yaml", ".yml":
		c.data[key] = parseYamlFile(path)
	case ".json":
		c.data[key] = parseJsonFile(path)
	}
}

func parseYamlFile(path string) map[string]interface{} {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	return parseYaml(string(content))
}

func parseJsonFile(path string) map[string]interface{} {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var result map[string]interface{}
	json.Unmarshal(content, &result)
	return result
}

func (c *Config) Get(key string) interface{} {
	keys := strings.Split(key, ".")
	var value interface{} = c.data

	for _, k := range keys {
		if m, ok := value.(map[string]interface{}); ok {
			value = m[k]
		} else {
			return nil
		}
	}

	return value
}

func (c *Config) Set(key string, value interface{}) *Config {
	keys := strings.Split(key, ".")
	m := c.data

	for i := 0; i < len(keys)-1; i++ {
		if _, ok := m[keys[i]]; !ok {
			m[keys[i]] = make(map[string]interface{})
		}
		if nested, ok := m[keys[i]].(map[string]interface{}); ok {
			m = nested
		} else {
			newMap := make(map[string]interface{})
			m[keys[i]] = newMap
			m = newMap
		}
	}

	m[keys[len(keys)-1]] = value
	return c
}

func (c *Config) All() map[string]interface{} {
	return c.data
}

func parseYaml(content string) map[string]interface{} {
	result := make(map[string]interface{})
	lines := strings.Split(content, "\n")
	var currentSection string

	for _, line := range lines {
		line = strings.TrimRight(line, " \t")

		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
			if currentSection == "" {
				continue
			}
			parts := strings.SplitN(strings.TrimSpace(line), ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.Trim(parts[1], "\"")
				if m, ok := result[currentSection].(map[string]interface{}); ok {
					m[key] = value
				}
			}
		} else {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.Trim(parts[1], "\"")
				if value == "" {
					result[key] = make(map[string]interface{})
					currentSection = key
				} else {
					result[key] = value
					currentSection = ""
				}
			}
		}
	}

	return result
}