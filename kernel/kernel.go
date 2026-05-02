package kernel

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Config struct {
	Host       string
	Port       string
	Env        string
	Debug      bool
	Timezone   string
	RootPath   string
	StoragePath string
}

type Kernel struct {
	config     *Config
	container  *Container
	router     http.Handler
	dispatcher *EventDispatcher
	bundles    []Bundler
	mux        *http.ServeMux
}

func New(cfg *Config) *Kernel {
	if cfg == nil {
		cfg = &Config{
			Host:     "0.0.0.0",
			Port:     "8080",
			Env:      "dev",
			Debug:    false,
			RootPath: getEnv("ROOT_PATH", "."),
		}
	}

	k := &Kernel{
		config:     cfg,
		container:  NewContainer(),
		dispatcher: NewEventDispatcher(),
		bundles:    make([]Bundler, 0),
		mux:        http.NewServeMux(),
	}

	k.registerCoreServices()
	return k
}

func (k *Kernel) Bootstrap(ctx context.Context) error {
	log.Printf("Bootstrapping GMCore Framework v1.0.0")
	log.Printf("Environment: %s", k.config.Env)

	if err := k.loadBundles(ctx); err != nil {
		return fmt.Errorf("failed to load bundles: %w", err)
	}

	k.router = k.buildRouter()
	return nil
}

func (k *Kernel) loadBundles(ctx context.Context) error {
	for _, bundle := range k.bundles {
		log.Printf("Booting bundle: %s", bundle.Name())
		if err := bundle.Boot(ctx); err != nil {
			return fmt.Errorf("bundle %s failed to boot: %w", bundle.Name(), err)
		}
	}
	return nil
}

func (k *Kernel) buildRouter() http.Handler {
	return k.mux
}

func (k *Kernel) Mux() http.Handler {
	return k.mux
}

func (k *Kernel) Router() http.Handler {
	return k.router
}

func (k *Kernel) Container() *Container {
	return k.container
}

func (k *Kernel) Dispatcher() *EventDispatcher {
	return k.dispatcher
}

func (k *Kernel) Config() *Config {
	return k.config
}

func (k *Kernel) AddBundle(bundle Bundler) {
	k.bundles = append(k.bundles, bundle)
}

func (k *Kernel) Handle(pattern string, handler http.Handler) {
	k.mux.Handle(pattern, handler)
}

func (k *Kernel) HandleFunc(pattern string, fn func(http.ResponseWriter, *http.Request)) {
	k.mux.HandleFunc(pattern, fn)
}

func (k *Kernel) Shutdown() {
	log.Println("Shutting down kernel...")
	for _, bundle := range k.bundles {
		if err := bundle.Shutdown(); err != nil {
			log.Printf("Error shutting down bundle %s: %v", bundle.Name(), err)
		}
	}
}

func (k *Kernel) Run() error {
	addr := fmt.Sprintf("%s:%s", k.config.Host, k.config.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: k.router,
	}

	go func() {
		log.Printf("Server starting on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}

type Bundler interface {
	Name() string
	Boot(ctx context.Context) error
	Shutdown() error
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

