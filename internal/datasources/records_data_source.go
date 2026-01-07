package datasources

import (
	"context"
	"fmt"
	"strings"

	"github.com/dnscale/terraform-provider-dnscale/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &RecordsDataSource{}

// NewRecordsDataSource creates a new records data source
func NewRecordsDataSource() datasource.DataSource {
	return &RecordsDataSource{}
}

// RecordsDataSource defines the data source implementation.
type RecordsDataSource struct {
	client *client.Client
}

// RecordsDataSourceModel describes the data source data model.
type RecordsDataSourceModel struct {
	ZoneID  types.String  `tfsdk:"zone_id"`
	Type    types.String  `tfsdk:"type"`
	Records []RecordModel `tfsdk:"records"`
}

// RecordModel describes a single record in the records list.
type RecordModel struct {
	ID       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Type     types.String `tfsdk:"type"`
	Content  types.String `tfsdk:"content"`
	TTL      types.Int64  `tfsdk:"ttl"`
	Priority types.Int64  `tfsdk:"priority"`
	Disabled types.Bool   `tfsdk:"disabled"`
}

// normalizeRecordContent normalizes record content based on type
// - TXT: strips surrounding quotes
// - MX: strips priority prefix (API returns "10 mail.example.com." but we store just "mail.example.com.")
// - SRV: keeps as-is since weight/port/target are all part of content
func normalizeRecordContent(recordType, content string, priority *int) string {
	switch recordType {
	case "TXT":
		if len(content) >= 2 {
			if (content[0] == '"' && content[len(content)-1] == '"') ||
				(content[0] == '\'' && content[len(content)-1] == '\'') {
				return content[1 : len(content)-1]
			}
		}
	case "MX":
		// DNScale API returns MX content with priority prefix embedded
		// Content like "10 mail.example.com." should become "mail.example.com."
		parts := strings.SplitN(content, " ", 2)
		if len(parts) == 2 && len(parts[0]) > 0 {
			// Check if first part is a number (priority)
			isNumber := true
			for _, c := range parts[0] {
				if c < '0' || c > '9' {
					isNumber = false
					break
				}
			}
			if isNumber {
				return parts[1]
			}
		}
	}
	return content
}

func (d *RecordsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_records"
}

func (d *RecordsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to list DNS records in a zone.",
		Attributes: map[string]schema.Attribute{
			"zone_id": schema.StringAttribute{
				Description: "Zone UUID to list records from.",
				Required:    true,
			},
			"type": schema.StringAttribute{
				Description: "Optional filter by record type (A, AAAA, CNAME, etc.).",
				Optional:    true,
			},
			"records": schema.ListNestedAttribute{
				Description: "List of DNS records.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Record ID.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Record name (FQDN with trailing dot).",
							Computed:    true,
						},
						"type": schema.StringAttribute{
							Description: "Record type.",
							Computed:    true,
						},
						"content": schema.StringAttribute{
							Description: "Record value.",
							Computed:    true,
						},
						"ttl": schema.Int64Attribute{
							Description: "Time-to-live in seconds.",
							Computed:    true,
						},
						"priority": schema.Int64Attribute{
							Description: "Priority for MX/SRV records.",
							Computed:    true,
						},
						"disabled": schema.BoolAttribute{
							Description: "Whether the record is disabled.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *RecordsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *RecordsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data RecordsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get all records from API
	records, err := d.client.ListRecords(ctx, data.ZoneID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list records, got error: %s", err))
		return
	}

	// Filter by type if specified
	typeFilter := ""
	if !data.Type.IsNull() && data.Type.ValueString() != "" {
		typeFilter = data.Type.ValueString()
	}

	// Map response to model
	data.Records = make([]RecordModel, 0)
	for _, record := range records {
		// Apply type filter if specified
		if typeFilter != "" && record.Type != typeFilter {
			continue
		}

		recordModel := RecordModel{
			ID:       types.StringValue(record.ID),
			Name:     types.StringValue(record.Name),
			Type:     types.StringValue(record.Type),
			Content:  types.StringValue(normalizeRecordContent(record.Type, record.Content, record.Priority)),
			TTL:      types.Int64Value(int64(record.TTL)),
			Disabled: types.BoolValue(record.Disabled),
		}

		if record.Priority != nil {
			recordModel.Priority = types.Int64Value(int64(*record.Priority))
		} else {
			recordModel.Priority = types.Int64Null()
		}

		data.Records = append(data.Records, recordModel)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
