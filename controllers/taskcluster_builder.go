package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/wellplayedgames/tiny-operator/pkg/helm"
	"helm.sh/helm/v3/pkg/chart/loader"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"strings"

	taskclusterv1beta1 "github.com/wellplayedgames/taskcluster-operator/api/v1beta1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	batchv2alpha1 "k8s.io/api/batch/v2alpha1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
)

type PostgresAccess struct {
	ReadDBURL  string `json:"read_db_url"`
	WriteDBURL string `json:"write_db_url"`
}

type PulseAccess struct {
	PulseUsername string `json:"pulse_username"`
	PulsePassword string `json:"pulse_password"`
}

type DBCryptoKey struct {
	ID        string `json:"id"`
	Algorithm string `json:"algo"`
	Key       string `json:"key"`
}

type CryptoConfig struct {
	AzureCryptoKey string        `json:"azure_crypto_key,omitempty"`
	DBCryptoKeys   []DBCryptoKey `json:"db_crypto_keys"`
}

type AWSAccess struct {
	AccessKeyID     string `json:"aws_access_key_id"`
	SecretAccessKey string `json:"aws_secret_access_key"`
}

type TaskClusterAccess struct {
	AccessToken string `json:"taskcluster_access_token"`
}

type AuthConfig struct {
	PostgresAccess
	PulseAccess
	CryptoConfig
	AzureAccounts       map[string]string                      `json:"azure_accounts"`
	StaticAccounts      []taskclusterv1beta1.StaticAccessToken `json:"static_clients"`
	WebSockTunnelSecret string                                 `json:"websocktunnel_secret"`
}

type BuiltInWorkersConfig struct {
	TaskClusterAccess
}

type GitHubConfig struct {
	TaskClusterAccess
	PostgresAccess
	PulseAccess
	BotUsername      string `json:"bot_username"`
	GitHubPrivatePEM string `json:"github_private_pem"`
	GitHubAppID      string `json:"github_app_id"`
	WebhookSecret    string `json:"webhook_secret"`
}

type HooksConfig struct {
	TaskClusterAccess
	PostgresAccess
	PulseAccess
	CryptoConfig
}

type IndexConfig struct {
	TaskClusterAccess
	PostgresAccess
	PulseAccess
}

type IRCConfig struct {
	Debug    bool   `json:"irc_debug,omitempty"`
	Nick     string `json:"irc_nick,omitempty"`
	Password string `json:"irc_password,omitempty"`
	Port     string `json:"irc_port,omitempty"`
	RealName string `json:"irc_real_name,omitempty"`
	Server   string `json:"irc_server,omitempty"`
	UserName string `json:"irc_user_name,omitempty"`
}

type MatrixConfig struct {
	AccessToken string `json:"matrix_access_token,omitempty"`
	BaseURL     string `json:"matrix_base_url,omitempty"`
	UserID      string `json:"matrix_user_id,omitempty"`
}

type SlackConfig struct {
	APIURL      string `json:"slack_api_url,omitempty"`
	AccessToken string `json:"slack_access_token,omitempty"`
}

type NotifyConfig struct {
	TaskClusterAccess
	PostgresAccess
	PulseAccess
	AWSAccess
	IRCConfig
	MatrixConfig
	SlackConfig
	EmailSourceAddress string `json:"email_source_address"`
}

type PurgeCacheConfig struct {
	TaskClusterAccess
	PostgresAccess
}

type QueueConfig struct {
	TaskClusterAccess
	PostgresAccess
	PulseAccess
	AWSAccess
	PublicArtifactBucket   string `json:"public_artifact_bucket"`
	PrivateArtifactBucket  string `json:"private_artifact_bucket"`
	SignPublicArtifactURLs bool   `json:"sign_public_artifact_urls"`
	ArtifactRegion         string `json:"artifact_region"`
}

type SecretsConfig struct {
	TaskClusterAccess
	PostgresAccess
	CryptoConfig
}

type GitHubLoginStrategy struct {
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
}

type UILoginStrategiesConfig struct {
	GitHub *GitHubLoginStrategy `json:"github"`
}

