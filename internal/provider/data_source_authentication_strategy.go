package provider

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type authenticationStrategyDataSourceType struct{}

func (t authenticationStrategyDataSourceType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Authentication Strategy data source",

		Attributes: map[string]tfsdk.Attribute{
			"key": {
				MarkdownDescription: "Key",
				Optional:            true,
				Type:                types.StringType,
			},
			"title": {
				MarkdownDescription: "Title",
				Type:                types.StringType,
				Computed:            true,
			},
			"id": {
				Type:     types.StringType,
				Computed: true,
			},
		},
	}, nil
}

func (t authenticationStrategyDataSourceType) NewDataSource(ctx context.Context, in tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
	provider, diags := convertProviderType(in)

	return authenticationStrategyDataSource{
		provider: provider,
	}, diags
}

type authenticationStrategyDataSourceData struct {
	Key   types.String `tfsdk:"key"`
	Title types.String `tfsdk:"title"`
	Id    types.String `tfsdk:"id"`
}

type authenticationStrategyDataSource struct {
	provider provider
}

func (d authenticationStrategyDataSource) Read(ctx context.Context, req tfsdk.ReadDataSourceRequest, resp *tfsdk.ReadDataSourceResponse) {
	var data authenticationStrategyDataSourceData

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	log.Printf("got here")

	if resp.Diagnostics.HasError() {
		return
	}

	log.Printf("got here")

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// example, err := d.provider.client.ReadExample(...)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read example, got error: %s", err))
	//     return
	// }

	// For the purposes of this example code, hardcoding a response value to
	// save into the Terraform state.
	data.Title = types.String{Value: "example-title"}
	data.Id = types.String{Value: "example-id"}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}
