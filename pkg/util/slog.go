package util

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"github.com/fatih/color"
	"github.com/monzo/slog"
)

type ColourLogger struct {
	Writer io.Writer
}

func (l ColourLogger) Log(evs ...slog.Event) {
	for _, e := range evs {
		switch e.Severity {
		case slog.TraceSeverity:
			fmt.Fprintf(l.Writer, "%s: %s\n", color.WhiteString("%s TRC", e.Timestamp.Format("15:04:05.000")), e.Message)
		case slog.DebugSeverity:
			fmt.Fprintf(l.Writer, "%s: %s\n", color.CyanString("%s DBG", e.Timestamp.Format("15:04:05.000")), e.Message)
		case slog.InfoSeverity:
			fmt.Fprintf(l.Writer, "%s: %s\n", color.BlueString("%s INF", e.Timestamp.Format("15:04:05.000")), e.Message)
		case slog.WarnSeverity:
			fmt.Fprintf(l.Writer, "%s: %s\n", color.YellowString("%s WRN", e.Timestamp.Format("15:04:05.000")), e.Message)
		case slog.ErrorSeverity:
			fmt.Fprintf(l.Writer, "%s: %s\n", color.RedString("%s ERR", e.Timestamp.Format("15:04:05.000")), e.Message)
		case slog.CriticalSeverity:
			fmt.Fprintf(l.Writer, "%s: %s\n", color.RedString("%s CRT", e.Timestamp.Format("15:04:05.000")), e.Message)
		}
	}
}

func (l ColourLogger) Flush() error { return nil }

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

	Params map[string]interface{} `json:"params,omitempty"`
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
