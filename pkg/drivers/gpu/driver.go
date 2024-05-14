package gpu

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	resourcev1alpha2 "k8s.io/api/resource/v1alpha2"
	"k8s.io/dynamic-resource-allocation/controller"
	dractrl "k8s.io/dynamic-resource-allocation/controller"

	gpuclientset "github.com/ihcsim/k8s-dra/pkg/apis/clientset/versioned"
	gpuapiv1alpha1 "github.com/ihcsim/k8s-dra/pkg/apis/gpu/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	DriverAPIGroup = gpuapiv1alpha1.GroupName
	DriverName     = "driver.gpu.resource.ihcsim"
)

var _ controller.Driver = &driver{}

// driver implemetns the controller.Driver interface, to provide the actual
// allocation and deallocation operations.
type driver struct {
	clientset gpuclientset.Interface
}

// NewDriver returns a new instance of the GPU driver.
func NewDriver() *driver {
	return &driver{}
}

// GetName returns the name of the driver.
func (d *driver) GetName() string {
	return DriverName
}

// see https://pkg.go.dev/k8s.io/dynamic-resource-allocation/controller#Driver
func (d *driver) GetClassParameters(ctx context.Context, class *resourcev1alpha2.ResourceClass) (interface{}, error) {
	if class.ParametersRef == nil {
		return &gpuapiv1alpha1.DeviceClassParametersSpec{
			DeviceSelector: []gpuapiv1alpha1.DeviceSelector{
				{
					Type: gpuapiv1alpha1.DeviceTypeGPU,
					Name: "*",
				},
			},
		}, nil
	}

	if class.ParametersRef.APIGroup != DriverAPIGroup {
		return nil, fmt.Errorf("incorrect API group: %v", class.ParametersRef.APIGroup)
	}

	dc, err := d.clientset.GpuV1alpha1().DeviceClassParameters().Get(ctx, class.ParametersRef.Name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error getting DeviceClassParameters called '%v': %w", class.ParametersRef.Name, err)
	}

	return &dc.Spec, nil
}

// see https://pkg.go.dev/k8s.io/dynamic-resource-allocation/controller#Driver
func (d *driver) GetClaimParameters(ctx context.Context, claim *resourcev1alpha2.ResourceClaim, class *resourcev1alpha2.ResourceClass, classParameters interface{}) (interface{}, error) {
	if claim.Spec.ParametersRef == nil {
		return gpuapiv1alpha1.ResourceClaimParametersSpec{
			Count: 1,
		}, nil
	}

	if claim.Spec.ParametersRef.APIGroup != DriverAPIGroup {
		return nil, fmt.Errorf("incorrect API group: %v", claim.Spec.ParametersRef.APIGroup)
	}

	switch claim.Spec.ParametersRef.Kind {
	case gpuapiv1alpha1.GPUClaimParametersKind:
		rc, err := d.clientset.GpuV1alpha1().ResourceClaimParameters(claim.Namespace).Get(ctx, claim.Spec.ParametersRef.Name, metav1.GetOptions{})
		if err != nil {
			return nil, fmt.Errorf("error getting GpuClaimParameters called '%v' in namespace '%v': %v", claim.Spec.ParametersRef.Name, claim.Namespace, err)
		}

		err = d.validateClaimParameters(&rc.Spec)
		if err != nil {
			return nil, fmt.Errorf("error validating ResourceClaimParameters called '%v' in namespace '%v': %w", claim.Spec.ParametersRef.Name, claim.Namespace, err)
		}

		return &rc.Spec, nil

	default:
		return nil, fmt.Errorf("unknown ResourceClaim.ParametersRef.Kind: %v", claim.Spec.ParametersRef.Kind)
	}
}

// see https://pkg.go.dev/k8s.io/dynamic-resource-allocation/controller#Driver
func (d *driver) Allocate(ctx context.Context, claims []*dractrl.ClaimAllocation, selectedNode string) {

}

// see https://pkg.go.dev/k8s.io/dynamic-resource-allocation/controller#Driver
func (d *driver) Deallocate(ctx context.Context, claim *resourcev1alpha2.ResourceClaim) error {
	return nil
}

// see https://pkg.go.dev/k8s.io/dynamic-resource-allocation/controller#Driver
func (d *driver) UnsuitableNodes(ctx context.Context, pod *v1.Pod, claims []*dractrl.ClaimAllocation, potentialNodes []string) error {
	return nil
}

func (d *driver) validateClaimParameters(claimParams *gpuapiv1alpha1.ResourceClaimParametersSpec) error {
	if claimParams.Count < 1 {
		return fmt.Errorf("invalid number of GPUs requested: %v", claimParams.Count)
	}

	return nil
}
