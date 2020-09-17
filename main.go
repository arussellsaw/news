package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/arussellsaw/news/dao"
	"github.com/arussellsaw/news/handler"
	"github.com/arussellsaw/news/pkg/util"
	"github.com/google/uuid"
	"github.com/monzo/slog"
)

func main() {
	ctx := context.Background()
	logger := util.ContextParamLogger{Logger: &util.StackDriverLogger{}}
	slog.SetDefaultLogger(logger)
	fmt.Println(uuid.New())

	err := dao.Init(ctx)
	if err != nil {
		slog.Error(ctx, "error initialising dao: %s", err)
		os.Exit(1)
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
