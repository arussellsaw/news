package util

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
)

func CloudContextMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := WithTrace(r.Context(), r)
		r = r.WithContext(ctx)
		h.ServeHTTP(w, r)
	})
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

type paramKey string

type paramContainer struct {
	mu     sync.Mutex
	params map[string]string
}

func WithParams(ctx context.Context, params map[string]string) context.Context {
	return context.WithValue(ctx, paramKey("params"), paramContainer{params: params})
}

func SetParam(ctx context.Context, key, value string) context.Context {
	v, ok := ctx.Value(paramKey("params")).(paramContainer)
	if !ok {
		return WithParams(ctx, map[string]string{key: value})
	}

	v.mu.Lock()
	defer v.mu.Unlock()
	v.params[key] = value

	return ctx
}

func Params(ctx context.Context) map[string]interface{} {
	container, ok := ctx.Value(paramKey("params")).(paramContainer)
	if !ok {
		return nil
	}

	container.mu.Lock()
	defer container.mu.Unlock()
	params := make(map[string]interface{})
	for k, v := range container.params {
		params[k] = v
	}

	return params
}
