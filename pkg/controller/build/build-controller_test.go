package build

import (
	"context"
	"fmt"
	"os"
	"strings"

	ign3types "github.com/coreos/ignition/v2/config/v3_4/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	mcfgv1 "github.com/openshift/api/machineconfiguration/v1"
	mcfgv1alpha1 "github.com/openshift/api/machineconfiguration/v1alpha1"
	fakeclientbuildv1 "github.com/openshift/client-go/build/clientset/versioned/fake"
	fakeclientmachineconfigv1 "github.com/openshift/client-go/machineconfiguration/clientset/versioned/fake"
	testfixtures "github.com/openshift/machine-config-operator/test/fixtures"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	fakecorev1client "k8s.io/client-go/kubernetes/fake"
	"k8s.io/klog/v2"

	corev1 "k8s.io/api/core/v1"

	"github.com/openshift/machine-config-operator/pkg/apihelpers"
	ctrlcommon "github.com/openshift/machine-config-operator/pkg/controller/common"
	ctrlcommonconsts "github.com/openshift/machine-config-operator/pkg/controller/common/constants"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"

	"testing"
)

const (
	expectedImageSHA             string = "sha256:628e4e8f0a78d91015c6cebeee95931ae2e8defe5dfb4ced4a82830e08937573"
	expectedImagePullspecWithTag string = "registry.hostname.com/org/repo:latest"
	expectedImagePullspecWithSHA string = "registry.hostname.com/org/repo@" + expectedImageSHA
)

type optInFunc func(context.Context, *testing.T, *Clients, string)

func TestMain(m *testing.M) {
	klog.InitFlags(nil)
	os.Exit(m.Run())
}

func TestBuildControllerSinglePool(t *testing.T) {
	t.Parallel()

	pool := "worker"

	t.Run("Happy Path", func(t *testing.T) {
		t.Parallel()

		newBuildControllerTestFixture(t).runTestFuncs(t, testFuncs{
			customPodBuilder: func(ctx context.Context, t *testing.T, cs *Clients) {
				testMCPCustomBuildPod(ctx, t, cs, pool)
			},
		})
	})

	t.Run("Happy Path Multiple Configs", func(t *testing.T) {
		t.Parallel()

		newBuildControllerTestFixture(t).runTestFuncs(t, testFuncs{
			customPodBuilder: func(ctx context.Context, t *testing.T, cs *Clients) {
				testMultipleConfigsAreRolledOut(ctx, t, cs, pool, testMCPCustomBuildPod)
			},
		})
	})

	t.Run("Build Failure", func(t *testing.T) {
		t.Parallel()

		newBuildControllerTestFixture(t).runTestFuncs(t, testFuncs{
			customPodBuilder: func(ctx context.Context, t *testing.T, cs *Clients) {
				mcp := newMachineConfigPool(pool)
				mosc := newMachineOSConfig(mcp)
				cs.mcfgclient.MachineconfigurationV1alpha1().MachineOSConfigs().Create(ctx, mosc, metav1.CreateOptions{})
				assertMOSBFollowsBuildPodStatus(ctx, t, cs, mcp, mosc, corev1.PodFailed)
				assertMachineConfigPoolReachesStateWithMsg(ctx, t, cs, pool, isMOSBBuildFailure, isMOSBBuildFailureMsg)
			},
		})
	})

	t.Run("Degraded Pool", func(t *testing.T) {
		t.Parallel()

		newBuildControllerTestFixture(t).runTestFuncs(t, testFuncs{
			customPodBuilder: testMCPIsDegraded,
		})
	})
}

