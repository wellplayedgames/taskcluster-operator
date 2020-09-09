package v1beta1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SQLDatabaseSpec defines the desired state of a SQLDatabase
type SQLDatabaseSpec struct {
	InstanceRef *corev1.LocalObjectReference `json:"instanceRef,omitempty"`
}

// SQLDatabaseStatus defines the observed state of a SQLDatabase
type SQLDatabaseStatus struct {
}

// +kubebuilder:object:root=true

// SQLDatabase is the Schema for the instances API
type SQLDatabase struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SQLDatabaseSpec   `json:"spec,omitempty"`
	Status SQLDatabaseStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SQLDatabaseList contains a list of SQLDatabases
type SQLDatabaseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SQLDatabase `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SQLDatabase{}, &SQLDatabaseList{})
}
