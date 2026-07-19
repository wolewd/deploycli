// Package output centralises progress and result reporting.
// Progress on stderr, machine-parseable results on stdout.
package output

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Logger reports progress and results, switching format based on JSON mode.
type Logger struct {
	JSON bool
}

// NewLogger returns a Logger. jsonMode toggles between human-readable stderr
// and single-line JSON on stdout.
func NewLogger(jsonMode bool) *Logger {
	return &Logger{JSON: jsonMode}
}

// Step logs a free-form human-readable progress message to stderr. It is a
// no-op in JSON mode, since JSON consumers only care about structured step
// results (free text would break line-delimited JSON parsing).
func (l *Logger) Step(format string, args ...any) {
	if l.JSON {
		return
	}
	fmt.Fprintf(os.Stderr, "==> "+format+"\n", args...)
}

// Error reports a failure. In JSON mode it emits a structured error line on
// stdout; otherwise it prints a human-readable line to stderr.
func (l *Logger) Error(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	if l.JSON {
		l.emit(map[string]any{
			"status": "error",
			"error":  msg,
		})
		return
	}
	fmt.Fprintf(os.Stderr, "ERROR: %s\n", msg)
}

// StepResult reports the outcome of a completed step, e.g.
// {"step":"build","status":"ok","duration_ms":4210}
func (l *Logger) StepResult(step string, ok bool, duration time.Duration, extra map[string]any) {
	status := "ok"
	if !ok {
		status = "error"
	}
	if l.JSON {
		payload := map[string]any{
			"step":        step,
			"status":      status,
			"duration_ms": duration.Milliseconds(),
		}
		for k, v := range extra {
			payload[k] = v
		}
		l.emit(payload)
		return
	}
	fmt.Fprintf(os.Stderr, "==> %s: %s (%dms)\n", step, status, duration.Milliseconds())
}

func (l *Logger) emit(payload map[string]any) {
	b, err := json.Marshal(payload)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: failed to marshal json output: %v\n", err)
		return
	}
	fmt.Fprintln(os.Stdout, string(b))
}
