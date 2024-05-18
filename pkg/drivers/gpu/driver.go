package gpu

import (
	"context"
	"errors"
	"fmt"

	"github.com/ihcsim/k8s-dra/pkg/apis"
	allocationv1alpha1 "github.com/ihcsim/k8s-dra/pkg/apis/allocation/v1alpha1"
	clientset "github.com/ihcsim/k8s-dra/pkg/apis/clientset/versioned"
	gpuv1alpha1 "github.com/ihcsim/k8s-dra/pkg/apis/gpu/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	resourcev1alpha2 "k8s.io/api/resource/v1alpha2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	dractrl "k8s.io/dynamic-resource-allocation/controller"
)

const (
	apiGroup   = apis.GroupName
	driverName = "driver.resources.ihcsim"
)

var _ dractrl.Driver = &driver{}

// driver implements the controller.Driver interface, to provide the actual
// allocation and deallocation operations.
type driver struct {
	clientset clientset.Interface
	gpu       GPUPlugin
}

// NewDriver returns a new instance of the GPU driver.
func NewDriver() *driver {
	return &driver{}
}

// GetName returns the name of the driver.
func (d *driver) GetName() string {
	return driverName
}

// see https://pkg.go.dev/k8s.io/dynamic-resource-allocation/controller#Driver
func (d *driver) GetClassParameters(ctx context.Context, class *resourcev1alpha2.ResourceClass) (interface{}, error) {
	if class.ParametersRef == nil {
		return &gpuv1alpha1.GPUDeviceClassParametersSpec{
			DeviceSelector: []gpuv1alpha1.DeviceSelector{
				{
					Type: gpuv1alpha1.DeviceTypeGPU,
					Name: "*",
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

	dc, err := d.clientset.GpuV1alpha1().GPUDeviceClassParameters().Get(ctx, class.ParametersRef.Name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error getting DeviceClassParameters called '%s': %w", class.ParametersRef.Name, err)
	}

	return &dc.Spec, nil
}

// see https://pkg.go.dev/k8s.io/dynamic-resource-allocation/controller#Driver
func (d *driver) GetClaimParameters(
	ctx context.Context,
	claim *resourcev1alpha2.ResourceClaim,
	class *resourcev1alpha2.ResourceClass,
	classParameters interface{}) (interface{}, error) {
	if claim.Spec.ParametersRef == nil {
		return gpuv1alpha1.GPUClaimParametersSpec{
			Count: 1,
		}, nil
	}

	if claim.Spec.ParametersRef.APIGroup != apiGroup {
		return nil, fmt.Errorf("incorrect API group: %s (vs. %s)", claim.Spec.ParametersRef.APIGroup, apiGroup)
	}

	switch claim.Spec.ParametersRef.Kind {
	case gpuv1alpha1.GPUClaimParametersKind:
		rc, err := d.clientset.GpuV1alpha1().GPUClaimParameters(claim.Namespace).Get(ctx, claim.Spec.ParametersRef.Name, metav1.GetOptions{})
		if err != nil {
			return nil, fmt.Errorf("error getting GPUClaimParameters called '%v' in namespace '%v': %v", claim.Spec.ParametersRef.Name, claim.Namespace, err)
		}

		if err := d.validateClaimParameters(&rc.Spec); err != nil {
			return nil, fmt.Errorf("error validating GPUClaimParameters called '%v' in namespace '%v': %w", claim.Spec.ParametersRef.Name, claim.Namespace, err)
		}
		return &rc.Spec, nil

	default:
		return nil, fmt.Errorf("unsupported resource claim kind: %v", claim.Spec.ParametersRef.Kind)
	}
}

// see https://pkg.go.dev/k8s.io/dynamic-resource-allocation/controller#Driver
func (d *driver) Allocate(ctx context.Context, claimAllocations []*dractrl.ClaimAllocation, selectedNode string) {
	for _, claimAllocation := range claimAllocations {
		if selectedNode == "" {
			claimAllocation.Error = fmt.Errorf("failed to allocate device: immediate allocation is not supported.")
			continue
		}

		deviceAllocation, err := d.nodeDeviceAllocation(ctx, claimAllocation.Claim, selectedNode)
		if err != nil {
			claimAllocation.Error = err
			continue
		}

		claimUID := claimAllocation.Claim.GetUID()
		if _, exists := deviceAllocation.Status.AllocatedClaims[string(claimUID)]; exists {
			claimAllocation.Allocation = buildAllocationResult(selectedNode, true)
			continue
		}

		releaseClaim := make(chan struct{})
		switch claimParams := claimAllocation.ClaimParameters.(type) {
		case *gpuv1alpha1.GPUClaimParametersSpec:
			claimParams, ok := claimAllocation.ClaimParameters.(*gpuv1alpha1.GPUClaimParametersSpec)
			if !ok {
				claimAllocation.Error = fmt.Errorf("invalid GPU claim parameters")
				continue
			}

			classParams, ok := claimAllocation.ClassParameters.(*gpuv1alpha1.GPUDeviceClassParametersSpec)
			if !ok {
				claimAllocation.Error = fmt.Errorf("invalid GPU class parameters")
				continue
			}

			claim := claimAllocation.Claim
			if err = d.allocateGPU(ctx, deviceAllocation, claim, claimParams, classParams, selectedNode, releaseClaim); err != nil {
				claimAllocation.Error = err
				continue
			}

		default:
			claimAllocation.Error = fmt.Errorf("unsupported claim parameters kind: %T", claimParams)
			continue
		}

		claimNamespace := claimAllocation.Claim.GetNamespace()
		updateOpts := metav1.UpdateOptions{}
		if _, err := d.clientset.AllocationV1alpha1().NodeDeviceAllocations(claimNamespace).Update(ctx, deviceAllocation, updateOpts); err != nil {
			claimAllocation.Error = err
			continue
		}
		claimAllocation.Allocation = buildAllocationResult(selectedNode, true)
		releaseClaim <- struct{}{}
	}
}

func (d *driver) nodeDeviceAllocation(ctx context.Context, claim *resourcev1alpha2.ResourceClaim, selectedNode string) (*allocationv1alpha1.NodeDeviceAllocation, error) {
	var (
		namespace = claim.GetNamespace()
		listOpts  = metav1.ListOptions{
			LabelSelector: "kubernetes.io/hostname=" + selectedNode,
		}
	)

	deviceAllocations, err := d.clientset.AllocationV1alpha1().NodeDeviceAllocations(namespace).List(ctx, listOpts)
	if err != nil {
		return nil, err
	}

	if len(deviceAllocations.Items) != 0 {
		return nil, fmt.Errorf("expected exactly one matching node for device allocation, got %d", len(deviceAllocations.Items))
	}
	deviceAllocation := deviceAllocations.Items[0]

	if deviceAllocation.Status.State != allocationv1alpha1.Ready {
		return nil, fmt.Errorf("failed to allocate device: device allocation not ready")
	}

	if deviceAllocation.Status.AllocatedClaims == nil {
		deviceAllocation.Status.AllocatedClaims = make(map[string]allocationv1alpha1.AllocatedDevices)
	}

	return deviceAllocation.DeepCopy(), nil
}

func (d *driver) allocateGPU(
	ctx context.Context,
	deviceAllocation *allocationv1alpha1.NodeDeviceAllocation,
	claim *resourcev1alpha2.ResourceClaim,
	claimParams *gpuv1alpha1.GPUClaimParametersSpec,
	classParams *gpuv1alpha1.GPUDeviceClassParametersSpec,
	selectedNode string,
	releaseClaim <-chan struct{}) error {
	claimUID := string(claim.GetUID())
	allocatedClaims, err := d.gpu.pendingAllocatedClaims(claimUID, selectedNode)
	if err != nil {
		return err
	}

	go func() {
		<-releaseClaim
		//nolint:staticcheck
		if err := d.gpu.removeAllocatedClaim(claimUID); err != nil {
			//@TODO log errors here
		}
	}()

	deviceAllocation.Status.AllocatedClaims[string(claim.GetUID())] = allocatedClaims
	return nil
}

// see https://pkg.go.dev/k8s.io/dynamic-resource-allocation/controller#Driver
func (d *driver) Deallocate(ctx context.Context, claim *resourcev1alpha2.ResourceClaim) error {
	if claim.Status.DriverName != d.GetName() {
		return nil
	}

	selectedNode := getSelectedNode(claim)
	if selectedNode == "" {
		return nil
	}

	deviceAllocation, err := d.nodeDeviceAllocation(ctx, claim, selectedNode)
	if err != nil {
		return err
	}

	if deviceAllocation.Status.AllocatedClaims == nil {
		return nil
	}

	claimUID := string(claim.GetUID())
	allocatedDevices, exists := deviceAllocation.Status.AllocatedClaims[claimUID]
	if !exists {
		return nil
	}

	var errs error
	if gpus := allocatedDevices.GPUs; gpus != nil {
		for _, gpu := range gpus.Devices {
			if err := d.deallocateGPU(ctx, claimUID, gpu); err != nil {
				errs = errors.Join(errs, err)
			}
		}
	}

	if errs != nil {
		return errs
	}

	delete(deviceAllocation.Status.AllocatedClaims, claimUID)

	claimNamespace := claim.GetNamespace()
	updateOpts := metav1.UpdateOptions{}
	_, err = d.clientset.AllocationV1alpha1().NodeDeviceAllocations(claimNamespace).Update(ctx, deviceAllocation, updateOpts)
	return err
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

func (d *driver) deallocateGPU(ctx context.Context, claimUID string, gpu allocationv1alpha1.AllocatedGPU) error {
	return d.gpu.removeAllocatedClaim(claimUID)
}

// see https://pkg.go.dev/k8s.io/dynamic-resource-allocation/controller#Driver
func (d *driver) UnsuitableNodes(ctx context.Context, pod *corev1.Pod, claims []*dractrl.ClaimAllocation, potentialNodes []string) error {
	return nil
}

func (d *driver) validateClaimParameters(claimParams *gpuv1alpha1.GPUClaimParametersSpec) error {
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