func TestBuildControllerMultipleOptedInPools(t *testing.T) {
	t.Parallel()

	pools := []string{"master", "worker"}

	// Tests that a single config is rolled out to the target MachineConfigPools.
	t.Run("Happy Path", func(t *testing.T) {
		t.Parallel()

		fixture := newBuildControllerTestFixture(t)

		for _, pool := range pools {
			pool := pool
			t.Run(pool, func(t *testing.T) {
				fixture.runTestFuncs(t, testFuncs{
					customPodBuilder: func(ctx context.Context, t *testing.T, cs *Clients) {
						t.Logf("Running in pool %s", pool)
						testMCPCustomBuildPod(ctx, t, cs, pool)
					},
				})
			})
		}
	})

	// Tests that multiple configs are serially rolled out to the target
	// MachineConfigPool and ensures that each config is rolled out before moving
	// onto the next one.
	t.Run("Happy Path Multiple Configs", func(t *testing.T) {
		t.Parallel()

		fixture := newBuildControllerTestFixture(t)

		for _, pool := range pools {
			pool := pool
			t.Run(pool, func(t *testing.T) {
				fixture.runTestFuncs(t, testFuncs{
					customPodBuilder: func(ctx context.Context, t *testing.T, cs *Clients) {
						testMultipleConfigsAreRolledOut(ctx, t, cs, pool, testMCPCustomBuildPod)
					},
				})
			})
		}
	})

	// Tests that a build failure degrades the MachineConfigPool
	t.Run("Build Failure", func(t *testing.T) {
		t.Parallel()

		fixture := newBuildControllerTestFixture(t)

		for _, pool := range pools {
			pool := pool
			t.Run(pool, func(t *testing.T) {
				fixture.runTestFuncs(t, testFuncs{
					customPodBuilder: func(ctx context.Context, t *testing.T, cs *Clients) {
						mcp := newMachineConfigPool(pool)
						mosc := newMachineOSConfig(mcp)
						cs.mcfgclient.MachineconfigurationV1alpha1().MachineOSConfigs().Create(ctx, mosc, metav1.CreateOptions{})
						assertMOSBFollowsBuildPodStatus(ctx, t, cs, mcp, mosc, corev1.PodFailed)
						assertMachineConfigPoolReachesStateWithMsg(ctx, t, cs, pool, isMOSBBuildFailure, isMOSBBuildFailureMsg)
					},
				})
			})
		}
	})
}

// Tests that if multiple goroutines attempt to create or update a
// canonicalized secret that all requests are successful.
func TestCanonicalizedSecrets(t *testing.T) {
	t.Parallel()

	clients := getClientsForTest()
	bc := NewWithCustomPodBuilder(BuildControllerConfig{MaxRetries: 1, UpdateDelay: testUpdateDelay}, clients)

	legacyPullSecretName := "final-image-push-secret"

	assertIsCanonicalSecret := func(t *testing.T, s *corev1.Secret) {
		// We expect the returned secret to contain the canonical suffix, to have
		// the canonical labels, and to generally be recognized as a canonical
		// secret.
		assert.Contains(t, s.Name, canonicalSecretSuffix)
		assert.True(t, isCanonicalizedSecret(s))
		assert.True(t, hasCanonicalizedSecretLabels(s))
		assert.Equal(t, s.Labels[OriginalSecretNameLabelKey], legacyPullSecretName)
		assert.True(t, IsObjectCreatedByBuildController(s))
	}

	testFunc := func(t *testing.T) {
		t.Parallel()

		out, err := bc.validatePullSecret(legacyPullSecretName)
		assert.NoError(t, err)

		assertIsCanonicalSecret(t, out)

		// Next, we change the original secret to validate that the canonicalized secret also gets the update.
		secret, err := clients.kubeclient.CoreV1().Secrets(ctrlcommonconsts.MCONamespace).Get(context.TODO(), legacyPullSecretName, metav1.GetOptions{})
		require.NoError(t, err)

		secret.Data[corev1.DockerConfigKey] = []byte(`{"updated.hostname.com": {"username": "user", "password": "s3kr1t", "auth": "s00pers3kr1t", "email": "user@hostname.com"}}`)

		_, err = clients.kubeclient.CoreV1().Secrets(ctrlcommonconsts.MCONamespace).Update(context.TODO(), secret, metav1.UpdateOptions{})
		require.NoError(t, err)

		// We look up the secret by its canonicalized name because once
		// BuildController canonicalizes a secret, it updates the MachineOSConfig
		// with the canonicalized name. That means that when it is retrieved again
		// for the next build, it is retrieved by its canonicalized name.
		out, err = bc.validatePullSecret(legacyPullSecretName + canonicalSecretSuffix)
		assert.NoError(t, err)

		assertIsCanonicalSecret(t, out)
		assert.Contains(t, string(out.Data[corev1.DockerConfigJsonKey]), "updated.hostname.com")

		// We also query the API server for the canonical version of this secret to
		// ensure that it is actually updated.
		secret, err = clients.kubeclient.CoreV1().Secrets(ctrlcommonconsts.MCONamespace).Get(context.TODO(), legacyPullSecretName+canonicalSecretSuffix, metav1.GetOptions{})
		require.NoError(t, err)

		assertIsCanonicalSecret(t, secret)
		assert.Contains(t, string(secret.Data[corev1.DockerConfigJsonKey]), "updated.hostname.com")
	}

	// We run the test function 10 times in parallel against the same
	// BuildController instance. Doing this simulates BuildController having to
	// handle multiple secret canonicalization attempts, similar to if someone
	// were to oc apply a YAML document containing at least two MachineOSConfigs
	// which reference the same legacy pull secret.
	for i := 0; i <= 10; i++ {
		t.Run("", testFunc)
	}
}

