package domain

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/google/uuid"
	"github.com/monzo/slog"
)

type analyticsEvent struct {
	UserID             string
	InsertionTimestamp time.Time
	Payload            map[string]string
}

// Save implements the ValueSaver interface.
func (e analyticsEvent) Save() (map[string]bigquery.Value, string, error) {
	p, err := json.Marshal(e.Payload)
	if err != nil {
		return nil, "", err
	}
	return map[string]bigquery.Value{
		"user_id":             e.UserID,
		"insertion_timestamp": e.InsertionTimestamp,
		"payload":             string(p),
	}, "", nil
}

func AnalyticsMiddleware(h http.HandlerFunc) (http.HandlerFunc, error) {
	projectID := "russellsaw"
	datasetID := "news"
	tableID := "analytics_raw"
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("bigquery.NewClient: %v", err)
	}
	defer client.Close()

	return func(w http.ResponseWriter, r *http.Request) {
		e := analyticsEvent{
			InsertionTimestamp: time.Now(),
		}
		c, err := r.Cookie("uid")
		if err == http.ErrNoCookie {
			c = &http.Cookie{Name: "uid", Value: uuid.New().String()}
			http.SetCookie(w, c)
		} else if err != nil {
			http.Error(w, "internal server error", 500)
			slog.Error(r.Context(), "Error setting user_id: %s", err)
			return
		}
		e.UserID = c.Value
		h(w, r)
		e.Payload = make(map[string]string)
		e.Payload["path"] = r.URL.Path
		e.Payload["referer"] = r.Referer()
		inserter := client.Dataset(datasetID).Table(tableID).Inserter()
		if err := inserter.Put(ctx, e); err != nil {
			slog.Error(r.Context(), "Error inserting data: %s", err)
		}
	}, nil
}
