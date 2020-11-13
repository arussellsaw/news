package handler

import (
	"net/http"
	"time"

	"github.com/arussellsaw/news/dao"
	"github.com/arussellsaw/news/domain"
	"github.com/monzo/slog"
)

func handleGenerateEdition(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	e, err := dao.GetEditionForTime(ctx, time.Now(), false)
	if err != nil {
		slog.Error(ctx, "Error getting edition: %s", err)
		http.Error(w, err.Error(), 500)
		return
	}
	if e != nil {
		slog.Info(ctx, "Edition %s - %s already exists and is within window", e.ID, e.Name)
		w.Write([]byte(e.ID))
		return
	}

	e, err = domain.NewEdition(ctx, time.Now())
	if err != nil {
		slog.Error(ctx, "Error creating edition: %s", err)
		http.Error(w, err.Error(), 500)
		return
	}
	err = dao.SetEdition(ctx, e)
	if err != nil {
		slog.Error(ctx, "Error storing edition: %s", err)
		http.Error(w, err.Error(), 500)
		return
	}
	slog.Info(ctx, "Created new edition: %s - %s", e.ID, e.Name)
	w.Write([]byte(e.ID))
}
