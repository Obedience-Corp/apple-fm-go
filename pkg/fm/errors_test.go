package fm

import (
	"errors"
	"os/exec"
	"testing"
)

func TestClassifyStderrPrefixes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		stderr string
		want   error
	}{
		{"Error: License not accepted. Run `sudo fm license`.", ErrLicense},
		{"fm: error: appleIntelligenceNotEnabled", ErrUnavailable},
		{"Error: model is not ready", ErrUnavailable},
		{"Error: exceededContextWindowSize(4096)", ErrContextWindow},
		{"guardrailViolation(\"unsafe\")", ErrGuardrail},
		{"Error: Unknown model 'pcc'", ErrUnknownModel},
		{"Error: the only model is 'system'", ErrUnknownModel},
		{"Error: invalid schema at schema.json", ErrSchema},
		{"Usage: fm respond <prompt>\nError: guardrailViolation", ErrGuardrail},
		{"Error: the prompt mentioned guardrailViolation", nil},
		{"Error: Unknown option '--frobnicate'", nil},
		{"", nil},
	}
	for _, tt := range tests {
		if got := classify(1, tt.stderr); !errors.Is(got, tt.want) || (tt.want == nil && got != nil) {
			t.Errorf("classify(%q) = %v, want %v", tt.stderr, got, tt.want)
		}
	}
}

func TestUnclassifiedFailureKeepsExitCodeAndStderr(t *testing.T) {
	t.Parallel()
	const stderr = "Error: Unknown option '--frobnicate'\nUsage: fm respond\n"
	client, _ := newFakeClient(t, fake{stderr: stderr, exit: 64})
	_, err := client.Respond(t.Context(), "hi", nil)
	fmErr := asFMError(t, err)
	if fmErr.Kind != nil || fmErr.ExitCode != 64 || fmErr.Stderr != stderr {
		t.Fatalf("error fields = %+v", fmErr)
	}
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("error does not wrap *exec.ExitError: %v", err)
	}
	const want = "fm respond: exit status 64: Error: Unknown option '--frobnicate'"
	if err.Error() != want {
		t.Fatalf("Error() = %q, want %q", err.Error(), want)
	}
}

func TestErrorMessageIncludesKind(t *testing.T) {
	t.Parallel()
	err := &Error{Op: "respond", Kind: ErrGuardrail, ExitCode: 1, Stderr: "\nguardrailViolation\n"}
	const want = "fm respond: guardrail violation: exit status 1: guardrailViolation"
	if err.Error() != want {
		t.Fatalf("Error() = %q, want %q", err.Error(), want)
	}
	if !errors.Is(err, ErrGuardrail) || errors.Is(err, ErrSchema) {
		t.Fatalf("errors.Is mismatch for %v", err)
	}
}
