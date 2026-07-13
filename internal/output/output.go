// Package output renders command results. Under --json every response uses the
// consistent envelope from ADR 0015; otherwise output is human-friendly.
package output

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/danmatthews/wt/internal/apperr"
)

// JSON toggles machine output; bound to the persistent --json flag.
var JSON bool

// Emit prints a success result. Under --json it writes {ok:true, data, ...extra};
// otherwise it invokes human.
func Emit(data any, extra map[string]any, human func()) {
	if !JSON {
		human()
		return
	}
	env := map[string]any{"ok": true, "data": data}
	for k, v := range extra {
		env[k] = v
	}
	writeJSON(os.Stdout, env)
}

// Fail prints a failure. Under --json it writes {ok:false, error}; otherwise a
// human message to stderr. Callers still exit non-zero.
func Fail(err error) {
	e := apperr.Coerce(err)
	if !JSON {
		fmt.Fprintf(os.Stderr, "error: %s (%s)\n", e.Message, e.Code)
		return
	}
	writeJSON(os.Stderr, map[string]any{"ok": false, "error": e})
}

func writeJSON(f *os.File, v any) {
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	enc.Encode(v)
}
