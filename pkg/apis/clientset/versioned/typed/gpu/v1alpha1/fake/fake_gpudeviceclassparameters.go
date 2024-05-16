/*
Copyright The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"
	json "encoding/json"
	"fmt"

	gpuv1alpha1 "github.com/ihcsim/k8s-dra/pkg/apis/applyconfiguration/gpu/v1alpha1"
	v1alpha1 "github.com/ihcsim/k8s-dra/pkg/apis/gpu/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeGPUDeviceClassParameters implements GPUDeviceClassParametersInterface
type FakeGPUDeviceClassParameters struct {
	Fake *FakeGpuV1alpha1
}

var gpudeviceclassparametersResource = v1alpha1.SchemeGroupVersion.WithResource("gpudeviceclassparameters")

var gpudeviceclassparametersKind = v1alpha1.SchemeGroupVersion.WithKind("GPUDeviceClassParameters")

// Get takes name of the gPUDeviceClassParameters, and returns the corresponding gPUDeviceClassParameters object, and an error if there is any.
func (c *FakeGPUDeviceClassParameters) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.GPUDeviceClassParameters, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootGetAction(gpudeviceclassparametersResource, name), &v1alpha1.GPUDeviceClassParameters{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.GPUDeviceClassParameters), err
}

// List takes label and field selectors, and returns the list of GPUDeviceClassParameters that match those selectors.
func (c *FakeGPUDeviceClassParameters) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.GPUDeviceClassParametersList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootListAction(gpudeviceclassparametersResource, gpudeviceclassparametersKind, opts), &v1alpha1.GPUDeviceClassParametersList{})
	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.GPUDeviceClassParametersList{ListMeta: obj.(*v1alpha1.GPUDeviceClassParametersList).ListMeta}
	for _, item := range obj.(*v1alpha1.GPUDeviceClassParametersList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested gPUDeviceClassParameters.
func (c *FakeGPUDeviceClassParameters) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewRootWatchAction(gpudeviceclassparametersResource, opts))
}

// Create takes the representation of a gPUDeviceClassParameters and creates it.  Returns the server's representation of the gPUDeviceClassParameters, and an error, if there is any.
func (c *FakeGPUDeviceClassParameters) Create(ctx context.Context, gPUDeviceClassParameters *v1alpha1.GPUDeviceClassParameters, opts v1.CreateOptions) (result *v1alpha1.GPUDeviceClassParameters, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootCreateAction(gpudeviceclassparametersResource, gPUDeviceClassParameters), &v1alpha1.GPUDeviceClassParameters{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.GPUDeviceClassParameters), err
}

// Update takes the representation of a gPUDeviceClassParameters and updates it. Returns the server's representation of the gPUDeviceClassParameters, and an error, if there is any.
func (c *FakeGPUDeviceClassParameters) Update(ctx context.Context, gPUDeviceClassParameters *v1alpha1.GPUDeviceClassParameters, opts v1.UpdateOptions) (result *v1alpha1.GPUDeviceClassParameters, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateAction(gpudeviceclassparametersResource, gPUDeviceClassParameters), &v1alpha1.GPUDeviceClassParameters{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.GPUDeviceClassParameters), err
}

// Delete takes name of the gPUDeviceClassParameters and deletes it. Returns an error if one occurs.
func (c *FakeGPUDeviceClassParameters) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewRootDeleteActionWithOptions(gpudeviceclassparametersResource, name, opts), &v1alpha1.GPUDeviceClassParameters{})
	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeGPUDeviceClassParameters) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewRootDeleteCollectionAction(gpudeviceclassparametersResource, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.GPUDeviceClassParametersList{})
	return err
}

// Patch applies the patch and returns the patched gPUDeviceClassParameters.
func (c *FakeGPUDeviceClassParameters) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.GPUDeviceClassParameters, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceAction(gpudeviceclassparametersResource, name, pt, data, subresources...), &v1alpha1.GPUDeviceClassParameters{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.GPUDeviceClassParameters), err
}

// Apply takes the given apply declarative configuration, applies it and returns the applied gPUDeviceClassParameters.
func (c *FakeGPUDeviceClassParameters) Apply(ctx context.Context, gPUDeviceClassParameters *gpuv1alpha1.GPUDeviceClassParametersApplyConfiguration, opts v1.ApplyOptions) (result *v1alpha1.GPUDeviceClassParameters, err error) {
	if gPUDeviceClassParameters == nil {
		return nil, fmt.Errorf("gPUDeviceClassParameters provided to Apply must not be nil")
	}
	data, err := json.Marshal(gPUDeviceClassParameters)
	if err != nil {
		return nil, err
	}
	name := gPUDeviceClassParameters.Name
	if name == nil {
		return nil, fmt.Errorf("gPUDeviceClassParameters.Name must be provided to Apply")
	}
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceAction(gpudeviceclassparametersResource, *name, types.ApplyPatchType, data), &v1alpha1.GPUDeviceClassParameters{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.GPUDeviceClassParameters), err
}
