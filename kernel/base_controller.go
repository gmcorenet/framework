package kernel

import (
	"context"
	"encoding/json"
	"net/http"
)

type BaseController struct {
	Response http.ResponseWriter
	Request  *http.Request
}

func (c *BaseController) JSON(ctx context.Context, statusCode int, data interface{}) {
	if c.Response == nil {
		panic("BaseController.Response is nil; set it before calling JSON")
	}
	c.Response.Header().Set("Content-Type", "application/json; charset=utf-8")
	c.Response.WriteHeader(statusCode)
	if err := json.NewEncoder(c.Response).Encode(data); err != nil {
		http.Error(c.Response, err.Error(), http.StatusInternalServerError)
	}
}
