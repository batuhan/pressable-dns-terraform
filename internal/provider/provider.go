package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/batuhan/terraform-provider-pressable/internal/pressable"
)

var _ provider.Provider = &pressableProvider{}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &pressableProvider{version: version}
	}
}

type pressableProvider struct {
	version string
}

type providerModel struct {
	BaseURL      types.String `tfsdk:"base_url"`
	AccessToken  types.String `tfsdk:"access_token"`
	ClientID     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
}

func (p *pressableProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "pressable"
	resp.Version = p.version
}

func (p *pressableProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manage Pressable sites, DNS, domains, settings, and public API operations.",
		Attributes: map[string]schema.Attribute{
			"base_url": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Pressable API base URL. Defaults to `https://my.pressable.com`. Can also be set with `PRESSABLE_BASE_URL`.",
			},
			"access_token": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "Pressable bearer token. Can also be set with `PRESSABLE_ACCESS_TOKEN`.",
			},
			"client_id": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "Pressable API application client id. Can also be set with `PRESSABLE_CLIENT_ID`.",
			},
			"client_secret": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "Pressable API application client secret. Can also be set with `PRESSABLE_CLIENT_SECRET`.",
			},
		},
	}
}

func (p *pressableProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config providerModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	baseURL := valueOrEnv(config.BaseURL, "PRESSABLE_BASE_URL")
	accessToken := valueOrEnv(config.AccessToken, "PRESSABLE_ACCESS_TOKEN")
	clientID := valueOrEnv(config.ClientID, "PRESSABLE_CLIENT_ID")
	clientSecret := valueOrEnv(config.ClientSecret, "PRESSABLE_CLIENT_SECRET")

	if accessToken == "" && (clientID == "" || clientSecret == "") {
		resp.Diagnostics.AddAttributeError(
			path.Root("access_token"),
			"Missing Pressable credentials",
			"Set access_token or both client_id and client_secret, either in provider configuration or PRESSABLE_* environment variables.",
		)
		return
	}

	client, err := pressable.New(baseURL, accessToken, clientID, clientSecret)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Pressable client configuration", err.Error())
		return
	}
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *pressableProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewAPIRequestDataSource,
		NewSitesDataSource,
		NewZonesDataSource,
		NewZoneRecordsDataSource,
	}
}

func (p *pressableProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewAPIActionResource,
		NewSiteResource,
		NewDNSRecordResource,
		NewSiteDomainResource,
		NewSiteSettingResource,
	}
}

func valueOrEnv(value types.String, envName string) string {
	if !value.IsNull() && !value.IsUnknown() {
		return value.ValueString()
	}
	return os.Getenv(envName)
}
