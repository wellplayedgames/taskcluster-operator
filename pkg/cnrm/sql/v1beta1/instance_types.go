package v1beta1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SecretSpec struct {
	Value     *string              `json:"value,omitempty"`
	ValueFrom *corev1.EnvVarSource `json:"valueFrom,omitempty"`
}

// SQLInstanceSpec defines the desired state of SQLInstance
type SQLInstanceSpec struct {
	RootPassword *SecretSpec `json:"rootPassword,omitempty"`
}

// SQLInstanceStatus defines the observed state of Instance
type SQLInstanceStatus struct {
	PublicIPAddress  string `json:"publicIpAddress,omitempty"`
	PrivateIPAddress string `json:"privateIpAddress,omitempty"`
}

// +kubebuilder:object:root=true

// SQLInstance is the Schema for the instances API
type SQLInstance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SQLInstanceSpec   `json:"spec,omitempty"`
	Status SQLInstanceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SQLInstanceList contains a list of Instance
type SQLInstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SQLInstance `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SQLInstance{}, &SQLInstanceList{})
}
