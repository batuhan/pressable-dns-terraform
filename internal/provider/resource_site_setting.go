package provider

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/batuhan/terraform-provider-pressable/internal/pressable"
)

var _ resource.Resource = &siteSettingResource{}
var _ resource.ResourceWithConfigure = &siteSettingResource{}
var _ resource.ResourceWithImportState = &siteSettingResource{}

func NewSiteSettingResource() resource.Resource { return &siteSettingResource{} }

type siteSettingResource struct{ client *pressable.Client }

type siteSettingResourceModel struct {
	ID          types.String `tfsdk:"id"`
	SiteID      types.Int64  `tfsdk:"site_id"`
	Setting     types.String `tfsdk:"setting"`
	RequestBody types.String `tfsdk:"request_body"`
	DataJSON    types.String `tfsdk:"data_json"`
}

func (r *siteSettingResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "pressable_site_setting"
}

func (r *siteSettingResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages durable Pressable site setting endpoints such as cdn, edge-cache, maintenance-mode, basic-authentication, wordpress-mcp, multisite_support, light_weight_404, php_fs_permissions, and woo-cart-cache.",
		Attributes: map[string]schema.Attribute{
			"id":           schema.StringAttribute{Computed: true},
			"site_id":      schema.Int64Attribute{Required: true},
			"setting":      schema.StringAttribute{Required: true},
			"request_body": schema.StringAttribute{Optional: true, Sensitive: true, MarkdownDescription: "JSON body sent to the setting endpoint."},
			"data_json":    schema.StringAttribute{Computed: true},
		},
	}
}

func (r *siteSettingResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *siteSettingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan siteSettingResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.apply(ctx, &plan); err != nil {
		resp.Diagnostics.AddError("Apply site setting failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *siteSettingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state siteSettingResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	path, readable := settingPath(state.SiteID.ValueInt64(), state.Setting.ValueString())
	if readable {
		envelope, _, err := r.client.Request(ctx, http.MethodGet, path, nil)
		if err == nil {
			state.DataJSON = types.StringValue(pressable.RawString(envelope.Data))
		}
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *siteSettingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan siteSettingResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.apply(ctx, &plan); err != nil {
		resp.Diagnostics.AddError("Apply site setting failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *siteSettingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state siteSettingResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	path, _ := settingPath(state.SiteID.ValueInt64(), state.Setting.ValueString())
	if settingDeleteSupported(state.Setting.ValueString()) {
		_, _, err := r.client.Request(ctx, http.MethodDelete, path, nil)
		if err != nil {
			resp.Diagnostics.AddError("Delete site setting failed", err.Error())
			return
		}
	}
	resp.State.RemoveResource(ctx)
}

func (r *siteSettingResource) apply(ctx context.Context, plan *siteSettingResourceModel) error {
	body, err := parseJSONBody(plan.RequestBody)
	if err != nil {
		return err
	}
	path, _ := settingPath(plan.SiteID.ValueInt64(), plan.Setting.ValueString())
	method := settingApplyMethod(plan.Setting.ValueString())
	envelope, _, err := r.client.Request(ctx, method, path, body)
	if err != nil {
		return err
	}
	plan.ID = types.StringValue(stableID(fmtInt(plan.SiteID.ValueInt64()), plan.Setting.ValueString()))
	plan.DataJSON = types.StringValue(pressable.RawString(envelope.Data))
	return nil
}

func settingPath(siteID int64, setting string) (string, bool) {
	readable := false
	switch setting {
	case "cdn":
		readable = true
	case "edge-cache":
		readable = true
	case "basic-authentication", "maintenance-mode", "multisite_support", "light_weight_404", "php_fs_permissions", "wordpress-mcp", "woo-cart-cache":
	default:
	}
	return fmt.Sprintf("/v1/sites/%d/%s", siteID, setting), readable
}

func settingApplyMethod(setting string) string {
	if setting == "cdn" {
		return http.MethodPost
	}
	return http.MethodPut
}

func settingDeleteSupported(setting string) bool {
	switch setting {
	case "cdn", "edge-cache":
		return true
	default:
		return false
	}
}

func (r *siteSettingResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts, err := splitImportID(req.ID, 2, "site_id/setting")
	if err != nil {
		resp.Diagnostics.AddError("Invalid import id", err.Error())
		return
	}
	siteID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		resp.Diagnostics.AddAttributeError(path.Root("site_id"), "Invalid site_id", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("site_id"), siteID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("setting"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), stableID(parts[0], parts[1]))...)
}
