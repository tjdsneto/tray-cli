//go:build unix

package commands

import (
	"os"
	"syscall"
)

func pidAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	if pid == os.Getpid() {
		return true
	}
	return syscall.Kill(pid, 0) == nil
}
