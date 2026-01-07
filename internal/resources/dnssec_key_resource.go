package resources

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/dnscale/terraform-provider-dnscale/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &DNSSECKeyResource{}
	_ resource.ResourceWithImportState = &DNSSECKeyResource{}
)

// NewDNSSECKeyResource creates a new DNSSEC key resource
func NewDNSSECKeyResource() resource.Resource {
	return &DNSSECKeyResource{}
}

// DNSSECKeyResource defines the resource implementation.
type DNSSECKeyResource struct {
	client *client.Client
}

// DNSSECKeyResourceModel describes the resource data model.
type DNSSECKeyResourceModel struct {
	ID        types.Int64  `tfsdk:"id"`
	ZoneID    types.String `tfsdk:"zone_id"`
	KeyType   types.String `tfsdk:"key_type"`
	Algorithm types.String `tfsdk:"algorithm"`
	Bits      types.Int64  `tfsdk:"bits"`
	Active    types.Bool   `tfsdk:"active"`
	Published types.Bool   `tfsdk:"published"`
	KeyTag    types.Int64  `tfsdk:"key_tag"`
	DNSKey    types.String `tfsdk:"dnskey"`
	DS        types.List   `tfsdk:"ds"`
}

func (r *DNSSECKeyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dnssec_key"
}

func (r *DNSSECKeyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a DNSSEC cryptographic key in DNScale.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "Cryptokey ID.",
				Computed:    true,
			},
			"zone_id": schema.StringAttribute{
				Description: "Zone UUID that this key belongs to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"key_type": schema.StringAttribute{
				Description: "Key type: KSK (Key-Signing-Key) or ZSK (Zone-Signing-Key).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("KSK", "ZSK", "ksk", "zsk"),
				},
			},
			"algorithm": schema.StringAttribute{
				Description: "Signing algorithm (e.g., ECDSAP256SHA256, RSASHA256).",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"bits": schema.Int64Attribute{
				Description: "Key size in bits.",
				Optional:    true,
				Computed:    true,
			},
			"active": schema.BoolAttribute{
				Description: "Whether the key is used for signing. Default: true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"published": schema.BoolAttribute{
				Description: "Whether the key is included in the zone. Default: true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"key_tag": schema.Int64Attribute{
				Description: "DNSSEC key tag.",
				Computed:    true,
			},
			"dnskey": schema.StringAttribute{
				Description: "DNSKEY record data.",
				Computed:    true,
			},
			"ds": schema.ListAttribute{
				Description: "DS records for parent zone delegation.",
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (r *DNSSECKeyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DNSSECKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DNSSECKeyResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build cryptokey input
	input := client.CryptokeyInput{
		KeyType: strings.ToLower(data.KeyType.ValueString()),
	}

	if !data.Algorithm.IsNull() && !data.Algorithm.IsUnknown() {
		input.Algorithm = data.Algorithm.ValueString()
	}

	if !data.Bits.IsNull() && !data.Bits.IsUnknown() {
		bits := int(data.Bits.ValueInt64())
		input.Bits = &bits
	}

	if !data.Active.IsNull() && !data.Active.IsUnknown() {
		active := data.Active.ValueBool()
		input.Active = &active
	}

	if !data.Published.IsNull() && !data.Published.IsUnknown() {
		published := data.Published.ValueBool()
		input.Published = &published
	}

	// Create cryptokey via API
	key, err := r.client.CreateCryptokey(ctx, data.ZoneID.ValueString(), input)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create DNSSEC key, got error: %s", err))
		return
	}

	// Map response to model
	data.ID = types.Int64Value(int64(key.ID))
	data.KeyType = types.StringValue(strings.ToUpper(key.KeyType))
	data.Algorithm = types.StringValue(key.Algorithm)
	data.Bits = types.Int64Value(int64(key.Bits))
	data.Active = types.BoolValue(key.Active)
	data.Published = types.BoolValue(key.Published)
	data.KeyTag = types.Int64Value(int64(key.KeyTag))
	data.DNSKey = types.StringValue(key.DNSKey)

	// Convert DS records to list
	dsValues := make([]types.String, len(key.DS))
	for i, ds := range key.DS {
		dsValues[i] = types.StringValue(ds)
	}
	dsList, diags := types.ListValueFrom(ctx, types.StringType, dsValues)
	resp.Diagnostics.Append(diags...)
	data.DS = dsList

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DNSSECKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DNSSECKeyResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get cryptokey from API
	key, err := r.client.GetCryptokey(ctx, data.ZoneID.ValueString(), int(data.ID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read DNSSEC key, got error: %s", err))
		return
	}

	// Map response to model
	data.ID = types.Int64Value(int64(key.ID))
	data.KeyType = types.StringValue(strings.ToUpper(key.KeyType))
	data.Algorithm = types.StringValue(key.Algorithm)
	data.Bits = types.Int64Value(int64(key.Bits))
	data.Active = types.BoolValue(key.Active)
	data.Published = types.BoolValue(key.Published)
	data.KeyTag = types.Int64Value(int64(key.KeyTag))
	data.DNSKey = types.StringValue(key.DNSKey)

	// Convert DS records to list
	dsValues := make([]types.String, len(key.DS))
	for i, ds := range key.DS {
		dsValues[i] = types.StringValue(ds)
	}
	dsList, diags := types.ListValueFrom(ctx, types.StringType, dsValues)
	resp.Diagnostics.Append(diags...)
	data.DS = dsList

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DNSSECKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DNSSECKeyResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build update input (only active and published can be updated)
	input := client.CryptokeyUpdate{}

	if !data.Active.IsNull() && !data.Active.IsUnknown() {
		active := data.Active.ValueBool()
		input.Active = &active
	}

	if !data.Published.IsNull() && !data.Published.IsUnknown() {
		published := data.Published.ValueBool()
		input.Published = &published
	}

	// Update cryptokey via API
	key, err := r.client.UpdateCryptokey(ctx, data.ZoneID.ValueString(), int(data.ID.ValueInt64()), input)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update DNSSEC key, got error: %s", err))
		return
	}

	// Map response to model
	data.ID = types.Int64Value(int64(key.ID))
	data.KeyType = types.StringValue(strings.ToUpper(key.KeyType))
	data.Algorithm = types.StringValue(key.Algorithm)
	data.Bits = types.Int64Value(int64(key.Bits))
	data.Active = types.BoolValue(key.Active)
	data.Published = types.BoolValue(key.Published)
	data.KeyTag = types.Int64Value(int64(key.KeyTag))
	data.DNSKey = types.StringValue(key.DNSKey)

	// Convert DS records to list
	dsValues := make([]types.String, len(key.DS))
	for i, ds := range key.DS {
		dsValues[i] = types.StringValue(ds)
	}
	dsList, diags := types.ListValueFrom(ctx, types.StringType, dsValues)
	resp.Diagnostics.Append(diags...)
	data.DS = dsList

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DNSSECKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DNSSECKeyResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete cryptokey via API
	err := r.client.DeleteCryptokey(ctx, data.ZoneID.ValueString(), int(data.ID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete DNSSEC key, got error: %s", err))
		return
	}
}

func (r *DNSSECKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: zone_id/key_id
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"Import ID must be in format: zone_id/key_id",
		)
		return
	}

	keyID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Key ID",
			"Key ID must be a valid integer",
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("zone_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), keyID)...)
}
