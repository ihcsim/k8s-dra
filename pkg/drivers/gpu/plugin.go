package gpu

import (
	"fmt"
	"sync"

	gpuv1alpha1 "github.com/ihcsim/k8s-dra/pkg/apis/gpu/v1alpha1"
	zlog "github.com/rs/zerolog"
	corev1 "k8s.io/api/core/v1"
	dractrl "k8s.io/dynamic-resource-allocation/controller"
)

type gpuPlugin struct {
	// pendingAllocatedGPUs is a map of resource claim UID to a map of node name to allocated
	// GPUs.
	pendingAllocatedGPUs map[nodeClaim][]*gpuv1alpha1.GPUDevice

	// mux is used to synchronized concurrent read-write accesses to allocations.
	mux sync.RWMutex

	log zlog.Logger
}

type nodeClaim struct {
	claimUID string
	nodeName string
}

func (n *nodeClaim) String() string {
	return fmt.Sprintf("%s/%s", n.nodeName, n.claimUID)
}

func newGPUPlugin(log zlog.Logger) *gpuPlugin {
	return &gpuPlugin{
		pendingAllocatedGPUs: map[nodeClaim][]*gpuv1alpha1.GPUDevice{},
		mux:                  sync.RWMutex{},
		log:                  log,
	}
}

func (p *gpuPlugin) allocate(
	claimUID, nodeName string,
	claimParams *gpuv1alpha1.GPUClaimParametersSpec,
	classParams *gpuv1alpha1.GPUClassParametersSpec) ([]*gpuv1alpha1.GPUDevice, func() error, error) {
	p.mux.RLock()
	defer p.mux.RUnlock()

	nodeClaim := nodeClaim{
		claimUID: claimUID,
		nodeName: nodeName,
	}
	allocatedGPUs, exists := p.pendingAllocatedGPUs[nodeClaim]
	if !exists {
		return nil, nil, fmt.Errorf("no allocations generated for node claim %s", nodeClaim)
	}

	// once the allocation is committed on K8s side, remove the devices from the
	// pending list
	commitAllocation := func() error {
		return p.deallocate(claimUID, nodeName)
	}

	return allocatedGPUs, commitAllocation, nil
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
	claims []*dractrl.ClaimAllocation,
	potentialNode string) error {
	p.mux.Lock()
	defer p.mux.Unlock()

	gpuClaims := []*dractrl.ClaimAllocation{}
	for _, claim := range claims {
		if _, ok := claim.ClaimParameters.(*gpuv1alpha1.GPUClaimParametersSpec); !ok {
			p.log.Info().Msgf("skipping unsupported claim parameters kind: %T", claim.ClaimParameters)
			continue
		}
		gpuClaims = append(gpuClaims, claim)
	}

	allocatableGPUs := p.allocatableGPUs(nodeDevices)
	allocatedGPUs := map[string][]string{}
	for _, gpuClaim := range gpuClaims {
		nodeClaim := nodeClaim{
			claimUID: string(gpuClaim.Claim.GetUID()),
			nodeName: potentialNode,
		}

		// if the nodeDevices has already allocated GPUs for the claim, add the
		// allocated GPUs to the result map
		if gpus, exists := p.pendingAllocatedGPUs[nodeClaim]; exists {
			nodeDevices.Status.AllocatedGPUs[nodeClaim.claimUID] = gpus
			for _, gpu := range gpus {
				allocatedGPUs[nodeClaim.claimUID] = append(allocatedGPUs[nodeClaim.claimUID], gpu.UUID)
			}
			continue
		}

		// otherwise, allocate up to claimParams.Count GPUs from the available pool
		claimParams, ok := gpuClaim.ClaimParameters.(*gpuv1alpha1.GPUClaimParametersSpec)
		if !ok {
			p.log.Info().Msgf("skipping unsupported claim parameters kind: %T", gpuClaim.ClaimParameters)
			continue
		}
		for _, gpu := range allocatableGPUs {
			allocatedGPUs[nodeClaim.claimUID] = append(allocatedGPUs[nodeClaim.claimUID], gpu.UUID)
			if len(allocatedGPUs[nodeClaim.claimUID]) >= claimParams.Count {
				break
			}
		}

		// if the number of allocated GPUs is less than the requested count, mark the node as
		// unsuitable
		if claimParams.Count != len(allocatedGPUs[nodeClaim.claimUID]) {
			p.log.Info().Msgf("insufficient GPUs on node %s for claim %s, marking node as unsuitable", potentialNode, nodeClaim.claimUID)
			gpuClaim.UnsuitableNodes = append(gpuClaim.UnsuitableNodes, potentialNode)
			continue
		}

		// otherwise, potentialNode is a suitable node
		if _, exists := p.pendingAllocatedGPUs[nodeClaim]; !exists {
			p.pendingAllocatedGPUs[nodeClaim] = []*gpuv1alpha1.GPUDevice{}
		}
		for _, gpu := range allocatedGPUs[nodeClaim.claimUID] {
			allocatedGPU := &gpuv1alpha1.GPUDevice{UUID: gpu}
			p.pendingAllocatedGPUs[nodeClaim] = append(p.pendingAllocatedGPUs[nodeClaim], allocatedGPU)
		}
	}

	return nil
}

func (p *gpuPlugin) allocatableGPUs(nodeDevices *gpuv1alpha1.NodeDevices) map[string]*gpuv1alpha1.GPUDevice {
	available := map[string]*gpuv1alpha1.GPUDevice{}
	for _, gpu := range nodeDevices.Spec.AllocatableGPUs {
		available[gpu.UUID] = gpu
	}

	for _, gpus := range nodeDevices.Status.AllocatedGPUs {
		for _, gpu := range gpus {
			delete(available, gpu.UUID)
		}
	}

	return available
}
