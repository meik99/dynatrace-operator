package version

import (
	"context"
	"testing"

	"github.com/Dynatrace/dynatrace-operator/pkg/api/scheme/fake"
	"github.com/Dynatrace/dynatrace-operator/pkg/api/status"
	dynatracev1beta2 "github.com/Dynatrace/dynatrace-operator/pkg/api/v1beta2/dynakube"
	dtclient "github.com/Dynatrace/dynatrace-operator/pkg/clients/dynatrace"
	dtclientmock "github.com/Dynatrace/dynatrace-operator/test/mocks/pkg/clients/dynatrace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestActiveGateUpdater(t *testing.T) {
	ctx := context.Background()
	testImage := dtclient.LatestImageInfo{
		Source: "some.registry.com",
		Tag:    "1.2.3.4-5",
	}

	t.Run("Getters work as expected", func(t *testing.T) {
		dynakube := &dynatracev1beta2.DynaKube{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					dynatracev1beta2.AnnotationFeatureDisableActiveGateUpdates: "true",
				},
			},
			Spec: dynatracev1beta2.DynaKubeSpec{
				ActiveGate: dynatracev1beta2.ActiveGateSpec{
					Capabilities: []dynatracev1beta2.CapabilityDisplayName{dynatracev1beta2.DynatraceApiCapability.DisplayName},
					CapabilityProperties: dynatracev1beta2.CapabilityProperties{
						Image: testImage.String(),
					},
				},
			},
		}
		mockClient := dtclientmock.NewClient(t)
		mockActiveGateImageInfo(mockClient, testImage)

		updater := newActiveGateUpdater(dynakube, fake.NewClient(), mockClient)

		assert.Equal(t, "activegate", updater.Name())
		assert.True(t, updater.IsEnabled())
		assert.Equal(t, dynakube.Spec.ActiveGate.Image, updater.CustomImage())
		assert.Equal(t, "", updater.CustomVersion())
		assert.False(t, updater.IsAutoUpdateEnabled())
		imageInfo, err := updater.LatestImageInfo(ctx)
		require.NoError(t, err)
		assert.Equal(t, testImage, *imageInfo)
	})
}

func TestActiveGateUseDefault(t *testing.T) {
	t.Run("Set according to defaults, unset previous status", func(t *testing.T) {
		dynakube := &dynatracev1beta2.DynaKube{
			Spec: dynatracev1beta2.DynaKubeSpec{
				APIURL: testApiUrl,
				ActiveGate: dynatracev1beta2.ActiveGateSpec{
					CapabilityProperties: dynatracev1beta2.CapabilityProperties{},
				},
			},
			Status: dynatracev1beta2.DynaKubeStatus{
				ActiveGate: dynatracev1beta2.ActiveGateStatus{
					VersionStatus: status.VersionStatus{
						Version: "prev",
					},
				},
			},
		}
		expectedVersion := "1.2.3.4-5"
		expectedImage := dynakube.DefaultActiveGateImage(expectedVersion)
		mockClient := dtclientmock.NewClient(t)

		mockClient.On("GetLatestActiveGateVersion", mock.AnythingOfType("context.backgroundCtx"), mock.Anything).Return(expectedVersion, nil)

		updater := newActiveGateUpdater(dynakube, fake.NewClient(), mockClient)

		err := updater.UseTenantRegistry(context.Background())
		require.NoError(t, err)
		assert.Equal(t, expectedImage, dynakube.Status.ActiveGate.ImageID)
		assert.Equal(t, expectedVersion, dynakube.Status.ActiveGate.Version)
	})
}

func TestActiveGateIsEnabled(t *testing.T) {
	t.Run("cleans up if not enabled", func(t *testing.T) {
		dynakube := &dynatracev1beta2.DynaKube{
			Status: dynatracev1beta2.DynaKubeStatus{
				ActiveGate: dynatracev1beta2.ActiveGateStatus{
					VersionStatus: status.VersionStatus{
						Version: "prev",
					},
				},
			},
		}

		updater := newActiveGateUpdater(dynakube, nil, nil)

		isEnabled := updater.IsEnabled()
		require.False(t, isEnabled)

		assert.Empty(t, updater.Target())
	})
}
