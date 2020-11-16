package util

import (
	"net/http"
	"time"

	"github.com/monzo/slog"
)

func HTTPLogParamsMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx = SetParam(ctx, "referrer", r.Referer())
		r = r.WithContext(ctx)
		slog.Info(ctx, "Handling request: %s %s", r.Method, r.URL.Path)
		start := time.Now()
		h.ServeHTTP(w, r)
		slog.Debug(ctx, "Request %s handled in %s", r.URL.Path, time.Since(start))
	})
}
