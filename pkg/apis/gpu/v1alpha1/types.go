package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true

// GPUDeviceClassParameters holds the set of parameters provided when creating a resource class for this driver.
type GPUDeviceClassParameters struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec GPUDeviceClassParametersSpec `json:"spec,omitempty"`
}

// GPUDeviceClassParametersSpec is the spec for the GPUDeviceClassParametersSpec CRD.
type GPUDeviceClassParametersSpec struct {
	DeviceSelector []DeviceSelector `json:"deviceSelector,omitempty"`
}

// DeviceSelector allows one to match on a specific type of Device as part of the class.
type DeviceSelector struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GPUDeviceClassParametersList represents the "plural" of a DeviceClassParameters CRD object.
type GPUDeviceClassParametersList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []GPUDeviceClassParameters `json:"items"`
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
