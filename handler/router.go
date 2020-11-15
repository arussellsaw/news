package handler

import (
	"context"
	"github.com/arussellsaw/news/domain"
	"github.com/monzo/slog"
	"net/http"
	"net/http/cookiejar"
	"os"

	"github.com/arussellsaw/news/pkg/util"
)

var (
	jar, _ = cookiejar.New(&cookiejar.Options{})
	c      = http.Client{
		Jar: jar, //binks
	}

	p domain.Publisher
)

func Init(ctx context.Context) http.Handler {
	m := http.NewServeMux()
	m.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))
	m.Handle("/image", http.HandlerFunc(handleDitherImage))
	m.Handle("/article", http.HandlerFunc(handleArticle))
	m.Handle("/favicon.ico", http.NotFoundHandler())
	m.Handle("/barcode", http.HandlerFunc(handleQRCode))
	m.Handle("/generate-edition", http.HandlerFunc(handleGenerateEdition))
	m.Handle("/events/source", http.HandlerFunc(handlePubsubSource))
	m.Handle("/events/article", http.HandlerFunc(handlePubsubArticle))
	m.Handle("/article/debug", http.HandlerFunc(handleDebugArticle))
	m.Handle("/poll", http.HandlerFunc(handlePoll))
	m.Handle("/", http.HandlerFunc(handleNews))

	h := util.CloudContextMiddleware(
		util.HTTPLogParamsMiddleware(
			m,
		),
	)
	h, err := domain.AnalyticsMiddleware(h.ServeHTTP)
	if err != nil {
		panic(err)
	}

	if os.Getenv("USER") == "alexrussell-saw" {
		slog.Info(ctx, "Using HTTP Publisher")
		p = &domain.HTTPPublisher{
			SourceURL:  "http://localhost:8080/events/source",
			ArticleURL: "http://localhost:8080/events/article",
		}
	} else {
		slog.Info(ctx, "Using PubSub Publisher")
		p, err = domain.NewPubSubPublisher(ctx)
		if err != nil {
			panic(err)
		}
	}

	return h
}