type WebServerConfig struct {
	TaskClusterAccess
	PostgresAccess
	PulseAccess
	CryptoConfig
	PublicURL                   string                  `json:"public_url"`
	AdditionalAllowedCORSOrigin string                  `json:"additional_allowed_cors_origin"`
	UILoginStrategies           UILoginStrategiesConfig `json:"ui_login_strategies"`
	SessionSecret               string                  `json:"session_secret"`
	RegisteredClients           []string                `json:"registered_clients"`
}

type WorkerManagerConfig struct {
	TaskClusterAccess
	PostgresAccess
	PulseAccess
	CryptoConfig
	Providers map[string]json.RawMessage `json:"providers"`
}

type UIConfig struct {
	GraphQLSubscriptionEndpoint string `json:"graphql_subscription_endpoint"`
	GraphQLEndpoint             string `json:"graphql_endpoint"`
	BannerMessage               string `json:"banner_message"`
	UILoginStrategyNames        string `json:"ui_login_strategy_names"`
}

type TaskClusterValues struct {
	Auth           AuthConfig           `json:"auth"`
	BuiltInWorkers BuiltInWorkersConfig `json:"built_in_workers"`
	GitHub         GitHubConfig         `json:"github"`
	Hooks          HooksConfig          `json:"hooks"`
	Index          IndexConfig          `json:"index"`
	Notify         NotifyConfig         `json:"notify"`
	PurgeCache     PurgeCacheConfig     `json:"purge_cache"`
	Queue          QueueConfig          `json:"queue"`
	Secrets        SecretsConfig        `json:"secrets"`
	WebServer      WebServerConfig      `json:"web_server"`
	WorkerManager  WorkerManagerConfig  `json:"worker_manager"`
	UI             UIConfig             `json:"ui"`

	RootURL             string `json:"rootUrl"`
	ApplicationName     string `json:"applicationName"`
	IngressStaticIPName string `json:"ingressStaticIpName"`
	IngressSecretName   string `json:"ingressSecretName"`
	IngressExternalDNS  string `json:"ingressExternalDNS"`
	PulseHostname       string `json:"pulseHostname"`
	PulseVHost          string `json:"pulseVhost"`
	DockerImage         string `json:"dockerImage"`
	AzureAccountID      string `json:"azureAccountId"`
}

func (o *TaskClusterOperations) renderChart(values *TaskClusterValues) ([]runtime.Object, error) {
	chrt, err := loader.LoadDir(o.ChartPath)
	if err != nil {
		return nil, err
	}

	js, err := json.Marshal(values)
	if err != nil {
		return nil, err
	}

	rawValues := map[string]interface{}{}
	err = json.Unmarshal(js, &rawValues)
	if err != nil {
		return nil, err
	}

	objects, err := helm.RenderChart(o.Scheme, chrt, rawValues, o.source.Namespace)
	if err != nil {
		return nil, err
	}

	for _, obj := range objects {
		acc, err := meta.Accessor(obj)
		if err != nil {
			return nil, err
		}

		acc.SetNamespace(o.source.Namespace)

		// Ensure that protocol is filled out.
		if d, ok := obj.(*appsv1.Deployment); ok {
			podSpec := &d.Spec.Template.Spec

			for idx := range podSpec.Containers {
				c := &podSpec.Containers[idx]
				for portIdx := range c.Ports {
					p := &c.Ports[portIdx]
					if p.Protocol == "" {
						p.Protocol = corev1.ProtocolTCP
					}
				}
			}

		}
	}

	return objects, nil
}

