//go:build applefm

package fm_test

import (
	"errors"
	"testing"

	"github.com/Obedience-Corp/apple-fm-go/pkg/fm"
)

func TestRealFMAvailable(t *testing.T) {
	status, err := fm.New("").Available(t.Context())
	switch {
	case errors.Is(err, fm.ErrNotFound):
		t.Skipf("fm is not installed: %v", err)
	case errors.Is(err, fm.ErrUnavailable), errors.Is(err, fm.ErrLicense):
		t.Skipf("fm reports the model is not usable: %v", err)
	case err != nil:
		t.Fatalf("Available: %v", err)
	}
	if !status.Ready {
		t.Fatalf("Available returned no error but Ready is false: %+v", status)
	}
	t.Logf("fm available: %q", status.Stdout)
}
