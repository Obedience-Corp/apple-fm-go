package fm

import (
	"errors"
	"os/exec"
	"slices"
	"strings"
	"testing"
)

func TestNewDefaultsToFM(t *testing.T) {
	t.Parallel()
	if got := New("").Bin; got != DefaultBin {
		t.Fatalf("New(\"\").Bin = %q, want %q", got, DefaultBin)
	}
	if got := New("/opt/fm").Bin; got != "/opt/fm" {
		t.Fatalf("New(path).Bin = %q", got)
	}
}

func TestAvailableReady(t *testing.T) {
	t.Parallel()
	client, argvPath := newFakeClient(t, fake{stdout: "available\n"})
	status, err := client.Available(t.Context())
	if err != nil {
		t.Fatalf("Available: %v", err)
	}
	if !status.Ready || status.Stdout != "available\n" {
		t.Fatalf("status = %+v", status)
	}
	if got := readArgv(t, argvPath); !slices.Equal(got, []string{"available"}) {
		t.Fatalf("argv = %q", got)
	}
}

func TestAvailableNotReady(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		stderr string
		want   error
	}{
		{"license", "Error: License not accepted. Run sudo fm license.\n", ErrLicense},
		{"model not ready", "Error: modelNotReady\n", ErrUnavailable},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			client, _ := newFakeClient(t, fake{stdout: "unavailable\n", stderr: tt.stderr, exit: 1})
			status, err := client.Available(t.Context())
			if !errors.Is(err, tt.want) {
				t.Fatalf("err = %v, want %v", err, tt.want)
			}
			if status.Ready || status.Stdout != "unavailable\n" {
				t.Fatalf("status = %+v", status)
			}
			if fmErr := asFMError(t, err); fmErr.ExitCode != 1 || fmErr.Stderr != tt.stderr {
				t.Fatalf("error fields = %+v", fmErr)
			}
		})
	}
}

func TestMissingBinaryReturnsErrNotFound(t *testing.T) {
	t.Parallel()
	for _, bin := range []string{"apple-fm-go-no-such-binary", "/nonexistent/apple-fm-go/fm"} {
		t.Run(bin, func(t *testing.T) {
			t.Parallel()
			client := New(bin)
			_, err := client.Available(t.Context())
			if !errors.Is(err, ErrNotFound) {
				t.Fatalf("err = %v, want ErrNotFound", err)
			}
			if !strings.Contains(err.Error(), bin) {
				t.Fatalf("error %q does not name %q", err, bin)
			}
			var execErr *exec.Error
			if !errors.As(err, &execErr) || execErr.Name != bin {
				t.Fatalf("error does not wrap *exec.Error for %q: %v", bin, err)
			}
		})
	}
}

func TestEnvNilPassesThroughProcessEnvironment(t *testing.T) {
	t.Setenv(envFake, "1")
	t.Setenv(envEchoEnv, "APPLE_FM_GO_PROBE")
	t.Setenv("APPLE_FM_GO_PROBE", "from-parent")
	client := New(testBinary(t))
	resp, err := client.Respond(t.Context(), "hi", nil)
	if err != nil {
		t.Fatalf("Respond: %v", err)
	}
	if resp.Stdout != "from-parent" {
		t.Fatalf("Stdout = %q, want the parent's value", resp.Stdout)
	}
}

func TestEnvSetReplacesProcessEnvironment(t *testing.T) {
	t.Setenv("APPLE_FM_GO_PROBE", "from-parent")
	client := New(testBinary(t))
	client.Env = []string{envFake + "=1", envEchoEnv + "=APPLE_FM_GO_PROBE"}
	resp, err := client.Respond(t.Context(), "hi", nil)
	if err != nil {
		t.Fatalf("Respond: %v", err)
	}
	if resp.Stdout != "" {
		t.Fatalf("Stdout = %q, want empty because Env omits the probe", resp.Stdout)
	}
}
