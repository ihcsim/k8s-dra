package gpu

import (
	"context"

	v1 "k8s.io/api/core/v1"
	resourcev1alpha2 "k8s.io/api/resource/v1alpha2"
	"k8s.io/dynamic-resource-allocation/controller"
	dractrl "k8s.io/dynamic-resource-allocation/controller"
)

// driver implemetns the controller.Driver interface, to provide the actual
// allocation and deallocation operations.
type driver struct{}

var _ controller.Driver = &driver{}

// NewDriver returns a new instance of the GPU driver.
func NewDriver() *driver {
	return &driver{}
}

// GetName returns the name of the driver.
func (d *driver) GetName() string {
	return "gpudrv.ihcsim.github.com"
}

// see https://pkg.go.dev/k8s.io/dynamic-resource-allocation/controller#Driver
func (d *driver) GetClassParameters(ctx context.Context, class *resourcev1alpha2.ResourceClass) (interface{}, error) {
	return nil, nil

}

// see https://pkg.go.dev/k8s.io/dynamic-resource-allocation/controller#Driver
func (d *driver) GetClaimParameters(ctx context.Context, claim *resourcev1alpha2.ResourceClaim, class *resourcev1alpha2.ResourceClass, classParameters interface{}) (interface{}, error) {
	return nil, nil
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
