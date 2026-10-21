//go:build !unix

package fm

import "os/exec"

func configureProcess(*exec.Cmd) {}
