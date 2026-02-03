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
				Config:             testAccStatusPageResourceConfigWithMonitors(slug, title, monitorName),
				ExpectNonEmptyPlan: false,
				ConfigStateChecks: []statecheck.StateCheck{
					// Verify that the server-assigned group ID is populated in state.
					statecheck.ExpectKnownValue(
						"uptimekuma_status_page.test",
						tfjsonpath.New("public_group_list").AtSliceIndex(0).AtMapKey("id"),
						knownvalue.NotNull(),
					),
				},
			},
			{
				// Identical config, no changes. Reproduces issue #223: without the
				// UseStateForUnknown plan modifier on public_group_list[].id, Terraform
				// marks the group ID as (known after apply) on every plan, producing a
				// perpetual diff.
				Config:             testAccStatusPageResourceConfigWithMonitors(slug, title, monitorName),
				ExpectNonEmptyPlan: false,
			},
		},
	})
}
