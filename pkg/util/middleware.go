package util

import (
	"net/http"

	"github.com/monzo/slog"
)

func HTTPLogParamsMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx = SetParam(ctx, "referrer", r.Referer())
		r = r.WithContext(ctx)
		slog.Info(ctx, "Handling request: %s %s", r.Method, r.URL.Path)
		h.ServeHTTP(w, r)
	})
}
