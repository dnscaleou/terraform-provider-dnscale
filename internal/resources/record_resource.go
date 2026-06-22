package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/dnscale/terraform-provider-dnscale/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &RecordResource{}
	_ resource.ResourceWithImportState = &RecordResource{}
)

// NewRecordResource creates a new record resource
func NewRecordResource() resource.Resource {
	return &RecordResource{}
}

// RecordResource defines the resource implementation.
type RecordResource struct {
	client *client.Client
}

// RecordResourceModel describes the resource data model.
type RecordResourceModel struct {
	ID       types.String `tfsdk:"id"`
	ZoneID   types.String `tfsdk:"zone_id"`
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

func (r *RecordResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_record"
}

func (r *RecordResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a DNS record in DNScale.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Record ID (base64 encoded).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"zone_id": schema.StringAttribute{
				Description: "Zone UUID that this record belongs to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Full record name with trailing dot (e.g., www.example.com.).",
				Required:    true,
			},
			"type": schema.StringAttribute{
				Description: "DNS record type (A, AAAA, CNAME, MX, TXT, NS, SRV, CAA, PTR, ALIAS, TLSA, SSHFP, HTTPS, SVCB).",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						"A", "AAAA", "CNAME", "MX", "TXT", "NS",
						"SRV", "CAA", "PTR", "ALIAS", "TLSA", "SSHFP",
						"HTTPS", "SVCB",
					),
				},
			},
			"content": schema.StringAttribute{
				Description: "Record value (IP address, hostname, text, etc.).",
				Required:    true,
			},
			"ttl": schema.Int64Attribute{
				Description: "Time-to-live in seconds. Must be between 300 and 86400. Default: 3600.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(3600),
				Validators: []validator.Int64{
					int64validator.AtLeast(300),
					int64validator.AtMost(86400),
				},
			},
			"priority": schema.Int64Attribute{
				Description: "Priority for MX and SRV records (0-65535).",
				Optional:    true,
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
					int64validator.AtMost(65535),
				},
			},
			"disabled": schema.BoolAttribute{
				Description: "Whether the record is disabled. Default: false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
		},
	}
}

func (r *RecordResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *RecordResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data RecordResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build record input
	input := client.RecordInput{
		Name:     data.Name.ValueString(),
		Type:     data.Type.ValueString(),
		Content:  data.Content.ValueString(),
		TTL:      int(data.TTL.ValueInt64()),
		Disabled: data.Disabled.ValueBool(),
	}

	if !data.Priority.IsNull() && !data.Priority.IsUnknown() {
		priority := int(data.Priority.ValueInt64())
		input.Priority = &priority
	}

	// Create record via API
	record, err := r.client.CreateRecord(ctx, data.ZoneID.ValueString(), input)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create record, got error: %s", err))
		return
	}

	// For TXT/MX/SRV records, the API may return a different ID or content than what
	// subsequent List/Get calls return. Fetch the canonical record to ensure we store
	// the correct ID and normalized content.
	recordType := data.Type.ValueString()
	if recordType == "TXT" || recordType == "MX" || recordType == "SRV" {
		records, err := r.client.ListRecords(ctx, data.ZoneID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list records after create, got error: %s", err))
			return
		}
		normalizedContent := normalizeRecordContent(record.Type, record.Content, record.Priority)
		for _, r := range records {
			if r.Name == record.Name && r.Type == record.Type &&
				normalizeRecordContent(r.Type, r.Content, r.Priority) == normalizedContent {
				record = &r
				break
			}
		}
	}

	// Map response to model
	data.ID = types.StringValue(record.ID)
	data.Name = types.StringValue(record.Name)
	data.Type = types.StringValue(record.Type)
	data.Content = types.StringValue(normalizeRecordContent(record.Type, record.Content, record.Priority))
	data.TTL = types.Int64Value(int64(record.TTL))
	data.Disabled = types.BoolValue(record.Disabled)

	if record.Priority != nil {
		data.Priority = types.Int64Value(int64(*record.Priority))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RecordResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data RecordResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get record from API
	record, err := r.client.GetRecord(ctx, data.ZoneID.ValueString(), data.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read record, got error: %s", err))
		return
	}

	// Map response to model
	data.ID = types.StringValue(record.ID)
	data.Name = types.StringValue(record.Name)
	data.Type = types.StringValue(record.Type)
	data.Content = types.StringValue(normalizeRecordContent(record.Type, record.Content, record.Priority))
	data.TTL = types.Int64Value(int64(record.TTL))
	data.Disabled = types.BoolValue(record.Disabled)

	if record.Priority != nil {
		data.Priority = types.Int64Value(int64(*record.Priority))
	} else if record.Type == "MX" {
		// For MX records, the API embeds priority in content (e.g., "10 mail.example.com.")
		// Extract it from content if not provided separately
		parts := strings.SplitN(record.Content, " ", 2)
		if len(parts) == 2 {
			var priority int
			if _, err := fmt.Sscanf(parts[0], "%d", &priority); err == nil {
				data.Priority = types.Int64Value(int64(priority))
			} else {
				data.Priority = types.Int64Null()
			}
		} else {
			data.Priority = types.Int64Null()
		}
	} else {
		data.Priority = types.Int64Null()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RecordResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data RecordResourceModel
	var state RecordResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build record input
	input := client.RecordInput{
		Name:     data.Name.ValueString(),
		Type:     data.Type.ValueString(),
		Content:  data.Content.ValueString(),
		TTL:      int(data.TTL.ValueInt64()),
		Disabled: data.Disabled.ValueBool(),
	}

	if !data.Priority.IsNull() && !data.Priority.IsUnknown() {
		priority := int(data.Priority.ValueInt64())
		input.Priority = &priority
	}

	// Update record via API (uses the state's ID)
	record, err := r.client.UpdateRecord(ctx, data.ZoneID.ValueString(), state.ID.ValueString(), input)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update record, got error: %s", err))
		return
	}

	// For TXT/MX/SRV records, fetch the canonical record to get the correct ID
	recordType := data.Type.ValueString()
	if recordType == "TXT" || recordType == "MX" || recordType == "SRV" {
		records, err := r.client.ListRecords(ctx, data.ZoneID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list records after update, got error: %s", err))
			return
		}
		normalizedContent := normalizeRecordContent(record.Type, record.Content, record.Priority)
		for _, r := range records {
			if r.Name == record.Name && r.Type == record.Type &&
				normalizeRecordContent(r.Type, r.Content, r.Priority) == normalizedContent {
				record = &r
				break
			}
		}
	}

	// Map response to model (ID may change if name/type/content changed)
	data.ID = types.StringValue(record.ID)
	data.Name = types.StringValue(record.Name)
	data.Type = types.StringValue(record.Type)
	data.Content = types.StringValue(normalizeRecordContent(record.Type, record.Content, record.Priority))
	data.TTL = types.Int64Value(int64(record.TTL))
	data.Disabled = types.BoolValue(record.Disabled)

	if record.Priority != nil {
		data.Priority = types.Int64Value(int64(*record.Priority))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RecordResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data RecordResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete record via API
	err := r.client.DeleteRecord(ctx, data.ZoneID.ValueString(), data.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete record, got error: %s", err))
		return
	}
}

func (r *RecordResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: zone_id/record_id
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"Import ID must be in format: zone_id/record_id",
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("zone_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}
