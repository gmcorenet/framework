package profiler

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"sort"
	"sync"
	"time"
)

type Profiler struct {
	enabled     bool
	requestData map[uint64]*RequestProfile
	mu          sync.RWMutex
	startTime   time.Time
}

type RequestProfile struct {
	ID          uint64
	StartTime   time.Time
	EndTime     time.Time
	Duration    time.Duration
	Method      string
	Path        string
	StatusCode  int
	MemoryStart runtime.MemStats
	MemoryEnd   runtime.MemStats
	Spans       []Span
}

type Span struct {
	Name      string
	StartTime time.Time
	EndTime   time.Time
	Duration  time.Duration
	Metadata  map[string]interface{}
}

func NewProfiler() *Profiler {
	return &Profiler{
		enabled:     true,
		requestData: make(map[uint64]*RequestProfile),
		startTime:   time.Now(),
	}
}

func (p *Profiler) Enable() {
	p.enabled = true
}

func (p *Profiler) Disable() {
	p.enabled = false
}

func (p *Profiler) IsEnabled() bool {
	return p.enabled
}

func (p *Profiler) StartRequest(req *http.Request) uint64 {
	if !p.enabled {
		return 0
	}

	var memStart runtime.MemStats
	runtime.ReadMemStats(&memStart)

	id := uint64(time.Now().UnixNano())

	profile := &RequestProfile{
		ID:          id,
		StartTime:   time.Now(),
		Method:      req.Method,
		Path:        req.URL.Path,
		MemoryStart: memStart,
		Spans:       make([]Span, 0),
	}

	p.mu.Lock()
	p.requestData[id] = profile
	p.mu.Unlock()

	return id
}

func (p *Profiler) EndRequest(id uint64, statusCode int) {
	if !p.enabled || id == 0 {
		return
	}

	var memEnd runtime.MemStats
	runtime.ReadMemStats(&memEnd)

	p.mu.Lock()
	defer p.mu.Unlock()

	if profile, ok := p.requestData[id]; ok {
		profile.EndTime = time.Now()
		profile.Duration = profile.EndTime.Sub(profile.StartTime)
		profile.StatusCode = statusCode
		profile.MemoryEnd = memEnd
	}
}

func (p *Profiler) AddSpan(id uint64, name string, metadata map[string]interface{}) {
	if !p.enabled || id == 0 {
		return
	}

	span := Span{
		Name:     name,
		Metadata: metadata,
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if profile, ok := p.requestData[id]; ok {
		profile.Spans = append(profile.Spans, span)
	}
}

func (p *Profiler) GetRequestProfile(id uint64) *RequestProfile {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.requestData[id]
}

func (p *Profiler) GetAllProfiles() []*RequestProfile {
	p.mu.RLock()
	defer p.mu.RUnlock()

	profiles := make([]*RequestProfile, 0, len(p.requestData))
	for _, profile := range p.requestData {
		profiles = append(profiles, profile)
	}

	sort.Slice(profiles, func(i, j int) bool {
		return profiles[i].StartTime.Before(profiles[j].StartTime)
	})

	return profiles
}

func (p *Profiler) GetSlowRequests(threshold time.Duration) []*RequestProfile {
	profiles := p.GetAllProfiles()
	slow := make([]*RequestProfile, 0)

	for _, profile := range profiles {
		if profile.Duration > threshold {
			slow = append(slow, profile)
		}
	}

	return slow
}

func (p *Profiler) GetMemoryUsage() MemoryUsage {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	return MemoryUsage{
		Alloc:      mem.Alloc,
		TotalAlloc: mem.TotalAlloc,
		Sys:        mem.Sys,
		NumGC:      mem.NumGC,
		Uptime:     time.Since(p.startTime),
	}
}

type MemoryUsage struct {
	Alloc      uint64
	TotalAlloc uint64
	Sys        uint64
	NumGC      uint32
	Uptime     time.Duration
}

func (p *Profiler) RenderProfile(id uint64) string {
	profile := p.GetRequestProfile(id)
	if profile == nil {
		return "Profile not found"
	}

	return fmt.Sprintf(`
Request Profile #%d
==================
Method: %s
Path: %s
Status: %d
Duration: %s
Memory Start: %d KB
Memory End: %d KB
Memory Delta: %d KB
Spans: %d
`, profile.ID, profile.Method, profile.Path, profile.StatusCode,
		profile.Duration, profile.MemoryStart.Alloc/1024,
		profile.MemoryEnd.Alloc/1024,
		(profile.MemoryEnd.Alloc-profile.MemoryStart.Alloc)/1024,
		len(profile.Spans))
}

func (p *Profiler) RenderSummary() string {
	profiles := p.GetAllProfiles()
	if len(profiles) == 0 {
		return "No requests profiled"
	}

	var totalDuration time.Duration
	var slowCount int
	slowThreshold := 500 * time.Millisecond

	for _, profile := range profiles {
		totalDuration += profile.Duration
		if profile.Duration > slowThreshold {
			slowCount++
		}
	}

	avgDuration := totalDuration / time.Duration(len(profiles))
	memory := p.GetMemoryUsage()

	return fmt.Sprintf(`
Profiler Summary
================
Total Requests: %d
Average Duration: %s
Slow Requests (>500ms): %d
Memory Alloc: %d KB
Memory Sys: %d KB
GC Runs: %d
Uptime: %s
`, len(profiles), avgDuration, slowCount,
		memory.Alloc/1024, memory.Sys/1024,
		memory.NumGC, memory.Uptime)
}

type contextKey string

const ProfilerContextKey contextKey = "profiler"

func SaveToContext(ctx context.Context, p *Profiler) context.Context {
	return context.WithValue(ctx, ProfilerContextKey, p)
}

func FromContext(ctx context.Context) *Profiler {
	if p, ok := ctx.Value(ProfilerContextKey).(*Profiler); ok {
		return p
	}
	return nil
}

func StartSpan(ctx context.Context, name string) (context.Context, func()) {
	p := FromContext(ctx)
	if p == nil {
		return ctx, func() {}
	}

	id := ctx.Value("request_id").(uint64)
	start := time.Now()

	return ctx, func() {
		p.AddSpan(id, name, map[string]interface{}{
			"start": start,
			"end":   time.Now(),
		})
	}
}

type Middleware struct {
	profiler *Profiler
}

func NewMiddleware(p *Profiler) *Middleware {
	return &Middleware{profiler: p}
}

func (m *Middleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !m.profiler.IsEnabled() {
			next.ServeHTTP(w, r)
			return
		}

		id := m.profiler.StartRequest(r)

		ctx := context.WithValue(r.Context(), "request_id", id)
		r = r.WithContext(ctx)

		start := time.Now()
		statusCode := 200

		next.ServeHTTP(w, r)

		m.profiler.EndRequest(id, statusCode)
		_ = start
	})
}
