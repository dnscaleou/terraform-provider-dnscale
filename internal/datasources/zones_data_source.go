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
var _ datasource.DataSource = &ZonesDataSource{}

// NewZonesDataSource creates a new zones data source
func NewZonesDataSource() datasource.DataSource {
	return &ZonesDataSource{}
}

// ZonesDataSource defines the data source implementation.
type ZonesDataSource struct {
	client *client.Client
}

// ZonesDataSourceModel describes the data source data model.
type ZonesDataSourceModel struct {
	Zones []ZoneModel `tfsdk:"zones"`
}

// ZoneModel describes a single zone in the zones list.
type ZoneModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Type       types.String `tfsdk:"type"`
	Region     types.String `tfsdk:"region"`
	Status     types.String `tfsdk:"status"`
	CustomerID types.String `tfsdk:"customer_id"`
	CreatedAt  types.String `tfsdk:"created_at"`
	UpdatedAt  types.String `tfsdk:"updated_at"`
}

func (d *ZonesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_zones"
}

func (d *ZonesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to list all DNS zones.",
		Attributes: map[string]schema.Attribute{
			"zones": schema.ListNestedAttribute{
				Description: "List of DNS zones.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Zone UUID identifier.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Zone name (FQDN).",
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
				},
			},
		},
	}
}

func (d *ZonesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ZonesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ZonesDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get all zones from API
	zones, err := d.client.ListZones(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list zones, got error: %s", err))
		return
	}

	// Map response to model
	data.Zones = make([]ZoneModel, len(zones))
	for i, zone := range zones {
		data.Zones[i] = ZoneModel{
			ID:         types.StringValue(zone.ID),
			Name:       types.StringValue(zone.Name),
			Type:       types.StringValue(zone.Type),
			Region:     types.StringValue(zone.Region),
			Status:     types.StringValue(zone.Status),
			CustomerID: types.StringValue(zone.CustomerID),
			CreatedAt:  types.StringValue(zone.CreatedAt.Format("2006-01-02T15:04:05Z")),
			UpdatedAt:  types.StringValue(zone.UpdatedAt.Format("2006-01-02T15:04:05Z")),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
