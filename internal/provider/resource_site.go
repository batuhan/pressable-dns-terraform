package provider

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/batuhan/terraform-provider-pressable/internal/pressable"
)

var _ resource.Resource = &siteResource{}
var _ resource.ResourceWithConfigure = &siteResource{}

func NewSiteResource() resource.Resource { return &siteResource{} }

type siteResource struct{ client *pressable.Client }

type siteResourceModel struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	PHPVersion        types.String `tfsdk:"php_version"`
	Install           types.String `tfsdk:"install"`
	Sandbox           types.Bool   `tfsdk:"sandbox"`
	Staging           types.Bool   `tfsdk:"staging"`
	Duplikit          types.Bool   `tfsdk:"duplikit"`
	DatacenterCode    types.String `tfsdk:"datacenter_code"`
	WPAdminUsername   types.String `tfsdk:"wp_admin_username"`
	WPAdminEmail      types.String `tfsdk:"wp_admin_email"`
	MultisiteSupport  types.Bool   `tfsdk:"multisite_support"`
	WPEnvironmentType types.String `tfsdk:"wp_environment_type"`
	ForceDelete       types.Bool   `tfsdk:"force_delete"`
	DataJSON          types.String `tfsdk:"data_json"`
}

func (r *siteResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "pressable_site"
}

func (r *siteResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Pressable site.",
		Attributes: map[string]schema.Attribute{
			"id":                  schema.StringAttribute{Computed: true},
			"name":                schema.StringAttribute{Required: true},
			"php_version":         schema.StringAttribute{Optional: true},
			"install":             schema.StringAttribute{Optional: true},
			"sandbox":             schema.BoolAttribute{Optional: true},
			"staging":             schema.BoolAttribute{Optional: true},
			"duplikit":            schema.BoolAttribute{Optional: true},
			"datacenter_code":     schema.StringAttribute{Optional: true},
			"wp_admin_username":   schema.StringAttribute{Optional: true},
			"wp_admin_email":      schema.StringAttribute{Optional: true},
			"multisite_support":   schema.BoolAttribute{Optional: true},
			"wp_environment_type": schema.StringAttribute{Optional: true},
			"force_delete":        schema.BoolAttribute{Optional: true},
			"data_json":           schema.StringAttribute{Computed: true},
		},
	}
}

func (r *siteResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *siteResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan siteResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body := compactMap(map[string]any{
		"name":                plan.Name.ValueString(),
		"php_version":         optionalString(plan.PHPVersion),
		"install":             optionalString(plan.Install),
		"sandbox":             optionalBool(plan.Sandbox),
		"staging":             optionalBool(plan.Staging),
		"duplikit":            optionalBool(plan.Duplikit),
		"datacenter_code":     optionalString(plan.DatacenterCode),
		"wp_admin_username":   optionalString(plan.WPAdminUsername),
		"wp_admin_email":      optionalString(plan.WPAdminEmail),
		"multisite_support":   optionalBool(plan.MultisiteSupport),
		"wp_environment_type": optionalString(plan.WPEnvironmentType),
	})
	envelope, _, err := r.client.Request(ctx, http.MethodPost, "/v1/sites", body)
	if err != nil {
		resp.Diagnostics.AddError("Create Pressable site failed", err.Error())
		return
	}
	id, err := idFromData(envelope.Data)
	if err != nil {
		resp.Diagnostics.AddError("Create Pressable site returned no id", err.Error())
		return
	}
	plan.ID = types.StringValue(id)
	plan.DataJSON = types.StringValue(pressable.RawString(envelope.Data))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *siteResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state siteResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	envelope, _, err := r.client.Request(ctx, http.MethodGet, "/v1/sites/"+state.ID.ValueString(), nil)
	if err != nil {
		resp.Diagnostics.AddError("Read Pressable site failed", err.Error())
		return
	}
	state.DataJSON = types.StringValue(pressable.RawString(envelope.Data))
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *siteResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan siteResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body := compactMap(map[string]any{
		"name":                optionalString(plan.Name),
		"php_version":         optionalString(plan.PHPVersion),
		"wp_environment_type": optionalString(plan.WPEnvironmentType),
	})
	envelope, _, err := r.client.Request(ctx, http.MethodPut, "/v1/sites/"+plan.ID.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("Update Pressable site failed", err.Error())
		return
	}
	plan.DataJSON = types.StringValue(pressable.RawString(envelope.Data))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *siteResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state siteResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	path := "/v1/sites/" + state.ID.ValueString()
	if !state.ForceDelete.IsNull() && !state.ForceDelete.IsUnknown() && state.ForceDelete.ValueBool() {
		path += "?force=true"
	}
	_, _, err := r.client.Request(ctx, http.MethodDelete, path, nil)
	if err != nil {
		resp.Diagnostics.AddError("Delete Pressable site failed", err.Error())
		return
	}
	resp.State.RemoveResource(ctx)
}
