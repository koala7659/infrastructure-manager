package controller

import (
	"fmt"

	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/onsi/gomega/types"
)

type finalizerMatcher struct {
	expectedFinalizer string
	actualFinalizers  []string
}

func haveFinalizer(v string) types.GomegaMatcher {
	return &finalizerMatcher{
		expectedFinalizer: v,
	}
}

func (m *finalizerMatcher) Match(actual any) (success bool, err error) {
	actualRt, ok := actual.(client.Object)
	if !ok {
		return false, fmt.Errorf("haveFinalizer matcher expects an *sigs.k8s.io/controller-runtime/pkg/client.Object")
	}

	m.actualFinalizers = actualRt.GetFinalizers()
	return ContainElement(m.expectedFinalizer).Match(actualRt.GetFinalizers())
}

func (m *finalizerMatcher) FailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected\n\t%v\nto contain\n\t%s", m.actualFinalizers, m.expectedFinalizer)
}

func (m *finalizerMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected\n\t%v\nnot to contain\n\t%s", m.actualFinalizers, m.expectedFinalizer)
}