// Holds a name and function to implement a given BuildController test.
type buildControllerTestFixture struct {
	ctx                    context.Context
	t                      *testing.T
	customPodBuilderClient *Clients
}

type testFuncs struct {
	customPodBuilder func(context.Context, *testing.T, *Clients)
}

func newBuildControllerTestFixtureWithContext(ctx context.Context, t *testing.T) *buildControllerTestFixture {
	b := &buildControllerTestFixture{
		ctx: ctx,
		t:   t,
	}

	b.customPodBuilderClient = b.startBuildControllerWithCustomPodBuilder()

	return b
}

func newBuildControllerTestFixture(t *testing.T) *buildControllerTestFixture {
	ctx, cancel := context.WithTimeout(context.Background(), maxWait)
	t.Cleanup(cancel)

	return newBuildControllerTestFixtureWithContext(ctx, t)
}

func (b *buildControllerTestFixture) runTestFuncs(t *testing.T, tf testFuncs) {
	t.Run("CustomBuildPod", func(t *testing.T) {
		t.Parallel()
		tf.customPodBuilder(b.ctx, t, b.customPodBuilderClient)
	})
}

func (b *buildControllerTestFixture) setupClients() *Clients {
	return getClientsForTest()
}

func (b *buildControllerTestFixture) getConfig() BuildControllerConfig {
	return BuildControllerConfig{
		MaxRetries:  1,
		UpdateDelay: testUpdateDelay,
	}
}

func (b *buildControllerTestFixture) startBuildControllerWithCustomPodBuilder() *Clients {
	clients := b.setupClients()

	ctrl := NewWithCustomPodBuilder(b.getConfig(), clients)

	go ctrl.Run(b.ctx, 5)

	return clients
}

func getClientsForTest() *Clients {
	objects := newMachineConfigPoolAndConfigs("master", "rendered-master-1")
	objects = append(objects, newMachineConfigPoolAndConfigs("worker", "rendered-worker-1")...)
	objects = append(objects, &mcfgv1.ControllerConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: "machine-config-controller",
		},
	})

	imagesConfigMap := getImagesConfigMap()

	osImageURLConfigMap := getOSImageURLConfigMap()

	legacyPullSecret := `{"registry.hostname.com": {"username": "user", "password": "s3kr1t", "auth": "s00pers3kr1t", "email": "user@hostname.com"}}`

	pullSecret := `{"auths":{"registry.hostname.com": {"username": "user", "password": "s3kr1t", "auth": "s00pers3kr1t", "email": "user@hostname.com"}}}`

	return &Clients{
		mcfgclient: fakeclientmachineconfigv1.NewSimpleClientset(objects...),
		kubeclient: fakecorev1client.NewSimpleClientset(
			imagesConfigMap,
			osImageURLConfigMap,
			&corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "final-image-push-secret",
					Namespace: ctrlcommonconsts.MCONamespace,
				},
				Data: map[string][]byte{
					corev1.DockerConfigKey: []byte(legacyPullSecret),
				},
				Type: corev1.SecretTypeDockercfg,
			},
			&corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "base-image-pull-secret",
					Namespace: ctrlcommonconsts.MCONamespace,
				},
				Data: map[string][]byte{
					corev1.DockerConfigJsonKey: []byte(pullSecret),
				},
				Type: corev1.SecretTypeDockerConfigJson,
			},
			&corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "current-image-pull-secret",
					Namespace: ctrlcommonconsts.MCONamespace,
				},
				Data: map[string][]byte{
					corev1.DockerConfigJsonKey: []byte(pullSecret),
				},
				Type: corev1.SecretTypeDockerConfigJson,
			},
			&corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "etc-pki-entitlement",
					Namespace: "openshift-config-managed",
				},
				Data: map[string][]byte{
					"entitlement-key.pem": []byte("abc"),
					"entitlement.pem":     []byte("123"),
				},
			},
			&corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "machine-config-operator",
					Namespace: ctrlcommonconsts.MCONamespace,
				},
			},
		),
		buildclient: fakeclientbuildv1.NewSimpleClientset(),
	}

}

