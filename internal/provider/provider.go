package provider

import (
	"context"
	"fmt"
	"os"

	"github.com/camjjack/terraform-provider-wikijs/wikijs"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// provider satisfies the tfsdk.Provider interface and usually is included
// with all Resource and DataSource implementations.
type provider struct {
	// client can contain the upstream provider SDK or HTTP client used to
	// communicate with the upstream service. Resource and DataSource
	// implementations can then make calls using this client.
	//
	// TODO: If appropriate, implement upstream provider SDK or HTTP client.
	// client vendorsdk.ExampleClient
	client *wikijs.WikijsClient

	// configured is set to true at the end of the Configure method.
	// This can be used in Resource and DataSource implementations to verify
	// that the provider was previously configured.
	configured bool

	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// providerData can be used to store data from the Terraform configuration.
type providerData struct {
	Host          types.String `tfsdk:"host"`
	Username      types.String `tfsdk:"username"`
	Password      types.String `tfsdk:"password"`
	InitialSetup  types.Bool   `tfsdk:"initial_setup"`
	ClientTimeout types.Int64  `tfsdk:"client_timeout"`
	CaCert        types.String `tfsdk:"ca_cert"`
}

func (p *provider) Configure(ctx context.Context, req tfsdk.ConfigureProviderRequest, resp *tfsdk.ConfigureProviderResponse) {

	var data providerData
	diags := req.Config.Get(ctx, &data)

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
	var host string
	if data.Host.Null {
		host = os.Getenv("WIKIJS_HOST")
	} else {
		host = data.Host.Value
	}

	var username string
	if data.Username.Null {
		username = os.Getenv("WIKIJS_USERNAME")
	} else {
		username = data.Username.Value
	}

	var password string
	if data.Password.Null {
		password = os.Getenv("WIKIJS_PASSWORD")
	} else {
		password = data.Password.Value
	}

	client, err := wikijs.NewWikijsClient(host, username, password, data.InitialSetup.Value, data.ClientTimeout.Value, data.CaCert.Value)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create client",
			"Unable to create wikijs client:\n\n"+err.Error(),
		)
		return
	}
	p.client = client
	p.configured = true

}

func (p *provider) GetResources(ctx context.Context) (map[string]tfsdk.ResourceType, diag.Diagnostics) {
	return map[string]tfsdk.ResourceType{
		//"scaffolding_example": exampleResourceType{},
	}, nil
}

func (p *provider) GetDataSources(ctx context.Context) (map[string]tfsdk.DataSourceType, diag.Diagnostics) {
	return map[string]tfsdk.DataSourceType{
		"wikijs_authentication_strategy": authenticationStrategyDataSourceType{},
	}, nil
}

func (p *provider) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"host": {
				MarkdownDescription: "wikijs host",
				Type:                types.StringType,
				Optional:            true,
			},
			"username": {
				MarkdownDescription: "wikijs administrator username",
				Type:                types.StringType,
				Optional:            true,
			},
			"password": {
				MarkdownDescription: "wikijs administrator password",
				Type:                types.StringType,
				Optional:            true,
				Sensitive:           true,
			},
			"initial_setup": {
				MarkdownDescription: "Conduct intial setup request",
				Type:                types.BoolType,
				Optional:            true,
				//Default:           true,
			},
			"client_timeout": {
				MarkdownDescription: "Timeout for client",
				Type:                types.Int64Type,
				Optional:            true,
			},
			"ca_cert": {
				MarkdownDescription: "Root CA certificate (useful for development purposes)",
				Type:                types.StringType,
				Optional:            true,
			},
		},
	}, nil
}

func New(version string) func() tfsdk.Provider {
	return func() tfsdk.Provider {
		return &provider{
			version: version,
		}
	}
}

// convertProviderType is a helper function for NewResource and NewDataSource
// implementations to associate the concrete provider type. Alternatively,
// this helper can be skipped and the provider type can be directly type
// asserted (e.g. provider: in.(*provider)), however using this can prevent
// potential panics.
func convertProviderType(in tfsdk.Provider) (provider, diag.Diagnostics) {
	var diags diag.Diagnostics

	p, ok := in.(*provider)

	if !ok {
		diags.AddError(
			"Unexpected Provider Instance Type",
			fmt.Sprintf("While creating the data source or resource, an unexpected provider type (%T) was received. This is always a bug in the provider code and should be reported to the provider developers.", p),
		)
		return provider{}, diags
	}

	if p == nil {
		diags.AddError(
			"Unexpected Provider Instance Type",
			"While creating the data source or resource, an unexpected empty provider instance was received. This is always a bug in the provider code and should be reported to the provider developers.",
		)
		return provider{}, diags
	}

	return *p, diags
}
