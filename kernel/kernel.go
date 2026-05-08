package kernel

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/gmcorenet/framework/container"
	"github.com/gmcorenet/framework/events"
	"github.com/gmcorenet/framework/router"
	"gopkg.in/yaml.v3"
)

type HandlerFunc = router.Handler

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
	config        *Config
	container     *container.Container
	routerHandler http.Handler
	dispatcher    *events.EventDispatcher
	eventManager  *eventManager
	bundles       []Bundler
	routeBuilder  *router.Router
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
		config:       cfg,
		container:    container.NewContainer(),
		dispatcher:   events.NewEventDispatcher(),
		eventManager: newEventManager(),
		bundles:      make([]Bundler, 0),
		routeBuilder: router.New(),
	}

	return k
}

func (k *Kernel) Bootstrap(ctx context.Context) error {
	log.Printf("Bootstrapping GMCore Framework v1.0.0")
	log.Printf("Environment: %s", k.config.Env)

	event := NewKernelEvent(ctx, nil, nil)
	k.dispatchKernelEvent(ctx, EventBoot, event)

	if err := k.loadBundles(ctx); err != nil {
		return fmt.Errorf("failed to load bundles: %w", err)
	}

	k.routerHandler = k.buildRouter()
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
	return k.routeBuilder
}

func (k *Kernel) Mux() http.Handler {
	return k.routerHandler
}

func (k *Kernel) Router() http.Handler {
	return k.routerHandler
}

func (k *Kernel) RouteBuilder() *router.Router {
	return k.routeBuilder
}

func (k *Kernel) Container() *container.Container {
	return k.container
}

func (k *Kernel) SetContainer(c *container.Container) {
	k.container = c
}

func (k *Kernel) SetRouter(r http.Handler) {
	k.routerHandler = r
}

func (k *Kernel) Dispatcher() *events.EventDispatcher {
	return k.dispatcher
}

func (k *Kernel) EventManager() *eventManager {
	return k.eventManager
}

func (k *Kernel) Config() *Config {
	return k.config
}

func (k *Kernel) Subscribe(event string, subscriber *EventSubscriber) {
	k.eventManager.Subscribe(event, subscriber)
}

func (k *Kernel) dispatchKernelEvent(ctx context.Context, event string, ke *KernelEvent) {
	k.eventManager.Dispatch(ctx, event, ke)
}

func (k *Kernel) AddBundle(bundle Bundler) {
	k.bundles = append(k.bundles, bundle)
}

func (k *Kernel) RegisterDefaultServices() {
	k.container.Set("config", map[string]interface{}{
		"app_name": "GMCore Application",
		"env": k.config.Env,
		"debug": k.config.Debug,
	})
	k.container.Set("dispatcher", k.dispatcher)
	k.container.Set("event_manager", k.eventManager)
}

func (k *Kernel) GET(pattern string, handler HandlerFunc) (*router.Route, error) {
	return k.routeBuilder.GET(pattern, handler)
}

func (k *Kernel) POST(pattern string, handler HandlerFunc) (*router.Route, error) {
	return k.routeBuilder.POST(pattern, handler)
}

func (k *Kernel) PUT(pattern string, handler HandlerFunc) (*router.Route, error) {
	return k.routeBuilder.PUT(pattern, handler)
}

func (k *Kernel) DELETE(pattern string, handler HandlerFunc) (*router.Route, error) {
	return k.routeBuilder.DELETE(pattern, handler)
}

func (k *Kernel) PATCH(pattern string, handler HandlerFunc) (*router.Route, error) {
	return k.routeBuilder.PATCH(pattern, handler)
}

func (k *Kernel) Any(pattern string, handler HandlerFunc) (*router.Route, error) {
	return k.routeBuilder.Any(pattern, handler)
}

func (k *Kernel) Match(methods []string, pattern string, handler HandlerFunc) (*router.Route, error) {
	return k.routeBuilder.Match(methods, pattern, handler)
}

func (k *Kernel) Group(prefix string, callback func(*router.Router), middlewares ...router.Middleware) *router.Router {
	return k.routeBuilder.Group(prefix, callback, middlewares...)
}

func (k *Kernel) Use(m router.Middleware) {
	k.routeBuilder.Use(m)
}

func (k *Kernel) HandleRequest(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	ctx = context.WithValue(ctx, "kernel", k)

	event := NewKernelEvent(ctx, req, w)
	k.dispatchKernelEvent(ctx, EventRequest, event)

	k.routerHandler.ServeHTTP(w, req.WithContext(ctx))
}

