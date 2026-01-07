package provider

import (
	"context"
	"os"

	"github.com/dnscale/terraform-provider-dnscale/internal/client"
	"github.com/dnscale/terraform-provider-dnscale/internal/datasources"
	"github.com/dnscale/terraform-provider-dnscale/internal/resources"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure DNScaleProvider satisfies various provider interfaces.
var _ provider.Provider = &DNScaleProvider{}

// DNScaleProvider defines the provider implementation.
type DNScaleProvider struct {
	version string
}

// DNScaleProviderModel describes the provider data model.
type DNScaleProviderModel struct {
	APIKey types.String `tfsdk:"api_key"`
	APIURL types.String `tfsdk:"api_url"`
}

// New creates a new provider factory function
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &DNScaleProvider{
			version: version,
		}
	}
}

func (p *DNScaleProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "dnscale"
	resp.Version = p.version
}

func (p *DNScaleProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Terraform provider for managing DNS zones and records on DNScale.",
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				Description: "DNScale API key for authentication. Can also be set via DNSCALE_API_KEY environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
			"api_url": schema.StringAttribute{
				Description: "DNScale API URL. Defaults to https://api.dnscale.eu. Can also be set via DNSCALE_API_URL environment variable.",
				Optional:    true,
			},
		},
	}
}

func (p *DNScaleProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config DNScaleProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get API key from config or environment
	apiKey := os.Getenv("DNSCALE_API_KEY")
	if !config.APIKey.IsNull() {
		apiKey = config.APIKey.ValueString()
	}

	if apiKey == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Missing DNScale API Key",
			"The provider cannot create the DNScale API client as there is a missing or empty value for the DNScale API key. "+
				"Set the api_key value in the configuration or use the DNSCALE_API_KEY environment variable.",
		)
		return
	}

	// Get API URL from config or environment
	apiURL := os.Getenv("DNSCALE_API_URL")
	if !config.APIURL.IsNull() {
		apiURL = config.APIURL.ValueString()
	}
	if apiURL == "" {
		apiURL = client.DefaultBaseURL
	}

	// Create DNScale client
	dnscaleClient := client.NewClient(apiKey, apiURL)

	// Make client available to data sources and resources
	resp.DataSourceData = dnscaleClient
	resp.ResourceData = dnscaleClient
}

func (p *DNScaleProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		resources.NewZoneResource,
		resources.NewRecordResource,
		resources.NewDNSSECKeyResource,
	}
}

func (p *DNScaleProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		datasources.NewZoneDataSource,
		datasources.NewZonesDataSource,
		datasources.NewRecordsDataSource,
		datasources.NewDNSSECStatusDataSource,
	}
}
