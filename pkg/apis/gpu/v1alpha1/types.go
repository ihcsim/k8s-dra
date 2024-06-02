package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
// +kubebuilder:resource:scope=Namespaced

// NodeDevices holds the spec of GPU devices on a node, and the devices'
// allocation state. A GPU device can be in one of three states: allocatable,
// allocated, or prepared.
// The name of the object is the name of the node.
type NodeDevices struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	AllocatableGPUs []*GPUDevice                   `json:"allocatedGPUs,omitempty"`
	Allocations     map[string][]*DeviceAllocation `json:"allocations,omitempty"`
	NodeSuitability map[string]NodeSuitability     `json:"nodeSuitability,omitempty"`
}

// DeviceAllocation represents the allocation state of a GPU device.
type DeviceAllocation struct {
	Claim  corev1.TypedLocalObjectReference `json:"claim"`
	Device *GPUDevice                       `json:"devices"`
	State  DeviceAllocationState            `json:"state"`
}

// DeviceAllocationState represents the state of a GPU device. A GPU device can
// be in one of three states: allocatable, allocated, or prepared.
type DeviceAllocationState string

const (
	// the kubelet plugin determines the allocatable devices on a node
	DeviceAllocationStateAllocatable = "allocatable"

	// the device driver places a temporary hold on a device if the host node
	// is deemed suitable for satisfying a pod's resource claim
	DeviceAllocationStateHold = "hold"

	// the device driver allocates a device to a pod based on the pod's resource
	// claim request
	DeviceAllocationStateAllocated = "allocated"

	// the kubelet plugin prepares an allocated device for use by a pod
	DeviceAllocationStatePrepared = "prepared"
)

// NodeSuitability describes the suitability of a node for running GPU workloads.
type NodeSuitability string

const (
	NodeSuitabilitySuitable   = "suitable"
	NodeSuitabilityUnsuitable = "unsuitable"
	NodeSuitabilityUnknown    = "unknown"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NodeDevicesList represents a list of NodeDevices CRD objects.
type NodeDevicesList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []NodeDevices `json:"items"`
}

// GPUDevice represents an allocatable GPU device on a node.
type GPUDevice struct {
	UUID        string `json:"uuid"`
	ProductName string `json:"productName"`
	Vendor      string `json:"vendor"`
}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true

// GPUClassParameters holds the set of parameters provided when creating a resource class for this driver.
type GPUClassParameters struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec GPUClassParametersSpec `json:"spec,omitempty"`
}

// GPUClassParametersSpec is the spec for the GPUClassParametersSpec CRD.
type GPUClassParametersSpec struct {
	DeviceSelector []DeviceSelector `json:"deviceSelector,omitempty"`
}

// DeviceSelector allows one to match on a specific type of Device as part of the class.
type DeviceSelector struct {
	Name   string `json:"name"`
	Vendor string `json:"vendor"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GPUClassParametersList represents the "plural" of a DeviceClassParameters CRD object.
type GPUClassParametersList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []GPUClassParameters `json:"items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
// +kubebuilder:resource:scope=Namespaced

// GPUClaimParameters holds the set of parameters provided when creating a resource claim for a GPU.
type GPUClaimParameters struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec GPUClaimParametersSpec `json:"spec,omitempty"`
}

// GPUClaimParametersSpec is the spec for the ResourceClaimParameters CRD.
type GPUClaimParametersSpec struct {
	Count int `json:"count,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GPUClaimParametersList represents the "plural" of a ResourceClaimParameters CRD object.
type GPUClaimParametersList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []GPUClaimParameters `json:"items"`
}
