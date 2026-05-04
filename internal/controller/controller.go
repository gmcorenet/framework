package controller

import (
	"context"
	"encoding/json"
	"net/http"
)

type ContextKey string

const (
	RequestKey  ContextKey = "request"
	ResponseKey ContextKey = "response"
	ParamsKey   ContextKey = "params"
)

type BaseController struct{}

func (c *BaseController) Request(ctx context.Context) *http.Request {
	if req, ok := ctx.Value(RequestKey).(*http.Request); ok {
		return req
	}
	return nil
}

func (c *BaseController) Response(ctx context.Context) http.ResponseWriter {
	if w, ok := ctx.Value(ResponseKey).(http.ResponseWriter); ok {
		return w
	}
	return nil
}

func (c *BaseController) Params(ctx context.Context) map[string]string {
	if params, ok := ctx.Value(ParamsKey).(map[string]string); ok {
		return params
	}
	return make(map[string]string)
}

func (c *BaseController) Param(ctx context.Context, key string) string {
	params := c.Params(ctx)
	if val, ok := params[key]; ok {
		return val
	}
	return ""
}

func (c *BaseController) Container(ctx context.Context) interface{} {
	if container, ok := ctx.Value("container").(interface{}); ok {
		return container
	}
	return nil
}

func (c *BaseController) GetQuery(ctx context.Context, key, defaultValue string) string {
	req := c.Request(ctx)
	if req == nil {
		return defaultValue
	}
	if value := req.URL.Query().Get(key); value != "" {
		return value
	}
	return defaultValue
}

func (c *BaseController) GetForm(ctx context.Context, key, defaultValue string) string {
	req := c.Request(ctx)
	if req == nil {
		return defaultValue
	}
	if value := req.FormValue(key); value != "" {
		return value
	}
	return defaultValue
}

func (c *BaseController) Redirect(ctx context.Context, url string, status int) {
	w := c.Response(ctx)
	if w != nil {
		http.Redirect(w, c.Request(ctx), url, status)
	}
}

func (c *BaseController) JSON(ctx context.Context, status int, data interface{}) {
	w := c.Response(ctx)
	if w != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		// Simple JSON encoding - in production use proper encoder
		encodeJSON(w, data)
	}
}

func encodeJSON(w http.ResponseWriter, data interface{}) {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		http.Error(w, "JSON encoding error", http.StatusInternalServerError)
	}
}

func (c *BaseController) Render(ctx context.Context, template string, data map[string]interface{}) {
	// Placeholder for view rendering integration
}

type ControllerInterface interface {
	Init(ctx context.Context)
}