// Helper that determines if the build is a success.
func isMOSBBuildSuccess(mosc *mcfgv1alpha1.MachineOSConfig, mosb *mcfgv1alpha1.MachineOSBuild, mcp *mcfgv1.MachineConfigPool) bool {
	moscState := ctrlcommon.NewMachineOSConfigState(mosc)
	mosbState := ctrlcommon.NewMachineOSBuildState(mosb)

	return moscState.HasOSImage() &&
		moscState.GetOSImage() == expectedImagePullspecWithSHA &&
		mosbState.IsBuildSuccess() &&
		mcp.Spec.Configuration.Name == mcp.Status.Configuration.Name
}

func isMOSBBuildInProgress(mosb *mcfgv1alpha1.MachineOSBuild) bool {
	mosbState := ctrlcommon.NewMachineOSBuildState(mosb)
	return mosbState.IsBuilding()
}

func isMOSBBuildSuccessMsg(mosc *mcfgv1alpha1.MachineOSConfig, mosb *mcfgv1alpha1.MachineOSBuild, mcp *mcfgv1.MachineConfigPool) string {
	sb := &strings.Builder{}

	lps := ctrlcommon.NewLayeredPoolState(mcp)
	moscState := ctrlcommon.NewMachineOSConfigState(mosc)
	mosbState := ctrlcommon.NewMachineOSBuildState(mosb)

	fmt.Fprintf(sb, "Has OS image? %v\n", moscState.HasOSImage())
	fmt.Fprintf(sb, "Matches expected pullspec (%s)? %v\n", expectedImagePullspecWithSHA, moscState.GetOSImage() == expectedImagePullspecWithSHA)
	fmt.Fprintf(sb, "Is build success? %v\n", mosbState.IsBuildSuccess())
	fmt.Fprintf(sb, "Is degraded? %v\n", lps.IsDegraded())
	fmt.Fprintf(sb, "Spec.Configuration == Status.Configuration? %v\n", mcp.Spec.Configuration.Name == mcp.Status.Configuration.Name)
	return sb.String()
}

// Helper that determines if the build was a failure.
func isMOSBBuildFailure(mosc *mcfgv1alpha1.MachineOSConfig, mosb *mcfgv1alpha1.MachineOSBuild, mcp *mcfgv1.MachineConfigPool) bool {
	mosbState := ctrlcommon.NewMachineOSBuildState(mosb)

	return mosbState.IsBuildFailure() &&
		mcp.Spec.Configuration.Name == mcp.Status.Configuration.Name
}

func isMOSBBuildFailureMsg(mosc *mcfgv1alpha1.MachineOSConfig, mosb *mcfgv1alpha1.MachineOSBuild, mcp *mcfgv1.MachineConfigPool) string {
	sb := &strings.Builder{}

	lps := ctrlcommon.NewLayeredPoolState(mcp)
	mosbState := ctrlcommon.NewMachineOSBuildState(mosb)

	fmt.Fprintf(sb, "Is build failure? %v\n", mosbState.IsBuildFailure())
	fmt.Fprintf(sb, "Is degraded? %v\n", lps.IsDegraded())
	fmt.Fprintf(sb, "Spec.Configuration == Status.Configuration? %v\n", mcp.Spec.Configuration.Name == mcp.Status.Configuration.Name)
	return sb.String()
}

