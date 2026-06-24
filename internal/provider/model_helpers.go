package provider

import (
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func optionalString(value types.String) any {
	if value.IsNull() || value.IsUnknown() || value.ValueString() == "" {
		return nil
	}
	return value.ValueString()
}

func optionalBool(value types.Bool) any {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	return value.ValueBool()
}

func optionalInt(value types.Int64) any {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	return value.ValueInt64()
}

func compactMap(input map[string]any) map[string]any {
	output := make(map[string]any, len(input))
	for key, value := range input {
		if value != nil {
			output[key] = value
		}
	}
	return output
}

func idFromData(raw json.RawMessage) (string, error) {
	var value map[string]any
	if err := json.Unmarshal(raw, &value); err != nil {
		return "", err
	}
	id, ok := value["id"]
	if !ok {
		return "", fmt.Errorf("missing id in response data")
	}
	switch typed := id.(type) {
	case float64:
		return fmtInt(int64(typed)), nil
	case string:
		return typed, nil
	default:
		return "", fmt.Errorf("unsupported id type %T", id)
	}
}
