package http

import (
	"context"
	"net/http"
)

type Router interface {
	Register(pattern string, handler Handler)
}

type Handler interface {
	Handle(ctx Context)
}

type Context interface {
	context.Context
	Request() *http.Request
	ResponseWriter() http.ResponseWriter
}
