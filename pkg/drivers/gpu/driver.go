package gpu

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/ihcsim/k8s-dra/pkg/apis"
	draclientset "github.com/ihcsim/k8s-dra/pkg/apis/clientset/versioned"
	gpuv1alpha1 "github.com/ihcsim/k8s-dra/pkg/apis/gpu/v1alpha1"
	zlog "github.com/rs/zerolog"
	corev1 "k8s.io/api/core/v1"
	resourcev1alpha2 "k8s.io/api/resource/v1alpha2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	dractrl "k8s.io/dynamic-resource-allocation/controller"
)

const (
	apiGroup   = apis.GroupName
	driverName = "driver.resources.ihcsim"
)

var _ dractrl.Driver = &driver{}

// driver implements the controller.Driver interface, to provide the actual
// allocation and deallocation operations of GPU resources.
type driver struct {
	clientsets draclientset.Interface
	namespace  string
	log        zlog.Logger
}

// NewDriver returns a new instance of the GPU driver.
func NewDriver(clientsets draclientset.Interface, namespace string, log zlog.Logger) (*driver, error) {
	return &driver{
		clientsets: clientsets,
		namespace:  namespace,
		log:        log,
	}, nil
}

// GetName returns the name of the driver.
func (d *driver) GetName() string {
	return driverName
}

// GetClassParameters retrieves the underlying concrete GPU resource class
// parameters.
// see https://pkg.go.dev/k8s.io/dynamic-resource-allocation/controller#Driver
func (d *driver) GetClassParameters(ctx context.Context, class *resourcev1alpha2.ResourceClass) (interface{}, error) {
	if class.ParametersRef == nil {
		d.log.Info().Msg("no class parameters found, so using default values")
		return &gpuv1alpha1.GPUClassParametersSpec{
			DeviceSelector: []gpuv1alpha1.DeviceSelector{
				{
					Vendor: "*",
					Name:   "*",
				},
			},
		}, nil
	}

	if class.DriverName != d.GetName() {
		return nil, fmt.Errorf("incorrect driver name %s (vs. %s)", class.DriverName, d.GetName())
	}

	if class.ParametersRef.APIGroup != apiGroup {
		return nil, fmt.Errorf("incorrect API group %s (vs. %s)", class.ParametersRef.APIGroup, apiGroup)
	}

	classParams, err := d.clientsets.GpuV1alpha1().GPUClassParameters().Get(ctx, class.ParametersRef.Name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error getting DeviceClassParameters called '%s': %w", class.ParametersRef.Name, err)
	}

	d.log.Info().Msgf("found class parameters %s", classParams.GetName())
	return &classParams.Spec, nil
}

// GetClaimParameters retrieves the underlying concrete GPU resource claim
// parameters.
// see https://pkg.go.dev/k8s.io/dynamic-resource-allocation/controller#Driver
func (d *driver) GetClaimParameters(
	ctx context.Context,
	claim *resourcev1alpha2.ResourceClaim,
	class *resourcev1alpha2.ResourceClass,
	classParameters interface{}) (interface{}, error) {
	if claim.Spec.ParametersRef == nil {
		d.log.Info().Msg("no claim parameters found, so using default values")
		return gpuv1alpha1.GPURequirementsSpec{
			Count: 1,
		}, nil
	}

	if class.DriverName != d.GetName() {
		return nil, fmt.Errorf("incorrect driver name %s (vs. %s)", class.DriverName, d.GetName())
	}

	if claim.Spec.ParametersRef.APIGroup != apiGroup {
		return nil, fmt.Errorf("incorrect API group: %s (vs. %s)", claim.Spec.ParametersRef.APIGroup, apiGroup)
	}

	if !strings.EqualFold(claim.Spec.ParametersRef.Kind, gpuv1alpha1.GPURequirementsKind) {
		return nil, fmt.Errorf("unsupported resource claim kind: %v", claim.Spec.ParametersRef.Kind)
	}

	claimParams, err := d.clientsets.GpuV1alpha1().GPURequirements(claim.Namespace).Get(ctx, claim.Spec.ParametersRef.Name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error getting GPURequirements called '%v' in namespace '%v': %v", claim.Spec.ParametersRef.Name, claim.Namespace, err)
	}

	if err := d.validateClaimParameters(&claimParams.Spec); err != nil {
		return nil, fmt.Errorf("error validating GPURequirements called '%v' in namespace '%v': %w", claim.Spec.ParametersRef.Name, claim.Namespace, err)
	}

	d.log.Info().Msgf("successfully retrieved claim parameters: %s", claimParams.GetName())
	return &claimParams.Spec, nil
}

