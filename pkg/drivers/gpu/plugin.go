package gpu

import (
	allocationapiv1alpha1 "github.com/ihcsim/k8s-dra/pkg/apis/allocation/v1alpha1"
	gpuv1alpha1 "github.com/ihcsim/k8s-dra/pkg/apis/gpu/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	dractrl "k8s.io/dynamic-resource-allocation/controller"
)

type gpuPlugin struct{}

func (p *gpuPlugin) pendingAllocatedClaims(
	claimUID, selectedNode string,
	claimParams *gpuv1alpha1.GPUClaimParameters,
	classParams *gpuv1alpha1.GPUClassParameters) (allocationapiv1alpha1.AllocatedDevices, error) {
	return allocationapiv1alpha1.AllocatedDevices{}, nil
}

func (p *gpuPlugin) removeAllocatedClaim(claimUID string) error {
	return nil
}

func (p *gpuPlugin) unsuitableNode(
	nodedeviceAllocation *allocationapiv1alpha1.NodeDeviceAllocation,
	pod *corev1.Pod,
	gpuClaims []*dractrl.ClaimAllocation,
	allClaims []*dractrl.ClaimAllocation,
	potentialNode string) error {
	return nil
}
