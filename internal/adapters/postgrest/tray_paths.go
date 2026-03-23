package postgrest

import (
	"net/url"
	"strings"
)

const trayMemberSelectColumns = "id,tray_id,user_id,joined_at,invited_via"

func traysByIDPath(trayID string) string {
	q := url.Values{}
	q.Set("id", "eq."+strings.TrimSpace(trayID))
	return "/rest/v1/trays?" + q.Encode()
}

func trayMembersListPath(trayID string) string {
	q := url.Values{}
	q.Set("tray_id", "eq."+strings.TrimSpace(trayID))
	q.Set("select", trayMemberSelectColumns)
	q.Set("order", "joined_at.asc")
	return "/rest/v1/tray_members?" + q.Encode()
}

func trayMembersDeletePath(trayID, userID string) string {
	q := url.Values{}
	q.Set("tray_id", "eq."+strings.TrimSpace(trayID))
	q.Set("user_id", "eq."+strings.TrimSpace(userID))
	return "/rest/v1/tray_members?" + q.Encode()
}
