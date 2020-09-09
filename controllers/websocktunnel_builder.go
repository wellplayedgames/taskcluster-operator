package controllers

import (
	"fmt"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"

	certmanagerv1alpha2 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	taskclusterv1beta1 "github.com/wellplayedgames/taskcluster-operator/api/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
)

const (
	image          = "taskcluster/websocktunnel"
	defaultVersion = "30.0.2"
)

type WebSockTunnelBuilder struct {
	logr.Logger

	Source *taskclusterv1beta1.WebSockTunnel
}

func (b *WebSockTunnelBuilder) Build() ([]runtime.Object, error) {
	name := b.Source.Name
	namespace := b.Source.Namespace
	spec := &b.Source.Spec

	objects := []runtime.Object{
		&certmanagerv1alpha2.Certificate{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
				Namespace: namespace,
			},
			Spec: certmanagerv1alpha2.CertificateSpec{
				SecretName: fmt.Sprintf("%s-tls", name),
				DNSNames:   []string{spec.DomainName},
				IssuerRef:  spec.CertificateIssuerRef,
			},
		},
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
				Namespace: namespace,
			},
			Spec: appsv1.DeploymentSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app.kubernetes.io/component": "websocktunnel",
						"app.kubernetes.io/name":      name,
					},
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"app.kubernetes.io/component": "websocktunnel",
							"app.kubernetes.io/name":      name,
						},
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "websocktunnel",
								Image: fmt.Sprintf("%s:%s", image, defaultVersion),
								Env: []corev1.EnvVar{
									{Name: "ENV", Value: "production"},
									{Name: "URL_PREFIX", Value: fmt.Sprintf("https://%s", spec.DomainName)},
									{Name: "AUDIENCE", Value: "taskcluster"},
									{
										Name: "TASKCLUSTER_PROXY_SECRET_A",
										ValueFrom: &corev1.EnvVarSource{
											SecretKeyRef: &corev1.SecretKeySelector{
												LocalObjectReference: spec.SecretRef,
												Key:                  keySecret,
											},
										},
									},
									{
										Name: "TASKCLUSTER_PROXY_SECRET_B",
										ValueFrom: &corev1.EnvVarSource{
											SecretKeyRef: &corev1.SecretKeySelector{
												LocalObjectReference: spec.SecretRef,
												Key:                  keySecretLast,
											},
										},
									},
								},
								Resources: corev1.ResourceRequirements{
									Requests: corev1.ResourceList{
										"cpu": *resource.NewMilliQuantity(10, resource.BinarySI),
									},
								},
								ReadinessProbe: &corev1.Probe{
									Handler: corev1.Handler{
										HTTPGet: &corev1.HTTPGetAction{
											Path: "/__lbheartbeat__",
											Port: intstr.FromInt(80),
										},
									},
									InitialDelaySeconds: 3,
									PeriodSeconds:       3,
								},
								LivenessProbe: &corev1.Probe{
									Handler: corev1.Handler{
										HTTPGet: &corev1.HTTPGetAction{
											Path: "/__lbheartbeat__",
											Port: intstr.FromInt(80),
										},
									},
									InitialDelaySeconds: 30,
									PeriodSeconds:       3,
								},
							},
						},
					},
				},
			},
		},
		&corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
				Namespace: namespace,
				Labels: map[string]string{
					"app.kubernetes.io/component": "websocktunnel",
					"app.kubernetes.io/name":      name,
				},
			},
			Spec: corev1.ServiceSpec{
				Type: corev1.ServiceTypeNodePort,
				Selector: map[string]string{
					"app.kubernetes.io/component": "websocktunnel",
					"app.kubernetes.io/name":      name,
				},
				Ports: []corev1.ServicePort{
					{
						Name:       "http",
						Protocol:   corev1.ProtocolTCP,
						Port:       80,
						TargetPort: intstr.FromInt(80),
					},
				},
			},
		},
		&networkingv1beta1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
				Namespace: namespace,
				Annotations: map[string]string{
					"kubernetes.io/ingress.allow-http": "false",
				},
			},
			Spec: networkingv1beta1.IngressSpec{
				TLS: []networkingv1beta1.IngressTLS{
					{
						Hosts:      []string{spec.DomainName},
						SecretName: fmt.Sprintf("%s-tls", name),
					},
				},
				Rules: []networkingv1beta1.IngressRule{
					{
						Host: spec.DomainName,
						IngressRuleValue: networkingv1beta1.IngressRuleValue{
							HTTP: &networkingv1beta1.HTTPIngressRuleValue{
								Paths: []networkingv1beta1.HTTPIngressPath{
									{
										Path: "/*",
										Backend: networkingv1beta1.IngressBackend{
											ServiceName: name,
											ServicePort: intstr.FromInt(80),
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	return objects, nil
}
