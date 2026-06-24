package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/batuhan/terraform-provider-pressable/internal/pressable"
)

var _ datasource.DataSource = &sitesDataSource{}
var _ datasource.DataSourceWithConfigure = &sitesDataSource{}

func NewSitesDataSource() datasource.DataSource { return &sitesDataSource{} }

type sitesDataSource struct{ client *pressable.Client }

type sitesDataSourceModel struct {
	ID            types.String `tfsdk:"id"`
	TagName       types.String `tfsdk:"tag_name"`
	FavoritesOnly types.Bool   `tfsdk:"favorites_only"`
	Page          types.Int64  `tfsdk:"page"`
	PerPage       types.Int64  `tfsdk:"per_page"`
	DataJSON      types.String `tfsdk:"data_json"`
}

func (d *sitesDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "pressable_sites"
}

func (d *sitesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists Pressable sites visible to the configured API credentials.",
		Attributes: map[string]schema.Attribute{
			"id":             schema.StringAttribute{Computed: true},
			"tag_name":       schema.StringAttribute{Optional: true},
			"favorites_only": schema.BoolAttribute{Optional: true},
			"page":           schema.Int64Attribute{Optional: true},
			"per_page":       schema.Int64Attribute{Optional: true},
			"data_json":      schema.StringAttribute{Computed: true},
		},
	}
}

func (d *sitesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*pressable.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data", "Expected *pressable.Client.")
		return
	}
	d.client = client
}

func (d *sitesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var plan sitesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	path := "/v1/sites"
	query := make([]string, 0, 4)
	if !plan.TagName.IsNull() && !plan.TagName.IsUnknown() && plan.TagName.ValueString() != "" {
		query = append(query, "tag_name="+urlQueryEscape(plan.TagName.ValueString()))
	}
	if !plan.FavoritesOnly.IsNull() && !plan.FavoritesOnly.IsUnknown() && plan.FavoritesOnly.ValueBool() {
		query = append(query, "favorites_only=true")
	}
	if !plan.Page.IsNull() && !plan.Page.IsUnknown() {
		query = append(query, "paginate=true", "page="+fmtInt(plan.Page.ValueInt64()))
	}
	if !plan.PerPage.IsNull() && !plan.PerPage.IsUnknown() {
		query = append(query, "paginate=true", "per_page="+fmtInt(plan.PerPage.ValueInt64()))
	}
	if len(query) > 0 {
		path += "?" + joinQuery(query)
	}
	data, diags := readRaw(ctx, d.client, path)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = types.StringValue(stableID(path))
	plan.DataJSON = data
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