func (k *Kernel) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	k.HandleRequest(w, req)
}

func (k *Kernel) Shutdown() {
	log.Println("Shutting down kernel...")
	for _, bundle := range k.bundles {
		if err := bundle.Shutdown(); err != nil {
			log.Printf("Error shutting down bundle %s: %v", bundle.Name(), err)
		}
	}
}

func (k *Kernel) RunServer() *http.Server {
	addr := fmt.Sprintf("%s:%s", k.config.Host, k.config.Port)

	k.ensureWelcomePage()
	k.ensureHealthEndpoint()

	k.routerHandler = k.buildRouter()

	certFile := filepath.Join(k.config.RootPath, "var", "keys", "cert.pem")
	keyFile := filepath.Join(k.config.RootPath, "var", "keys", "key.pem")

	if _, err := os.Stat(certFile); err == nil {
		if _, err := os.Stat(keyFile); err == nil {
			tlsCfg, err := loadTLSConfig(certFile, keyFile)
			if err == nil {
				return &http.Server{
					Addr:      addr,
					Handler:   k,
					TLSConfig: tlsCfg,
				}
			}
		}
	}

	return &http.Server{
		Addr:    addr,
		Handler: k,
	}
}

func loadTLSConfig(certFile, keyFile string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}
	return &tls.Config{Certificates: []tls.Certificate{cert}}, nil
}

func (k *Kernel) ensureWelcomePage() {
	k.GET("/", func(w http.ResponseWriter, r *http.Request, params map[string]string) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(k.welcomeHTML()))
	})
}

func (k *Kernel) ensureHealthEndpoint() {
	k.GET("/health", func(w http.ResponseWriter, r *http.Request, params map[string]string) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})
}

func (k *Kernel) Run() error {
	srv := k.RunServer()

	errCh := make(chan error, 1)
	go func() {
		log.Printf("Server starting on %s", srv.Addr)
		var err error
		if srv.TLSConfig != nil {
			err = srv.ListenAndServeTLS("", "")
		} else {
			err = srv.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	select {
	case <-quit:
		log.Println("Shutdown signal received")
	case err := <-errCh:
		log.Printf("Server error: %v", err)
		return err
	}

	log.Println("Shutting down server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}

func (k *Kernel) welcomeHTML() string {
	appName := "GMCore"
	appVersion := ""
	if data, err := os.ReadFile(filepath.Join(k.config.RootPath, "manifest.yaml")); err == nil {
		var mf struct {
			Name    string `yaml:"name"`
			Version string `yaml:"version"`
		}
		if yaml.Unmarshal(data, &mf) == nil {
			if mf.Name != "" {
				appName = mf.Name
			}
			if mf.Version != "" {
				appVersion = mf.Version
			}
		}
	}
	if appVersion != "" {
		appVersion = "v" + appVersion
	} else {
		appVersion = "v1.0.0"
	}

	return `<!DOCTYPE html><html lang="en">
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>` + appName + `</title>
<style>*{margin:0;padding:0;box-sizing:border-box}
body{background:#1a1a2e;color:#e0e0e0;font:16px/1.6 system-ui,sans-serif;display:flex;align-items:center;justify-content:center;min-height:100vh}
.card{background:#16213e;border-radius:16px;padding:48px;max-width:600px;width:90%;text-align:center;box-shadow:0 20px 60px rgba(0,0,0,.3)}
h1{font-size:2.2em;color:#e94560;margin-bottom:8px}
.version{color:#888;font-size:.9em;margin-bottom:24px}
.info{display:grid;grid-template-columns:1fr 1fr;gap:12px;text-align:left;margin-top:24px}
.info div{background:#0f3460;padding:12px 16px;border-radius:8px}
.info strong{color:#e94560;display:block;font-size:.8em;text-transform:uppercase;margin-bottom:4px}
.info span{color:#ccc}
.footer{margin-top:32px;color:#555;font-size:.8em}
</style></head><body><div class="card">
<h1>🚀 ` + appName + `</h1><p class="version">` + appVersion + ` — ` + k.config.Env + ` mode</p>
<p>Welcome to GMCore. Create your first controller to replace this page.</p>
<div class="info"><div><strong>Go</strong><span>` + runtime.Version() + `</span></div>
<div><strong>OS</strong><span>` + runtime.GOOS + `/` + runtime.GOARCH + `</span></div></div>
<p class="footer">GMCore Framework</p></div></body></html>`
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

