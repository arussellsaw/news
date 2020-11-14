package domain

import (
	"context"
	"sort"
	"time"
	"unicode/utf8"

	"github.com/pkg/errors"

	"github.com/google/uuid"
)

var (
	morningEdition time.Duration = 6 * time.Hour
	eveningEdition time.Duration = 17 * time.Hour
)

type Edition struct {
	ID         string
	Name       string
	Sources    []Source
	Articles   []Article
	Categories []string
	Date       string

	StartTime time.Time
	EndTime   time.Time

	Metadata map[string]string

	Article Article
	claimed map[string]bool
}

func (e *Edition) GetArticle(size int, image bool) Article {
	if e.claimed == nil {
		e.claimed = make(map[string]bool)
	}
top:
	for _, a := range e.Articles {
		if a.ID == "bar" {
			continue
		}
		if a.Size() >= size {
			if a.ImageURL == "" && image {
				continue
			}
			if e.claimed[a.ID] {
				continue
			}
			e.claimed[a.ID] = true
			a.ImageURL = func() string {
				if image {
					return a.ImageURL
				}
				return ""
			}()
			a.Trim(size)
			a.Layout = Layout{}
			return a
		}
	}
	if size >= 0 {
		size -= 100
		goto top
	}
	return Article{
		Title:  "Not Found",
		Author: "404",
	}
}

func NewEdition(ctx context.Context, now time.Time) (*Edition, error) {
	morning := now.Truncate(24 * time.Hour).Add(morningEdition)
	evening := now.Truncate(24 * time.Hour).Add(eveningEdition)

	catMap := make(map[string]struct{})
	for _, s := range sources {
		for _, c := range s.Categories {
			catMap[c] = struct{}{}
		}
	}
	cats := make([]string, 0, len(catMap))
	for c := range catMap {
		cats = append(cats, c)
	}

	e := Edition{
		ID:         uuid.New().String(),
		Sources:    sources,
		Categories: cats,
	}

	e.Date = time.Now().Format("Monday January 02 2006")

	switch {
	case now.After(morning) && now.Before(evening):
		e.StartTime = morning
		e.EndTime = evening
		e.Name = "Morning Edition"
	case now.Before(morning) || now.After(evening):
		e.StartTime = evening
		e.EndTime = morning.Add(24 * time.Hour)
		e.Name = "Evening Edition"
	}

	articles, err := FetchArticles(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "error fetching articles")
	}

	newArticles := []Article{}
L:
	for _, a := range articles {
		if time.Since(a.Timestamp) > 100*time.Hour {
			continue
		}
		for _, e := range a.Content {
			if !utf8.Valid([]byte(e.Value)) {
				continue L
			}
		}
		newArticles = append(newArticles, a)
	}
	e.Articles = newArticles

	bySource := make(map[string][]Article)
	for _, a := range e.Articles {
		bySource[a.Source.Name] = append(bySource[a.Source.Name], a)
	}
	newArticles = nil
	for _, as := range bySource {
		sort.Slice(as, func(i, j int) bool {
			return as[i].Timestamp.After(as[j].Timestamp)
		})
	}
top:
	for s, as := range bySource {
		newArticles = append(newArticles, as[0])
		bySource[s] = as[1:]
		if len(bySource[s]) == 0 {
			delete(bySource, s)
			goto top
		}
	}
	if len(bySource) != 0 {
		goto top
	}
	e.Articles = newArticles

	return &e, nil
}