// Creates an MOSC and and MOSB and asserts that the MOSB reaches the desired state.
func testMCPCustomBuildPod(ctx context.Context, t *testing.T, cs *Clients, poolName string) {

	mcp, err := cs.mcfgclient.MachineconfigurationV1().MachineConfigPools().Get(ctx, poolName, metav1.GetOptions{})
	require.NoError(t, err)

	var mosc *mcfgv1alpha1.MachineOSConfig

	mosc, err = getMachineOSConfig(ctx, cs, mosc, mcp)
	require.NoError(t, err)
	if mosc == nil {
		mosc = newMachineOSConfig(mcp)
		_, err = cs.mcfgclient.MachineconfigurationV1alpha1().MachineOSConfigs().Create(ctx, mosc, metav1.CreateOptions{})
		require.NoError(t, err)
	}

	assertMOSBFollowsBuildPodStatus(ctx, t, cs, mcp, mosc, corev1.PodSucceeded)
	assertMachineConfigPoolReachesStateWithMsg(ctx, t, cs, poolName, isMOSBBuildSuccess, isMOSBBuildSuccessMsg)
}

func testRebuildDoesNothing(ctx context.Context, t *testing.T, cs *Clients, poolName string) {

	// Set an unrelated label to force a sync.
	mcpList, err := cs.mcfgclient.MachineconfigurationV1().MachineConfigPools().List(ctx, metav1.ListOptions{})
	require.NoError(t, err)

	for _, mcp := range mcpList.Items {
		mcp := mcp
		mcp.Labels[ctrlcommonconsts.RebuildPoolLabel] = ""
		_, err := cs.mcfgclient.MachineconfigurationV1().MachineConfigPools().Update(ctx, &mcp, metav1.UpdateOptions{})
		require.NoError(t, err)
	}

	mcpList, err = cs.mcfgclient.MachineconfigurationV1().MachineConfigPools().List(ctx, metav1.ListOptions{})
	require.NoError(t, err)

	for _, mcp := range mcpList.Items {
		mcp := mcp
		// ps := newPoolState(&mcp)
		ps := ctrlcommon.NewLayeredPoolState(&mcp)
		// assert.False(t, ps.IsLayered())
		assert.False(t, ps.HasOSImage())
	}

}

// Rolls out multiple configs to a given pool, asserting that each config is completely rolled out before moving onto the next.
func testMultipleConfigsAreRolledOut(ctx context.Context, t *testing.T, cs *Clients, poolName string, optInFunc optInFunc) {
	for i := 1; i < 10; i++ {
		config := fmt.Sprintf("rendered-%s-%d", poolName, i)

		mcp, err := cs.mcfgclient.MachineconfigurationV1().MachineConfigPools().Get(ctx, poolName, metav1.GetOptions{})
		require.NoError(t, err)

		mcp.Spec.Configuration.Name = config
		mcp.Status.Configuration.Name = config

		renderedMC := testfixtures.NewMachineConfig(
			config,
			map[string]string{
				ctrlcommonconsts.GeneratedByControllerVersionAnnotationKey: "version-number",
				"machineconfiguration.openshift.io/role":                   poolName,
			},
			"",
			[]ign3types.File{})

		_, err = cs.mcfgclient.MachineconfigurationV1().MachineConfigs().Create(ctx, renderedMC, metav1.CreateOptions{})
		if err != nil && !k8serrors.IsAlreadyExists(err) {
			require.NoError(t, err)
		}

		_, err = cs.mcfgclient.MachineconfigurationV1().MachineConfigPools().Update(ctx, mcp, metav1.UpdateOptions{})
		require.NoError(t, err)

		_, err = cs.mcfgclient.MachineconfigurationV1().MachineConfigPools().UpdateStatus(ctx, mcp, metav1.UpdateOptions{})
		require.NoError(t, err)

		optInFunc(ctx, t, cs, poolName)

		checkFunc := func(mosc *mcfgv1alpha1.MachineOSConfig, mosb *mcfgv1alpha1.MachineOSBuild, pool *mcfgv1.MachineConfigPool) bool {
			return pool.Spec.Configuration.Name == config && isMOSBBuildSuccess(mosc, mosb, pool)
		}

		msgFunc := func(mosc *mcfgv1alpha1.MachineOSConfig, mosb *mcfgv1alpha1.MachineOSBuild, pool *mcfgv1.MachineConfigPool) string {
			sb := &strings.Builder{}
			fmt.Fprintln(sb, isMOSBBuildFailureMsg(mosc, mosb, pool))
			fmt.Fprintf(sb, "Configuration name equals config? %v. Expected: %s\n, Actual: %s\n", pool.Spec.Configuration.Name == config, config, pool.Spec.Configuration.Name)
			return sb.String()
		}

		assertMachineConfigPoolReachesStateWithMsg(ctx, t, cs, poolName, checkFunc, msgFunc)
	}
}

