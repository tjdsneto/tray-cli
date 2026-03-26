package commands

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/tjdsneto/tray-cli/internal/adapters/postgrest/pghttp"
	"github.com/tjdsneto/tray-cli/internal/auth"
	"github.com/tjdsneto/tray-cli/internal/cli/errs"
	"github.com/tjdsneto/tray-cli/internal/config"
	"github.com/tjdsneto/tray-cli/internal/credentials"
	"github.com/tjdsneto/tray-cli/internal/domain"
)

func refreshSessionAfter401(ctx context.Context) (domain.Session, error) {
	rawURL := config.SupabaseURL()
	anon := config.SupabaseAnonKey()
	if rawURL == "" || anon == "" {
		return domain.Session{}, fmt.Errorf("%w", errs.MissingBackendConfig)
	}
	f, err := credentials.Load(cmdDeps.ConfigDir())
	if err != nil {
		return domain.Session{}, err
	}
	cctx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()
	f, err = auth.RefreshSessionTokens(cctx, rawURL, anon, http.DefaultClient, cmdDeps.ConfigDir(), f)
	if err != nil {
		return domain.Session{}, err
	}
	return f.Session(), nil
}

func listItemsWithRetry(ctx context.Context, svcs domain.Services, sess domain.Session, q domain.ListItemsQuery) ([]domain.Item, domain.Session, error) {
	items, err := svcs.Items.List(ctx, sess, q)
	if err == nil {
		return items, sess, nil
	}
	if !errors.Is(err, pghttp.ErrUnauthorized) {
		return nil, sess, err
	}
	ns, err2 := refreshSessionAfter401(ctx)
	if err2 != nil {
		return nil, sess, err2
	}
	items, err = svcs.Items.List(ctx, ns, q)
	return items, ns, err
}

func listOutboxWithRetry(ctx context.Context, svcs domain.Services, sess domain.Session) ([]domain.Item, domain.Session, error) {
	items, err := svcs.Items.ListOutbox(ctx, sess)
	if err == nil {
		return items, sess, nil
	}
	if !errors.Is(err, pghttp.ErrUnauthorized) {
		return nil, sess, err
	}
	ns, err2 := refreshSessionAfter401(ctx)
	if err2 != nil {
		return nil, sess, err2
	}
	items, err = svcs.Items.ListOutbox(ctx, ns)
	return items, ns, err
}