func (o *TaskClusterOperations) createDBUpgradeJob() []runtime.Object {
	secretName := fmt.Sprintf("%s-db-admin", o.source.Name)
	dbUrl := o.dbInfo.ConnectionString(true, true)

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: o.source.Namespace,
			Name:      secretName,
		},
		Data: map[string][]byte{
			"ADMIN_DB_URL": []byte(dbUrl),
		},
	}

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: o.source.Namespace,
			Name:      fmt.Sprintf("%s-db-upgrade", o.source.Name),
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"hive.wellplayed.games/enabled": "false",
					},
				},
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyNever,
					Containers: []corev1.Container{
						{
							Name:  "db-upgrade",
							Image: o.dockerImage(),
							Args:  []string{"script/db:upgrade"},
							EnvFrom: []corev1.EnvFromSource{
								{
									SecretRef: &corev1.SecretEnvSource{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: secretName,
										},
									},
								},
							},
							Env: []corev1.EnvVar{
								{Name: "NODE_ENV", Value: "production"},
								{Name: "USERNAME_PREFIX", Value: o.source.Spec.PostgresUserPrefix},
							},
						},
					},
				},
			},
		},
	}

	objects := []runtime.Object{
		secret,
		job,
	}
	return objects
}

func (o *TaskClusterOperations) patchResources(objects []runtime.Object) {
	// Make CronJobs replace.
	for _, obj := range objects {
		if job, ok := obj.(*batchv1beta1.CronJob); ok {
			meta := &job.Spec.JobTemplate.Spec.Template.ObjectMeta
			if meta.Annotations == nil {
				meta.Annotations = map[string]string{}
			}

			meta.Annotations["hive.wellplayed.games/enabled"] = "false"
		}

		if job, ok := obj.(*batchv2alpha1.CronJob); ok {
			meta := &job.Spec.JobTemplate.Spec.Template.ObjectMeta
			if meta.Annotations == nil {
				meta.Annotations = map[string]string{}
			}

			meta.Annotations["hive.wellplayed.games/enabled"] = "false"
		}

		if deployment, ok := obj.(*appsv1.Deployment); ok {
			if o.source.Spec.IRCSecretRef == nil && deployment.Name == "taskcluster-notify-irc" {
				numReplicas := int32(0)
				deployment.Spec.Replicas = &numReplicas
			}
		}

		// Enable cert-manager on ingress.
		gk := obj.GetObjectKind().GroupVersionKind().GroupKind()
		extGk := schema.GroupKind{Group: extensionsv1beta1.GroupName, Kind: "Ingress"}
		appsGk := schema.GroupKind{Group: appsv1.GroupName, Kind: "Ingress"}

		if gk == extGk || gk == appsGk {
			acc, err := meta.Accessor(obj)
			if err != nil {
				panic(err)
			}

			annotations := acc.GetAnnotations()
			if annotations == nil {
				annotations = map[string]string{}
			}

			issuerRef := &o.source.Spec.Ingress.IssuerRef

			if issuerRef.Kind == "ClusterIssuer" {
				annotations["cert-manager.io/cluster-issuer"] = issuerRef.Name
			} else if issuerRef.Kind == "Issuer" {
				annotations["cert-manager.io/issuer"] = issuerRef.Name
			}

			// Remove the pre-shared-cert annotation, since we want to use Kubernetes secrets
			// and not GCP SSL certificates.
			delete(annotations, "ingress.gcp.kubernetes.io/pre-shared-cert")
			acc.SetAnnotations(annotations)
		}

		// Set Ingress TLS secret natively using ingress fields.
		tlsSecretRef := o.source.Spec.Ingress.TLSSecretRef
		if tlsSecretRef != nil {
			domain := o.source.Spec.RootURL
			domain = strings.Replace(domain, "https://", "", -1)
			domain = strings.Replace(domain, "http://", "", -1)

			if ingress, ok := obj.(*extensionsv1beta1.Ingress); ok {
				ingress.Spec.TLS = []extensionsv1beta1.IngressTLS{
					{
						Hosts:      []string{domain},
						SecretName: tlsSecretRef.Name,
					},
				}
			}
		}
	}
}

func (o *TaskClusterOperations) Build(ctx context.Context) ([]runtime.Object, error) {
	values, err := o.RenderValues(ctx)
	if err != nil {
		return nil, err
	}

	objects, err := o.renderChart(values)
	if err != nil {
		return nil, err
	}

	objects = append(objects, o.createDBUpgradeJob()...)

	o.patchResources(objects)
	return objects, nil
}
