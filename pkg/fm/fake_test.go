package fm

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

const (
	envFake         = "APPLE_FM_GO_FAKE"
	envArgv         = "APPLE_FM_GO_FAKE_ARGV"
	envStdout       = "APPLE_FM_GO_FAKE_STDOUT"
	envStderr       = "APPLE_FM_GO_FAKE_STDERR"
	envExit         = "APPLE_FM_GO_FAKE_EXIT"
	envSleep        = "APPLE_FM_GO_FAKE_SLEEP"
	envEchoEnv      = "APPLE_FM_GO_FAKE_ECHO_ENV"
	envPIDFile      = "APPLE_FM_GO_FAKE_PID_FILE"
	envChildPIDFile = "APPLE_FM_GO_FAKE_CHILD_PID_FILE"
)

func TestMain(m *testing.M) {
	if os.Getenv(envFake) == "1" {
		os.Exit(runFake(os.Args[1:]))
	}
	os.Exit(m.Run())
}

func runFake(args []string) int {
	if path := os.Getenv(envArgv); path != "" {
		data, err := json.Marshal(args)
		if err != nil {
			return 97
		}
		if err := os.WriteFile(path, data, 0o600); err != nil {
			return 97
		}
	}
	if path := os.Getenv(envPIDFile); path != "" {
		if err := writeAtomic(path, strconv.Itoa(os.Getpid())); err != nil {
			return 98
		}
	}
	if path := os.Getenv(envChildPIDFile); path != "" {
		child := exec.Command(os.Args[0])
		child.Env = append(os.Environ(), envChildPIDFile+"=", envArgv+"=", envPIDFile+"="+path)
		if err := child.Start(); err != nil {
			return 99
		}
	}
	stdout := os.Getenv(envStdout)
	if key := os.Getenv(envEchoEnv); key != "" {
		stdout = os.Getenv(key) + stdout
	}
	if _, err := os.Stdout.WriteString(stdout); err != nil {
		return 96
	}
	if _, err := os.Stderr.WriteString(os.Getenv(envStderr)); err != nil {
		return 96
	}
	if d, err := time.ParseDuration(os.Getenv(envSleep)); err == nil {
		time.Sleep(d)
	}
	code, _ := strconv.Atoi(os.Getenv(envExit))
	return code
}

func writeAtomic(path, content string) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(content), 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

type fake struct {
	stdout       string
	stderr       string
	exit         int
	sleep        time.Duration
	echoEnv      string
	pidFile      string
	childPIDFile string
}

func (f fake) env(argvPath string) []string {
	env := append(os.Environ(),
		envFake+"=1",
		envArgv+"="+argvPath,
		envStdout+"="+f.stdout,
		envStderr+"="+f.stderr,
		envExit+"="+strconv.Itoa(f.exit),
		envEchoEnv+"="+f.echoEnv,
		envPIDFile+"="+f.pidFile,
		envChildPIDFile+"="+f.childPIDFile,
	)
	if f.sleep > 0 {
		env = append(env, envSleep+"="+f.sleep.String())
	}
	return env
}

func testBinary(t *testing.T) string {
	t.Helper()
	bin, err := os.Executable()
	if err != nil {
		t.Fatalf("os.Executable: %v", err)
	}
	return bin
}

func newFakeClient(t *testing.T, f fake) (*Client, string) {
	t.Helper()
	argvPath := filepath.Join(t.TempDir(), "argv.json")
	client := New(testBinary(t))
	client.Env = f.env(argvPath)
	return client, argvPath
}

func readArgv(t *testing.T, path string) []string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("fake fm did not record argv: %v", err)
	}
	var args []string
	if err := json.Unmarshal(data, &args); err != nil {
		t.Fatalf("decode argv %q: %v", data, err)
	}
	return args
}

func assertNotRun(t *testing.T, argvPath string) {
	t.Helper()
	if _, err := os.Stat(argvPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("fake fm ran, argv file stat err = %v", err)
	}
}

func asFMError(t *testing.T, err error) *Error {
	t.Helper()
	var fmErr *Error
	if !errors.As(err, &fmErr) {
		t.Fatalf("error %v (%T) is not *fm.Error", err, err)
	}
	return fmErr
}
