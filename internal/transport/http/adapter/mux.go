package adapter

import (
	"context"
	"net/http"

	transport "github.com/art-es/queue-service/internal/transport/http"
)

type MuxRouter struct {
	Mux *http.ServeMux
}

type muxHandler struct {
	handler transport.Handler
}

type muxContext struct {
	context.Context
	w http.ResponseWriter
	r *http.Request
}

func NewMuxRouter() *MuxRouter {
	return &MuxRouter{
		Mux: http.NewServeMux(),
	}
}

func (r *MuxRouter) Register(pattern string, handler transport.Handler) {
	r.Mux.Handle(pattern, &muxHandler{handler: handler})
}

func (h *muxHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.handler.Handle(&muxContext{
		Context: r.Context(),
		r:       r,
		w:       w,
	})
}

func (c *muxContext) Request() *http.Request              { return c.r }
func (c *muxContext) ResponseWriter() http.ResponseWriter { return c.w }
