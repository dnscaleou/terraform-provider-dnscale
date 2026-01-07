package datasources

import (
	"context"
	"fmt"

	"github.com/dnscale/terraform-provider-dnscale/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &ZoneDataSource{}

// NewZoneDataSource creates a new zone data source
func NewZoneDataSource() datasource.DataSource {
	return &ZoneDataSource{}
}

// ZoneDataSource defines the data source implementation.
type ZoneDataSource struct {
	client *client.Client
}

// ZoneDataSourceModel describes the data source data model.
type ZoneDataSourceModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Type       types.String `tfsdk:"type"`
	Region     types.String `tfsdk:"region"`
	Status     types.String `tfsdk:"status"`
	CustomerID types.String `tfsdk:"customer_id"`
	CreatedAt  types.String `tfsdk:"created_at"`
	UpdatedAt  types.String `tfsdk:"updated_at"`
}

func (d *ZoneDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_zone"
}

func (d *ZoneDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to look up an existing DNS zone by ID or name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Zone UUID identifier. Either id or name must be specified.",
				Optional:    true,
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Zone name (FQDN). Either id or name must be specified.",
				Optional:    true,
				Computed:    true,
			},
			"type": schema.StringAttribute{
				Description: "Zone type (master or slave).",
				Computed:    true,
			},
			"region": schema.StringAttribute{
				Description: "Zone region (EU, GLOBAL, or EU_GLOBAL).",
				Computed:    true,
			},
			"status": schema.StringAttribute{
				Description: "Zone status.",
				Computed:    true,
			},
			"customer_id": schema.StringAttribute{
				Description: "Customer UUID that owns this zone.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the zone was created.",
				Computed:    true,
			},
			"updated_at": schema.StringAttribute{
				Description: "Timestamp when the zone was last updated.",
				Computed:    true,
			},
		},
	}
}

func (d *ZoneDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *ZoneDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ZoneDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var zone *client.Zone
	var err error

	// Look up by ID if provided, otherwise by name
	if !data.ID.IsNull() && data.ID.ValueString() != "" {
		zone, err = d.client.GetZone(ctx, data.ID.ValueString())
	} else if !data.Name.IsNull() && data.Name.ValueString() != "" {
		zone, err = d.client.GetZoneByName(ctx, data.Name.ValueString())
	} else {
		resp.Diagnostics.AddError(
			"Missing Required Attribute",
			"Either 'id' or 'name' must be specified to look up a zone.",
		)
		return
	}

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read zone, got error: %s", err))
		return
	}

	// Map response to model
	data.ID = types.StringValue(zone.ID)
	data.Name = types.StringValue(zone.Name)
	data.Type = types.StringValue(zone.Type)
	data.Region = types.StringValue(zone.Region)
	data.Status = types.StringValue(zone.Status)
	data.CustomerID = types.StringValue(zone.CustomerID)
	data.CreatedAt = types.StringValue(zone.CreatedAt.Format("2006-01-02T15:04:05Z"))
	data.UpdatedAt = types.StringValue(zone.UpdatedAt.Format("2006-01-02T15:04:05Z"))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
