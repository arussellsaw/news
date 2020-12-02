package dao

import (
	"context"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"sync"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/arussellsaw/news/domain"
	"google.golang.org/api/iterator"
)

var (
	client       *firestore.Client
	mu           sync.RWMutex
	articleCache = make(map[string]domain.Article)
)

func Init(ctx context.Context) error {
	c, err := firestore.NewClient(ctx, "russellsaw")
	if err != nil {
		return err
	}

	client = c

	return nil
}

type storedEdition struct {
	ID         string
	Name       string
	Date       string
	StartTime  time.Time
	EndTime    time.Time
	Created    time.Time
	Sources    []domain.Source
	Articles   []string
	Categories []string
	Metadata   map[string]string
}

func GetEditionForTime(ctx context.Context, t time.Time, allowRecent bool) (*domain.Edition, error) {
	iter := client.Collection("editions").
		Where("EndTime", ">", t).
		Documents(ctx)
	candidates := []*domain.Edition{}
	var maxEdition storedEdition
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		s := storedEdition{}
		err = doc.DataTo(&s)
		if err != nil {
			return nil, err
		}
		if s.EndTime.After(maxEdition.EndTime) {
			maxEdition = s
		}
		if s.EndTime.After(t) {
			e, err := editionFromStored(ctx, s)
			if err != nil {
				return nil, err
			}
			candidates = append(candidates, e)
		}
	}
	if len(candidates) == 0 {
		if maxEdition.ID != "" && allowRecent {
			return editionFromStored(ctx, maxEdition)
		}
	}

	selected := &domain.Edition{}
	for _, e := range candidates {
		if e.Created.After(selected.Created) {
			selected = e
		}
	}
	if selected.ID == "" {
		return nil, nil
	}
	return selected, nil
}

func SetEdition(ctx context.Context, e *domain.Edition) error {
	for _, a := range e.Articles {
		err := SetArticle(ctx, &a)
		if err != nil {
			return errors.Wrap(err, "storing article: "+a.ID)
		}
	}
	stored := editionToStored(e)
	_, err := client.Collection("editions").Doc(e.ID).Set(ctx, stored)
	return err
}

func editionToStored(e *domain.Edition) storedEdition {
	s := storedEdition{
		ID:         e.ID,
		Name:       e.Name,
		Date:       e.Date,
		Sources:    e.Sources,
		StartTime:  e.StartTime,
		EndTime:    e.EndTime,
		Created:    e.Created,
		Categories: e.Categories,
		Metadata:   e.Metadata,
	}

	var IDs []string
	for _, a := range e.Articles {
		IDs = append(IDs, a.ID)
	}
	s.Articles = IDs
	return s
}

func editionFromStored(ctx context.Context, s storedEdition) (*domain.Edition, error) {
	e := domain.Edition{
		ID:         s.ID,
		Name:       s.Name,
		Date:       s.Date,
		Sources:    s.Sources,
		StartTime:  s.StartTime,
		EndTime:    s.EndTime,
		Created:    s.Created,
		Categories: s.Categories,
		Metadata:   s.Metadata,
	}
	e.Articles = make([]domain.Article, len(s.Articles))
	g := errgroup.Group{}
	for i, id := range s.Articles {
		i, id := i, id
		g.Go(func() error {
			a, err := GetArticle(ctx, id)
			if err != nil {
				return err
			}
			e.Articles[i] = *a
			return nil
		})
	}
	return &e, g.Wait()
}

func GetEdition(ctx context.Context, id string) (*domain.Edition, error) {
	d, err := client.Collection("editions").Doc(id).Get(ctx)
	if err != nil {
		return nil, err
	}
	s := storedEdition{}
	err = d.DataTo(&s)
	if err != nil {
		return nil, err
	}
	e, err := editionFromStored(ctx, s)
	return e, err
}

func SetArticle(ctx context.Context, a *domain.Article) error {
	_, err := client.Collection("articles").Doc(a.ID).Set(ctx, a)
	mu.Lock()
	delete(articleCache, a.ID)
	mu.Unlock()
	return err
}

func GetArticle(ctx context.Context, id string) (*domain.Article, error) {
	mu.RLock()
	a, ok := articleCache[id]
	if ok {
		mu.RUnlock()
		return &a, nil
	}
	mu.RUnlock()

	d, err := client.Collection("articles").Doc(id).Get(ctx)
	if err != nil {
		return nil, err
	}
	a = domain.Article{}
	err = d.DataTo(&a)
	mu.Lock()
	articleCache[a.ID] = a
	mu.Unlock()
	return &a, err
}

