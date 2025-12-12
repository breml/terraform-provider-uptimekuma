package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccStatusPageGroupIDs(t *testing.T) {
	slug := acctest.RandomWithPrefix("test-groupid")
	title := "Status Page Group ID Test"
	monitorName := acctest.RandomWithPrefix("test-monitor")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStatusPageResourceConfigWithMonitors(slug, title, monitorName),
			},
		},
	})
}
