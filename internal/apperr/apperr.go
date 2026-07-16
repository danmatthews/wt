// Package apperr defines the structured error contract (ADR 0015): every
// failure carries a stable machine `code`, a human `message`, and optional
// `details`. All failures exit non-zero (1); callers branch on Code.
package apperr

import "fmt"

// Stable error codes. Additive only — never repurpose an existing code.
const (
	CodeNotInWorktree          = "not_in_worktree"
	CodeWorktreeNotRegistered  = "worktree_not_registered"
	CodeWorktreeNotFound       = "worktree_not_found"
	CodeNameConflict           = "name_conflict"
	CodeEntryPointNotFound     = "entry_point_not_found"
	CodeEntryPointNameConflict = "entry_point_name_conflict"
	CodeUnknownEntryPointType  = "unknown_entry_point_type"
	CodeLockTimeout            = "lock_timeout"
	CodeIOError                = "io_error"
	CodeUsage                  = "usage"
	CodeGitUnavailable         = "git_unavailable"
	CodeGitError               = "git_error"
)

// Error is a wt domain error.
type Error struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

func (e *Error) Error() string { return e.Message }

// New builds an Error with a printf-style message.
func New(code, format string, args ...any) *Error {
	return &Error{Code: code, Message: fmt.Sprintf(format, args...)}
}

// WithDetail attaches a context key/value and returns the error for chaining.
func (e *Error) WithDetail(key string, value any) *Error {
	if e.Details == nil {
		e.Details = map[string]any{}
	}
	e.Details[key] = value
	return e
}

// Coerce turns any error into an *Error, defaulting unknowns to io_error.
func Coerce(err error) *Error {
	if err == nil {
		return nil
	}
	if e, ok := err.(*Error); ok {
		return e
	}
	return New(CodeIOError, "%s", err.Error())
}
