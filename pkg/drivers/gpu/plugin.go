package gpu

import (
	allocationapiv1alpha1 "github.com/ihcsim/k8s-dra/pkg/apis/allocation/v1alpha1"
	gpuv1alpha1 "github.com/ihcsim/k8s-dra/pkg/apis/gpu/v1alpha1"
)

type GPUPlugin struct{}

func (p *GPUPlugin) pendingAllocatedClaims(
	claimUID, selectedNode string,
	claimParams *gpuv1alpha1.GPUClaimParameters,
	classParams *gpuv1alpha1.GPUClassParameters) (allocationapiv1alpha1.AllocatedDevices, error) {
	return allocationapiv1alpha1.AllocatedDevices{}, nil
}

func (p *GPUPlugin) removeAllocatedClaim(claimUID string) error {
	return nil
}
