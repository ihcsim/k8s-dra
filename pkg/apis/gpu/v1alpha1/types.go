package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true

// DeviceClassParameters holds the set of parameters provided when creating a resource class for this driver.
type DeviceClassParameters struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec DeviceClassParametersSpec `json:"spec,omitempty"`
}

// DeviceClassParametersSpec is the spec for the DeviceClassParametersSpec CRD.
type DeviceClassParametersSpec struct {
	DeviceSelector []DeviceSelector `json:"deviceSelector,omitempty"`
}

// DeviceSelector allows one to match on a specific type of Device as part of the class.
type DeviceSelector struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DeviceClassParametersList represents the "plural" of a DeviceClassParameters CRD object.
type DeviceClassParametersList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []DeviceClassParameters `json:"items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
// +kubebuilder:resource:scope=Namespaced

// ResourceClaimParameters holds the set of parameters provided when creating a resource claim for a GPU.
type ResourceClaimParameters struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ResourceClaimParametersSpec `json:"spec,omitempty"`
}

// ResourceClaimParametersSpec is the spec for the ResourceClaimParameters CRD.
type ResourceClaimParametersSpec struct {
	Count int `json:"count,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ResourceClaimParametersList represents the "plural" of a ResourceClaimParameters CRD object.
type ResourceClaimParametersList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []ResourceClaimParameters `json:"items"`
}