package daemonset

import (
	"fmt"
	"testing"

	"github.com/Dynatrace/dynatrace-operator/pkg/api/status"
	dynatracev1beta2 "github.com/Dynatrace/dynatrace-operator/pkg/api/v1beta2/dynakube"
	"github.com/Dynatrace/dynatrace-operator/pkg/controllers/dynakube/deploymentmetadata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testUID   = "test-uid"
	testKey   = "test-key"
	testValue = "test-value"

	testClusterID = "test-cluster-id"
	testURL       = "https://testing.dev.dynatracelabs.com/api"
	testName      = "test-name"

	testNewHostGroupName     = "newhostgroup"
	testOldHostGroupArgument = "--set-host-group=oldhostgroup"
	testNewHostGroupArgument = "--set-host-group=newhostgroup"
)

func TestArguments(t *testing.T) {
	t.Run("returns default arguments if hostInjection is nil", func(t *testing.T) {
		builder := builder{
			dk: &dynatracev1beta2.DynaKube{},
		}
		arguments, _ := builder.arguments()

		expectedDefaultArguments := []string{
			"--set-host-property=OperatorVersion=$(DT_OPERATOR_VERSION)",
			"--set-server={$(DT_SERVER)}",
			"--set-tenant=$(DT_TENANT)",
		}
		assert.Equal(t, expectedDefaultArguments, arguments)
	})
	t.Run("classic fullstack", func(t *testing.T) {
		instance := dynatracev1beta2.DynaKube{
			Spec: dynatracev1beta2.DynaKubeSpec{
				APIURL: testURL,
				OneAgent: dynatracev1beta2.OneAgentSpec{
					ClassicFullStack: &dynatracev1beta2.HostInjectSpec{
						Args: []string{testValue},
					},
				},
			},
		}
		dsBuilder := classicFullStack{
			builder{
				dk:             &instance,
				hostInjectSpec: instance.Spec.OneAgent.ClassicFullStack,
				clusterID:      testClusterID,
			},
		}
		podSpecs, _ := dsBuilder.podSpec()
		assert.NotNil(t, podSpecs)
		assert.NotEmpty(t, podSpecs.Containers)
		assert.Contains(t, podSpecs.Containers[0].Args, testValue)
	})
	t.Run("when injected arguments are provided then they are appended at the end of the arguments", func(t *testing.T) {
		args := []string{testValue}
		builder := builder{
			dk:             &dynatracev1beta2.DynaKube{},
			hostInjectSpec: &dynatracev1beta2.HostInjectSpec{Args: args},
		}

		arguments, _ := builder.arguments()

		expectedDefaultArguments := []string{
			"--set-host-property=OperatorVersion=$(DT_OPERATOR_VERSION)",
			"--set-server={$(DT_SERVER)}",
			"--set-tenant=$(DT_TENANT)",
			"test-value",
		}
		assert.Equal(t, expectedDefaultArguments, arguments)
	})
	t.Run("when injected arguments are provided then they come last", func(t *testing.T) {
		args := []string{
			"--set-app-log-content-access=true",
			"--set-host-id-source=lustiglustig",
			"--set-host-group=APP_LUSTIG_PETER",
			"--set-server=https://hyper.super.com:9999",
		}
		builder := builder{
			dk:             &dynatracev1beta2.DynaKube{},
			hostInjectSpec: &dynatracev1beta2.HostInjectSpec{Args: args},
		}

		arguments, _ := builder.arguments()

		expectedDefaultArguments := []string{
			"--set-app-log-content-access=true",
			"--set-host-group=APP_LUSTIG_PETER",
			"--set-host-id-source=lustiglustig",
			"--set-host-property=OperatorVersion=$(DT_OPERATOR_VERSION)",
			"--set-server={$(DT_SERVER)}",
			"--set-server=https://hyper.super.com:9999",
			"--set-tenant=$(DT_TENANT)",
		}
		assert.Equal(t, expectedDefaultArguments, arguments)
	})
	t.Run("--set-proxy is not set with OneAgent version >=1.271.0", func(t *testing.T) {
		builder := builder{
			dk: &dynatracev1beta2.DynaKube{
				Status: dynatracev1beta2.DynaKubeStatus{
					OneAgent: dynatracev1beta2.OneAgentStatus{
						VersionStatus: status.VersionStatus{
							Version: "1.285.0.20240122-141707",
						},
					},
				},
			},
		}
		arguments, _ := builder.arguments()

		expectedDefaultArguments := []string{
			"--set-host-property=OperatorVersion=$(DT_OPERATOR_VERSION)",
			"--set-server={$(DT_SERVER)}",
			"--set-tenant=$(DT_TENANT)",
		}
		assert.Equal(t, expectedDefaultArguments, arguments)
	})
	t.Run("multiple set-host-property entries are possible", func(t *testing.T) {
		args := []string{
			"--set-app-log-content-access=true",
			"--set-host-id-source=lustiglustig",
			"--set-host-group=APP_LUSTIG_PETER",
			"--set-server=https://hyper.super.com:9999",
			"--set-host-property=item0=value0",
			"--set-host-property=item1=value1",
			"--set-host-property=item2=value2",
		}
		builder := builder{
			dk:             &dynatracev1beta2.DynaKube{},
			hostInjectSpec: &dynatracev1beta2.HostInjectSpec{Args: args},
		}

		arguments, _ := builder.arguments()

		expectedDefaultArguments := []string{
			"--set-app-log-content-access=true",
			"--set-host-group=APP_LUSTIG_PETER",
			"--set-host-id-source=lustiglustig",
			"--set-host-property=OperatorVersion=$(DT_OPERATOR_VERSION)",
			"--set-host-property=item0=value0",
			"--set-host-property=item1=value1",
			"--set-host-property=item2=value2",
			"--set-server={$(DT_SERVER)}",
			"--set-server=https://hyper.super.com:9999",
			"--set-tenant=$(DT_TENANT)",
		}
		assert.Equal(t, expectedDefaultArguments, arguments)
	})
}

