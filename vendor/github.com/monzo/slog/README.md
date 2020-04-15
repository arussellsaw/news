# `slog`
**S**tructured **log**ging.

slog is a library for capturing structured log information. In contrast to "traditional" logging libraries, slog:

* captures a [Context](https://golang.org/pkg/context/) for each event
* captures arbitrary key-value metadata on each log event

slog forwards messages to [`log`](https://golang.org/pkg/log/) by default. But you probably want to write a a custom output to make use of the context and metadata. At [Monzo](https://monzo.com/), slog captures events both on a per-service and a per-request basis (using the context information) and sends them to a centralised logging system. This lets us view all the logs for a given request across all the micro-services it touches.
