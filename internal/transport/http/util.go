package http

import (
	"encoding/json"
	"net/http"

	"github.com/art-es/queue-service/internal/infra/ops"
)

const (
	ReasonEmpty    = "EMPTY"
	ReasonTooLarge = "TOO_LARGE"
	ReasonTooSmall = "TOO_SMALL"
	ReasonInvalid  = "INVALID"
)

const (
	messageInternalError = "Internal error"
)

type CommonResponseBody struct {
	Message string                    `json:"message,omitempty"`
	Fields  []CommonResponseBodyField `json:"fields,omitempty"`
}

type CommonResponseBodyField struct {
	Name    string `json:"name"`
	Reason  string `json:"reason"`
	Message string `json:"message,omitempty"`
}

func WriteInvalidRequestBody(ctx Context) {
	WriteBadRequest(ctx, "Invalid request body")
}

func WriteBadRequest(ctx Context, msg string) {
	Write(ctx, http.StatusBadRequest, &CommonResponseBody{
		Message: msg,
	})
}

func WriteBadRequestFields(ctx Context, fields ...CommonResponseBodyField) {
	Write(ctx, http.StatusBadRequest, &CommonResponseBody{
		Fields: fields,
	})
}

func WriteInternalError(ctx Context) {
	Write(ctx, http.StatusInternalServerError, &CommonResponseBody{
		Message: messageInternalError,
	})
}

func Write(ctx Context, code int, body any) {
	w := ctx.ResponseWriter()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(body)
}

func WriteEmpty(ctx Context, code int) {
	w := ctx.ResponseWriter()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write([]byte("{}"))
}

func GetIdempotencyKey(ctx Context) *string {
	return ops.PointerOrNil(ctx.Request().Header.Get("X-IdempotencyKey"))
}
