package util

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

func CloudContextMiddleware(h func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := WithTrace(r.Context(), r)
		r = r.WithContext(ctx)
		h(w, r)
	}
}

type traceKey string

func WithTrace(ctx context.Context, r *http.Request) context.Context {
	var trace string

	traceHeader := r.Header.Get("X-Cloud-Trace-Context")

	traceParts := strings.Split(traceHeader, "/")
	if len(traceParts) > 0 && len(traceParts[0]) > 0 {
		trace = fmt.Sprintf("projects/russellsaw/traces/%s", traceParts[0])
	}

	return context.WithValue(ctx, traceKey("trace"), trace)
}

func Trace(ctx context.Context) string {
	v, ok := ctx.Value(traceKey("trace")).(string)
	if !ok {
		return ""
	}
	return v
}
