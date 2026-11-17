package api

func ActionsFor(scope, status string) []string {
	switch {
	case scope == "local" && status == "open":
		return []string{"complete"}
	case scope == "remote" && status == "pending":
		return []string{"accept", "decline", "snooze"}
	case scope == "remote" && status == "accepted":
		return []string{"complete", "archive"}
	case scope == "remote" && status == "snoozed":
		return []string{"archive"}
	default:
		return []string{}
	}
}