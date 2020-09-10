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
)

const (
	image          = "taskcluster/websocktunnel"
	defaultVersion = "30.0.2"

	envoyConfig = `
admin:
  access_log_path: /dev/stdout
  address:
    socket_address: { address: 0.0.0.0, port_value: 9901 }

static_resources:
  secrets:
  - name: server_cert
    tls_certificate:
      certificate_chain:
        filename: /tls/tls.crt
      private_key:
        filename: /tls/tls.key
  listeners:
  - name: listener_0
    address:
      socket_address: { address: 0.0.0.0, port_value: 443 }
    filter_chains:
    - filters:
      - name: envoy.tcp_proxy
        config:
          stat_prefix: ingress_tcp
          cluster: service
          max_connect_attempts: 100
          access_log:
            - name: envoy.file_access_log
              config:
                path: /dev/stdout
      tls_context:
        common_tls_context:
          tls_certificate_sds_secret_configs:
          - name: server_cert
  clusters:
  - name: service
    connect_timeout: 120s
    type: STATIC
    dns_lookup_family: V4_ONLY
    lb_policy: ROUND_ROBIN
    load_assignment:
      cluster_name: service
      endpoints:
      - lb_endpoints:
        - endpoint:
            address:
              socket_address:
                address: 127.0.0.1
                port_value: 80
`
)

type WebSockTunnelBuilder struct {
	logr.Logger

	Source *taskclusterv1beta1.WebSockTunnel
}

func (b *WebSockTunnelBuilder) Build() ([]runtime.Object, error) {
	name := b.Source.Name
	namespace := b.Source.Namespace
	spec := &b.Source.Spec
	tlsSecretName := fmt.Sprintf("%s-tls", name)
	envoyConfigName := fmt.Sprintf("%s-envoy", name)

	objects := []runtime.Object{
		&certmanagerv1alpha2.Certificate{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
				Namespace: namespace,
			},
			Spec: certmanagerv1alpha2.CertificateSpec{
				SecretName: tlsSecretName,
				DNSNames:   []string{spec.DomainName},
				IssuerRef:  spec.CertificateIssuerRef,
			},
		},
		&corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name: envoyConfigName,
				Namespace: namespace,
			},
			Data: map[string]string{
				"config.yaml": envoyConfig,
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
										corev1.ResourceCPU: *resource.NewMilliQuantity(10, resource.BinarySI),
									},
								},
								ReadinessProbe: &corev1.Probe{
									Handler: corev1.Handler{
										HTTPGet: &corev1.HTTPGetAction{
											Path: "/__lbheartbeat__",
											Port: intstr.FromInt(80),
											Scheme: corev1.URISchemeHTTP,
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
											Scheme: corev1.URISchemeHTTP,
										},
									},
									InitialDelaySeconds: 30,
									PeriodSeconds:       3,
								},
							},
							{
								Name: "tls-terminate",
								Image: "envoyproxy/envoy:v1.11.1",
								Args: []string{"-c", "/etc/envoy/config.yaml"},
								VolumeMounts: []corev1.VolumeMount{
									{
										Name: "tls",
										MountPath: "/tls",
										ReadOnly: true,
									},
									{
										Name: "envoy-config",
										MountPath: "/etc/envoy",
										ReadOnly: true,
									},
								},
								Resources: corev1.ResourceRequirements{
									Requests: corev1.ResourceList{
										corev1.ResourceCPU: *resource.NewMilliQuantity(10, resource.BinarySI),
									},
								},
							},
						},
						Volumes: []corev1.Volume{
							{
								Name: "tls",
								VolumeSource: corev1.VolumeSource{
									Secret: &corev1.SecretVolumeSource{
										SecretName:  tlsSecretName,
									},
								},
							},
							{
								Name: "envoy-config",
								VolumeSource: corev1.VolumeSource{
									ConfigMap: &corev1.ConfigMapVolumeSource{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: envoyConfigName,
										},
									},
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
				Annotations: map[string]string{
					"external-dns.alpha.kubernetes.io/hostname": spec.DomainName,
				},
			},
			Spec: corev1.ServiceSpec{
				Type: corev1.ServiceTypeLoadBalancer,
				Selector: map[string]string{
					"app.kubernetes.io/component": "websocktunnel",
					"app.kubernetes.io/name":      name,
				},
				Ports: []corev1.ServicePort{
					{
						Name:       "https",
						Protocol:   corev1.ProtocolTCP,
						Port:       443,
						TargetPort: intstr.FromInt(443),
					},
				},
			},
		},
	}

	return objects, nil
}
