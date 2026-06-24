package provider

import (
	"encoding/json"
	"fmt"
)

func idFromDomainData(raw json.RawMessage, name string) (string, error) {
	var one map[string]any
	if err := json.Unmarshal(raw, &one); err == nil {
		if id, ok := one["id"]; ok {
			return scalarID(id)
		}
	}
	var many struct {
		Domains []map[string]any `json:"domains"`
	}
	if err := json.Unmarshal(raw, &many); err == nil {
		for _, domain := range many.Domains {
			if domain["domainName"] == name || domain["domain_name"] == name {
				return scalarID(domain["id"])
			}
		}
	}
	return "", fmt.Errorf("domain id for %q was not present in response", name)
}

func scalarID(value any) (string, error) {
	switch typed := value.(type) {
	case float64:
		return fmtInt(int64(typed)), nil
	case string:
		return typed, nil
	default:
		return "", fmt.Errorf("unsupported id type %T", value)
	}
}