// Allocate is called when all same-driver ResourceClaims for Pod are ready to be
// allocated.
// see https://pkg.go.dev/k8s.io/dynamic-resource-allocation/controller#Driver
func (d *driver) Allocate(ctx context.Context, claimAllocations []*dractrl.ClaimAllocation, selectedNode string) {
	d.log.Debug().Msg("attempting to allocate GPUs...")
	for _, ca := range claimAllocations {
		if selectedNode == "" {
			ca.Error = fmt.Errorf("immediate allocation is not supported.")
			continue
		}

		if err := d.allocate(ctx, ca, selectedNode); err != nil {
			ca.Error = err
			continue
		}
		ca.Allocation = buildAllocationResult(selectedNode, true)
	}
}

func (d *driver) allocate(
	ctx context.Context,
	claimAllocation *dractrl.ClaimAllocation,
	selectedNode string) error {
	var (
		claim    = claimAllocation.Claim
		claimUID = string(claim.GetUID())
		log      = d.log.With().
				Str("podClaimName", claimAllocation.PodClaimName).
				Str("selectedNode", selectedNode).
				Str("claimUID", claimUID).
				Logger()
	)

	getOpts := metav1.GetOptions{}
	nodeDevices, err := d.clientsets.GpuV1alpha1().NodeGPUSlices(d.namespace).Get(ctx, selectedNode, getOpts)
	if err != nil {
		return err
	}

	log.Info().Msg("allocating GPUs...")
	allocatableGPUs, err := d.findAllocatableGPUs(nodeDevices, claimAllocation, selectedNode)
	if err != nil {
		return err
	}

	for _, allocatable := range allocatableGPUs {
		apiGroup := apis.GroupName
		newDeviceAllocation := &gpuv1alpha1.DeviceAllocation{
			Claim: corev1.TypedLocalObjectReference{
				APIGroup: &apiGroup,
				Kind:     gpuv1alpha1.GPURequirementsKind,
				Name:     claimUID,
			},
			Device: allocatable,
			State:  gpuv1alpha1.DeviceAllocationStateAllocated,
		}
		nodeDevices.Allocations[claimUID] = append(nodeDevices.Allocations[claimUID], newDeviceAllocation)
	}

	if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		_, err := d.clientsets.GpuV1alpha1().NodeGPUSlices(d.namespace).Update(ctx, nodeDevices, metav1.UpdateOptions{})
		return err
	}); err != nil {
		return err
	}

	log.Info().Msg("allocation completed")
	return nil
}

