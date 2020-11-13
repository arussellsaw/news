package handler

import (
	"github.com/arussellsaw/news/domain"
	"net/http"

	"github.com/arussellsaw/news/pkg/util"
)

func Init() http.Handler {
	m := http.NewServeMux()
	m.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))
	m.Handle("/image", http.HandlerFunc(handleDitherImage))
	m.Handle("/article", http.HandlerFunc(handleArticle))
	m.Handle("/favicon.ico", http.NotFoundHandler())
	m.Handle("/barcode", http.HandlerFunc(handleQRCode))
	m.Handle("/generate-edition", http.HandlerFunc(handleGenerateEdition))
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
	return h
}
