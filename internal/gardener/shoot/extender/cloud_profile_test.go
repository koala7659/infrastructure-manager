package extender

import (
	"testing"

	imv1 "github.com/kyma-project/infrastructure-manager/api/v1"
	"github.com/kyma-project/infrastructure-manager/internal/gardener/shoot/hyperscaler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtendWithCloudProfile(t *testing.T) {
	for _, testCase := range []struct {
		name            string
		providerType    string
		expectedProfile string
	}{
		{
			name:            "Set cloud profile name for aws",
			providerType:    hyperscaler.TypeAWS,
			expectedProfile: DefaultAWSCloudProfileName,
		},
		{
			name:            "Set cloud profile name for azure",
			providerType:    hyperscaler.TypeAzure,
			expectedProfile: DefaultAzureCloudProfileName,
		},
		{
			name:            "Set cloud profile for gcp",
			providerType:    hyperscaler.TypeGCP,
			expectedProfile: DefaultGCPCloudProfileName,
		},
		{
			name:            "Set cloud profile for openstack",
			providerType:    hyperscaler.TypeOpenStack,
			expectedProfile: DefaultOpenStackCloudProfileName,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			// given
			runtime := imv1.Runtime{
				Spec: imv1.RuntimeSpec{
					Shoot: imv1.RuntimeShoot{
						Name: "myshoot",
						Provider: imv1.Provider{
							Type: testCase.providerType,
						},
					},
				},
			}
			shoot := fixEmptyGardenerShoot("test", "dev")

			// when
			err := ExtendWithCloudProfile(runtime, &shoot)

			// then
			require.NoError(t, err)
			assert.Equal(t, testCase.expectedProfile, shoot.Spec.CloudProfileName)
		})
	}

	t.Run("", func(t *testing.T) {
		// given
		runtime := imv1.Runtime{
			Spec: imv1.RuntimeSpec{
				Shoot: imv1.RuntimeShoot{
					Name: "myshoot",
					Provider: imv1.Provider{
						Type: "unknown",
					},
				},
			},
		}
		shoot := fixEmptyGardenerShoot("test", "dev")

		// when
		err := ExtendWithCloudProfile(runtime, &shoot)

		// then
		require.Error(t, err)
	})
}
