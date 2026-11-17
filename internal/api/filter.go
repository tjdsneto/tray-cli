package api

import "sort"

func SelectItems(items []Item, filter string) []Item {
	out := make([]Item, 0, len(items))
	for _, it := range items {
		if filter == "all" || openWork(it) {
			out = append(out, it)
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})
	return out
}

func openWork(it Item) bool {
	switch it.Scope {
	case "local":
		return it.Status == "open"
	case "remote":
		switch it.Status {
		case "pending", "accepted", "snoozed":
			return true
		}
	}
	return false
}