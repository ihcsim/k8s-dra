package gpu

import (
	"errors"
	"fmt"
	"sync"

	allocationv1alpha1 "github.com/ihcsim/k8s-dra/pkg/apis/allocation/v1alpha1"
	gpuv1alpha1 "github.com/ihcsim/k8s-dra/pkg/apis/gpu/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	dractrl "k8s.io/dynamic-resource-allocation/controller"
)

type gpuPlugin struct {
	// pendingAllocatedClaims is a map of resource claim UID to a map of node name to allocated
	// devices.
	pendingAllocatedClaims map[string]map[string]allocationv1alpha1.AllocatedDevices

	// mux is used to synchronized concurrent read-write accesses to allocations.
	mux sync.RWMutex
}

func newGPUPlugin() *gpuPlugin {
	return &gpuPlugin{
		pendingAllocatedClaims: map[string]map[string]allocationv1alpha1.AllocatedDevices{},
		mux:                    sync.RWMutex{},
	}
}

func (p *gpuPlugin) allocate(
	claimUID string,
	selectedNode string,
	claimParams *gpuv1alpha1.GPUClaimParametersSpec,
	classParams *gpuv1alpha1.GPUClassParametersSpec) (allocationv1alpha1.AllocatedDevices, error) {
	p.mux.RLock()
	defer p.mux.RUnlock()

	if _, exists := p.pendingAllocatedClaims[claimUID]; !exists {
		return allocationv1alpha1.AllocatedDevices{}, fmt.Errorf("no allocations generated for claim %s on node %s", claimUID, selectedNode)
	}

	allocatedDevices, exists := p.pendingAllocatedClaims[claimUID][selectedNode]
	if !exists {
		return allocationv1alpha1.AllocatedDevices{}, fmt.Errorf("no allocations generated for claim %s on node %s", claimUID, selectedNode)
	}

	return allocatedDevices, nil
}

func (p *gpuPlugin) deallocate(claimUID string) error {
	p.mux.Lock()
	defer p.mux.Unlock()

	delete(p.pendingAllocatedClaims, claimUID)
	return nil
}

func (p *gpuPlugin) unsuitableNode(
	nodeDeviceAllocation *allocationv1alpha1.NodeDeviceAllocation,
	pod *corev1.Pod,
	gpuClaims []*dractrl.ClaimAllocation,
	allClaims []*dractrl.ClaimAllocation,
	potentialNode string) error {

	var errs error
	for claimUID := range p.pendingAllocatedClaims {
		if allocation, exists := p.pendingAllocatedClaims[claimUID][potentialNode]; exists {
			if _, exists := nodeDeviceAllocation.Status.AllocatedClaims[claimUID]; exists {
				if err := p.deallocate(claimUID); err != nil {
					errs = errors.Join(errs, err)
					continue
				}
			} else {
				nodeDeviceAllocation.Status.AllocatedClaims[claimUID] = allocation
			}
		}
	}

	allocated := p.findAllocated(nodeDeviceAllocation, gpuClaims)

	for _, gpuClaim := range gpuClaims {
		claimUID := string(gpuClaim.Claim.GetUID())
		claimParams, ok := gpuClaim.ClaimParameters.(*gpuv1alpha1.GPUClaimParametersSpec)
		if !ok {
			errs = errors.Join(errs, fmt.Errorf("failed to cast claim parameters to GPUClaimParametersSpec"))
			continue
		}

		if claimParams.Count != len(allocated[claimUID]) {
			for _, claim := range allClaims {
				claim.UnsuitableNodes = append(claim.UnsuitableNodes, potentialNode)
			}
			return nil
		}

		var devices []allocationv1alpha1.AllocatedGPU
		for _, gpu := range allocated[claimUID] {
			devices = append(devices, allocationv1alpha1.AllocatedGPU{
				UUID: gpu,
			})
		}

		allocatedDevices := allocationv1alpha1.AllocatedDevices{
			GPUs: &allocationv1alpha1.AllocatedGPUs{
				Devices: devices,
			},
		}

		if _, exists := p.pendingAllocatedClaims[claimUID]; !exists {
			p.pendingAllocatedClaims[claimUID] = map[string]allocationv1alpha1.AllocatedDevices{}
		}
		p.pendingAllocatedClaims[claimUID][potentialNode] = allocatedDevices
	}

	return errs
}

func (p *gpuPlugin) findAllocated(nodeDeviceAllocation *allocationv1alpha1.NodeDeviceAllocation, gpuClaims []*dractrl.ClaimAllocation) map[string][]string {
	available := map[string]*allocationv1alpha1.AllocatableGPU{}
	for _, device := range nodeDeviceAllocation.Status.AllocatableDevices {
		available[device.GPU.UUID] = device.GPU
	}

	for _, allocation := range nodeDeviceAllocation.Status.AllocatedClaims {
		for _, gpu := range allocation.GPUs.Devices {
			delete(available, gpu.UUID)
		}
	}

	allocated := map[string][]string{}
	for _, gpuClaim := range gpuClaims {
		claimUID := string(gpuClaim.Claim.GetUID())
		if _, exists := nodeDeviceAllocation.Status.AllocatedClaims[claimUID]; exists {
			devices := nodeDeviceAllocation.Status.AllocatedClaims[claimUID].GPUs.Devices
			for _, device := range devices {
				allocated[claimUID] = append(allocated[claimUID], device.UUID)
			}
		}

		claimParams := gpuClaim.ClaimParameters.(*gpuv1alpha1.GPUClaimParametersSpec)
		devices := []string{}
		for i := 0; i < claimParams.Count; i++ {
			for _, device := range available {
				devices = append(devices, device.UUID)
				delete(available, device.UUID)
				break
			}
		}
		allocated[claimUID] = devices
	}

	return allocated
}
