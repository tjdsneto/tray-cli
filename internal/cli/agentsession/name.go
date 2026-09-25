package agentsession

import (
	"fmt"
	"strings"
)

const NamePrefix = "agent-session:"

func NormalizeID(id string) (string, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return "", fmt.Errorf("agent session id is empty")
	}
	return id, nil
}

func TrayName(id string) (string, error) {
	id, err := NormalizeID(id)
	if err != nil {
		return "", err
	}
	return NamePrefix + id, nil
}

func ParseTrayName(name string) (id string, ok bool) {
	name = strings.TrimSpace(name)
	if !strings.HasPrefix(name, NamePrefix) {
		return "", false
	}
	id = strings.TrimSpace(strings.TrimPrefix(name, NamePrefix))
	if id == "" {
		return "", false
	}
	return id, true
}

func IsAgentSessionTrayName(name string) bool {
	_, ok := ParseTrayName(name)
	return ok
}
