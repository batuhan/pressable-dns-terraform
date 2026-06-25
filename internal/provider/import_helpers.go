package provider

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func splitImportID(id string, parts int, example string) ([]string, error) {
	pieces := strings.Split(id, "/")
	if len(pieces) != parts {
		return nil, fmt.Errorf("expected import id %q, got %q", example, id)
	}
	for _, piece := range pieces {
		if strings.TrimSpace(piece) == "" {
			return nil, fmt.Errorf("expected import id %q, got %q", example, id)
		}
	}
	return pieces, nil
}

func hydrateSiteStateFromData(raw json.RawMessage, state *siteResourceModel) {
	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		return
	}
	setStringIfPresent(data, "name", &state.Name)
	setStringIfPresent(data, "display_name", &state.Name)
	setStringIfPresent(data, "displayName", &state.Name)
	setStringIfPresent(data, "php_version", &state.PHPVersion)
	setStringIfPresent(data, "phpVersion", &state.PHPVersion)
	setStringIfPresent(data, "wp_environment_type", &state.WPEnvironmentType)
	setStringIfPresent(data, "wpEnvironmentType", &state.WPEnvironmentType)
	setBoolIfPresent(data, "multisite_support", &state.MultisiteSupport)
	setBoolIfPresent(data, "multisiteSupport", &state.MultisiteSupport)
}

func hydrateRecordStateFromData(raw json.RawMessage, state *dnsRecordResourceModel) {
	var zone struct {
		Records []map[string]any `json:"records"`
	}
	if err := json.Unmarshal(raw, &zone); err != nil {
		return
	}
	for _, record := range zone.Records {
		id, err := scalarID(record["id"])
		if err != nil || id != state.ID.ValueString() {
			continue
		}
		setStringIfPresent(record, "type", &state.Type)
		setStringIfPresent(record, "name", &state.Name)
		setStringIfPresent(record, "value", &state.Value)
		setIntIfPresent(record, "ttl", &state.TTL)
		setIntIfPresent(record, "priority", &state.Priority)
		setIntIfPresent(record, "weight", &state.Weight)
		setIntIfPresent(record, "port", &state.Port)
		setIntIfPresent(record, "site_id", &state.SiteID)
		setIntIfPresent(record, "siteId", &state.SiteID)
		return
	}
}

func hydrateDomainStateFromData(raw json.RawMessage, state *siteDomainResourceModel) {
	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		return
	}
	setStringIfPresent(data, "domain_name", &state.Name)
	setStringIfPresent(data, "domainName", &state.Name)
	setBoolIfPresent(data, "primary", &state.Primary)
}

func setStringIfPresent(data map[string]any, key string, target *types.String) {
	value, ok := data[key].(string)
	if ok && value != "" {
		*target = types.StringValue(value)
	}
}

func setBoolIfPresent(data map[string]any, key string, target *types.Bool) {
	value, ok := data[key].(bool)
	if ok {
		*target = types.BoolValue(value)
	}
}

func setIntIfPresent(data map[string]any, key string, target *types.Int64) {
	value, ok := data[key]
	if !ok || value == nil {
		return
	}
	switch typed := value.(type) {
	case float64:
		*target = types.Int64Value(int64(typed))
	case int64:
		*target = types.Int64Value(typed)
	case int:
		*target = types.Int64Value(int64(typed))
	}
}
