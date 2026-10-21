package fm

import (
	"errors"
	"slices"
	"testing"
)

func TestCountTokens(t *testing.T) {
	t.Parallel()
	client, argvPath := newFakeClient(t, fake{stdout: "  42\n"})
	n, err := client.CountTokens(t.Context(), CountRequest{Text: "- a\n- b"})
	if err != nil {
		t.Fatalf("CountTokens: %v", err)
	}
	if n != 42 {
		t.Fatalf("n = %d, want 42", n)
	}
	if got, want := readArgv(t, argvPath), []string{"count-tokens", "--text=- a\n- b"}; !slices.Equal(got, want) {
		t.Fatalf("argv\n got %q\nwant %q", got, want)
	}
}

func TestCountTokensGarbage(t *testing.T) {
	t.Parallel()
	for _, stdout := range []string{"42 tokens\n", "", "forty-two", "-1", "4 2"} {
		t.Run(stdout, func(t *testing.T) {
			t.Parallel()
			client, _ := newFakeClient(t, fake{stdout: stdout})
			n, err := client.CountTokens(t.Context(), CountRequest{Text: "doc"})
			if !errors.Is(err, ErrTokenCount) {
				t.Fatalf("CountTokens(%q) = %d, %v; want ErrTokenCount", stdout, n, err)
			}
			if fmErr := asFMError(t, err); fmErr.Stdout != stdout || fmErr.Op != "count-tokens" {
				t.Fatalf("error fields = %+v", fmErr)
			}
		})
	}
}

func TestCountTokensFailure(t *testing.T) {
	t.Parallel()
	client, _ := newFakeClient(t, fake{stderr: "Error: modelNotReady\n", exit: 2})
	if _, err := client.CountTokens(t.Context(), CountRequest{Text: "doc"}); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("err = %v, want ErrUnavailable", err)
	}
}
