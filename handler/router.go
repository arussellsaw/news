package handler

import (
	"net/http"

	"github.com/arussellsaw/news/pkg/util"
)

func Init() http.Handler {
	m := http.NewServeMux()
	m.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))
	m.Handle("/image", http.HandlerFunc(handleDitherImage))
	m.Handle("/article", http.HandlerFunc(handleArticle))
	m.Handle("/", http.HandlerFunc(handleNews))

	return http.HandlerFunc(util.CloudContextMiddleware(m.ServeHTTP))
}
