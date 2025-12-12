package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
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
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("uptimekuma_status_page.test", tfjsonpath.New("public_group_list[0].id"), knownvalue.NotNull()),
				},
			},
		},
	})
}
