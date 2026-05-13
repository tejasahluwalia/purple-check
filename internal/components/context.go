package components

import (
	"context"
	"net/http"
)

type contextKey string

const RequestContextKey contextKey = "request-context"

func GetRequestContext(ctx context.Context) *http.Request {
	request, _ := ctx.Value(RequestContextKey).(*http.Request)
	return request
}
