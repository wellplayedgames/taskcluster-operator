/*
Copyright 2020 Well Played Games Ltd.

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

package v1beta1

import (
	cmmeta "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// WebSockTunnelSpec defines the desired state of WebSockTunnel
type WebSockTunnelSpec struct {
	DomainName string `json:"domainName"`

	SecretRef            corev1.LocalObjectReference `json:"secretRef"`
	CertificateIssuerRef cmmeta.ObjectReference      `json:"certificateIssuerRef"`

	// TODO: Add rotation schedule here?
}

// WebSockTunnelConditionType represents the type enum of a condition.
type WebSockTunnelConditionType string

const (
	// WebSockTunnelProgressing is used when the instance is not blocked by an
	// external dependency or reconcile error.
	WebSockTunnelProgressing WebSockTunnelConditionType = "Progressing"
)

// WebSockTunnelCondition represents a condition of an Instance
type WebSockTunnelCondition struct {
	Type   WebSockTunnelConditionType  `json:"type"`
	Status corev1.ConditionStatus `json:"status"`
	// Last time the condition transitioned from one status to another.
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
	// Unique, this should be a short, machine understandable string that gives the reason
	// for condition's last transition. If it reports "ResizeStarted" that means the underlying
	// persistent volume is being resized.
	// +optional
	Reason string `json:"reason,omitempty"`
	// Human-readable message indicating details about last transition.
	// +optional
	Message string `json:"message,omitempty"`
}

// WebSockTunnelStatus defines the observed state of WebSockTunnel
type WebSockTunnelStatus struct {
	Conditions []WebSockTunnelCondition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// WebSockTunnel is the Schema for the websocktunnels API
type WebSockTunnel struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WebSockTunnelSpec   `json:"spec,omitempty"`
	Status WebSockTunnelStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// WebSockTunnelList contains a list of WebSockTunnel
type WebSockTunnelList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []WebSockTunnel `json:"items"`
}

func init() {
	SchemeBuilder.Register(&WebSockTunnel{}, &WebSockTunnelList{})
}