// Deallocate gets called when a ResourceClaim is ready to be freed.
// see https://pkg.go.dev/k8s.io/dynamic-resource-allocation/controller#Driver
func (d *driver) Deallocate(ctx context.Context, claim *resourcev1alpha2.ResourceClaim) error {
	d.log.Debug().Msg("attempting to deallocate GPUs...")
	if !claim.Status.DeallocationRequested {
		return fmt.Errorf("unexpected deallocation request")
	}

	selectedNode := getSelectedNode(claim)
	if selectedNode == "" {
		d.log.Info().Msg("no selected node found, skipping deallocation")
		return nil
	}

	var (
		claimUID = string(claim.GetUID())
		log      = d.log.With().
				Str("selectedNode", selectedNode).
				Str("claimUID", claimUID).
				Logger()
	)

	nodeDevices, err := d.clientsets.GpuV1alpha1().NodeGPUSlices(d.namespace).Get(ctx, selectedNode, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if _, exists := nodeDevices.Allocations[claimUID]; !exists {
		log.Info().Msg("no GPUs allocated, skipping deallocation")
		return nil
	}

	log.Info().Msg("deallocating claimed GPUs...")
	delete(nodeDevices.Allocations, claimUID)

	if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		_, err := d.clientsets.GpuV1alpha1().NodeGPUSlices(d.namespace).Update(ctx, nodeDevices, metav1.UpdateOptions{})
		return err
	}); err != nil {
		return err
	}

	log.Info().Msg("deallocation completed")
	return nil
}

func getSelectedNode(claim *resourcev1alpha2.ResourceClaim) string {
	if claim.Status.Allocation == nil {
		return ""
	}

	if claim.Status.Allocation.AvailableOnNodes == nil {
		return ""
	}

	return claim.Status.Allocation.AvailableOnNodes.NodeSelectorTerms[0].MatchFields[0].Values[0]
}

// UnsuitableNodes checks all pending claims with delayed allocation for a pod.
// see https://pkg.go.dev/k8s.io/dynamic-resource-allocation/controller#Driver
func (d *driver) UnsuitableNodes(ctx context.Context, pod *corev1.Pod, claims []*dractrl.ClaimAllocation, potentialNodes []string) error {
	d.log.Debug().Msg("assessing potential node's suitability...")
	var errs error
	for _, potentialNode := range potentialNodes {
		nodeDevices, err := d.clientsets.GpuV1alpha1().NodeGPUSlices(d.namespace).Get(ctx, potentialNode, metav1.GetOptions{})
		if err != nil {
			for _, claim := range claims {
				claim.UnsuitableNodes = append(claim.UnsuitableNodes, potentialNode)
			}
			return nil
		}

		nodeDevicesUpdated, err := d.unsuitableNode(nodeDevices, pod, claims, potentialNode)
		if err != nil {
			errs = errors.Join(errs, err)
			continue
		}

		if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			_, err := d.clientsets.GpuV1alpha1().NodeGPUSlices(d.namespace).Update(ctx, nodeDevicesUpdated, metav1.UpdateOptions{})
			return err
		}); err != nil {
			errs = errors.Join(errs, err)
			continue
		}
	}

	d.log.Info().Msg("unsuitable nodes check completed")
	return errs
}

func (d *driver) validateClaimParameters(claimParams *gpuv1alpha1.GPURequirementsSpec) error {
	if claimParams.Count < 1 {
		return fmt.Errorf("invalid number of GPUs requested: %v", claimParams.Count)
	}

	return nil
}

func buildAllocationResult(selectedNode string, shareable bool) *resourcev1alpha2.AllocationResult {
	nodeSelector := &corev1.NodeSelector{
		NodeSelectorTerms: []corev1.NodeSelectorTerm{
			{
				MatchFields: []corev1.NodeSelectorRequirement{
					{
						Key:      "metadata.name",
						Operator: "In",
						Values:   []string{selectedNode},
					},
				},
			},
		},
	}

	return &resourcev1alpha2.AllocationResult{
		AvailableOnNodes: nodeSelector,
		Shareable:        shareable,
	}
}

