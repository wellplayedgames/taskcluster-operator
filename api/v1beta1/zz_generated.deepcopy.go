// +build !ignore_autogenerated

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

// Code generated by controller-gen. DO NOT EDIT.

package v1beta1

import (
	"k8s.io/api/core/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AccessToken) DeepCopyInto(out *AccessToken) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AccessToken.
func (in *AccessToken) DeepCopy() *AccessToken {
	if in == nil {
		return nil
	}
	out := new(AccessToken)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AccessToken) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AccessTokenList) DeepCopyInto(out *AccessTokenList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]AccessToken, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AccessTokenList.
func (in *AccessTokenList) DeepCopy() *AccessTokenList {
	if in == nil {
		return nil
	}
	out := new(AccessTokenList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AccessTokenList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AccessTokenSpec) DeepCopyInto(out *AccessTokenSpec) {
	*out = *in
	out.InstanceRef = in.InstanceRef
	if in.Scopes != nil {
		in, out := &in.Scopes, &out.Scopes
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AccessTokenSpec.
func (in *AccessTokenSpec) DeepCopy() *AccessTokenSpec {
	if in == nil {
		return nil
	}
	out := new(AccessTokenSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AccessTokenStatus) DeepCopyInto(out *AccessTokenStatus) {
	*out = *in
	if in.ObservedGeneration != nil {
		in, out := &in.ObservedGeneration, &out.ObservedGeneration
		*out = new(int64)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AccessTokenStatus.
func (in *AccessTokenStatus) DeepCopy() *AccessTokenStatus {
	if in == nil {
		return nil
	}
	out := new(AccessTokenStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GitHubSpec) DeepCopyInto(out *GitHubSpec) {
	*out = *in
	if in.SecretRef != nil {
		in, out := &in.SecretRef, &out.SecretRef
		*out = new(v1.LocalObjectReference)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GitHubSpec.
func (in *GitHubSpec) DeepCopy() *GitHubSpec {
	if in == nil {
		return nil
	}
	out := new(GitHubSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Instance) DeepCopyInto(out *Instance) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Instance.
func (in *Instance) DeepCopy() *Instance {
	if in == nil {
		return nil
	}
	out := new(Instance)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Instance) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *InstanceCondition) DeepCopyInto(out *InstanceCondition) {
	*out = *in
	in.LastTransitionTime.DeepCopyInto(&out.LastTransitionTime)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new InstanceCondition.
func (in *InstanceCondition) DeepCopy() *InstanceCondition {
	if in == nil {
		return nil
	}
	out := new(InstanceCondition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *InstanceIngressSpec) DeepCopyInto(out *InstanceIngressSpec) {
	*out = *in
	if in.TLSSecretRef != nil {
		in, out := &in.TLSSecretRef, &out.TLSSecretRef
		*out = new(v1.LocalObjectReference)
		**out = **in
	}
	out.IssuerRef = in.IssuerRef
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new InstanceIngressSpec.
func (in *InstanceIngressSpec) DeepCopy() *InstanceIngressSpec {
	if in == nil {
		return nil
	}
	out := new(InstanceIngressSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *InstanceList) DeepCopyInto(out *InstanceList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Instance, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new InstanceList.
func (in *InstanceList) DeepCopy() *InstanceList {
	if in == nil {
		return nil
	}
	out := new(InstanceList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *InstanceList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *InstanceSpec) DeepCopyInto(out *InstanceSpec) {
	*out = *in
	if in.WebSockTunnelSecretRef != nil {
		in, out := &in.WebSockTunnelSecretRef, &out.WebSockTunnelSecretRef
		*out = new(v1.LocalObjectReference)
		**out = **in
	}
	if in.AWSSecretRef != nil {
		in, out := &in.AWSSecretRef, &out.AWSSecretRef
		*out = new(v1.LocalObjectReference)
		**out = **in
	}
	if in.AzureSecretRef != nil {
		in, out := &in.AzureSecretRef, &out.AzureSecretRef
		*out = new(v1.LocalObjectReference)
		**out = **in
	}
	if in.WorkerManagerProvidersSecretRef != nil {
		in, out := &in.WorkerManagerProvidersSecretRef, &out.WorkerManagerProvidersSecretRef
		*out = new(v1.LocalObjectReference)
		**out = **in
	}
	if in.DatabaseRef != nil {
		in, out := &in.DatabaseRef, &out.DatabaseRef
		*out = new(v1.LocalObjectReference)
		**out = **in
	}
	if in.AuthSecretRef != nil {
		in, out := &in.AuthSecretRef, &out.AuthSecretRef
		*out = new(v1.LocalObjectReference)
		**out = **in
	}
	if in.AccessTokensSecretRef != nil {
		in, out := &in.AccessTokensSecretRef, &out.AccessTokensSecretRef
		*out = new(v1.LocalObjectReference)
		**out = **in
	}
	if in.IRCSecretRef != nil {
		in, out := &in.IRCSecretRef, &out.IRCSecretRef
		*out = new(v1.LocalObjectReference)
		**out = **in
	}
	if in.MatrixSecretRef != nil {
		in, out := &in.MatrixSecretRef, &out.MatrixSecretRef
		*out = new(v1.LocalObjectReference)
		**out = **in
	}
	if in.SlackSecretRef != nil {
		in, out := &in.SlackSecretRef, &out.SlackSecretRef
		*out = new(v1.LocalObjectReference)
		**out = **in
	}
	in.GitHub.DeepCopyInto(&out.GitHub)
	in.Pulse.DeepCopyInto(&out.Pulse)
	in.Ingress.DeepCopyInto(&out.Ingress)
	if in.LoginStrategies != nil {
		in, out := &in.LoginStrategies, &out.LoginStrategies
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new InstanceSpec.
func (in *InstanceSpec) DeepCopy() *InstanceSpec {
	if in == nil {
		return nil
	}
	out := new(InstanceSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *InstanceStatus) DeepCopyInto(out *InstanceStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]InstanceCondition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new InstanceStatus.
func (in *InstanceStatus) DeepCopy() *InstanceStatus {
	if in == nil {
		return nil
	}
	out := new(InstanceStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PulseSpec) DeepCopyInto(out *PulseSpec) {
	*out = *in
	if in.AdminSecretRef != nil {
		in, out := &in.AdminSecretRef, &out.AdminSecretRef
		*out = new(v1.LocalObjectReference)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PulseSpec.
func (in *PulseSpec) DeepCopy() *PulseSpec {
	if in == nil {
		return nil
	}
	out := new(PulseSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StaticAccessToken) DeepCopyInto(out *StaticAccessToken) {
	*out = *in
	if in.Scopes != nil {
		in, out := &in.Scopes, &out.Scopes
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StaticAccessToken.
func (in *StaticAccessToken) DeepCopy() *StaticAccessToken {
	if in == nil {
		return nil
	}
	out := new(StaticAccessToken)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WebSockTunnel) DeepCopyInto(out *WebSockTunnel) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WebSockTunnel.
func (in *WebSockTunnel) DeepCopy() *WebSockTunnel {
	if in == nil {
		return nil
	}
	out := new(WebSockTunnel)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *WebSockTunnel) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WebSockTunnelCondition) DeepCopyInto(out *WebSockTunnelCondition) {
	*out = *in
	in.LastTransitionTime.DeepCopyInto(&out.LastTransitionTime)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WebSockTunnelCondition.
func (in *WebSockTunnelCondition) DeepCopy() *WebSockTunnelCondition {
	if in == nil {
		return nil
	}
	out := new(WebSockTunnelCondition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WebSockTunnelList) DeepCopyInto(out *WebSockTunnelList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]WebSockTunnel, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WebSockTunnelList.
func (in *WebSockTunnelList) DeepCopy() *WebSockTunnelList {
	if in == nil {
		return nil
	}
	out := new(WebSockTunnelList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *WebSockTunnelList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WebSockTunnelSpec) DeepCopyInto(out *WebSockTunnelSpec) {
	*out = *in
	out.SecretRef = in.SecretRef
	out.CertificateIssuerRef = in.CertificateIssuerRef
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WebSockTunnelSpec.
func (in *WebSockTunnelSpec) DeepCopy() *WebSockTunnelSpec {
	if in == nil {
		return nil
	}
	out := new(WebSockTunnelSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WebSockTunnelStatus) DeepCopyInto(out *WebSockTunnelStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]WebSockTunnelCondition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WebSockTunnelStatus.
func (in *WebSockTunnelStatus) DeepCopy() *WebSockTunnelStatus {
	if in == nil {
		return nil
	}
	out := new(WebSockTunnelStatus)
	in.DeepCopyInto(out)
	return out
}
