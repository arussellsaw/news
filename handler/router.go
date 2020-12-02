package handler

import (
	"context"
	"github.com/arussellsaw/news/domain"
	"github.com/monzo/slog"
	"html/template"
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
	m.Handle("/", http.HandlerFunc(handleNews))
	m.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))
	m.Handle("/image", http.HandlerFunc(handleDitherImage))
	m.Handle("/article", http.HandlerFunc(handleArticle))
	m.Handle("/login", http.HandlerFunc(handleLogin))
	m.Handle("/settings", http.HandlerFunc(handleSettings))
	m.Handle("/logout", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{
			Name: "sess",
		})
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	}))
	m.Handle("/favicon.ico", http.NotFoundHandler())
	m.Handle("/generate-edition", http.HandlerFunc(handleGenerateEdition))
	m.Handle("/events/source", http.HandlerFunc(handlePubsubSource))
	m.Handle("/events/article", http.HandlerFunc(handlePubsubArticle))
	m.Handle("/poll", http.HandlerFunc(handlePoll))
	m.Handle("/article/debug", http.HandlerFunc(handleDebugArticle))
	m.Handle("/article/refresh", http.HandlerFunc(handleRefreshArticle))
	m.Handle("/settings/source", genericHandler("tmpl/settings_source.html", sourceSettingsData))

	h := util.CloudContextMiddleware(
		util.HTTPLogParamsMiddleware(
			m,
		),
	)
	h, err := domain.AnalyticsMiddleware(h.ServeHTTP)
	if err != nil {
		panic(err)
	}

	h = sessionMiddleware(h)

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

type genericPage struct {
	Data interface{}
	base
}

func genericHandler(path string, data func(w http.ResponseWriter, r *http.Request) (interface{}, error)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		u := domain.UserFromContext(ctx)
		p := genericPage{
			base: base{
				User: u,
			},
		}
		t := template.New("frame.html")
		t, err := t.ParseFiles("tmpl/frame.html", path)
		if err != nil {
			slog.Error(ctx, "Error parsing template: %s", err)
			http.Error(w, err.Error(), 500)
			return
		}

		p.Data, err = data(w, r)
		if err != nil {
			p.Error = err.Error()
		}

		err = t.Execute(w, &p)
		if err != nil {
			slog.Error(ctx, "Error executing template: %s", err)
			http.Error(w, err.Error(), 500)
			return
		}

	})
}
