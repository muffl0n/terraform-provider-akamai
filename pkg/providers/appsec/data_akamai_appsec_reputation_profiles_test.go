package appsec

import (
	"encoding/json"
	"testing"

	"github.com/akamai/AkamaiOPEN-edgegrid-golang/v2/pkg/appsec"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/stretchr/testify/mock"
)

func TestAccAkamaiReputationProfiles_data_basic(t *testing.T) {
	t.Run("match by ReputationProfiles ID", func(t *testing.T) {
		client := &mockappsec{}

		cv := appsec.GetReputationProfilesResponse{}
		expectJS := compactJSON(loadFixtureBytes("testdata/TestDSReputationProfiles/ReputationProfiles.json"))
		json.Unmarshal([]byte(expectJS), &cv)

		client.On("GetReputationProfiles",
			mock.Anything, // ctx is irrelevant for this test
			appsec.GetReputationProfilesRequest{ConfigID: 43253, ConfigVersion: 7, ReputationProfileId: 12345},
		).Return(&cv, nil)

		useClient(client, func() {
			resource.Test(t, resource.TestCase{
				IsUnitTest: true,
				Providers:  testAccProviders,
				Steps: []resource.TestStep{
					{
						Config: loadFixtureString("testdata/TestDSReputationProfiles/match_by_id.tf"),
						Check: resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttr("data.akamai_appsec_reputation_profiles.test", "id", "43253"),
						),
					},
				},
			})
		})

		client.AssertExpectations(t)
	})

}