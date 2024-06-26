//go:build !ignore_autogenerated
// +build !ignore_autogenerated

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

// Code generated by deepcopy-gen. DO NOT EDIT.

package v1alpha1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DeviceAllocation) DeepCopyInto(out *DeviceAllocation) {
	*out = *in
	in.Claim.DeepCopyInto(&out.Claim)
	if in.Device != nil {
		in, out := &in.Device, &out.Device
		*out = new(GPUDevice)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DeviceAllocation.
func (in *DeviceAllocation) DeepCopy() *DeviceAllocation {
	if in == nil {
		return nil
	}
	out := new(DeviceAllocation)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DeviceSelector) DeepCopyInto(out *DeviceSelector) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DeviceSelector.
func (in *DeviceSelector) DeepCopy() *DeviceSelector {
	if in == nil {
		return nil
	}
	out := new(DeviceSelector)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GPUClassParameters) DeepCopyInto(out *GPUClassParameters) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GPUClassParameters.
func (in *GPUClassParameters) DeepCopy() *GPUClassParameters {
	if in == nil {
		return nil
	}
	out := new(GPUClassParameters)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *GPUClassParameters) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GPUClassParametersList) DeepCopyInto(out *GPUClassParametersList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]GPUClassParameters, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GPUClassParametersList.
func (in *GPUClassParametersList) DeepCopy() *GPUClassParametersList {
	if in == nil {
		return nil
	}
	out := new(GPUClassParametersList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *GPUClassParametersList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GPUClassParametersSpec) DeepCopyInto(out *GPUClassParametersSpec) {
	*out = *in
	if in.DeviceSelector != nil {
		in, out := &in.DeviceSelector, &out.DeviceSelector
		*out = make([]DeviceSelector, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GPUClassParametersSpec.
func (in *GPUClassParametersSpec) DeepCopy() *GPUClassParametersSpec {
	if in == nil {
		return nil
	}
	out := new(GPUClassParametersSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GPUDevice) DeepCopyInto(out *GPUDevice) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GPUDevice.
func (in *GPUDevice) DeepCopy() *GPUDevice {
	if in == nil {
		return nil
	}
	out := new(GPUDevice)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GPURequirements) DeepCopyInto(out *GPURequirements) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GPURequirements.
func (in *GPURequirements) DeepCopy() *GPURequirements {
	if in == nil {
		return nil
	}
	out := new(GPURequirements)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *GPURequirements) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GPURequirementsList) DeepCopyInto(out *GPURequirementsList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]GPURequirements, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GPURequirementsList.
func (in *GPURequirementsList) DeepCopy() *GPURequirementsList {
	if in == nil {
		return nil
	}
	out := new(GPURequirementsList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *GPURequirementsList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GPURequirementsSpec) DeepCopyInto(out *GPURequirementsSpec) {
	*out = *in
	out.Memory = in.Memory.DeepCopy()
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GPURequirementsSpec.
func (in *GPURequirementsSpec) DeepCopy() *GPURequirementsSpec {
	if in == nil {
		return nil
	}
	out := new(GPURequirementsSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NodeGPUSlices) DeepCopyInto(out *NodeGPUSlices) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	if in.AllocatableGPUs != nil {
		in, out := &in.AllocatableGPUs, &out.AllocatableGPUs
		*out = make([]*GPUDevice, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(GPUDevice)
				**out = **in
			}
		}
	}
	if in.Allocations != nil {
		in, out := &in.Allocations, &out.Allocations
		*out = make(map[string][]*DeviceAllocation, len(*in))
		for key, val := range *in {
			var outVal []*DeviceAllocation
			if val == nil {
				(*out)[key] = nil
			} else {
				in, out := &val, &outVal
				*out = make([]*DeviceAllocation, len(*in))
				for i := range *in {
					if (*in)[i] != nil {
						in, out := &(*in)[i], &(*out)[i]
						*out = new(DeviceAllocation)
						(*in).DeepCopyInto(*out)
					}
				}
			}
			(*out)[key] = outVal
		}
	}
	if in.NodeSuitability != nil {
		in, out := &in.NodeSuitability, &out.NodeSuitability
		*out = make(map[string]NodeSuitability, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NodeGPUSlices.
func (in *NodeGPUSlices) DeepCopy() *NodeGPUSlices {
	if in == nil {
		return nil
	}
	out := new(NodeGPUSlices)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *NodeGPUSlices) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NodeGPUSlicesList) DeepCopyInto(out *NodeGPUSlicesList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]NodeGPUSlices, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NodeGPUSlicesList.
func (in *NodeGPUSlicesList) DeepCopy() *NodeGPUSlicesList {
	if in == nil {
		return nil
	}
	out := new(NodeGPUSlicesList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *NodeGPUSlicesList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}
