//go:build unix

package fm

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
	"testing"
	"time"
)

func TestCancelKillsProcessGroup(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	childPIDFile := filepath.Join(dir, "child.pid")
	grandchildPIDFile := filepath.Join(dir, "grandchild.pid")
	client, _ := newFakeClient(t, fake{
		sleep:        time.Minute,
		pidFile:      childPIDFile,
		childPIDFile: grandchildPIDFile,
	})

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	done := make(chan error, 1)
	go func() {
		_, err := client.Respond(ctx, "sleep", nil)
		done <- err
	}()

	pids := []int{waitForPIDFile(t, childPIDFile, done), waitForPIDFile(t, grandchildPIDFile, done)}
	t.Cleanup(func() {
		for _, pid := range pids {
			_ = syscall.Kill(pid, syscall.SIGKILL)
		}
	})

	start := time.Now()
	cancel()
	var err error
	select {
	case err = <-done:
	case <-time.After(15 * time.Second):
		t.Fatal("Respond did not return within 15s of cancel")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
	if elapsed := time.Since(start); elapsed > 10*time.Second {
		t.Fatalf("Respond returned %v after cancel, want prompt kill", elapsed)
	}
	for _, pid := range pids {
		waitForExit(t, pid)
	}
}

func waitForPIDFile(t *testing.T, path string, done <-chan error) int {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		select {
		case err := <-done:
			t.Fatalf("Respond returned before the fake started: %v", err)
		default:
		}
		if data, err := os.ReadFile(path); err == nil {
			pid, err := strconv.Atoi(string(data))
			if err != nil {
				t.Fatalf("pid file %s: %v", path, err)
			}
			return pid
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s", path)
	return 0
}

func waitForExit(t *testing.T, pid int) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		if err := syscall.Kill(pid, 0); errors.Is(err, syscall.ESRCH) {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("process %d still exists after cancel", pid)
}
