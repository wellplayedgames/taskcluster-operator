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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// StaticAccessToken contains a taskcluster access token definition.
type StaticAccessToken struct {
	ClientID    string   `json:"clientId"`
	AccessToken string   `json:"accessToken"`
	Description string   `json:"description,omitempty"`
	Scopes      []string `json:"scopes,omitempty"`
}

// PulseSpec contains the pulse connection details.
type PulseSpec struct {
	AdminSecretRef *corev1.LocalObjectReference `json:"adminSecretRef"`
	Host           string                       `json:"host"`
	Vhost          string                       `json:"vhost"`
}

// GitHubSpec contains the desired GitHub integration configuration.
type GitHubSpec struct {
	BotUsername string                       `json:"botUsername"`
	SecretRef   *corev1.LocalObjectReference `json:"secretRef"`
}

// InstanceIngressSpec contains the desired ingress configuration.
type InstanceIngressSpec struct {
	StaticIPName    string                       `json:"staticIpName"`
	ExternalDNSName string                       `json:"externalDNSName"`
	TLSSecretRef    *corev1.LocalObjectReference `json:"tlsSecretRef"`
	IssuerRef       corev1.ObjectReference       `json:"issuerRef"`
}

// InstanceSpec defines the desired state of Instance
type InstanceSpec struct {
	WebSockTunnelSecretRef          *corev1.LocalObjectReference `json:"webSockTunnelSecretRef"`
	AWSSecretRef                    *corev1.LocalObjectReference `json:"awsSecretRef"`
	AzureSecretRef                  *corev1.LocalObjectReference `json:"azureSecretRef"`
	WorkerManagerProvidersSecretRef *corev1.LocalObjectReference `json:"workerManagerProvidersSecretRef"`
	DatabaseRef                     *corev1.LocalObjectReference `json:"databaseRef"`
	AuthSecretRef                   *corev1.LocalObjectReference `json:"authSecretRef"`
	AccessTokensSecretRef           *corev1.LocalObjectReference `json:"accessTokensSecretRef"`

	GitHub  GitHubSpec          `json:"github"`
	Pulse   PulseSpec           `json:"pulse"`
	Ingress InstanceIngressSpec `json:"ingress"`

	RootURL                     string   `json:"rootUrl"`
	ApplicationName             string   `json:"applicationName"`
	BannerMessage               string   `json:"bannerMessage"`
	EmailSourceAddress          string   `json:"emailSourceAddress"`
	PublicArtifactBucket        string   `json:"publicArtifactBucket"`
	PrivateArtifactBucket       string   `json:"privateArtifactBucket"`
	ArtifactRegion              string   `json:"artifactRegion"`
	AdditionalAllowedCORSOrigin string   `json:"additionalAllowedCorsOrigin"`
	LoginStrategies             []string `json:"loginStrategies"`
	AzureAccountID              string   `json:"azureAccountId"`
	DockerImage                 string   `json:"dockerImage"`
	PostgresUserPrefix          string   `json:"postgresUserPrefix"`
}

// InstanceConditionType represents the type enum of a condition.
type InstanceConditionType string

const (
	// InstanceProgressing is used when the instance is not blocked by an
	// external dependency or reconcile error.
	InstanceProgressing InstanceConditionType = "Progressing"
)

// InstanceCondition represents a condition of an Instance
type InstanceCondition struct {
	Type   InstanceConditionType  `json:"type"`
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

// InstanceStatus defines the observed state of Instance
type InstanceStatus struct {
	Conditions []InstanceCondition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Instance is the Schema for the instances API
type Instance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   InstanceSpec   `json:"spec,omitempty"`
	Status InstanceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// InstanceList contains a list of Instance
type InstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Instance `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Instance{}, &InstanceList{})
}
