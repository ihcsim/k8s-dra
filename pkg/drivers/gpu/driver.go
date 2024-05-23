package gpu

import (
	"context"
	"errors"
	"fmt"
	"strings"

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
// allocation and deallocation operations of GPU resources.
type driver struct {
	clientset clientset.Interface
	gpu       gpuPlugin
	namespace string
}

// NewDriver returns a new instance of the GPU driver.
func NewDriver(namespace string) *driver {
	return &driver{
		gpu:       gpuPlugin{},
		namespace: namespace,
	}
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
		return &gpuv1alpha1.GPUClassParametersSpec{
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

	dc, err := d.clientset.GpuV1alpha1().GPUClassParameters().Get(ctx, class.ParametersRef.Name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error getting DeviceClassParameters called '%s': %w", class.ParametersRef.Name, err)
	}

	return &dc, nil
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
		return gpuv1alpha1.GPUClaimParametersSpec{
			Count: 1,
		}, nil
	}

	if class.DriverName != d.GetName() {
		return nil, fmt.Errorf("incorrect driver name %s (vs. %s)", class.DriverName, d.GetName())
	}

	if claim.Spec.ParametersRef.APIGroup != apiGroup {
		return nil, fmt.Errorf("incorrect API group: %s (vs. %s)", claim.Spec.ParametersRef.APIGroup, apiGroup)
	}

	if !strings.EqualFold(claim.Spec.ParametersRef.Kind, gpuv1alpha1.GPUClaimParametersKind) {
		return nil, fmt.Errorf("unsupported resource claim kind: %v", claim.Spec.ParametersRef.Kind)
	}

	rc, err := d.clientset.GpuV1alpha1().GPUClaimParameters(claim.Namespace).Get(ctx, claim.Spec.ParametersRef.Name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error getting GPUClaimParameters called '%v' in namespace '%v': %v", claim.Spec.ParametersRef.Name, claim.Namespace, err)
	}

	if err := d.validateClaimParameters(&rc.Spec); err != nil {
		return nil, fmt.Errorf("error validating GPUClaimParameters called '%v' in namespace '%v': %w", claim.Spec.ParametersRef.Name, claim.Namespace, err)
	}

	return &rc, nil
}

// see https://pkg.go.dev/k8s.io/dynamic-resource-allocation/controller#Driver
func (d *driver) Allocate(ctx context.Context, claimAllocations []*dractrl.ClaimAllocation, selectedNode string) {
	for _, ca := range claimAllocations {
		if selectedNode == "" {
			ca.Error = fmt.Errorf("failed to allocate device: immediate allocation is not supported.")
			continue
		}

		if err := d.allocateGPU(ctx, ca, selectedNode); err != nil {
			ca.Error = err
			continue
		}
		ca.Allocation = buildAllocationResult(selectedNode, true)
	}
}

func (d *driver) nodeDeviceAllocation(ctx context.Context, namespace, selectedNode string) (*allocationv1alpha1.NodeDeviceAllocation, error) {
	listOpts := metav1.ListOptions{
		LabelSelector: "kubernetes.io/hostname=" + selectedNode,
	}

	deviceAllocations, err := d.clientset.AllocationV1alpha1().NodeDeviceAllocations(namespace).List(ctx, listOpts)
	if err != nil {
		return nil, err
	}

	if len(deviceAllocations.Items) != 0 {
		return nil, fmt.Errorf("expected exactly one matching node for device allocation for node %s, got %d", selectedNode, len(deviceAllocations.Items))
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
	claimAllocation *dractrl.ClaimAllocation,
	selectedNode string) error {
	var (
		claim          = claimAllocation.Claim
		claimUID       = string(claim.GetUID())
		claimNamespace = claim.GetNamespace()
	)

	deviceAllocation, err := d.nodeDeviceAllocation(ctx, d.namespace, selectedNode)
	if err != nil {
		return err
	}

	// if there is an on-going allocation, let it finish
	if _, exists := deviceAllocation.Status.AllocatedClaims[claimUID]; exists {
		return nil
	}

	claimParams, ok := claimAllocation.ClaimParameters.(*gpuv1alpha1.GPUClaimParameters)
	if !ok {
		return fmt.Errorf("unsupported claim parameters kind: %T", claimAllocation.ClaimParameters)
	}

	classParams, ok := claimAllocation.ClassParameters.(*gpuv1alpha1.GPUClassParameters)
	if !ok {
		return fmt.Errorf("unsupported class parameters kind: %T", claimAllocation.ClassParameters)
	}

	allocatedClaims, err := d.gpu.pendingAllocatedClaims(
		claimUID,
		selectedNode,
		claimParams.DeepCopy(),
		classParams.DeepCopy())
	if err != nil {
		return err
	}

	deviceAllocation.Status.AllocatedClaims[claimUID] = allocatedClaims

	updateOpts := metav1.UpdateOptions{}
	if _, err := d.clientset.AllocationV1alpha1().NodeDeviceAllocations(claimNamespace).Update(ctx, deviceAllocation, updateOpts); err != nil {
		return err
	}

	return d.gpu.removeAllocatedClaim(claimUID)
}

// see https://pkg.go.dev/k8s.io/dynamic-resource-allocation/controller#Driver
func (d *driver) Deallocate(ctx context.Context, claim *resourcev1alpha2.ResourceClaim) error {
	selectedNode := getSelectedNode(claim)
	if selectedNode == "" {
		return nil
	}

	deviceAllocation, err := d.nodeDeviceAllocation(ctx, d.namespace, selectedNode)
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

	claimNamespace := claim.GetNamespace()
	updateOpts := metav1.UpdateOptions{}
	if _, err := d.clientset.AllocationV1alpha1().NodeDeviceAllocations(claimNamespace).Update(ctx, deviceAllocation, updateOpts); err != nil {
		return err
	}

	delete(deviceAllocation.Status.AllocatedClaims, claimUID)
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

func (d *driver) deallocateGPU(ctx context.Context, claimUID string, gpu allocationv1alpha1.AllocatedGPU) error {
	return d.gpu.removeAllocatedClaim(claimUID)
}

// see https://pkg.go.dev/k8s.io/dynamic-resource-allocation/controller#Driver
func (d *driver) UnsuitableNodes(ctx context.Context, pod *corev1.Pod, claims []*dractrl.ClaimAllocation, potentialNodes []string) error {
	var (
		errs      error
		gpuClaims = []*dractrl.ClaimAllocation{}
	)
	for _, potentialNode := range potentialNodes {
		deviceAllocation, err := d.nodeDeviceAllocation(ctx, d.namespace, potentialNode)
		if err != nil {
			for _, claim := range claims {
				claim.UnsuitableNodes = append(claim.UnsuitableNodes, potentialNode)
			}
			return nil
		}

		for _, claim := range claims {
			if _, ok := claim.ClaimParameters.(*gpuv1alpha1.GPUClaimParameters); !ok {
				errs = errors.Join(errs, fmt.Errorf("unsupported claim parameters kind: %T", claim.ClaimParameters))
				continue
			}
			gpuClaims = append(gpuClaims, claim)
		}

		if err := d.gpu.unsuitableNode(deviceAllocation, pod, gpuClaims, claims, potentialNode); err != nil {
			errs = errors.Join(errs, err)
		}
	}

	return errs
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
