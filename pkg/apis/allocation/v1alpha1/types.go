package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
// +kubebuilder:resource:scope=Namespaced

// NodeDeviceAllocation holds the state required for allocation on a node.
type NodeDeviceAllocation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NodeDeviceAllocationSpec   `json:"spec,omitempty"`
	Status NodeDeviceAllocationStatus `json:"status,omitempty"`
}

// NodeDeviceAllocationSpec is the spec for the DeviceAllocation CRD.
type NodeDeviceAllocationSpec struct{}

// DeviceAllocationState is the status for the DeviceAllocation CRD.
type NodeDeviceAllocationStatus struct {
	State              NodeDeviceAllocationState   `json:"state"`
	AllocatableDevices []AllocatableDevice         `json:"allocatableDevices,omitempty"`
	AllocatedClaims    map[string]AllocatedDevices `json:"allocatedClaims,omitempty"`
	PreparedClaims     map[string]PreparedDevices  `json:"preparedClaims,omitempty"`
}

type NodeDeviceAllocationState int

const (
	Ready NodeDeviceAllocationState = iota
	NotReady
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NodeDeviceAllocationList represents the "plural" of a DeviceAllocation CRD object.
type NodeDeviceAllocationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []NodeDeviceAllocation `json:"items"`
}

// AllocatableDevice represents an allocatable device on a node.
type AllocatableDevice struct {
	GPU *AllocatableGPU `json:"gpu,omitempty"`
}

// AllocatableGPU represents an allocatable GPU on a node.
type AllocatableGPU struct {
	UUID        string `json:"uuid"`
	ProductName string `json:"productName"`
}

// AllocatedDevices represents a set of allocated devices.
type AllocatedDevices struct {
	GPUs *AllocatedGPUs `json:"gpus,omitempty"`
}

// AllocatedGPUs represents a set of allocated GPUs.
type AllocatedGPUs struct {
	Devices []AllocatedGPU `json:"devices"`
}

// AllocatedGPU represents an allocated GPU.
type AllocatedGPU struct {
	UUID string `json:"uuid,omitempty"`
}

// PreparedGpu represents a prepared GPU on a node.
type PreparedGPU struct {
	UUID string `json:"uuid"`
}

// PreparedGpus represents a set of prepared GPUs on a node.
type PreparedGPUs struct {
	Devices []PreparedGPU `json:"devices"`
}

// PreparedDevices represents a set of prepared devices on a node.
type PreparedDevices struct {
	GPUs *PreparedGPUs `json:"gpus,omitempty"`
}
