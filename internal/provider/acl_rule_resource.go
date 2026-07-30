package provider

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/terrasquid/terraform-provider-terrasquid/internal/client"
	"github.com/terrasquid/terraform-provider-terrasquid/internal/model"
)

var _ resource.Resource = &ACLRuleResource{}

type ACLRuleResource struct {
	client *client.APIClient
}

type ACLRuleResourceModel struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	Priority          types.Int64  `tfsdk:"priority"`
	Comment           types.String `tfsdk:"comment"`
	Sources           types.List   `tfsdk:"sources"`
	Destinations      types.List   `tfsdk:"destinations"`
	DestinationGroups types.List   `tfsdk:"destination_groups"`
	Service           types.String `tfsdk:"service"`
	CreatedAt         types.String `tfsdk:"created_at"`
	UpdatedAt         types.String `tfsdk:"updated_at"`
}

func NewACLRuleResource() resource.Resource {
	return &ACLRuleResource{}
}

func (r *ACLRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_acl_rule", req.ProviderTypeName)
}

func (r *ACLRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 63),
					stringvalidator.RegexMatches(regexp.MustCompile(`^[a-zA-Z0-9_-]+$`), ""),
				},
			},
			"priority": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(100),
			},
			"comment": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(""),
			},
			"sources": schema.ListAttribute{
				ElementType: types.StringType,
				Required:    true,
			},
			"destinations": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Default:     listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
			},
			"destination_groups": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Default:     listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
			},
			"service": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (r *ACLRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	c, err := configureClientResource(resp, req)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", err.Error())
		return
	}
	if c != nil {
		r.client = c
	}
}

func (r *ACLRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ACLRuleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var sources []string
	resp.Diagnostics.Append(plan.Sources.ElementsAs(ctx, &sources, false)...)
	destinations := aclRuleListElements(ctx, plan.Destinations, &resp.Diagnostics)
	destinationGroups := aclRuleListElements(ctx, plan.DestinationGroups, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	input := model.ACLRuleInput{
		Name:              plan.Name.ValueString(),
		Priority:          int(plan.Priority.ValueInt64()),
		Comment:           plan.Comment.ValueString(),
		Sources:           sources,
		Destinations:      destinations,
		DestinationGroups: destinationGroups,
	}

	result, err := r.client.CreateACLRule(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to create ACL rule: %s", err))
		return
	}

	plan.ID = types.StringValue(result.ID)
	plan.Name = types.StringValue(result.Name)
	plan.Priority = types.Int64Value(int64(result.Priority))
	plan.Comment = types.StringValue(result.Comment)
	plan.Service = types.StringValue(result.Service)
	plan.CreatedAt = types.StringValue(result.CreatedAt.Format(time.RFC3339))
	plan.UpdatedAt = types.StringValue(result.UpdatedAt.Format(time.RFC3339))

	sourcesList, diags := types.ListValueFrom(ctx, types.StringType, result.Sources)
	resp.Diagnostics.Append(diags...)
	plan.Sources = sourcesList

	destinationsList, diags := types.ListValueFrom(ctx, types.StringType, result.Destinations)
	resp.Diagnostics.Append(diags...)
	plan.Destinations = destinationsList

	destinationGroupsList, diags := types.ListValueFrom(ctx, types.StringType, result.DestinationGroups)
	resp.Diagnostics.Append(diags...)
	plan.DestinationGroups = destinationGroupsList

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ACLRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ACLRuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.GetACLRule(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to read ACL rule: %s", err))
		return
	}

	state.ID = types.StringValue(result.ID)
	state.Name = types.StringValue(result.Name)
	state.Priority = types.Int64Value(int64(result.Priority))
	state.Comment = types.StringValue(result.Comment)
	state.Service = types.StringValue(result.Service)
	state.CreatedAt = types.StringValue(result.CreatedAt.Format(time.RFC3339))
	state.UpdatedAt = types.StringValue(result.UpdatedAt.Format(time.RFC3339))

	sourcesList, diags := types.ListValueFrom(ctx, types.StringType, result.Sources)
	resp.Diagnostics.Append(diags...)
	state.Sources = sourcesList

	destinationsList, diags := types.ListValueFrom(ctx, types.StringType, result.Destinations)
	resp.Diagnostics.Append(diags...)
	state.Destinations = destinationsList

	destinationGroupsList, diags := types.ListValueFrom(ctx, types.StringType, result.DestinationGroups)
	resp.Diagnostics.Append(diags...)
	state.DestinationGroups = destinationGroupsList

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ACLRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ACLRuleResourceModel
	var state ACLRuleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var sources []string
	resp.Diagnostics.Append(plan.Sources.ElementsAs(ctx, &sources, false)...)
	destinations := aclRuleListElements(ctx, plan.Destinations, &resp.Diagnostics)
	destinationGroups := aclRuleListElements(ctx, plan.DestinationGroups, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	input := model.ACLRuleInput{
		Name:              plan.Name.ValueString(),
		Priority:          int(plan.Priority.ValueInt64()),
		Comment:           plan.Comment.ValueString(),
		Sources:           sources,
		Destinations:      destinations,
		DestinationGroups: destinationGroups,
	}

	id := plan.ID.ValueString()
	if id == "" {
		id = state.ID.ValueString()
	}
	result, err := r.client.UpdateACLRule(ctx, id, input)
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to update ACL rule: %s", err))
		return
	}

	plan.ID = types.StringValue(result.ID)
	plan.Name = types.StringValue(result.Name)
	plan.Priority = types.Int64Value(int64(result.Priority))
	plan.Comment = types.StringValue(result.Comment)
	plan.Service = types.StringValue(result.Service)
	plan.CreatedAt = types.StringValue(result.CreatedAt.Format(time.RFC3339))
	plan.UpdatedAt = types.StringValue(result.UpdatedAt.Format(time.RFC3339))

	sourcesList, diags := types.ListValueFrom(ctx, types.StringType, result.Sources)
	resp.Diagnostics.Append(diags...)
	plan.Sources = sourcesList

	destinationsList, diags := types.ListValueFrom(ctx, types.StringType, result.Destinations)
	resp.Diagnostics.Append(diags...)
	plan.Destinations = destinationsList

	destinationGroupsList, diags := types.ListValueFrom(ctx, types.StringType, result.DestinationGroups)
	resp.Diagnostics.Append(diags...)
	plan.DestinationGroups = destinationGroupsList

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ACLRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ACLRuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteACLRule(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFoundError(err) {
			return
		}
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to delete ACL rule: %s", err))
		return
	}
}

func (r *ACLRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func aclRuleListElements(ctx context.Context, value types.List, diagnostics *diag.Diagnostics) []string {
	if value.IsNull() {
		return []string{}
	}
	var values []string
	diagnostics.Append(value.ElementsAs(ctx, &values, false)...)
	return values
}
