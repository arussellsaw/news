package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/arussellsaw/news/domain"
	"github.com/arussellsaw/news/handler"
	"github.com/arussellsaw/news/pkg/util"
	"github.com/monzo/slog"
)

func main() {
	ctx := context.Background()
	logger := &util.StackDriverLogger{}
	slog.SetDefaultLogger(logger)

	var (
		err      error
		articles []domain.Article
	)
	for {
		articles, err = domain.GetArticles(ctx)
		if err != nil {
			slog.Error(ctx, "Error fetching articles: %s", err)
			continue
		}
		break
	}
	if len(articles) == 0 {
		log.Fatal(err)
	}

	var addr string
	if os.Getenv("NEWS_ENV") == "debug" {
		addr = ":8081"
	} else {
		addr = ":8080"
	}

	slog.Info(ctx, "ready, listening on addr: %s", addr)
	slog.Error(ctx, "serving: %s", http.ListenAndServe(addr, handler.Init()))
}
