package autoloader

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Autoloader struct {
	namespaces map[string]string
	classMap   map[string]string
	mu         sync.RWMutex
}

var (
	instance *Autoloader
	once     sync.Once
)

func GetInstance() *Autoloader {
	once.Do(func() {
		instance = &Autoloader{
			namespaces: make(map[string]string),
			classMap:   make(map[string]string),
		}
	})
	return instance
}

func (a *Autoloader) RegisterNamespace(prefix, baseDir string) *Autoloader {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.namespaces[strings.TrimRight(prefix, "\\")] = filepath.Clean(baseDir)
	return a
}

func (a *Autoloader) RegisterClassMap(classMap map[string]string) *Autoloader {
	a.mu.Lock()
	defer a.mu.Unlock()
	for k, v := range classMap {
		a.classMap[k] = v
	}
	return a
}

func (a *Autoloader) LoadClass(class string) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if path, ok := a.classMap[class]; ok {
		_, err := os.Stat(path)
		if err == nil {
			return true
		}
	}

	class = strings.TrimLeft(class, "\\")
	for prefix, baseDir := range a.namespaces {
		if strings.HasPrefix(class, prefix) {
			relativeClass := strings.TrimPrefix(class, prefix)
			file := baseDir + strings.ReplaceAll(relativeClass, "\\", "/") + ".go"
			if _, err := os.Stat(file); err == nil {
				return true
			}
		}
	}

	return false
}

func (a *Autoloader) Register() {
	GetInstance()
}