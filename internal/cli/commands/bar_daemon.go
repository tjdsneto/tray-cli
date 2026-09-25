package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

const (
	barPidFile = "bar.pid"
	barLogFile = "bar.log"
)

func acquireBarDaemon(configDir string) (cleanup func(), err error) {
	pidPath := filepath.Join(configDir, barPidFile)
	if b, err := os.ReadFile(pidPath); err == nil {
		old, _ := strconv.Atoi(strings.TrimSpace(string(b)))
		if old > 0 && pidAlive(old) {
			return nil, fmt.Errorf("bar daemon already running (pid %d)", old)
		}
		_ = os.Remove(pidPath)
	} else if !os.IsNotExist(err) {
		return nil, err
	}

	f, err := os.OpenFile(pidPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return nil, err
	}
	_, werr := fmt.Fprintf(f, "%d\n", os.Getpid())
	if cerr := f.Close(); werr != nil || cerr != nil {
		_ = os.Remove(pidPath)
		if werr != nil {
			return nil, werr
		}
		return nil, cerr
	}

	return func() { _ = os.Remove(pidPath) }, nil
}

func redirectBarDaemonLog(cmd *cobra.Command, configDir string) (cleanup func(), err error) {
	logPath := filepath.Join(configDir, barLogFile)
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return nil, err
	}
	cmd.SetOut(f)
	cmd.SetErr(f)
	return func() { _ = f.Close() }, nil
}
