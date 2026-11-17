package commands

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/tjdsneto/tray-cli/internal/api"
	"github.com/tjdsneto/tray-cli/internal/domain"
)

func cmdAPI() *cobra.Command {
	return &cobra.Command{
		Use:   "api",
		Short: "Serve the local desktop API on 127.0.0.1",
		RunE:  runAPI,
	}
}

func runAPI(cmd *cobra.Command, _ []string) error {
	token, err := randomToken()
	if err != nil {
		return err
	}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	port := ln.Addr().(*net.TCPAddr).Port
	fmt.Fprintln(cmd.OutOrStdout(), handshakeLine(port, token))

	h := api.NewHandler(token, newAPISource(cmdDeps.ConfigDir, cmdDeps.RequireAuth), 30*time.Second)
	h.Rebuild(cmd.Context())
	srv := &http.Server{Handler: h}
	go h.Loop(cmd.Context())

	ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		_, _ = io.Copy(io.Discard, cmd.InOrStdin())
		_ = srv.Shutdown(context.Background())
	}()
	go func() {
		<-ctx.Done()
		_ = srv.Shutdown(context.Background())
	}()
	err = srv.Serve(ln)
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}

func handshakeLine(port int, token string) string {
	b, _ := json.Marshal(struct {
		Port  int    `json:"port"`
		Token string `json:"token"`
	}{Port: port, Token: token})
	return string(b)
}

func randomToken() (string, error) {
	var b [32]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}

// newAPISource is a compile stub. Task 8 replaces it in api_source.go and deletes this function.
func newAPISource(func() string, func() (domain.Services, domain.Session, error)) api.Source {
	return nil
}
