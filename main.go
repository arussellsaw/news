package main

import (
	"html/template"
	"log"
	"net/http"
	"time"
)

var cache Cache = &memoryCache{m: make(map[string][2]string)}

func main() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))
	http.Handle("/favicon.ico", http.NotFoundHandler())
	http.Handle("/image", http.HandlerFunc(handleDitherImage))
	http.Handle("/", http.HandlerFunc(handleNews))
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleNews(w http.ResponseWriter, r *http.Request) {
	t := template.New("index.html")
	t, err := t.ParseFiles("tmpl/index.html")
	if err != nil {
		log.Fatal("parsing", err)
	}
	articles, err := getArticles()
	if err != nil {
		log.Fatal("parsing", err)
	}
	articles = LayoutArticles(articles)
	h := Homepage{
		Title:    "The Webpage",
		Date:     time.Now().Format("Mon 02 Jan 2006"),
		Sources:  sources,
		Articles: articles,
	}
	err = t.Execute(w, h)
	if err != nil {
		log.Fatal("executing", err)
	}
}

type Homepage struct {
	Title    string
	Date     string
	Sources  []Source
	Articles []Article
}
