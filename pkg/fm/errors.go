package fm

import (
	"errors"
	"strconv"
	"strings"
)

// Sentinel errors returned by Client methods. They arrive wrapped in an *Error
// that carries the exit code and the raw output, so test them with errors.Is.
var (
	// ErrNotFound means the fm executable could not be found.
	ErrNotFound = errors.New("binary not found")
	// ErrLicense means the fm command-line terms have not been accepted.
	// Accepting them is an operator step: sudo fm license.
	ErrLicense = errors.New("license not accepted")
	// ErrUnavailable means the model is not ready: assets are missing, Apple
	// Intelligence is off, or the device is not eligible.
	ErrUnavailable = errors.New("model unavailable")
	// ErrContextWindow means the prompt does not fit the model's context window.
	ErrContextWindow = errors.New("context window exceeded")
	// ErrGuardrail means the model refused the prompt.
	ErrGuardrail = errors.New("guardrail violation")
	// ErrUnknownModel means fm rejected the model name.
	ErrUnknownModel = errors.New("unknown model")
	// ErrSchema means fm rejected the schema, or a schema field could not be
	// expressed as fm flags.
	ErrSchema = errors.New("schema rejected")
	// ErrSchemaOutput means fm printed something other than JSON for a
	// response or a schema.
	ErrSchemaOutput = errors.New("output is not JSON")
	// ErrTokenCount means fm count-tokens printed something other than one
	// non-negative integer.
	ErrTokenCount = errors.New("token count is not an integer")
)

// Error describes a failed fm invocation. Kind is one of the sentinel errors,
// or nil when the failure did not match a known category.
type Error struct {
	// Op is the fm subcommand, such as "respond" or "schema object".
	Op string
	// Kind is the matched sentinel error, or nil.
	Kind error
	// ExitCode is fm's exit status, or -1 when fm did not exit on its own.
	ExitCode int
	// Stdout is everything fm wrote to standard output.
	Stdout string
	// Stderr is everything fm wrote to standard error.
	Stderr string
	// Err is the underlying cause, such as an *exec.ExitError or a context error.
	Err error
}

// Error formats the operation, category, cause, and first line of stderr.
func (e *Error) Error() string {
	parts := []string{"fm " + e.Op}
	if e.Kind != nil {
		parts = append(parts, e.Kind.Error())
	}
	if e.Err != nil {
		parts = append(parts, e.Err.Error())
	} else if e.ExitCode > 0 {
		parts = append(parts, "exit status "+strconv.Itoa(e.ExitCode))
	}
	if line := firstLine(e.Stderr); line != "" {
		parts = append(parts, line)
	}
	return strings.Join(parts, ": ")
}

// Unwrap returns Kind and Err so errors.Is matches either.
func (e *Error) Unwrap() []error {
	var errs []error
	for _, err := range []error{e.Kind, e.Err} {
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func classify(exitCode int, stderr string) error {
	if kind, ok := exitCodeKinds[exitCode]; ok {
		return kind
	}
	for line := range strings.Lines(stderr) {
		line = stripLeadIns(strings.ToLower(strings.TrimSpace(line)))
		if line == "" {
			continue
		}
		for _, entry := range stderrPrefixes {
			for _, prefix := range entry.prefixes {
				if strings.HasPrefix(line, prefix) {
					return entry.kind
				}
			}
		}
	}
	return nil
}

func stripLeadIns(line string) string {
	for {
		trimmed := line
		for _, lead := range stderrLeadIns {
			if rest, ok := strings.CutPrefix(trimmed, lead); ok {
				trimmed = strings.TrimSpace(rest)
			}
		}
		if trimmed == line {
			return line
		}
		line = trimmed
	}
}

func firstLine(s string) string {
	for line := range strings.Lines(s) {
		if line = strings.TrimSpace(line); line != "" {
			return line
		}
	}
	return ""
}