func GetArticleByURL(ctx context.Context, url string) (*domain.Article, error) {
	iter := client.Collection("articles").Where("Link", "==", url).Documents(ctx)
	docs, err := iter.GetAll()
	if err != nil {
		return nil, err
	}
	if len(docs) == 0 {
		return nil, nil
	}
	a := domain.Article{}
	err = docs[0].DataTo(&a)
	mu.Lock()
	articleCache[a.ID] = a
	mu.Unlock()
	return &a, err
}

func GetArticlesByTime(ctx context.Context, start, end time.Time) ([]domain.Article, error) {
	iter := client.Collection("articles").
		Where("Timestamp", ">", start).
		Where("Timestamp", "<", end).
		Documents(ctx)
	docs, err := iter.GetAll()
	if err != nil {
		return nil, err
	}

	out := []domain.Article{}
	for _, doc := range docs {
		a := domain.Article{}
		err = doc.DataTo(&a)
		mu.Lock()
		articleCache[a.ID] = a
		mu.Unlock()
		out = append(out, a)
	}
	return out, nil
}

func GetUser(ctx context.Context, id string) (*domain.User, error) {
	d, err := client.Collection("users").Doc(id).Get(ctx)
	if err != nil {
		return nil, err
	}
	u := domain.User{}
	err = d.DataTo(&u)
	return &u, err
}

func SetUser(ctx context.Context, u *domain.User) error {
	_, err := client.Collection("users").Doc(u.ID).Set(ctx, u)
	return err
}

func GetUserByName(ctx context.Context, name string) (*domain.User, error) {
	iter := client.Collection("users").Where("Name", "==", name).Documents(ctx)
	docs, err := iter.GetAll()
	if err != nil {
		return nil, err
	}
	if len(docs) == 0 {
		return nil, nil
	}
	u := domain.User{}
	err = docs[0].DataTo(&u)
	return &u, err
}

func GetSource(ctx context.Context, id string) (*domain.Source, error) {
	d, err := client.Collection("sources").Doc(id).Get(ctx)
	if err != nil {
		return nil, err
	}
	s := domain.Source{}
	err = d.DataTo(&s)
	return &s, err
}

func SetSource(ctx context.Context, s *domain.Source) error {
	_, err := client.Collection("sources").Doc(s.ID).Set(ctx, s)
	return err
}

func DeleteSource(ctx context.Context, id string) error {
	_, err := client.Collection("sources").Doc(id).Delete(ctx)
	return err
}

func GetSources(ctx context.Context, ownerID string) ([]domain.Source, error) {
	iter := client.Collection("sources").Where("OwnerID", "==", ownerID).Documents(ctx)
	docs, err := iter.GetAll()
	if err != nil {
		return nil, err
	}
	if len(docs) == 0 {
		return nil, nil
	}
	sources := make([]domain.Source, 0, len(docs))
	for _, doc := range docs {
		s := domain.Source{}
		err = doc.DataTo(&s)
		if err != nil {
			return nil, err
		}
		sources = append(sources, s)
	}
	return sources, nil
}

func GetAllSources(ctx context.Context) ([]domain.Source, error) {
	iter := client.Collection("sources").Documents(ctx)
	docs, err := iter.GetAll()
	if err != nil {
		return nil, err
	}
	if len(docs) == 0 {
		return nil, nil
	}
	sources := make([]domain.Source, 0, len(docs))
	for _, doc := range docs {
		s := domain.Source{}
		err = doc.DataTo(&s)
		if err != nil {
			return nil, err
		}
		sources = append(sources, s)
	}
	return sources, nil
}

func GetArticlesForOwner(ctx context.Context, ownerID string, start, end time.Time) ([]domain.Article, []domain.Source, error) {
	sources, err := GetSources(ctx, ownerID)
	if err != nil {
		return nil, nil, err
	}
	g := errgroup.Group{}
	articles := make(chan domain.Article, 1024)
	for _, s := range sources {
		s := s
		g.Go(func() error {
			docs, err := client.Collection("articles").
				Where("Source.FeedURL", "==", s.FeedURL).
				Where("Timestamp", ">", start).
				Where("Timestamp", "<", end).
				Documents(ctx).
				GetAll()
			if err != nil {
				return err
			}
			for _, doc := range docs {
				a := domain.Article{}
				err = doc.DataTo(&a)
				if err != nil {
					return err
				}
				articles <- a
			}
			return nil
		})
	}
	go func() {
		err = g.Wait()
		close(articles)
	}()
	out := []domain.Article{}
	for article := range articles {
		out = append(out, article)
	}
	if err != nil {
		return nil, nil, err
	}
	return out, sources, nil
}
