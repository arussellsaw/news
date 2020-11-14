package domain

import (
	"bytes"
	"cloud.google.com/go/pubsub"
	"context"
	"encoding/json"
	"github.com/monzo/slog"
	"net/http"
)

// PubSubMessage is the payload of a Pub/Sub event.
type PubSubMessage struct {
	Message      pubsub.Message `json:"message"`
	Subscription string         `json:"subscription"`
}

type Publisher interface {
	Publish(ctx context.Context, topic string, payload interface{}) error
}

func NewPubSubPublisher(ctx context.Context) (Publisher, error) {
	ps, err := pubsub.NewClient(ctx, "russellsaw")
	if err != nil {
		return nil, err
	}
	return &PubSubPublisher{
		ps: ps,
	}, nil
}

type PubSubPublisher struct {
	ps *pubsub.Client
}

func (p *PubSubPublisher) Publish(ctx context.Context, topic string, payload interface{}) error {
	t := p.ps.Topic(topic)
	msg, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	result := t.Publish(ctx, &pubsub.Message{
		Data: msg,
	})
	_, err = result.Get(ctx)
	return err
}

type HTTPPublisher struct {
	ArticleURL string
	SourceURL  string
}

func (p *HTTPPublisher) Publish(ctx context.Context, topic string, payload interface{}) error {
	var url string
	switch topic {
	case "articles":
		url = p.ArticleURL
	case "sources":
		url = p.SourceURL
	}

	pbuf, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	buf, err := json.Marshal(PubSubMessage{Message: pubsub.Message{Data: pbuf}})
	if err != nil {
		return err
	}
	go func() {
		_, err := http.Post(url, "application/json", bytes.NewReader(buf))
		if err != nil {
			slog.Error(ctx, "Error publishing event: %s", err)
		}
	}()
	return nil
}
