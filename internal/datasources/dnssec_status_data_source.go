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
var _ datasource.DataSource = &DNSSECStatusDataSource{}

// NewDNSSECStatusDataSource creates a new DNSSEC status data source
func NewDNSSECStatusDataSource() datasource.DataSource {
	return &DNSSECStatusDataSource{}
}

// DNSSECStatusDataSource defines the data source implementation.
type DNSSECStatusDataSource struct {
	client *client.Client
}

// DNSSECStatusDataSourceModel describes the data source data model.
type DNSSECStatusDataSourceModel struct {
	ZoneID    types.String `tfsdk:"zone_id"`
	Enabled   types.Bool   `tfsdk:"enabled"`
	KeysCount types.Int64  `tfsdk:"keys_count"`
	HasKSK    types.Bool   `tfsdk:"has_ksk"`
	HasZSK    types.Bool   `tfsdk:"has_zsk"`
}

func (d *DNSSECStatusDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dnssec_status"
}

func (d *DNSSECStatusDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to get the DNSSEC status of a zone.",
		Attributes: map[string]schema.Attribute{
			"zone_id": schema.StringAttribute{
				Description: "Zone UUID to get DNSSEC status for.",
				Required:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether DNSSEC is enabled for the zone.",
				Computed:    true,
			},
			"keys_count": schema.Int64Attribute{
				Description: "Number of DNSSEC keys in the zone.",
				Computed:    true,
			},
			"has_ksk": schema.BoolAttribute{
				Description: "Whether the zone has a Key-Signing-Key (KSK).",
				Computed:    true,
			},
			"has_zsk": schema.BoolAttribute{
				Description: "Whether the zone has a Zone-Signing-Key (ZSK).",
				Computed:    true,
			},
		},
	}
}

func (d *DNSSECStatusDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DNSSECStatusDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DNSSECStatusDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get DNSSEC status from API
	status, err := d.client.GetDNSSECStatus(ctx, data.ZoneID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get DNSSEC status, got error: %s", err))
		return
	}

	// Map response to model
	data.Enabled = types.BoolValue(status.Enabled)
	data.KeysCount = types.Int64Value(int64(status.KeysCount))
	data.HasKSK = types.BoolValue(status.HasKSK)
	data.HasZSK = types.BoolValue(status.HasZSK)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