func TestPodSpec_Arguments(t *testing.T) {
	instance := &dynatracev1beta2.DynaKube{
		Spec: dynatracev1beta2.DynaKubeSpec{
			OneAgent: dynatracev1beta2.OneAgentSpec{
				ClassicFullStack: &dynatracev1beta2.HostInjectSpec{
					Args: []string{testKey, testValue, testUID},
				},
			},
		},
	}
	hostInjectSpecs := instance.Spec.OneAgent.ClassicFullStack
	dsBuilder := classicFullStack{
		builder{
			dk:             instance,
			hostInjectSpec: hostInjectSpecs,
			clusterID:      testClusterID,
			deploymentType: deploymentmetadata.ClassicFullStackDeploymentType,
		},
	}

	instance.Annotations = map[string]string{}
	podSpecs, _ := dsBuilder.podSpec()
	require.NotNil(t, podSpecs)
	require.NotEmpty(t, podSpecs.Containers)

	for _, arg := range hostInjectSpecs.Args {
		assert.Contains(t, podSpecs.Containers[0].Args, arg)
	}

	assert.Contains(t, podSpecs.Containers[0].Args, fmt.Sprintf("--set-host-property=OperatorVersion=$(%s)", deploymentmetadata.EnvDtOperatorVersion))

	// deprecated
	t.Run(`has proxy arg`, func(t *testing.T) {
		instance.Status.OneAgent.Version = "1.272.0.0-0"
		instance.Spec.Proxy = &dynatracev1beta2.DynaKubeProxy{Value: testValue}
		podSpecs, _ = dsBuilder.podSpec()
		assert.Contains(t, podSpecs.Containers[0].Args, "--set-proxy=$(https_proxy)")

		instance.Spec.Proxy = nil
		instance.Status.OneAgent.Version = ""
		podSpecs, _ = dsBuilder.podSpec()
		assert.NotContains(t, podSpecs.Containers[0].Args, "--set-proxy=$(https_proxy)")
	})
	// deprecated
	t.Run(`has proxy arg but feature flag to ignore is enabled`, func(t *testing.T) {
		instance.Spec.Proxy = &dynatracev1beta2.DynaKubeProxy{Value: testValue}
		instance.Annotations[dynatracev1beta2.AnnotationFeatureOneAgentIgnoreProxy] = "true"
		podSpecs, _ = dsBuilder.podSpec()
		assert.NotContains(t, podSpecs.Containers[0].Args, "--set-proxy=$(https_proxy)")
	})
	t.Run(`has network zone arg`, func(t *testing.T) {
		instance.Spec.NetworkZone = testValue
		podSpecs, _ = dsBuilder.podSpec()
		assert.Contains(t, podSpecs.Containers[0].Args, "--set-network-zone="+testValue)

		instance.Spec.NetworkZone = ""
		podSpecs, _ = dsBuilder.podSpec()
		assert.NotContains(t, podSpecs.Containers[0].Args, "--set-network-zone="+testValue)
	})
	t.Run(`has host-id-source arg for classic fullstack`, func(t *testing.T) {
		daemonset, _ := dsBuilder.BuildDaemonSet()
		podSpecs = daemonset.Spec.Template.Spec
		assert.Contains(t, podSpecs.Containers[0].Args, "--set-host-id-source=auto")
	})
	t.Run(`has host-id-source arg for hostMonitoring`, func(t *testing.T) {
		hostMonInstance := &dynatracev1beta2.DynaKube{
			Spec: dynatracev1beta2.DynaKubeSpec{
				OneAgent: dynatracev1beta2.OneAgentSpec{
					HostMonitoring: &dynatracev1beta2.HostInjectSpec{
						Args: []string{testKey, testValue, testUID},
					},
				},
			},
		}

		hostMonInjectSpec := hostMonInstance.Spec.OneAgent.HostMonitoring

		dsBuilder := hostMonitoring{
			builder{
				dk:             hostMonInstance,
				hostInjectSpec: hostMonInjectSpec,
				clusterID:      testClusterID,
			},
		}
		daemonset, _ := dsBuilder.BuildDaemonSet()
		podSpecs := daemonset.Spec.Template.Spec
		assert.Contains(t, podSpecs.Containers[0].Args, "--set-host-id-source=k8s-node-name")
	})
	t.Run(`has host-id-source arg for cloudNativeFullstack`, func(t *testing.T) {
		cloudNativeInstance := &dynatracev1beta2.DynaKube{
			Spec: dynatracev1beta2.DynaKubeSpec{
				OneAgent: dynatracev1beta2.OneAgentSpec{
					CloudNativeFullStack: &dynatracev1beta2.CloudNativeFullStackSpec{
						HostInjectSpec: dynatracev1beta2.HostInjectSpec{Args: []string{testKey, testValue, testUID}},
					},
				},
			},
		}

		dsBuilder := hostMonitoring{
			builder{
				dk:             cloudNativeInstance,
				hostInjectSpec: &cloudNativeInstance.Spec.OneAgent.CloudNativeFullStack.HostInjectSpec,
				clusterID:      testClusterID,
			},
		}
		daemonset, _ := dsBuilder.BuildDaemonSet()
		podSpecs := daemonset.Spec.Template.Spec
		assert.Contains(t, podSpecs.Containers[0].Args, "--set-host-id-source=k8s-node-name")
	})
	t.Run(`has host-group for classicFullstack`, func(t *testing.T) {
		classicInstance := &dynatracev1beta2.DynaKube{
			Spec: dynatracev1beta2.DynaKubeSpec{
				OneAgent: dynatracev1beta2.OneAgentSpec{
					HostGroup: testNewHostGroupName,
					ClassicFullStack: &dynatracev1beta2.HostInjectSpec{
						Args: []string{testOldHostGroupArgument},
					},
				},
			},
		}

		dsBuilder := hostMonitoring{
			builder{
				dk:             classicInstance,
				hostInjectSpec: classicInstance.Spec.OneAgent.ClassicFullStack,
			},
		}
		arguments, err := dsBuilder.arguments()
		require.NoError(t, err)
		assert.Contains(t, arguments, testNewHostGroupArgument)
	})
	t.Run(`has host-group for cloudNativeFullstack`, func(t *testing.T) {
		cloudNativeInstance := &dynatracev1beta2.DynaKube{
			Spec: dynatracev1beta2.DynaKubeSpec{
				OneAgent: dynatracev1beta2.OneAgentSpec{
					HostGroup: testNewHostGroupName,
					CloudNativeFullStack: &dynatracev1beta2.CloudNativeFullStackSpec{
						HostInjectSpec: dynatracev1beta2.HostInjectSpec{Args: []string{testOldHostGroupArgument}},
					},
				},
			},
		}

		dsBuilder := hostMonitoring{
			builder{
				dk:             cloudNativeInstance,
				hostInjectSpec: &cloudNativeInstance.Spec.OneAgent.CloudNativeFullStack.HostInjectSpec,
			},
		}
		arguments, err := dsBuilder.arguments()
		require.NoError(t, err)
		assert.Contains(t, arguments, testNewHostGroupArgument)
	})
	t.Run(`has host-group for HostMonitoring`, func(t *testing.T) {
		hostMonitoringInstance := &dynatracev1beta2.DynaKube{
			Spec: dynatracev1beta2.DynaKubeSpec{
				OneAgent: dynatracev1beta2.OneAgentSpec{
					HostGroup: testNewHostGroupName,
					HostMonitoring: &dynatracev1beta2.HostInjectSpec{
						Args: []string{testOldHostGroupArgument},
					},
				},
			},
		}

		dsBuilder := hostMonitoring{
			builder{
				dk:             hostMonitoringInstance,
				hostInjectSpec: hostMonitoringInstance.Spec.OneAgent.HostMonitoring,
			},
		}
		arguments, err := dsBuilder.arguments()
		require.NoError(t, err)
		assert.Contains(t, arguments, testNewHostGroupArgument)
	})
}
