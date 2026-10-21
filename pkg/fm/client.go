// Package fm is an unofficial Go client for the fm command-line tool that
// ships with macOS 27 and runs Apple's on-device Foundation Model.
//
// The client shells out to fm. It does not link FoundationModels or embed a
// model. The subcommand and flag spellings it sends are provisional until they
// are checked against fm --help on macOS 27.
package fm

import (
	"bytes"
	"cmp"
	"context"
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"slices"
	"strings"
	"time"
)

// DefaultBin is the executable name looked up on PATH when Client.Bin is empty.
const DefaultBin = "fm"

const waitDelay = 5 * time.Second

// Client runs the non-interactive fm subcommands.
type Client struct {
	// Bin is the fm executable, either a name looked up on PATH or a path.
	// An empty Bin means DefaultBin.
	Bin string
	// Env is the child process environment in os/exec form. A nil Env passes
	// through os.Environ().
	Env []string
}

// New returns a Client that runs bin. An empty bin means DefaultBin on PATH.
func New(bin string) *Client {
	return &Client{Bin: cmp.Or(bin, DefaultBin)}
}

type output struct {
	stdout string
	stderr string
}

func (c *Client) run(ctx context.Context, args []string) (output, error) {
	op := opName(args)
	path, err := exec.LookPath(cmp.Or(c.Bin, DefaultBin))
	if err != nil {
		return output{}, &Error{Op: op, Kind: ErrNotFound, ExitCode: -1, Err: err}
	}

	cmd := exec.CommandContext(ctx, path, args...)
	cmd.Env = c.Env
	if cmd.Env == nil {
		cmd.Env = os.Environ()
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.WaitDelay = waitDelay
	configureProcess(cmd)

	runErr := cmd.Run()
	out := output{stdout: stdout.String(), stderr: stderr.String()}
	if runErr == nil {
		return out, nil
	}

	e := &Error{Op: op, ExitCode: -1, Stdout: out.stdout, Stderr: out.stderr, Err: runErr}
	if ctxErr := ctx.Err(); ctxErr != nil {
		e.Err = ctxErr
		return out, e
	}
	var exitErr *exec.ExitError
	switch {
	case errors.As(runErr, &exitErr):
		e.ExitCode = exitErr.ExitCode()
		e.Kind = classify(e.ExitCode, out.stderr)
	case errors.Is(runErr, fs.ErrNotExist):
		e.Kind = ErrNotFound
	}
	return out, e
}

func opName(args []string) string {
	end := slices.IndexFunc(args, func(a string) bool { return strings.HasPrefix(a, "-") })
	if end < 0 {
		end = len(args)
	}
	return strings.Join(args[:end], " ")
}

// Availability reports whether the on-device model can answer prompts.
type Availability struct {
	// Ready is true when fm available exited with status 0.
	Ready bool
	// Stdout is fm available's output, unparsed.
	Stdout string
}

// Available runs fm available. When the model is not ready, Available returns
// Ready false together with an error matching ErrLicense or ErrUnavailable.
// A non-zero exit that matches neither is returned unclassified with its
// stderr attached. A missing executable returns ErrNotFound.
func (c *Client) Available(ctx context.Context) (Availability, error) {
	out, err := c.run(ctx, []string{cmdAvailable})
	status := Availability{Stdout: out.stdout}
	if err != nil {
		return status, err
	}
	status.Ready = true
	return status, nil
}