func (d *driver) findAllocatableGPUs(
	nodeDevices *gpuv1alpha1.NodeGPUSlices,
	claimAllocation *dractrl.ClaimAllocation,
	selectedNode string) ([]*gpuv1alpha1.GPUDevice, error) {
	claimParams, ok := claimAllocation.ClaimParameters.(*gpuv1alpha1.GPURequirementsSpec)
	if !ok {
		return nil, fmt.Errorf("unsupported claim parameters kind: %T", claimAllocation.ClaimParameters)
	}

	availableGPUs := d.availableGPUs(nodeDevices)
	if len(availableGPUs) < claimParams.Count {
		return nil, fmt.Errorf("insufficient GPUs on node %s for claim %s", selectedNode, claimAllocation.Claim.GetUID())
	}

	classParams, ok := claimAllocation.ClassParameters.(*gpuv1alpha1.GPUClassParametersSpec)
	if !ok {
		return nil, fmt.Errorf("unsupported class parameters kind: %T", claimAllocation.ClassParameters)
	}

	allocatableGPUs := []*gpuv1alpha1.GPUDevice{}
	for _, availableGPU := range availableGPUs {
		if len(classParams.DeviceSelector) > 0 {
			for _, selector := range classParams.DeviceSelector {
				if selector.Name == availableGPU.ProductName && selector.Vendor == availableGPU.Vendor {
					allocatableGPUs = append(allocatableGPUs, availableGPU)
				}
			}
			continue
		}
		allocatableGPUs = append(allocatableGPUs, availableGPU)
	}

	return allocatableGPUs, nil
}

func (d *driver) unsuitableNode(
	nodeDevices *gpuv1alpha1.NodeGPUSlices,
	pod *corev1.Pod,
	claims []*dractrl.ClaimAllocation,
	potentialNode string) (*gpuv1alpha1.NodeGPUSlices, error) {
	nodeDeviceClone := nodeDevices.DeepCopy()
	for _, claim := range claims {
		claimParams, ok := claim.ClaimParameters.(*gpuv1alpha1.GPURequirementsSpec)
		if !ok {
			d.log.Info().Msgf("skipping unsupported claim parameters kind: %T", claim.ClaimParameters)
			continue
		}

		// if the number of allocated GPUs is less than the requested count, mark the node as
		// unsuitable
		claimUID := string(claim.Claim.GetUID())
		allocatedCount := d.allocatedCount(nodeDevices, claimUID)
		if claimParams.Count != allocatedCount {
			d.log.Info().Msgf("insufficient GPUs on node %s for claim %s, marking node as unsuitable", potentialNode, claimUID)
			claim.UnsuitableNodes = append(claim.UnsuitableNodes, potentialNode)
			nodeDeviceClone.NodeSuitability[claimUID] = gpuv1alpha1.NodeSuitabilityUnsuitable
			continue
		}
		nodeDeviceClone.NodeSuitability[claimUID] = gpuv1alpha1.NodeSuitabilitySuitable
	}

	return nodeDeviceClone, nil
}

func (d *driver) allocatedCount(nodeDevices *gpuv1alpha1.NodeGPUSlices, claimUID string) int {
	// if existing allocated GPUs are found for this claim, return the count
	if allocations, exists := nodeDevices.Allocations[claimUID]; exists {
		allocatedCount := 0
		for _, allocation := range allocations {
			if allocation.State == gpuv1alpha1.DeviceAllocationStateAllocated {
				allocatedCount++
			}
		}
		return allocatedCount
	}

	return len(d.availableGPUs(nodeDevices))
}

func (d *driver) availableGPUs(
	nodeDevices *gpuv1alpha1.NodeGPUSlices) map[string]*gpuv1alpha1.GPUDevice {
	available := map[string]*gpuv1alpha1.GPUDevice{}
	for _, gpu := range nodeDevices.AllocatableGPUs {
		d.log.Info().Msgf("found allocatable GPU %s", gpu.UUID)
		available[gpu.UUID] = gpu
	}

	// find the GPUs that are already allocated, regardless of their claims
	// and remove them from the available list
	for _, allocations := range nodeDevices.Allocations {
		for _, allocation := range allocations {
			if allocation.State == gpuv1alpha1.DeviceAllocationStateAllocated {
				d.log.Info().Msgf("remove allocated GPU %s from available list", allocation.Device.UUID)
				delete(available, allocation.Device.UUID)
			}
		}
	}

	return available
}
