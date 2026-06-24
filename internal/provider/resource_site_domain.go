package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/batuhan/terraform-provider-pressable/internal/pressable"
)

var _ resource.Resource = &siteDomainResource{}
var _ resource.ResourceWithConfigure = &siteDomainResource{}

func NewSiteDomainResource() resource.Resource { return &siteDomainResource{} }

type siteDomainResource struct{ client *pressable.Client }

type siteDomainResourceModel struct {
	ID       types.String `tfsdk:"id"`
	SiteID   types.Int64  `tfsdk:"site_id"`
	Name     types.String `tfsdk:"name"`
	Primary  types.Bool   `tfsdk:"primary"`
	DataJSON types.String `tfsdk:"data_json"`
}

func (r *siteDomainResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "pressable_site_domain"
}

func (r *siteDomainResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Attaches a domain to a Pressable site and can mark it primary.",
		Attributes: map[string]schema.Attribute{
			"id":        schema.StringAttribute{Computed: true},
			"site_id":   schema.Int64Attribute{Required: true},
			"name":      schema.StringAttribute{Required: true},
			"primary":   schema.BoolAttribute{Optional: true},
			"data_json": schema.StringAttribute{Computed: true},
		},
	}
}

func (r *siteDomainResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *siteDomainResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan siteDomainResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	envelope, _, err := r.client.Request(ctx, http.MethodPost, fmt.Sprintf("/v1/sites/%d/domains", plan.SiteID.ValueInt64()), map[string]any{
		"name": plan.Name.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Create site domain failed", err.Error())
		return
	}
	id, err := idFromDomainData(envelope.Data, plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Create site domain returned no id", err.Error())
		return
	}
	plan.ID = types.StringValue(id)
	plan.DataJSON = types.StringValue(pressable.RawString(envelope.Data))
	if !plan.Primary.IsNull() && !plan.Primary.IsUnknown() && plan.Primary.ValueBool() {
		if err := r.setPrimary(ctx, plan.SiteID.ValueInt64(), id); err != nil {
			resp.Diagnostics.AddError("Set primary site domain failed", err.Error())
			return
		}
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *siteDomainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state siteDomainResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	envelope, _, err := r.client.Request(ctx, http.MethodGet, fmt.Sprintf("/v1/sites/%d/domains/%s", state.SiteID.ValueInt64(), state.ID.ValueString()), nil)
	if err != nil {
		resp.Diagnostics.AddError("Read site domain failed", err.Error())
		return
	}
	state.DataJSON = types.StringValue(pressable.RawString(envelope.Data))
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *siteDomainResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan siteDomainResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !plan.Primary.IsNull() && !plan.Primary.IsUnknown() && plan.Primary.ValueBool() {
		if err := r.setPrimary(ctx, plan.SiteID.ValueInt64(), plan.ID.ValueString()); err != nil {
			resp.Diagnostics.AddError("Set primary site domain failed", err.Error())
			return
		}
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *siteDomainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state siteDomainResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	_, _, err := r.client.Request(ctx, http.MethodDelete, fmt.Sprintf("/v1/sites/%d/domains/%s", state.SiteID.ValueInt64(), state.ID.ValueString()), nil)
	if err != nil {
		resp.Diagnostics.AddError("Delete site domain failed", err.Error())
		return
	}
	resp.State.RemoveResource(ctx)
}

func (r *siteDomainResource) setPrimary(ctx context.Context, siteID int64, domainID string) error {
	_, _, err := r.client.Request(ctx, http.MethodPut, fmt.Sprintf("/v1/sites/%d/domains/%s/primary", siteID, domainID), nil)
	return err
}
