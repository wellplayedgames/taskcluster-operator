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
	ClientID    string   `json:"clientId,omitempty"`
	AccessToken string   `json:"accessToken,omitempty"`
	Description string   `json:"description,omitempty"`
	Scopes      []string `json:"scopes,omitempty"`
}

// PulseSpec contains the pulse connection details.
type PulseSpec struct {
	AdminSecretRef *corev1.LocalObjectReference `json:"adminSecretRef,omitempty"`
	Host           string                       `json:"host,omitempty"`
	Vhost          string                       `json:"vhost,omitempty"`
}

// GitHubSpec contains the desired GitHub integration configuration.
type GitHubSpec struct {
	BotUsername string                       `json:"botUsername,omitempty"`
	SecretRef   *corev1.LocalObjectReference `json:"secretRef,omitempty"`
}

// InstanceIngressSpec contains the desired ingress configuration.
type InstanceIngressSpec struct {
	StaticIPName    string                       `json:"staticIpName,omitempty"`
	ExternalDNSName string                       `json:"externalDNSName,omitempty"`
	TLSSecretRef    *corev1.LocalObjectReference `json:"tlsSecretRef,omitempty"`
	IssuerRef       corev1.ObjectReference       `json:"issuerRef,omitempty"`
}

// InstanceSpec defines the desired state of Instance
type InstanceSpec struct {
	WebSockTunnelSecretRef          *corev1.LocalObjectReference `json:"webSockTunnelSecretRef,omitempty"`
	AWSSecretRef                    *corev1.LocalObjectReference `json:"awsSecretRef,omitempty"`
	AzureSecretRef                  *corev1.LocalObjectReference `json:"azureSecretRef,omitempty"`
	WorkerManagerProvidersSecretRef *corev1.LocalObjectReference `json:"workerManagerProvidersSecretRef,omitempty"`
	DatabaseRef                     *corev1.LocalObjectReference `json:"databaseRef,omitempty"`
	AuthSecretRef                   *corev1.LocalObjectReference `json:"authSecretRef,omitempty"`
	AccessTokensSecretRef           *corev1.LocalObjectReference `json:"accessTokensSecretRef,omitempty"`

	// Notifications backends
	IRCSecretRef    *corev1.LocalObjectReference `json:"ircSecretRef,omitempty"`
	MatrixSecretRef *corev1.LocalObjectReference `json:"matrixSecretRef,omitempty"`
	SlackSecretRef  *corev1.LocalObjectReference `json:"slackSecretRef,omitempty"`

	GitHub  GitHubSpec          `json:"github,omitempty"`
	Pulse   PulseSpec           `json:"pulse,omitempty"`
	Ingress InstanceIngressSpec `json:"ingress,omitempty"`

	RootURL                     string   `json:"rootUrl,omitempty"`
	ApplicationName             string   `json:"applicationName,omitempty"`
	BannerMessage               string   `json:"bannerMessage,omitempty"`
	EmailSourceAddress          string   `json:"emailSourceAddress,omitempty"`
	PublicArtifactBucket        string   `json:"publicArtifactBucket,omitempty"`
	PrivateArtifactBucket       string   `json:"privateArtifactBucket,omitempty"`
	ArtifactRegion              string   `json:"artifactRegion,omitempty"`
	AdditionalAllowedCORSOrigin string   `json:"additionalAllowedCorsOrigin,omitempty"`
	LoginStrategies             []string `json:"loginStrategies,omitempty"`
	AzureAccountID              string   `json:"azureAccountId,omitempty"`
	DockerImage                 string   `json:"dockerImage,omitempty"`
	PostgresUserPrefix          string   `json:"postgresUserPrefix,omitempty"`
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
