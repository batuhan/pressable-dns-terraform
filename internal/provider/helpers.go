package provider

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/batuhan/terraform-provider-pressable/internal/pressable"
)

func clientFromResource(req any, data any, diags *diag.Diagnostics) *pressable.Client {
	client, ok := data.(*pressable.Client)
	if !ok {
		diags.AddError("Missing Pressable client", "Provider did not configure a Pressable client.")
		return nil
	}
	return client
}

func mustJSONString(value any) types.String {
	payload, err := json.Marshal(value)
	if err != nil {
		return types.StringValue("{}")
	}
	return types.StringValue(string(payload))
}

func parseJSONBody(body types.String) (any, error) {
	if body.IsNull() || body.IsUnknown() || body.ValueString() == "" {
		return nil, nil
	}
	var value any
	if err := json.Unmarshal([]byte(body.ValueString()), &value); err != nil {
		return nil, err
	}
	return value, nil
}

func stableID(parts ...string) string {
	hash := sha256.Sum256([]byte(fmt.Sprint(parts)))
	return hex.EncodeToString(hash[:])[:24]
}

func int64String(value int64) types.String {
	return types.StringValue(strconv.FormatInt(value, 10))
}

func idAsInt64(id types.String) (int64, error) {
	return strconv.ParseInt(id.ValueString(), 10, 64)
}

func envelopeStrings(envelope *pressable.Envelope) (types.String, types.String, types.String) {
	if envelope == nil {
		return types.StringValue(""), types.StringValue("null"), types.StringValue("null")
	}
	return types.StringValue(envelope.Message), types.StringValue(pressable.RawString(envelope.Data)), types.StringValue(pressable.RawString(envelope.Errors))
}

func readRaw(ctx context.Context, client *pressable.Client, path string) (types.String, diag.Diagnostics) {
	var diags diag.Diagnostics
	envelope, _, err := client.Request(ctx, "GET", path, nil)
	if err != nil {
		diags.AddError("Pressable read failed", err.Error())
		return types.StringValue("null"), diags
	}
	return types.StringValue(pressable.RawString(envelope.Data)), diags
}
