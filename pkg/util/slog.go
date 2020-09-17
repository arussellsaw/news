package util

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/monzo/slog"
)

type ContextParamLogger struct {
	slog.Logger
}

func (l ContextParamLogger) Log(evs ...slog.Event) {
	for i, e := range evs {
		params := Params(e.Context)
		if params == nil {
			continue
		}

		for k, v := range e.Metadata {
			params[k] = v
		}
		evs[i].Metadata = params
	}
	l.Logger.Log(evs...)
}

// StackDriverLogger is an implementation of monzo/slog.Logger
// that emits stackdriver compatible events
type StackDriverLogger struct {
	mu     sync.Mutex
	buffer []slog.Event
}

func (l *StackDriverLogger) Log(evs ...slog.Event) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, e := range evs {
		fmt.Println(NewEntry(e))
	}
}

func (l *StackDriverLogger) Flush() error {
	return nil
}

// Entry ...
type Entry struct {
	Message  string `json:"message"`
	Severity string `json:"severity,omitempty"`
	Trace    string `json:"logging.googleapis.com/trace,omitempty"`

	Params map[string]string `json:"params,omitempty"`
}

// String renders an entry structure to the JSON format expected by Stackdriver.
func (e Entry) String() string {
	if e.Severity == "" {
		e.Severity = "INFO"
	}
	out, err := json.Marshal(e)
	if err != nil {
		fmt.Println("json.Marshal:", err)
	}
	return string(out)
}

func NewEntry(e slog.Event) Entry {
	return Entry{
		Trace:    Trace(e.Context),
		Severity: e.Severity.String(),
		Message:  e.Message,
		Params:   e.Metadata,
	}
}
