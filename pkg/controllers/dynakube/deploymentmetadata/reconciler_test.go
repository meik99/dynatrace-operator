package deploymentmetadata

import (
	"context"
	"testing"

	dynatracev1beta2 "github.com/Dynatrace/dynatrace-operator/pkg/api/v1beta2/dynakube"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	testName      = "test-name"
	testNamespace = "test-namespace"
)

func createTestDynakubeObjectMeta() metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Namespace: testNamespace,
		Name:      testName,
	}
}

func createTestDynakube(spec *dynatracev1beta2.DynaKubeSpec) *dynatracev1beta2.DynaKube {
	dynakube := &dynatracev1beta2.DynaKube{ObjectMeta: createTestDynakubeObjectMeta()}
	if spec != nil {
		dynakube.Spec = *spec
	}

	return dynakube
}

func TestReconcile(t *testing.T) {
	clusterID := "test"

	t.Run(`don't create anything, if no mode is configured`, func(t *testing.T) {
		dynakube := createTestDynakube(nil)
		fakeClient := fake.NewClientBuilder().Build()
		r := NewReconciler(fakeClient, fakeClient, *dynakube, clusterID)
		err := r.Reconcile(context.Background())
		require.NoError(t, err)

		var actualConfigMap corev1.ConfigMap
		err = fakeClient.Get(context.Background(), client.ObjectKey{Name: GetDeploymentMetadataConfigMapName(testName), Namespace: testNamespace}, &actualConfigMap)
		require.Error(t, err)
	})
	t.Run(`delete configmap, if no mode is configured`, func(t *testing.T) {
		dynakube := createTestDynakube(nil)
		fakeClient := fake.NewClientBuilder().WithObjects(
			&corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      GetDeploymentMetadataConfigMapName(testName),
					Namespace: testNamespace,
				},
			},
		).Build()
		r := NewReconciler(fakeClient, fakeClient, *dynakube, clusterID)
		err := r.Reconcile(context.Background())
		require.NoError(t, err)

		var actualConfigMap corev1.ConfigMap
		err = fakeClient.Get(context.Background(), client.ObjectKey{Name: GetDeploymentMetadataConfigMapName(testName), Namespace: testNamespace}, &actualConfigMap)
		require.Error(t, err)
	})

	t.Run(`create configmap with 1 key, if only oneagent is needed`, func(t *testing.T) {
		dynakube := createTestDynakube(
			&dynatracev1beta2.DynaKubeSpec{
				OneAgent: dynatracev1beta2.OneAgentSpec{
					CloudNativeFullStack: &dynatracev1beta2.CloudNativeFullStackSpec{},
				},
			})

		fakeClient := fake.NewClientBuilder().Build()
		r := NewReconciler(fakeClient, fakeClient, *dynakube, clusterID)
		err := r.Reconcile(context.Background())
		require.NoError(t, err)

		var actualConfigMap corev1.ConfigMap
		err = fakeClient.Get(context.Background(), client.ObjectKey{Name: GetDeploymentMetadataConfigMapName(testName), Namespace: testNamespace}, &actualConfigMap)
		require.NoError(t, err)
		require.NotEmpty(t, actualConfigMap.Data)
		assert.NotEmpty(t, actualConfigMap.Data[OneAgentMetadataKey])
	})

	t.Run(`create configmap with 1 key, if only activegate is needed`, func(t *testing.T) {
		dynakube := createTestDynakube(
			&dynatracev1beta2.DynaKubeSpec{
				ActiveGate: dynatracev1beta2.ActiveGateSpec{
					Capabilities: []dynatracev1beta2.CapabilityDisplayName{
						dynatracev1beta2.KubeMonCapability.DisplayName,
					},
				},
			})

		fakeClient := fake.NewClientBuilder().Build()
		r := NewReconciler(fakeClient, fakeClient, *dynakube, clusterID)
		err := r.Reconcile(context.Background())
		require.NoError(t, err)

		var actualConfigMap corev1.ConfigMap
		err = fakeClient.Get(context.Background(), client.ObjectKey{Name: GetDeploymentMetadataConfigMapName(testName), Namespace: testNamespace}, &actualConfigMap)
		require.NoError(t, err)
		require.NotEmpty(t, actualConfigMap.Data)
		assert.NotEmpty(t, actualConfigMap.Data[ActiveGateMetadataKey])
	})
	t.Run(`create configmap with 2 keys, if both oneagent and activegate is needed`, func(t *testing.T) {
		dynakube := createTestDynakube(
			&dynatracev1beta2.DynaKubeSpec{
				OneAgent: dynatracev1beta2.OneAgentSpec{
					CloudNativeFullStack: &dynatracev1beta2.CloudNativeFullStackSpec{},
				},
				ActiveGate: dynatracev1beta2.ActiveGateSpec{
					Capabilities: []dynatracev1beta2.CapabilityDisplayName{
						dynatracev1beta2.KubeMonCapability.DisplayName,
					},
				},
			})

		fakeClient := fake.NewClientBuilder().Build()
		r := NewReconciler(fakeClient, fakeClient, *dynakube, clusterID)
		err := r.Reconcile(context.Background())
		require.NoError(t, err)

		var actualConfigMap corev1.ConfigMap
		err = fakeClient.Get(context.Background(), client.ObjectKey{Name: GetDeploymentMetadataConfigMapName(testName), Namespace: testNamespace}, &actualConfigMap)
		require.NoError(t, err)
		require.NotEmpty(t, actualConfigMap.Data)
		assert.NotEmpty(t, actualConfigMap.Data[OneAgentMetadataKey])
		assert.NotEmpty(t, actualConfigMap.Data[ActiveGateMetadataKey])
	})
}
