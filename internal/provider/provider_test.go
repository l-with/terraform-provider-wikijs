package provider

import (
	"os"
	"sync"
	"testing"

	"github.com/camjjack/terraform-provider-wikijs/wikijs"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var wikijsClient *wikijs.WikijsClient
var testAccProvider tfsdk.Provider
var testAccProtoV6ProviderFactories map[string]func() (tfprotov6.ProviderServer, error)
var clientConnOnce sync.Once

func init() {
	clientConnOnce.Do(func() {
		wikijsClient, _ = wikijs.NewWikijsClient(os.Getenv("WIKIJS_HOST"), os.Getenv("WIKIJS_USERNAME"), os.Getenv("WIKIJS_PASSWORD"), true, 30, "")
		testAccProvider = New("test", wikijsClient)()
	})
	testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
		"wikijs": providerserver.NewProtocol6WithError(testAccProvider),
	}
}

func testAccPreCheck(t *testing.T) {
	testEnvVars := []string{"WIKIJS_HOST", "WIKIJS_USERNAME", "WIKIJS_PASSWORD"}
	for _, env := range testEnvVars {
		if os.Getenv(env) == "" {
			t.Fatalf("%s must be set for acceptance tests", env)
		}
	}
}

func TestAccInitialSetup(t *testing.T) {
	wikijsClient, err := wikijs.NewWikijsClient(os.Getenv("WIKIJS_HOST"), os.Getenv("WIKIJS_USERNAME"), os.Getenv("WIKIJS_PASSWORD"), true, 30, "")
	if err != nil {
		t.Fatalf("%s", err)
	}

	res, err := wikijsClient.SetupDone()
	if res == false {
		t.Fatalf("%s", err)
	}
	wikijsClient.Cleanup()
}
