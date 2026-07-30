package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/terrasquid/terraform-provider-terrasquid/internal/client"
)

var _ datasource.DataSource = &DestinationGroupDataSource{}

type DestinationGroupDataSource struct{ client *client.APIClient }

type DestinationGroupDataSourceModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Destinations types.Set    `tfsdk:"destinations"`
	Comment      types.String `tfsdk:"comment"`
	Service      types.String `tfsdk:"service"`
	CreatedAt    types.String `tfsdk:"created_at"`
	UpdatedAt    types.String `tfsdk:"updated_at"`
}

func NewDestinationGroupDataSource() datasource.DataSource { return &DestinationGroupDataSource{} }

func (d *DestinationGroupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_destination_group", req.ProviderTypeName)
}

func (d *DestinationGroupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id":           schema.StringAttribute{Computed: true},
		"name":         schema.StringAttribute{Required: true},
		"destinations": schema.SetAttribute{Computed: true, ElementType: types.StringType},
		"comment":      schema.StringAttribute{Computed: true},
		"service":      schema.StringAttribute{Computed: true},
		"created_at":   schema.StringAttribute{Computed: true},
		"updated_at":   schema.StringAttribute{Computed: true},
	}}
}

func (d *DestinationGroupDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	c, err := configureClient(resp, req)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", err.Error())
		return
	}
	if c != nil {
		d.client = c
	}
}

func (d *DestinationGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config DestinationGroupDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	groups, err := d.client.ListDestinationGroups(ctx, config.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to read destination group: %s", err))
		return
	}
	if len(groups) == 0 {
		resp.Diagnostics.AddError("Destination Group Not Found", fmt.Sprintf("No destination group named %q exists.", config.Name.ValueString()))
		return
	}
	group := groups[0]
	destinations, diags := types.SetValueFrom(ctx, types.StringType, sortedDestinations(group.Destinations))
	resp.Diagnostics.Append(diags...)
	state := DestinationGroupDataSourceModel{
		ID: types.StringValue(group.ID), Name: types.StringValue(group.Name), Destinations: destinations,
		Comment: types.StringValue(group.Comment), Service: types.StringValue(group.Service),
		CreatedAt: types.StringValue(group.CreatedAt.Format(time.RFC3339)), UpdatedAt: types.StringValue(group.UpdatedAt.Format(time.RFC3339)),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
