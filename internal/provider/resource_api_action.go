package provider

import (
	"context"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/batuhan/terraform-provider-pressable/internal/pressable"
)

var _ resource.Resource = &apiActionResource{}
var _ resource.ResourceWithConfigure = &apiActionResource{}

func NewAPIActionResource() resource.Resource { return &apiActionResource{} }

type apiActionResource struct{ client *pressable.Client }

type apiActionResourceModel struct {
	ID           types.String `tfsdk:"id"`
	Method       types.String `tfsdk:"method"`
	Path         types.String `tfsdk:"path"`
	RequestBody  types.String `tfsdk:"request_body"`
	TriggersJSON types.String `tfsdk:"triggers_json"`
	StatusCode   types.Int64  `tfsdk:"status_code"`
	Message      types.String `tfsdk:"message"`
	DataJSON     types.String `tfsdk:"data_json"`
	ErrorsJSON   types.String `tfsdk:"errors_json"`
}

func (r *apiActionResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "pressable_api_action"
}

func (r *apiActionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Runs any Pressable public API mutation. Use typed resources for durable objects; use this for public API actions that Terraform cannot model directly.",
		Attributes: map[string]schema.Attribute{
			"id":            schema.StringAttribute{Computed: true},
			"method":        schema.StringAttribute{Required: true},
			"path":          schema.StringAttribute{Required: true},
			"request_body":  schema.StringAttribute{Optional: true, Sensitive: true},
			"triggers_json": schema.StringAttribute{Optional: true, MarkdownDescription: "JSON string used to force reruns when its value changes."},
			"status_code":   schema.Int64Attribute{Computed: true},
			"message":       schema.StringAttribute{Computed: true},
			"data_json":     schema.StringAttribute{Computed: true},
			"errors_json":   schema.StringAttribute{Computed: true},
		},
	}
}

func (r *apiActionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *apiActionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan apiActionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	r.run(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *apiActionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state apiActionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *apiActionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan apiActionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	r.run(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *apiActionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.State.RemoveResource(ctx)
}

func (r *apiActionResource) run(ctx context.Context, plan *apiActionResourceModel, diags *diag.Diagnostics) {
	body, err := parseJSONBody(plan.RequestBody)
	if err != nil {
		diags.AddError("Invalid request_body JSON", err.Error())
		return
	}
	method := strings.ToUpper(plan.Method.ValueString())
	if method == "" {
		method = http.MethodPost
	}
	envelope, status, err := r.client.Request(ctx, method, plan.Path.ValueString(), body)
	if err != nil {
		diags.AddError("Pressable API action failed", err.Error())
		return
	}
	message, dataJSON, errorsJSON := envelopeStrings(envelope)
	plan.ID = types.StringValue(stableID(method, plan.Path.ValueString(), plan.RequestBody.ValueString(), plan.TriggersJSON.ValueString()))
	plan.StatusCode = types.Int64Value(int64(status))
	plan.Message = message
	plan.DataJSON = dataJSON
	plan.ErrorsJSON = errorsJSON
}
