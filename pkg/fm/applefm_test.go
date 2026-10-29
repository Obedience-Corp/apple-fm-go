//go:build applefm

package fm_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Obedience-Corp/apple-fm-go/pkg/fm"
)

// Native tests use no filesystem fixtures. Fake-process tests run in Docker.
func nativeContext(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Minute)
	t.Cleanup(cancel)
	return ctx
}

func nativeClient(t *testing.T) *fm.Client {
	t.Helper()
	client := fm.New("")
	status, err := client.Available(nativeContext(t))
	if err != nil {
		if os.Getenv("APPLE_FM_GO_REQUIRE_NATIVE") != "1" &&
			(errors.Is(err, fm.ErrNotFound) || errors.Is(err, fm.ErrUnavailable) || errors.Is(err, fm.ErrLicense)) {
			t.Skipf("native prerequisite unavailable (not compatibility evidence): %v", err)
		}
		t.Fatalf("Available: %v", err)
	}
	if !status.Ready {
		t.Fatal("Available returned no error but Ready is false")
	}
	return client
}

func TestRealFMAvailable(t *testing.T) {
	nativeClient(t)
}

func TestRealFMRespond(t *testing.T) {
	client := nativeClient(t)
	response, err := client.Respond(nativeContext(t), "Say hello in one short sentence.", &fm.RespondOptions{
		Instructions: "Answer briefly in plain text.",
	})
	if err != nil {
		t.Fatalf("Respond: %v", err)
	}
	if strings.TrimSpace(response.Stdout) == "" {
		t.Fatal("Respond returned an empty reply")
	}
}

func TestRealFMSchemaObject(t *testing.T) {
	client := nativeClient(t)
	schema, err := client.SchemaObject(nativeContext(t), fm.ObjectSchema{
		Name:   "Items",
		Fields: []fm.Field{{Name: "names", Type: fm.String, Repeated: true}},
	})
	if err != nil {
		t.Fatalf("SchemaObject: %v", err)
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(schema, &object); err != nil || len(object) == 0 {
		t.Fatalf("SchemaObject did not return a nonempty JSON object: %q (%v)", schema, err)
	}
}

func TestRealFMCountTokens(t *testing.T) {
	client := nativeClient(t)
	n, err := client.CountTokens(nativeContext(t), fm.CountRequest{Text: "Hello from Go."})
	if err != nil {
		t.Fatalf("CountTokens: %v", err)
	}
	if n <= 0 {
		t.Fatalf("CountTokens = %d for nonempty input, want a positive count", n)
	}
}
