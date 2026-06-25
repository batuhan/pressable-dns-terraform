package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/batuhan/terraform-provider-pressable/internal/pressable"
)

var _ resource.Resource = &dnsRecordResource{}
var _ resource.ResourceWithConfigure = &dnsRecordResource{}
var _ resource.ResourceWithImportState = &dnsRecordResource{}

func NewDNSRecordResource() resource.Resource { return &dnsRecordResource{} }

type dnsRecordResource struct{ client *pressable.Client }

type dnsRecordResourceModel struct {
	ID       types.String `tfsdk:"id"`
	ZoneID   types.Int64  `tfsdk:"zone_id"`
	Type     types.String `tfsdk:"type"`
	Name     types.String `tfsdk:"name"`
	Value    types.String `tfsdk:"value"`
	Priority types.Int64  `tfsdk:"priority"`
	Weight   types.Int64  `tfsdk:"weight"`
	Port     types.Int64  `tfsdk:"port"`
	TTL      types.Int64  `tfsdk:"ttl"`
	SiteID   types.Int64  `tfsdk:"site_id"`
	DataJSON types.String `tfsdk:"data_json"`
}

func (r *dnsRecordResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "pressable_dns_record"
}

func (r *dnsRecordResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a DNS record in a Pressable zone. The public API has create/delete semantics, so updates are applied by replacing the record.",
		Attributes: map[string]schema.Attribute{
			"id":        schema.StringAttribute{Computed: true},
			"zone_id":   schema.Int64Attribute{Required: true},
			"type":      schema.StringAttribute{Optional: true, Computed: true},
			"name":      schema.StringAttribute{Optional: true, Computed: true},
			"value":     schema.StringAttribute{Optional: true, Computed: true},
			"priority":  schema.Int64Attribute{Optional: true, Computed: true},
			"weight":    schema.Int64Attribute{Optional: true, Computed: true},
			"port":      schema.Int64Attribute{Optional: true, Computed: true},
			"ttl":       schema.Int64Attribute{Optional: true, Computed: true},
			"site_id":   schema.Int64Attribute{Optional: true},
			"data_json": schema.StringAttribute{Computed: true},
		},
	}
}

func (r *dnsRecordResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*pressable.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data", "Expected *pressable.Client.")
		return
	}
	r.client = client
}

func (r *dnsRecordResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan dnsRecordResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.create(ctx, &plan); err != nil {
		resp.Diagnostics.AddError("Create DNS record failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *dnsRecordResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state dnsRecordResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	envelope, _, err := r.client.Request(ctx, http.MethodGet, fmt.Sprintf("/v1/zones/%d/records", state.ZoneID.ValueInt64()), nil)
	if err != nil {
		resp.Diagnostics.AddError("Read DNS records failed", err.Error())
		return
	}
	state.DataJSON = types.StringValue(pressable.RawString(envelope.Data))
	hydrateRecordStateFromData(envelope.Data, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *dnsRecordResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan dnsRecordResourceModel
	var state dnsRecordResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if state.ID.ValueString() != "" {
		_, _, err := r.client.Request(ctx, http.MethodDelete, fmt.Sprintf("/v1/zones/%d/records/%s", state.ZoneID.ValueInt64(), state.ID.ValueString()), nil)
		if err != nil {
			resp.Diagnostics.AddError("Replace DNS record delete failed", err.Error())
			return
		}
	}
	if err := r.create(ctx, &plan); err != nil {
		resp.Diagnostics.AddError("Replace DNS record create failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *dnsRecordResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state dnsRecordResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	_, _, err := r.client.Request(ctx, http.MethodDelete, fmt.Sprintf("/v1/zones/%d/records/%s", state.ZoneID.ValueInt64(), state.ID.ValueString()), nil)
	if err != nil {
		resp.Diagnostics.AddError("Delete DNS record failed", err.Error())
		return
	}
	resp.State.RemoveResource(ctx)
}

func (r *dnsRecordResource) create(ctx context.Context, plan *dnsRecordResourceModel) error {
	if plan.Type.IsNull() || plan.Type.IsUnknown() || plan.Type.ValueString() == "" {
		return fmt.Errorf("type is required when creating a DNS record")
	}
	if plan.Name.IsNull() || plan.Name.IsUnknown() || plan.Name.ValueString() == "" {
		return fmt.Errorf("name is required when creating a DNS record")
	}
	if plan.Value.IsNull() || plan.Value.IsUnknown() || plan.Value.ValueString() == "" {
		return fmt.Errorf("value is required when creating a DNS record")
	}
	body := compactMap(map[string]any{
		"type":     plan.Type.ValueString(),
		"name":     plan.Name.ValueString(),
		"value":    plan.Value.ValueString(),
		"priority": optionalInt(plan.Priority),
		"weight":   optionalInt(plan.Weight),
		"port":     optionalInt(plan.Port),
		"ttl":      optionalInt(plan.TTL),
		"site_id":  optionalInt(plan.SiteID),
	})
	envelope, _, err := r.client.Request(ctx, http.MethodPost, fmt.Sprintf("/v1/zones/%d/records", plan.ZoneID.ValueInt64()), body)
	if err != nil {
		return err
	}
	id, err := recordIDFromZoneResponse(envelope.Data, plan)
	if err != nil {
		return err
	}
	plan.ID = types.StringValue(id)
	plan.DataJSON = types.StringValue(pressable.RawString(envelope.Data))
	return nil
}

func recordIDFromZoneResponse(raw json.RawMessage, plan *dnsRecordResourceModel) (string, error) {
	var zone struct {
		Records []struct {
			ID       int64  `json:"id"`
			Type     string `json:"type"`
			Name     string `json:"name"`
			Value    string `json:"value"`
			TTL      int64  `json:"ttl"`
			Priority *int64 `json:"priority"`
		} `json:"records"`
	}
	if err := json.Unmarshal(raw, &zone); err != nil {
		return "", err
	}
	for _, record := range zone.Records {
		if record.Type == plan.Type.ValueString() && record.Name == plan.Name.ValueString() && record.Value == plan.Value.ValueString() {
			return fmtInt(record.ID), nil
		}
	}
	return "", fmt.Errorf("created DNS record was not present in zone response")
}

func (r *dnsRecordResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts, err := splitImportID(req.ID, 2, "zone_id/record_id")
	if err != nil {
		resp.Diagnostics.AddError("Invalid import id", err.Error())
		return
	}
	zoneID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		resp.Diagnostics.AddAttributeError(path.Root("zone_id"), "Invalid zone_id", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("zone_id"), zoneID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}
