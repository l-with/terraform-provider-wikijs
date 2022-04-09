package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAuthenticationStrategyDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccAuthenticationStrategyDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.wikijs_authentication_strategy.test", "title", "example-title"),
				),
			},
		},
	})
}

const testAccAuthenticationStrategyDataSourceConfig = `
data "wikijs_authentication_strategy" "test" {
	key = "key"
}
`
