package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/batuhan/terraform-provider-pressable/internal/pressable"
)

var _ datasource.DataSource = &zonesDataSource{}
var _ datasource.DataSourceWithConfigure = &zonesDataSource{}
var _ datasource.DataSource = &zoneRecordsDataSource{}
var _ datasource.DataSourceWithConfigure = &zoneRecordsDataSource{}

func NewZonesDataSource() datasource.DataSource       { return &zonesDataSource{} }
func NewZoneRecordsDataSource() datasource.DataSource { return &zoneRecordsDataSource{} }

type zonesDataSource struct{ client *pressable.Client }
type zoneRecordsDataSource struct{ client *pressable.Client }

type zonesDataSourceModel struct {
	ID       types.String `tfsdk:"id"`
	DataJSON types.String `tfsdk:"data_json"`
}

type zoneRecordsDataSourceModel struct {
	ID       types.String `tfsdk:"id"`
	ZoneID   types.Int64  `tfsdk:"zone_id"`
	DataJSON types.String `tfsdk:"data_json"`
}

func (d *zonesDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "pressable_zones"
}

func (d *zonesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists Pressable DNS zones.",
		Attributes: map[string]schema.Attribute{
			"id":        schema.StringAttribute{Computed: true},
			"data_json": schema.StringAttribute{Computed: true},
		},
	}
}

func (d *zonesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *zonesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var plan zonesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data, diags := readRaw(ctx, d.client, "/v1/zones")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = types.StringValue("zones")
	plan.DataJSON = data
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (d *zoneRecordsDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "pressable_zone_records"
}

func (d *zoneRecordsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists DNS records for a Pressable zone.",
		Attributes: map[string]schema.Attribute{
			"id":        schema.StringAttribute{Computed: true},
			"zone_id":   schema.Int64Attribute{Required: true},
			"data_json": schema.StringAttribute{Computed: true},
		},
	}
}

func (d *zoneRecordsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *zoneRecordsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var plan zoneRecordsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	path := "/v1/zones/" + fmtInt(plan.ZoneID.ValueInt64()) + "/records"
	data, diags := readRaw(ctx, d.client, path)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = types.StringValue(stableID(path))
	plan.DataJSON = data
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
