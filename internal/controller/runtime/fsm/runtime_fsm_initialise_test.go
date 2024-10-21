package fsm

import (
	"context"
	"time"

	gardener "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	imv1 "github.com/kyma-project/infrastructure-manager/api/v1"
	"github.com/kyma-project/infrastructure-manager/internal/controller/metrics/mocks"
	. "github.com/onsi/ginkgo/v2" //nolint:revive
	. "github.com/onsi/gomega"    //nolint:revive
	"github.com/onsi/gomega/types"
	"github.com/stretchr/testify/mock"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	util "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var withTestFinalizer = withFinalizer("test-me-plz")

var _ = Describe("KIM sFnInitialise", func() {
	now := metav1.NewTime(time.Now())

	testCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// GIVEN

	testScheme := runtime.NewScheme()
	util.Must(imv1.AddToScheme(testScheme))

	withTestSchemeAndObjects := func(objs ...client.Object) fakeFSMOpt {
		return func(fsm *fsm) error {
			return withFakedK8sClient(testScheme, objs...)(fsm)
		}
	}

	withMockedMetrics := func() fakeFSMOpt {
		m := &mocks.Metrics{}
		m.On("SetRuntimeStates", mock.Anything).Return()
		m.On("CleanUpRuntimeGauge", mock.Anything).Return()
		m.On("IncRuntimeFSMStopCounter").Return()
		return withMetrics(m)
	}

	testRt := imv1.Runtime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-instance",
			Namespace: "default",
		},
	}

	testRtWithFinalizerNoProvisioningCondition := imv1.Runtime{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-instance",
			Namespace:  "default",
			Finalizers: []string{"test-me-plz"},
			Labels: map[string]string{
				imv1.LabelControlledByProvisioner: "false",
			},
		},
	}

	testRtWithFinalizerAndProvisioningCondition := imv1.Runtime{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-instance",
			Namespace:  "default",
			Finalizers: []string{"test-me-plz"},
			Labels: map[string]string{
				imv1.LabelControlledByProvisioner: "false",
			},
		},
	}

	testDryRunRtWithFinalizerAndProvisioningCondition := imv1.Runtime{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-instance",
			Namespace:  "default",
			Finalizers: []string{"test-me-plz"},
			Labels: map[string]string{
				imv1.LabelControlledByProvisioner: "true",
			},
		},
	}

	testDryRunRtWithFinalizerAndProvisioningReadyCondition := imv1.Runtime{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-instance",
			Namespace:  "default",
			Finalizers: []string{"test-me-plz"},
			Labels: map[string]string{
				imv1.LabelControlledByProvisioner: "true",
			},
		},
	}

	provisioningCondition := metav1.Condition{
		Type:               string(imv1.ConditionTypeRuntimeProvisioned),
		Status:             metav1.ConditionUnknown,
		LastTransitionTime: now,
		Reason:             "Test reason",
		Message:            "Test message",
	}
	meta.SetStatusCondition(&testRtWithFinalizerAndProvisioningCondition.Status.Conditions, provisioningCondition)

	provisioningDryRunCondition := metav1.Condition{
		Type:               string(imv1.ConditionTypeRuntimeProvisionedDryRun),
		Status:             metav1.ConditionUnknown,
		LastTransitionTime: now,
		Reason:             "Test reason",
		Message:            "Test message",
	}
	meta.SetStatusCondition(&testDryRunRtWithFinalizerAndProvisioningCondition.Status.Conditions, provisioningDryRunCondition)

	provisioningDryRunConditionReady := metav1.Condition{
		Type:               string(imv1.ConditionTypeRuntimeProvisionedDryRun),
		Status:             metav1.ConditionTrue,
		LastTransitionTime: now,
		Reason:             "Test reason",
		Message:            "Test message",
	}
	meta.SetStatusCondition(&testDryRunRtWithFinalizerAndProvisioningReadyCondition.Status.Conditions, provisioningDryRunConditionReady)

	testRtWithDeletionTimestamp := imv1.Runtime{
		ObjectMeta: metav1.ObjectMeta{
			DeletionTimestamp: &now,
			Labels: map[string]string{
				imv1.LabelControlledByProvisioner: "false",
			},
		},
	}

	testRtWithDeletionTimestampAndFinalizer := imv1.Runtime{
		ObjectMeta: metav1.ObjectMeta{
			DeletionTimestamp: &now,
			Finalizers:        []string{"test-me-plz"},
			Labels: map[string]string{
				imv1.LabelControlledByProvisioner: "false",
			},
		},
	}

	testShoot := gardener.Shoot{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-instance",
			Namespace: "default",
		},
	}

	testFunction := buildTestFunction(sFnInitialize)

	// WHEN/THAN

	DescribeTable(
		"transition graph validation",
		testFunction,
		Entry(
			"should return nothing when CR is being deleted without finalizer and shoot is missing",
			testCtx,
			must(newFakeFSM, withTestFinalizer, withMockedMetrics(), withDefaultReconcileDuration()),
			&systemState{instance: testRtWithDeletionTimestamp},
			testOpts{
				MatchExpectedErr: BeNil(),
				MatchNextFnState: haveName("stopWithMetrics"),
			},
		),
		Entry(
			"should return sFnUpdateStatus when CR is being deleted with finalizer and shoot is missing - Remove finalizer",
			testCtx,
			must(newFakeFSM, withTestFinalizer, withTestSchemeAndObjects(&testRt), withMockedMetrics(), withDefaultReconcileDuration()),
			&systemState{instance: testRtWithDeletionTimestampAndFinalizer},
			testOpts{
				MatchExpectedErr: BeNil(),
				MatchNextFnState: haveName("sFnUpdateStatus"),
			},
		),
		Entry(
			"should return sFnDeleteShoot and no error when CR is being deleted with finalizer and shoot exists",
			testCtx,
			must(newFakeFSM, withTestFinalizer, withMockedMetrics(), withDefaultReconcileDuration()),
			&systemState{instance: testRtWithDeletionTimestampAndFinalizer, shoot: &testShoot},
			testOpts{
				MatchExpectedErr: BeNil(),
				MatchNextFnState: haveName("sFnDeleteKubeconfig"),
			},
		),
		Entry(
			"should return sFnUpdateStatus and no error when CR has been created without finalizer - Add finalizer",
			testCtx,
			must(newFakeFSM, withTestFinalizer, withTestSchemeAndObjects(&testRt), withMockedMetrics(), withDefaultReconcileDuration()),
			&systemState{instance: testRt},
			testOpts{
				MatchExpectedErr: BeNil(),
				MatchNextFnState: BeNil(),
				StateMatch:       []types.GomegaMatcher{haveFinalizer("test-me-plz")},
			},
		),
		Entry(
			"should return sFnUpdateStatus and no error when there is no Provisioning Condition - Add condition",
			testCtx,
			must(newFakeFSM, withTestFinalizer, withMockedMetrics(), withDefaultReconcileDuration()),
			&systemState{instance: testRtWithFinalizerNoProvisioningCondition},
			testOpts{
				MatchExpectedErr: BeNil(),
				MatchNextFnState: haveName("sFnUpdateStatus"),
			},
		),
		Entry(
			"should return sFnCreateShoot and no error when exists Provisioning Condition and shoot is missing",
			testCtx,
			must(newFakeFSM, withTestFinalizer, withMockedMetrics(), withDefaultReconcileDuration()),
			&systemState{instance: testRtWithFinalizerAndProvisioningCondition},
			testOpts{
				MatchExpectedErr: BeNil(),
				MatchNextFnState: haveName("sFnCreateShoot"),
			},
		),
		Entry(
			"should return sFnCreateShootDryRun and no error when exists Provisioning Condition and shoot is missing",
			testCtx,
			must(newFakeFSM, withTestFinalizer, withMockedMetrics(), withDefaultReconcileDuration()),
			&systemState{instance: testDryRunRtWithFinalizerAndProvisioningCondition},
			testOpts{
				MatchExpectedErr: BeNil(),
				MatchNextFnState: haveName("sFnCreateShootDryRun"),
			},
		),
		Entry(
			"should stop when sFnCreateShootDryRun was already executed",
			testCtx,
			must(newFakeFSM, withTestFinalizer, withMockedMetrics(), withDefaultReconcileDuration()),
			&systemState{instance: testDryRunRtWithFinalizerAndProvisioningReadyCondition},
			testOpts{
				MatchExpectedErr: BeNil(),
				MatchNextFnState: haveName("stopWithMetrics"),
			},
		),
		Entry(
			"should return sFnSelectShootProcessing and no error when exists Provisioning Condition and shoot exists",
			testCtx,
			must(newFakeFSM, withTestFinalizer, withMockedMetrics(), withDefaultReconcileDuration()),
			&systemState{instance: testRtWithFinalizerAndProvisioningCondition, shoot: &testShoot},
			testOpts{
				MatchExpectedErr: BeNil(),
				MatchNextFnState: haveName("sFnSelectShootProcessing"),
			},
		),
	)
})

type testOpts struct {
	MatchExpectedErr types.GomegaMatcher
	MatchNextFnState types.GomegaMatcher
	StateMatch       []types.GomegaMatcher
}

func buildTestFunction(fn stateFn) func(context.Context, *fsm, *systemState, testOpts) {
	return func(ctx context.Context, r *fsm, s *systemState, ops testOpts) {
		sFn, _, err := fn(ctx, r, s)

		Expect(err).To(ops.MatchExpectedErr)
		Expect(sFn).To(ops.MatchNextFnState)

		for _, match := range ops.StateMatch {
			Expect(&s.instance).Should(match)
		}
	}
}
