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

// WebSockTunnelStatus defines the observed state of WebSockTunnel
type WebSockTunnelStatus struct {
}

// +kubebuilder:object:root=true

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
