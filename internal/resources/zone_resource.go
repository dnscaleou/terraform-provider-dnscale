package resources

import (
	"context"
	"fmt"

	"github.com/dnscale/terraform-provider-dnscale/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &ZoneResource{}
	_ resource.ResourceWithImportState = &ZoneResource{}
)

// NewZoneResource creates a new zone resource
func NewZoneResource() resource.Resource {
	return &ZoneResource{}
}

// ZoneResource defines the resource implementation.
type ZoneResource struct {
	client *client.Client
}

// ZoneResourceModel describes the resource data model.
type ZoneResourceModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Type       types.String `tfsdk:"type"`
	Region     types.String `tfsdk:"region"`
	Status     types.String `tfsdk:"status"`
	CustomerID types.String `tfsdk:"customer_id"`
	CreatedAt  types.String `tfsdk:"created_at"`
	UpdatedAt  types.String `tfsdk:"updated_at"`
}

func (r *ZoneResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_zone"
}

func (r *ZoneResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a DNS zone in DNScale.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Zone UUID identifier.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Zone name (FQDN, e.g., example.com).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				Description: "Zone type. Valid values: master, slave. Default: master.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("master"),
				Validators: []validator.String{
					stringvalidator.OneOf("master", "slave"),
				},
			},
			"region": schema.StringAttribute{
				Description: "Zone region. Valid values: EU, GLOBAL, EU_GLOBAL.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("EU", "GLOBAL", "EU_GLOBAL"),
				},
			},
			"status": schema.StringAttribute{
				Description: "Zone status. Valid values: active, paused. Default: active.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("active"),
				Validators: []validator.String{
					stringvalidator.OneOf("active", "paused", "pending", "error"),
				},
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

func (r *ZoneResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *ZoneResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ZoneResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Create zone input
	input := client.ZoneInput{
		Name: data.Name.ValueString(),
	}

	if !data.Type.IsNull() && !data.Type.IsUnknown() {
		input.Type = data.Type.ValueString()
	}
	if !data.Region.IsNull() && !data.Region.IsUnknown() {
		input.Region = data.Region.ValueString()
	}
	if !data.Status.IsNull() && !data.Status.IsUnknown() {
		input.Status = data.Status.ValueString()
	}

	// Create zone via API
	zone, err := r.client.CreateZone(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create zone, got error: %s", err))
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

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ZoneResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ZoneResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get zone from API
	zone, err := r.client.GetZone(ctx, data.ID.ValueString())
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

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ZoneResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ZoneResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Create zone input for update
	input := client.ZoneInput{}

	if !data.Type.IsNull() && !data.Type.IsUnknown() {
		input.Type = data.Type.ValueString()
	}
	if !data.Region.IsNull() && !data.Region.IsUnknown() {
		input.Region = data.Region.ValueString()
	}
	if !data.Status.IsNull() && !data.Status.IsUnknown() {
		input.Status = data.Status.ValueString()
	}

	// Update zone via API
	zone, err := r.client.UpdateZone(ctx, data.ID.ValueString(), input)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update zone, got error: %s", err))
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

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ZoneResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ZoneResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Delete zone via API
	err := r.client.DeleteZone(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete zone, got error: %s", err))
		return
	}
}

func (r *ZoneResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
