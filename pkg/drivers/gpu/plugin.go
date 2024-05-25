package gpu

import (
	"errors"
	"fmt"
	"sync"

	gpuv1alpha1 "github.com/ihcsim/k8s-dra/pkg/apis/gpu/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	dractrl "k8s.io/dynamic-resource-allocation/controller"
)

type gpuPlugin struct {
	// pendingAllocatedGPUs is a map of resource claim UID to a map of node name to allocated
	// GPUs.
	pendingAllocatedGPUs map[nodeClaim][]*gpuv1alpha1.GPUDevice

	// mux is used to synchronized concurrent read-write accesses to allocations.
	mux sync.RWMutex
}

type nodeClaim struct {
	claimUID string
	nodeName string
}

func (n *nodeClaim) String() string {
	return fmt.Sprintf("%s/%s", n.nodeName, n.claimUID)
}

func newGPUPlugin() *gpuPlugin {
	return &gpuPlugin{
		pendingAllocatedGPUs: map[nodeClaim][]*gpuv1alpha1.GPUDevice{},
		mux:                  sync.RWMutex{},
	}
}

func (p *gpuPlugin) allocate(
	claimUID, nodeName string,
	claimParams *gpuv1alpha1.GPUClaimParametersSpec,
	classParams *gpuv1alpha1.GPUClassParametersSpec) ([]*gpuv1alpha1.GPUDevice, error) {
	p.mux.RLock()
	defer p.mux.RUnlock()

	nodeClaim := nodeClaim{
		claimUID: claimUID,
		nodeName: nodeName,
	}
	if _, exists := p.pendingAllocatedGPUs[nodeClaim]; !exists {
		return nil, fmt.Errorf("no allocations generated for node claim %s", nodeClaim)
	}

	allocatedGPUs, exists := p.pendingAllocatedGPUs[nodeClaim]
	if !exists {
		return nil, fmt.Errorf("no allocations generated for node claim %s", nodeClaim)
	}

	return allocatedGPUs, nil
}

func (p *gpuPlugin) deallocate(claimUID, nodeName string) error {
	p.mux.Lock()
	defer p.mux.Unlock()

	nodeClaim := nodeClaim{
		claimUID: claimUID,
		nodeName: nodeName,
	}
	delete(p.pendingAllocatedGPUs, nodeClaim)
	return nil
}

func (p *gpuPlugin) unsuitableNode(
	nodeDevices *gpuv1alpha1.NodeDevices,
	pod *corev1.Pod,
	gpuClaims []*dractrl.ClaimAllocation,
	allClaims []*dractrl.ClaimAllocation,
	potentialNode string) error {
	p.mux.Lock()
	defer p.mux.Unlock()

	var errs error
	for nodeClaim := range p.pendingAllocatedGPUs {
		if allocation, exists := p.pendingAllocatedGPUs[nodeClaim]; exists {
			claimUID := nodeClaim.claimUID
			if _, exists := nodeDevices.Status.AllocatedGPUs[claimUID]; exists {
				if err := p.deallocate(claimUID, potentialNode); err != nil {
					errs = errors.Join(errs, err)
					continue
				}
			} else {
				nodeDevices.Status.AllocatedGPUs[claimUID] = allocation
			}
		}
	}
	if errs != nil {
		return errs
	}

	availableGPUs := p.availableGPUs(nodeDevices)
	allocatedGPUs := map[string][]string{}
	for _, gpuClaim := range gpuClaims {
		// if the nodeDevices has already allocated GPUs for the claim, add the
		// allocated GPUs to the result map
		claimUID := string(gpuClaim.Claim.GetUID())
		if gpus, exists := nodeDevices.Status.AllocatedGPUs[claimUID]; exists {
			for _, gpu := range gpus {
				allocatedGPUs[claimUID] = append(allocatedGPUs[claimUID], gpu.UUID)
			}
			continue
		}

		// otherwise, allocate up to claimParams.Count GPUs from the available pool
		claimParams, ok := gpuClaim.ClaimParameters.(*gpuv1alpha1.GPUClaimParametersSpec)
		if !ok {
			errs = errors.Join(errs, fmt.Errorf("failed to cast claim parameters to GPUClaimParametersSpec"))
			continue
		}
		for _, gpu := range availableGPUs {
			allocatedGPUs[claimUID] = append(allocatedGPUs[claimUID], gpu.UUID)
			if len(allocatedGPUs[claimUID]) >= claimParams.Count {
				break
			}
		}

		// if the number of allocated GPUs is less than the requested count, mark the node as
		// unsuitable
		if claimParams.Count != len(allocatedGPUs[claimUID]) {
			for _, claim := range allClaims {
				claim.UnsuitableNodes = append(claim.UnsuitableNodes, potentialNode)
			}
			return nil
		}

		// otherwise, potentialNode is a suitable node
		nodeClaim := nodeClaim{
			claimUID: claimUID,
			nodeName: potentialNode,
		}
		if _, exists := p.pendingAllocatedGPUs[nodeClaim]; !exists {
			p.pendingAllocatedGPUs[nodeClaim] = []*gpuv1alpha1.GPUDevice{}
		}
		for _, gpu := range allocatedGPUs[claimUID] {
			allocatedGPU := &gpuv1alpha1.GPUDevice{UUID: gpu}
			p.pendingAllocatedGPUs[nodeClaim] = append(p.pendingAllocatedGPUs[nodeClaim], allocatedGPU)
		}
	}

	return errs
}

func (p *gpuPlugin) availableGPUs(nodeDevices *gpuv1alpha1.NodeDevices) map[string]*gpuv1alpha1.GPUDevice {
	available := map[string]*gpuv1alpha1.GPUDevice{}
	for _, gpu := range nodeDevices.Spec.AvailableGPUs {
		available[gpu.UUID] = gpu
	}

	for _, gpus := range nodeDevices.Status.AllocatedGPUs {
		for _, gpu := range gpus {
			delete(available, gpu.UUID)
		}
	}

	return available
}
