package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/batuhan/terraform-provider-pressable/internal/pressable"
)

var _ datasource.DataSource = &apiRequestDataSource{}
var _ datasource.DataSourceWithConfigure = &apiRequestDataSource{}

func NewAPIRequestDataSource() datasource.DataSource {
	return &apiRequestDataSource{}
}

type apiRequestDataSource struct {
	client *pressable.Client
}

type apiRequestDataSourceModel struct {
	ID           types.String `tfsdk:"id"`
	Method       types.String `tfsdk:"method"`
	Path         types.String `tfsdk:"path"`
	RequestBody  types.String `tfsdk:"request_body"`
	StatusCode   types.Int64  `tfsdk:"status_code"`
	Message      types.String `tfsdk:"message"`
	DataJSON     types.String `tfsdk:"data_json"`
	ErrorsJSON   types.String `tfsdk:"errors_json"`
	ResponseJSON types.String `tfsdk:"response_json"`
}

func (d *apiRequestDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "pressable_api_request"
}

func (d *apiRequestDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Calls any Pressable public API endpoint and returns the standard response envelope. Prefer typed data sources when one exists.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{Computed: true},
			"method": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "HTTP method. Defaults to `GET`.",
			},
			"path": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "API path, for example `/v1/sites`.",
			},
			"request_body": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "JSON request body for non-GET reads.",
			},
			"status_code": schema.Int64Attribute{Computed: true},
			"message":     schema.StringAttribute{Computed: true},
			"data_json":   schema.StringAttribute{Computed: true},
			"errors_json": schema.StringAttribute{Computed: true},
			"response_json": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "JSON object with `message`, `data`, and `errors`.",
			},
		},
	}
}

func (d *apiRequestDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *apiRequestDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var plan apiRequestDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	method := strings.ToUpper(plan.Method.ValueString())
	if method == "" {
		method = http.MethodGet
	}
	body, err := parseJSONBody(plan.RequestBody)
	if err != nil {
		resp.Diagnostics.AddAttributeError(path.Root("request_body"), "Invalid request_body JSON", err.Error())
		return
	}

	envelope, status, err := d.client.Request(ctx, method, plan.Path.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("Pressable API request failed", err.Error())
		return
	}
	message, dataJSON, errorsJSON := envelopeStrings(envelope)
	plan.ID = types.StringValue(stableID(method, plan.Path.ValueString(), plan.RequestBody.ValueString()))
	plan.StatusCode = types.Int64Value(int64(status))
	plan.Message = message
	plan.DataJSON = dataJSON
	plan.ErrorsJSON = errorsJSON
	plan.ResponseJSON = mustJSONString(map[string]any{
		"message": envelope.Message,
		"data":    jsonRaw(envelope.Data),
		"errors":  jsonRaw(envelope.Errors),
	})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func jsonRaw(raw []byte) any {
	if len(raw) == 0 {
		return nil
	}
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return string(raw)
	}
	return value
}
