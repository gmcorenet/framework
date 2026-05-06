package debug

import (
	"fmt"
	"net/http"
	"runtime"
	"time"
)

type DebugBundle struct {
	enabled bool
}

func NewDebugBundle() *DebugBundle {
	return &DebugBundle{enabled: true}
}

func (d *DebugBundle) Enable() {
	d.enabled = true
}

func (d *DebugBundle) Disable() {
	d.enabled = false
}

func (d *DebugBundle) IsEnabled() bool {
	return d.enabled
}

type DebugData struct {
	StartTime time.Time
	Requests  int
	Errors    int
	Memory    runtime.MemStats
}

func NewDebugData() *DebugData {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	return &DebugData{
		StartTime: time.Now(),
		Memory:    stats,
	}
}

func (d *DebugData) UpTime() time.Duration {
	return time.Since(d.StartTime)
}

func (d *DebugData) IncrementRequests() {
	d.Requests++
}

func (d *DebugData) IncrementErrors() {
	d.Errors++
}

type DebugPanel struct {
	data *DebugData
}

func NewDebugPanel() *DebugPanel {
	return &DebugPanel{data: NewDebugData()}
}

func (p *DebugPanel) Render() string {
	return fmt.Sprintf(`
<div class="debug-panel">
	<h3>Debug Information</h3>
	<p>Uptime: %s</p>
	<p>Requests: %d</p>
	<p>Errors: %d</p>
	<p>Memory: %d KB</p>
</div>
`, p.data.UpTime(), p.data.Requests, p.data.Errors, p.data.Memory.Alloc/1024)
}

type ExceptionData struct {
	Timestamp time.Time
	Message   string
	Stack     string
	Request   *http.Request
}

func (e *ExceptionData) String() string {
	return fmt.Sprintf("[%s] %s\n%s", e.Timestamp.Format(time.RFC3339), e.Message, e.Stack)
}

type ExceptionHandler struct {
	exceptions []ExceptionData
}

func NewExceptionHandler() *ExceptionHandler {
	return &ExceptionHandler{
		exceptions: make([]ExceptionData, 0),
	}
}

func (h *ExceptionHandler) Handle(err error, req *http.Request) {
	exc := ExceptionData{
		Timestamp: time.Now(),
		Message:   err.Error(),
		Stack:     "stack trace",
		Request:   req,
	}
	h.exceptions = append(h.exceptions, exc)
}

func (h *ExceptionHandler) GetExceptions() []ExceptionData {
	return h.exceptions
}

func (h *ExceptionHandler) Clear() {
	h.exceptions = make([]ExceptionData, 0)
}

type ToolbarData struct {
	Request  *http.Request
	Time     time.Time
	Duration time.Duration
	Memory   runtime.MemStats
	Queries  []string
}

func NewToolbarData(req *http.Request) *ToolbarData {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	return &ToolbarData{
		Request:  req,
		Time:     time.Now(),
		Memory:   stats,
		Queries:  make([]string, 0),
	}
}

func (t *ToolbarData) SetDuration(d time.Duration) {
	t.Duration = d
}

func (t *ToolbarData) AddQuery(query string) {
	t.Queries = append(t.Queries, query)
}

func (t *ToolbarData) Render() string {
	return fmt.Sprintf(`
<!-- Debug Toolbar -->
<div id="debug-toolbar" style="position:fixed;bottom:0;left:0;right:0;background:#333;color:#fff;padding:8px;font-size:12px;z-index:99999;">
	<strong>GMCore Debug</strong> |
	Time: %s |
	Duration: %s |
	Memory: %d KB |
	Queries: %d
</div>
`, t.Time.Format("15:04:05"), t.Duration, t.Memory.Alloc/1024, len(t.Queries))
}

type WebDebugBundle struct {
	*DebugBundle
	panel   *DebugPanel
	handler *ExceptionHandler
}

func NewWebDebugBundle() *WebDebugBundle {
	return &WebDebugBundle{
		DebugBundle: NewDebugBundle(),
		panel:       NewDebugPanel(),
		handler:     NewExceptionHandler(),
	}
}

func (b *WebDebugBundle) RenderPanel() string {
	return b.panel.Render()
}

func (b *WebDebugBundle) HandleException(err error, req *http.Request) {
	b.handler.Handle(err, req)
}

func (b *WebDebugBundle) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		duration := time.Since(start)

		data := NewToolbarData(r)
		data.SetDuration(duration)
		w.Write([]byte(data.Render()))
	})
}

type LogEntry struct {
	Timestamp time.Time
	Level     string
	Message   string
	Context   map[string]interface{}
}

func (l *LogEntry) String() string {
	return fmt.Sprintf("[%s] %s: %s", l.Timestamp.Format(time.RFC3339), l.Level, l.Message)
}

type DevelopmentTools struct {
	logs     []LogEntry
	enabled  bool
}

func NewDevelopmentTools() *DevelopmentTools {
	return &DevelopmentTools{
		logs:    make([]LogEntry, 0),
		enabled: true,
	}
}

func (d *DevelopmentTools) Log(level, message string, ctx map[string]interface{}) {
	d.logs = append(d.logs, LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
		Context:   ctx,
	})
}

func (d *DevelopmentTools) GetLogs() []LogEntry {
	return d.logs
}

func (d *DevelopmentTools) ClearLogs() {
	d.logs = make([]LogEntry, 0)
}
