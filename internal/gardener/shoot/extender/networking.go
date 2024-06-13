package extender

import (
	gardener "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	imv1 "github.com/kyma-project/infrastructure-manager/api/v1"
)

func ExtendWithNetworking(runtime imv1.Runtime, shoot *gardener.Shoot) error {
	runtimeNetworking := runtime.Spec.Shoot.Networking

	shoot.Spec.Networking = &gardener.Networking{
		Type:     runtimeNetworking.Type,
		Nodes:    &runtimeNetworking.Nodes,
		Pods:     &runtimeNetworking.Pods,
		Services: &runtimeNetworking.Services,
	}

	return nil
}