// Tests that if a MachineConfigPool is degraded, that a build (object / pod) is not created.
func testMCPIsDegraded(ctx context.Context, t *testing.T, cs *Clients) {
	mcp, err := cs.mcfgclient.MachineconfigurationV1().MachineConfigPools().Get(ctx, "worker", metav1.GetOptions{})
	require.NoError(t, err)

	mcp.Labels[ctrlcommonconsts.LayeringEnabledPoolLabel] = ""

	condition := apihelpers.NewMachineConfigPoolCondition(mcfgv1.MachineConfigPoolDegraded, corev1.ConditionTrue, "", "")
	apihelpers.SetMachineConfigPoolCondition(&mcp.Status, *condition)

	_, err = cs.mcfgclient.MachineconfigurationV1().MachineConfigPools().Update(ctx, mcp, metav1.UpdateOptions{})
	require.NoError(t, err)

	assertMachineConfigPoolReachesState(ctx, t, cs, "worker", func(mcp *mcfgv1.MachineConfigPool) bool {
		// TODO: Should we fail the build without even starting it if the pool is degraded?
		// for _, condition := range getMachineConfigPoolBuildConditions() {
		// 	if apihelpers.IsMachineConfigPoolConditionTrue(mcp.Status.Conditions, condition) {
		// 		return false
		// 	}
		// }

		return apihelpers.IsMachineConfigPoolConditionTrue(mcp.Status.Conditions, mcfgv1.MachineConfigPoolDegraded) &&
			assertNoBuildPods(ctx, t, cs)
	})
}

// Tests that a label update or similar does not cause a build to occur.
func testBuiltPoolGetsUnrelatedUpdate(ctx context.Context, t *testing.T, cs *Clients, optInFunc optInFunc) {
	optInFunc(ctx, t, cs, "worker")

	pool, err := cs.mcfgclient.MachineconfigurationV1().MachineConfigPools().Get(ctx, "worker", metav1.GetOptions{})
	require.NoError(t, err)

	pool.Annotations["unrelated-annotation"] = "hello"
	pool.Labels["unrelated-label"] = ""
	_, err = cs.mcfgclient.MachineconfigurationV1().MachineConfigPools().Update(ctx, pool, metav1.UpdateOptions{})
	require.NoError(t, err)

	assertMachineConfigPoolReachesState(ctx, t, cs, "worker", func(mcp *mcfgv1.MachineConfigPool) bool {
		return assert.Equal(t, mcp.Status.Conditions, pool.Status.Conditions) &&
			assertNoBuildPods(ctx, t, cs)
	})
}

// TO-DO: update the test for BuildDueToPoolChange

// // Mocks whether a given build is running.
// type mockIsBuildRunning bool

// func (m *mockIsBuildRunning) IsBuildRunning(*mcfgv1.MachineConfigPool) (bool, error) {
// 	return bool(*m), nil
// }

// // Tests if we should do a build for a variety of edge-cases and circumstances.
// func TestShouldWeDoABuild(t *testing.T) {
// 	// Mutators which mutate the given MachineConfigPool.
// 	toLayeredPool := func(mcp *mcfgv1.MachineConfigPool) *mcfgv1.MachineConfigPool {
// 		mcp.Labels[ctrlcommonconsts.LayeringEnabledPoolLabel] = ""
// 		return mcp
// 	}

// 	toLayeredPoolWithImagePullspec := func(mcp *mcfgv1.MachineConfigPool) *mcfgv1.MachineConfigPool {
// 		mcp = toLayeredPool(mcp)
// 		ps := newPoolState(mcp)
// 		ps.SetImagePullspec("image-pullspec")
// 		return ps.MachineConfigPool()
// 	}

// 	toLayeredPoolWithConditionsSet := func(mcp *mcfgv1.MachineConfigPool, conditions []mcfgv1.MachineConfigPoolCondition) *mcfgv1.MachineConfigPool {
// 		mcp = toLayeredPoolWithImagePullspec(mcp)
// 		ps := newPoolState(mcp)
// 		ps.SetBuildConditions(conditions)
// 		return ps.MachineConfigPool()
// 	}

