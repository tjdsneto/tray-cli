//go:build windows

package listenhook

import (
	"os/exec"
)

func runCmd(name string, args []string, env []string) error {
	c := exec.Command(name, args...)
	c.Env = env
	return c.Run()
}
