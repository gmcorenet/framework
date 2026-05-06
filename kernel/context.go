package kernel

import (
	"context"
	"net/http"

	"github.com/gmcorenet/framework/container"
)

type Context struct {
	context.Context
	Request  *http.Request
	Response http.ResponseWriter
	Kernel   *Kernel
}

func NewContext(ctx context.Context, w http.ResponseWriter, r *http.Request, k *Kernel) *Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return &Context{
		Context:  ctx,
		Request:  r,
		Response: w,
		Kernel:   k,
	}
}

func (c *Context) Container() *container.Container {
	return c.Kernel.Container()
}

func (c *Context) Value(key interface{}) interface{} {
	if val, ok := c.Context.Value(key).(interface{}); ok {
		return val
	}
	return nil
}