// 	type shouldWeBuildTestCase struct {
// 		name         string
// 		oldPool      *mcfgv1.MachineConfigPool
// 		curPool      *mcfgv1.MachineConfigPool
// 		buildRunning bool
// 		expected     bool
// 	}

// 	testCases := []shouldWeBuildTestCase{
// 		{
// 			name:     "Non-layered pool",
// 			oldPool:  newMachineConfigPool("worker", "rendered-worker-1"),
// 			curPool:  newMachineConfigPool("worker", "rendered-worker-1"),
// 			expected: false,
// 		},
// 		{
// 			name:     "Layered pool config change with missing image pullspec",
// 			oldPool:  toLayeredPool(newMachineConfigPool("worker", "rendered-worker-1")),
// 			curPool:  toLayeredPool(newMachineConfigPool("worker", "rendered-worker-2")),
// 			expected: true,
// 		},
// 		{
// 			name:     "Layered pool with no config change and missing image pullspec",
// 			oldPool:  toLayeredPool(newMachineConfigPool("worker", "rendered-worker-1")),
// 			curPool:  toLayeredPool(newMachineConfigPool("worker", "rendered-worker-1")),
// 			expected: true,
// 		},
// 		{
// 			name:    "Layered pool with image pullspec",
// 			oldPool: toLayeredPoolWithImagePullspec(newMachineConfigPool("worker", "rendered-worker-1")),
// 			curPool: toLayeredPoolWithImagePullspec(newMachineConfigPool("worker", "rendered-worker-1")),
// 		},
// 		{
// 			name:         "Layered pool with build pod",
// 			oldPool:      toLayeredPoolWithImagePullspec(newMachineConfigPool("worker", "rendered-worker-1")),
// 			curPool:      toLayeredPoolWithImagePullspec(newMachineConfigPool("worker", "rendered-worker-1")),
// 			buildRunning: true,
// 			expected:     false,
// 		},
// 		{
// 			name: "Layered pool with prior successful build and config change",
// 			oldPool: toLayeredPoolWithConditionsSet(newMachineConfigPool("worker", "rendered-worker-1"), []mcfgv1.MachineConfigPoolCondition{
// 				{
// 					Type:   mcfgv1.MachineConfigPoolBuildSuccess,
// 					Status: corev1.ConditionTrue,
// 				},
// 			}),
// 			curPool:  toLayeredPoolWithImagePullspec(newMachineConfigPool("worker", "rendered-worker-2")),
// 			expected: true,
// 		},
// 	}

// 	// Generate additional test cases programmatically.
// 	buildStates := map[mcfgv1.MachineConfigPoolConditionType]string{
// 		mcfgv1.MachineConfigPoolBuildFailed:    "failed",
// 		mcfgv1.MachineConfigPoolBuildPending:   "pending",
// 		mcfgv1.MachineConfigPoolBuilding:       "in progress",
// 		mcfgv1.MachineConfigPoolDegraded:       "degraded",
// 		mcfgv1.MachineConfigPoolNodeDegraded:   "node degraded",
// 		mcfgv1.MachineConfigPoolRenderDegraded: "render degraded",
// 	}

// 	for conditionType, name := range buildStates {
// 		conditions := []mcfgv1.MachineConfigPoolCondition{
// 			{
// 				Type:   conditionType,
// 				Status: corev1.ConditionTrue,
// 			},
// 		}

// 		testCases = append(testCases, shouldWeBuildTestCase{
// 			name:     fmt.Sprintf("Layered pool with %s build", name),
// 			oldPool:  toLayeredPoolWithConditionsSet(newMachineConfigPool("worker", "rendered-worker-1"), conditions),
// 			curPool:  toLayeredPoolWithConditionsSet(newMachineConfigPool("worker", "rendered-worker-1"), conditions),
// 			expected: false,
// 		})
// 	}

// 	for _, testCase := range testCases {
// 		testCase := testCase
// 		t.Run(testCase.name, func(t *testing.T) {
// 			t.Parallel()

// 			mb := mockIsBuildRunning(testCase.buildRunning)

// 			doABuild, err := shouldWeDoABuild(&mb, testCase.oldPool, testCase.curPool)
// 			assert.NoError(t, err)
// 			assert.Equal(t, testCase.expected, doABuild)
// 		})
// 	}
// }
