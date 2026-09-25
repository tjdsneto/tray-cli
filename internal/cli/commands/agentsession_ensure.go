package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/tjdsneto/tray-cli/internal/cli/agentsession"
	"github.com/tjdsneto/tray-cli/internal/cli/trayref"
	"github.com/tjdsneto/tray-cli/internal/domain"
)

// ensureAgentSessionTray returns the tray id for agent-session:<id>.
// Prefers an owned tray (create-on-add under the writer's account). If none exists but a
// joined tray with that name is visible via ListMine, returns that id for contribute
// without creating. Otherwise creates under the current user.
func ensureAgentSessionTray(ctx context.Context, svcs domain.Services, sess domain.Session, agentSessionID string) (string, error) {
	name, err := agentsession.TrayName(agentSessionID)
	if err != nil {
		return "", err
	}

	owned, err := svcs.Trays.ListOwned(ctx, sess)
	if err != nil {
		return "", err
	}
	id, err := pickSingleTrayID(owned, name)
	if err != nil {
		return "", err
	}
	if id != "" {
		return id, nil
	}

	mine, err := svcs.Trays.ListMine(ctx, sess)
	if err != nil {
		return "", err
	}
	id, err = pickSingleTrayID(mine, name)
	if err != nil {
		return "", err
	}
	if id != "" {
		return id, nil
	}

	created, err := svcs.Trays.Create(ctx, sess, name, nil)
	if err != nil {
		owned, listErr := svcs.Trays.ListOwned(ctx, sess)
		if listErr != nil {
			return "", err
		}
		id, pickErr := pickSingleTrayID(owned, name)
		if pickErr != nil {
			return "", pickErr
		}
		if id != "" {
			return id, nil
		}
		return "", err
	}
	return strings.TrimSpace(created.ID), nil
}

// pickSingleTrayID returns the id of the single name match, "" if none, or an error if many.
func pickSingleTrayID(trays []domain.Tray, name string) (string, error) {
	matches := trayref.FindTraysByNameFold(trays, name)
	switch len(matches) {
	case 0:
		return "", nil
	case 1:
		return strings.TrimSpace(matches[0].ID), nil
	default:
		return "", fmt.Errorf("multiple trays match %q — rename one in the dashboard or use a more specific name", strings.TrimSpace(name))
	}
}
