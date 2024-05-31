package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
// +kubebuilder:resource:scope=Namespaced

// NodeDevices holds the availability and allocation states of GPUs
// on a node. The name of the object is the name of the node.
type NodeDevices struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NodeDevicesSpec   `json:"spec,omitempty"`
	Status NodeDevicesStatus `json:"status,omitempty"`
}

// NodeDevicesSpec is the spec for the DeviceAllocation CRD.
type NodeDevicesSpec struct {
	AllocatableGPUs []*GPUDevice `json:"availableGpus,omitempty"`
}

// DeviceAllocationState is the status for the DeviceAllocation CRD.
type NodeDevicesStatus struct {
	State         NodeDevicesAllocationState `json:"state"`
	AllocatedGPUs map[string][]*GPUDevice    `json:"allocatedGpus,omitempty"`
	PreparedGPUs  map[string][]*GPUDevice    `json:"preparedGpus,omitempty"`
}

type NodeDevicesAllocationState int

const (
	NodeDevicesAllocationStateReady NodeDevicesAllocationState = iota
	NodeDevicesAllocationStateNotReady
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
