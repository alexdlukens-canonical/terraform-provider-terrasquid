package provider

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/terrasquid/terraform-provider-terrasquid/internal/client"
	"github.com/terrasquid/terraform-provider-terrasquid/internal/model"
)

var _ resource.Resource = &DestinationGroupResource{}

type DestinationGroupResource struct {
	client *client.APIClient
}

type DestinationGroupResourceModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Destinations types.Set    `tfsdk:"destinations"`
	Comment      types.String `tfsdk:"comment"`
	Service      types.String `tfsdk:"service"`
	CreatedAt    types.String `tfsdk:"created_at"`
	UpdatedAt    types.String `tfsdk:"updated_at"`
}

func NewDestinationGroupResource() resource.Resource {
	return &DestinationGroupResource{}
}

func (r *DestinationGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_destination_group", req.ProviderTypeName)
}

func (r *DestinationGroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"name": schema.StringAttribute{Required: true, Validators: []validator.String{
			stringvalidator.LengthBetween(1, 63),
			stringvalidator.RegexMatches(regexp.MustCompile(`^[a-zA-Z0-9_-]+$`), ""),
		}},
		"destinations": schema.SetAttribute{ElementType: types.StringType, Required: true, Validators: []validator.Set{setvalidator.SizeAtLeast(1)}},
		"comment":      schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("")},
		"service":      schema.StringAttribute{Computed: true},
		"created_at":   schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"updated_at":   schema.StringAttribute{Computed: true},
	}}
}

func (r *DestinationGroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	c, err := configureClientResource(resp, req)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", err.Error())
		return
	}
	if c != nil {
		r.client = c
	}
}

func (r *DestinationGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan DestinationGroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	result, err := r.client.CreateDestinationGroup(ctx, destinationGroupInput(ctx, plan, &resp.Diagnostics))
	if resp.Diagnostics.HasError() {
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to create destination group: %s", err))
		return
	}
	setDestinationGroupState(ctx, &plan, result, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *DestinationGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state DestinationGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	result, err := r.client.GetDestinationGroup(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to read destination group: %s", err))
		return
	}
	setDestinationGroupState(ctx, &state, result, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *DestinationGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan DestinationGroupResourceModel
	var state DestinationGroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	result, err := r.client.UpdateDestinationGroup(ctx, state.ID.ValueString(), destinationGroupInput(ctx, plan, &resp.Diagnostics))
	if resp.Diagnostics.HasError() {
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to update destination group: %s", err))
		return
	}
	setDestinationGroupState(ctx, &plan, result, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *DestinationGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state DestinationGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteDestinationGroup(ctx, state.ID.ValueString()); err != nil && !client.IsNotFoundError(err) {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to delete destination group: %s", err))
	}
}

func (r *DestinationGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func destinationGroupInput(ctx context.Context, value DestinationGroupResourceModel, diagnostics *diag.Diagnostics) model.DestinationGroupInput {
	var destinations []string
	diagnostics.Append(value.Destinations.ElementsAs(ctx, &destinations, false)...)
	return model.DestinationGroupInput{Name: value.Name.ValueString(), Destinations: sortedDestinations(destinations), Comment: value.Comment.ValueString()}
}

func setDestinationGroupState(ctx context.Context, state *DestinationGroupResourceModel, result *model.DestinationGroup, diagnostics *diag.Diagnostics) {
	state.ID = types.StringValue(result.ID)
	state.Name = types.StringValue(result.Name)
	state.Comment = types.StringValue(result.Comment)
	state.Service = types.StringValue(result.Service)
	state.CreatedAt = types.StringValue(result.CreatedAt.Format(time.RFC3339))
	state.UpdatedAt = types.StringValue(result.UpdatedAt.Format(time.RFC3339))
	destinations, diags := types.SetValueFrom(ctx, types.StringType, sortedDestinations(result.Destinations))
	diagnostics.Append(diags...)
	state.Destinations = destinations
}

func sortedDestinations(destinations []string) []string {
	sorted := append([]string(nil), destinations...)
	sort.Strings(sorted)
	return sorted
}
