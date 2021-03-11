package controllers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/jackc/pgx/v4"
	rabbithole "github.com/michaelklishin/rabbit-hole"
	taskclusterv1beta1 "github.com/wellplayedgames/taskcluster-operator/api/v1beta1"
	sqlv1beta1 "github.com/wellplayedgames/taskcluster-operator/pkg/cnrm/sql/v1beta1"
	"github.com/wellplayedgames/taskcluster-operator/pkg/pwgen"
	"github.com/wellplayedgames/tiny-operator/pkg/errors"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"strconv"
	"strings"
	"time"
)

const (
	defaultDockerRepo = "taskcluster/taskcluster"
	defaultVersion    = "42.1.1"
	stateKey          = "state"
	fieldOwner        = "taskcluster.wellplayed.games"
	hashAnnotation    = fieldOwner + "/hash"
)

var (
	postgresServices = []string{
		"",
		"github",
		"auth",
		"hooks",
		"index",
		"notify",
		"object",
		"purge_cache",
		"queue",
		"secrets",
		"web_server",
		"worker_manager",
	}

	pulseServices = []string{
		"auth",
		"github",
		"hooks",
		"index",
		"notify",
		"queue",
		"web_server",
		"worker_manager",
	}

	accessTokenServices = []string{
		"built_in_workers",
		"github",
		"hooks",
		"index",
		"notify",
		"object",
		"purge_cache",
		"queue",
		"root",
		"secrets",
		"web_server",
		"worker_manager",
	}

	cryptoServices = []string{
		"auth",
		"hooks",
		"object",
		"secrets",
		"web_server",
		"worker_manager",
	}
)

type ServiceAccount struct {
	AccessToken      string `json:"accessToken,omitempty"`
	PostgresPassword string `json:"postgresPassword,omitempty"`
	PulsePassword    string `json:"pulsePassword,omitempty"`
	CryptoConfig
}

type TaskClusterState struct {
	ServiceAccounts map[string]*ServiceAccount `json:"serviceAccounts,omitempty"`
	SessionSecret   string                     `json:"sessionSecret,omitempty"`
}

type PostgresDatabase struct {
	PublicIP  string `json:"publicIp"`
	PrivateIP string `json:"privateIp"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	Database  string `json:"database"`
}

// ConnectionString returns an admin Postgres connection string for this database.
func (d *PostgresDatabase) ConnectionString(public, noVerify bool) string {
	ip := d.PrivateIP
	if public {
		ip = d.PublicIP
	}

	params := "sslmode=require"
	if noVerify {
		// sslmode=verify is off since we have to pass in a CA certificate since v38
		params = "ssl=1&sslmode=no-verify"
	}

	return fmt.Sprintf("postgres://%s:%s@%s/%s?%s", d.Username, d.Password, ip, d.Database, params)
}

type TaskClusterOperations struct {
	logr.Logger
	client.Client
	Scheme *runtime.Scheme
	types.NamespacedName

	ChartPath    string
	UsePublicIPs bool

	source taskclusterv1beta1.Instance
	state  TaskClusterState
	dbInfo PostgresDatabase
	db     *pgx.Conn
	pulse  *rabbithole.Client

	dbUpgradeHash string
	dbUpgradeJob  *batchv1.Job

	accessTokenObjects []taskclusterv1beta1.StaticAccessToken
}

func (o *TaskClusterOperations) Prepare(ctx context.Context) error {
	if err := o.Client.Get(ctx, o.NamespacedName, &o.source); err != nil {
		return err
	}

	if err := o.readState(ctx); err != nil {
		return err
	}

	return nil
}

func (o *TaskClusterOperations) Close(ctx context.Context) error {
	var err error

	if db := o.db; db != nil {
		o.db = nil
		err = errors.Append(err, o.db.Close(ctx))
	}

	o.pulse = nil
	return nil
}

func (o *TaskClusterOperations) connectToPostgres(ctx context.Context) (*pgx.Conn, error) {
	if o.db != nil {
		return o.db, nil
	}

	dbRef := o.source.Spec.DatabaseRef
	if dbRef == nil {
		return nil, fmt.Errorf("no database ref specified")
	}

	ctx2, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	dbName := types.NamespacedName{
		Namespace: o.Namespace,
		Name:      dbRef.Name,
	}

	dbInfo, err := o.fetchPostgresDatabase(ctx2, dbName)
	if err != nil {
		return nil, err
	}

	// Ensure database exists
	{
		dbInfoWithoutDB := dbInfo
		dbInfoWithoutDB.Database = ""
		conn, err := pgx.Connect(ctx2, dbInfoWithoutDB.ConnectionString(o.UsePublicIPs, false))
		if err != nil {
			return nil, err
		}

		dbName := pgx.Identifier{dbInfo.Database}.Sanitize()

		rows, err := conn.Query(ctx2, "SELECT 1 FROM pg_database WHERE datname = $1", pgx.QuerySimpleProtocol(true), dbInfo.Database)
		if err != nil {
			return nil, fmt.Errorf("error checking database: %w", err)
		}

		defer conn.Close(ctx)
		defer rows.Close()

		if !rows.Next() {
			if _, err := conn.Exec(ctx2, "CREATE DATABASE "+dbName); err != nil {
				return nil, fmt.Errorf("error creating database: %w", err)
			}
		}
	}

	conn, err := pgx.Connect(ctx2, dbInfo.ConnectionString(o.UsePublicIPs, false))
	if err != nil {
		return nil, err
	}

	o.dbInfo = dbInfo
	o.db = conn
	return conn, nil
}

func (o *TaskClusterOperations) fetchPostgresDatabase(ctx context.Context, name types.NamespacedName) (PostgresDatabase, error) {
	var database sqlv1beta1.SQLDatabase
	if err := o.Client.Get(ctx, name, &database); err != nil {
		return PostgresDatabase{}, err
	}

	if database.Spec.InstanceRef == nil {
		return PostgresDatabase{}, fmt.Errorf("SQL database missing instanceRef")
	}

	instanceName := types.NamespacedName{
		Name:      database.Spec.InstanceRef.Name,
		Namespace: name.Namespace,
	}

	var instance sqlv1beta1.SQLInstance
	if err := o.Client.Get(ctx, instanceName, &instance); err != nil {
		return PostgresDatabase{}, nil
	}

	rootPasswordObj := instance.Spec.RootPassword
	if rootPasswordObj == nil {
		return PostgresDatabase{}, fmt.Errorf("SQL instance root password not specified")
	}

	rootPassword := ""
	if rootPasswordObj.Value != nil {
		rootPassword = *rootPasswordObj.Value
	} else if rootPasswordObj.ValueFrom == nil {
		return PostgresDatabase{}, fmt.Errorf("SQL instance password must set value or valueFrom")
	} else if rootPasswordObj.ValueFrom.SecretKeyRef != nil {
		secretName := types.NamespacedName{
			Namespace: name.Namespace,
			Name:      rootPasswordObj.ValueFrom.SecretKeyRef.Name,
		}

		var dbSecret corev1.Secret
		if err := o.Client.Get(ctx, secretName, &dbSecret); err != nil {
			return PostgresDatabase{}, err
		}

		rootPassword = (string)(dbSecret.Data[rootPasswordObj.ValueFrom.SecretKeyRef.Key])
	} else {
		return PostgresDatabase{}, fmt.Errorf("only secrets are supported for valueFrom")
	}

	publicIp := instance.Status.PublicIPAddress
	privateIp := instance.Status.PrivateIPAddress
	if publicIp == "" || privateIp == "" {
		return PostgresDatabase{}, fmt.Errorf("SQL instance has no IP address")
	}

	return PostgresDatabase{
		PublicIP:  publicIp,
		PrivateIP: privateIp,
		Username:  "postgres",
		Password:  rootPassword,
		Database:  name.Name,
	}, nil
}

func (o *TaskClusterOperations) connectToPulse(ctx context.Context) (*rabbithole.Client, error) {
	username := "guest"
	password := "guest"

	if pulseSecretRef := o.source.Spec.Pulse.AdminSecretRef; pulseSecretRef != nil {
		pulseSecretName := types.NamespacedName{
			Namespace: o.Namespace,
			Name:      pulseSecretRef.Name,
		}

		var secret corev1.Secret
		if err := o.Client.Get(ctx, pulseSecretName, &secret); err != nil {
			return nil, err
		}

		if secret.Data == nil {
			secret.Data = map[string][]byte{}
		}

		if s := (string)(secret.Data["admin-username"]); s != "" {
			username = s
		}

		if s := (string)(secret.Data["admin-password"]); s != "" {
			password = s
		}
	}

	// This is hardcoded to HTTPS because TaskCluster requires HTTPS anyway.
	endpoint := fmt.Sprintf("https://%s", o.source.Spec.Pulse.Host)
	client, err := rabbithole.NewClient(endpoint, username, password)
	if err != nil {
		return nil, err
	}

	o.pulse = client
	return client, nil
}

func (o *TaskClusterOperations) readState(ctx context.Context) error {
	name := o.NamespacedName
	name.Name = fmt.Sprintf("%s-state", name.Name)

	var secret corev1.Secret
	if err := client.IgnoreNotFound(o.Client.Get(ctx, name, &secret)); err != nil {
		return err
	}

	rawState, ok := secret.Data[stateKey]
	if !ok {
		return nil
	}

	return json.Unmarshal(rawState, &o.state)
}

func (o *TaskClusterOperations) writeState(ctx context.Context) error {
	rawState, err := json.Marshal(&o.state)
	if err != nil {
		return err
	}

	// Do not set the controller of this secret, as if the instance gets
	// deleted, the DB will become inaccessible.
	var secret corev1.Secret
	secret.TypeMeta.SetGroupVersionKind(corev1.SchemeGroupVersion.WithKind("Secret"))
	secret.Namespace = o.Namespace
	secret.Name = fmt.Sprintf("%s-state", o.Name)
	secret.Data = map[string][]byte{
		stateKey: rawState,
	}

	return o.Client.Patch(ctx, &secret, client.Apply, client.ForceOwnership, client.FieldOwner(fieldOwner))
}

func (o *TaskClusterOperations) migrateAccessTokenResources(ctx context.Context) error {
	instanceName := types.NamespacedName{
		Namespace: o.source.Namespace,
		Name:      o.source.Name,
	}

	var accessTokens taskclusterv1beta1.AccessTokenList
	if err := o.Client.List(ctx, &accessTokens, client.MatchingFields{fieldInstanceRef: instanceName.String()}); err != nil {
		return err
	}

	for idx := range accessTokens.Items {
		accessToken := &accessTokens.Items[idx]
		accessTokenName := types.NamespacedName{
			Namespace: accessToken.Namespace,
			Name:      accessToken.Name,
		}

		needsUpdate := false
		needsCreate := false

		var secret corev1.Secret
		err := o.Client.Get(ctx, accessTokenName, &secret)
		if apierrors.IsNotFound(err) {
			needsUpdate = true
			needsCreate = true

			secret.Namespace = accessToken.Namespace
			secret.Name = accessToken.Name
			secret.Data = map[string][]byte{}

			if err := controllerutil.SetControllerReference(accessToken, &secret, o.Scheme); err != nil {
				return err
			}
		} else if err != nil {
			return err
		}

		oldClientID := string(secret.Data["client-id"])
		accessTokenStr := string(secret.Data["access-token"])

		if !needsUpdate && (oldClientID != accessToken.Spec.ClientID) {
			needsUpdate = true
		}

		if needsUpdate {
			accessTokenStr = pwgen.AlphaNumeric(30)
			secret.Data["client-id"] = []byte(accessToken.Spec.ClientID)
			secret.Data["access-token"] = []byte(accessTokenStr)

			if needsCreate {
				if err := o.Client.Create(ctx, &secret); err != nil {
					return err
				}
			} else {
				if err := o.Client.Update(ctx, &secret); err != nil {
					return err
				}
			}
		}

		if !accessToken.Status.Created ||
			accessToken.Status.ObservedGeneration == nil ||
			*accessToken.Status.ObservedGeneration != accessToken.Generation {
			accessToken.Status.Created = true
			accessToken.Status.ObservedGeneration = &accessToken.Generation

			if err := o.Client.Status().Update(ctx, accessToken); err != nil {
				return err
			}
		}

		o.accessTokenObjects = append(o.accessTokenObjects, taskclusterv1beta1.StaticAccessToken{
			ClientID:    accessToken.Spec.ClientID,
			AccessToken: accessTokenStr,
			Description: accessToken.Spec.Description,
			Scopes:      accessToken.Spec.Scopes,
		})
	}

	return nil
}

func (o *TaskClusterOperations) MigrateState(ctx context.Context) error {
	// Ensure VHost
	pulse, err := o.connectToPulse(ctx)
	if err != nil {
		return err
	}

	_, err = pulse.PutVhost(o.source.Spec.Pulse.Vhost, rabbithole.VhostSettings{
		Tracing: false,
	})
	if err != nil {
		return err
	}

	if o.state.ServiceAccounts == nil {
		o.state.ServiceAccounts = map[string]*ServiceAccount{}
	}

	if o.state.SessionSecret == "" {
		o.state.SessionSecret = pwgen.AlphaNumeric(20)
	}

	for _, svc := range postgresServices {
		if err := o.ensurePostgresAccess(ctx, svc); err != nil {
			return err
		}
	}

	for _, svc := range pulseServices {
		if err := o.ensurePulseAccess(ctx, svc); err != nil {
			return err
		}
	}

	for _, svc := range cryptoServices {
		o.ensureCrypto(svc)
	}

	for _, svc := range accessTokenServices {
		o.ensureAccessToken(svc)
	}

	if err := o.migrateAccessTokenResources(ctx); err != nil {
		return err
	}

	return o.writeState(ctx)
}

func (o *TaskClusterOperations) dockerImage() string {
	src := o.source.Spec.DockerImage

	if src == "" {
		return fmt.Sprintf("%s:v%s", defaultDockerRepo, defaultVersion)
	} else if !strings.Contains(src, ":") {
		return fmt.Sprintf("%s:v%s", src, defaultVersion)
	} else {
		return src
	}
}

func (o *TaskClusterOperations) RenderValues(ctx context.Context) (*TaskClusterValues, error) {
	spec := &o.source.Spec
	rootURL := spec.RootURL
	if strings.HasSuffix(rootURL, "/") {
		rootURL = rootURL[:len(rootURL)-1]
	}

	values := &TaskClusterValues{
		Auth: AuthConfig{
			PostgresAccess: o.getPostgresAccess("auth"),
			PulseAccess:    o.getPulseAccess("auth"),
			CryptoConfig:   o.getCrypto("auth"),
			StaticAccounts: []taskclusterv1beta1.StaticAccessToken{
				o.getStaticAccessToken("built_in_workers"),
				o.getStaticAccessToken("github"),
				o.getStaticAccessToken("hooks"),
				o.getStaticAccessToken("index"),
				o.getStaticAccessToken("notify"),
				o.getStaticAccessToken("object"),
				o.getStaticAccessToken("purge_cache"),
				o.getStaticAccessToken("queue"),
				o.getStaticAccessToken("secrets"),
				o.getStaticAccessToken("web_server"),
				o.getStaticAccessToken("worker_manager"),
				o.getStaticAccessToken("root"),
			},
		},
		BuiltInWorkers: BuiltInWorkersConfig{
			TaskClusterAccess: o.getTaskClusterAccess("built_in_workers"),
		},
		GitHub: GitHubConfig{
			TaskClusterAccess: o.getTaskClusterAccess("github"),
			PostgresAccess:    o.getPostgresAccess("github"),
			PulseAccess:       o.getPulseAccess("github"),
			BotUsername:       spec.GitHub.BotUsername,
		},
		Hooks: HooksConfig{
			TaskClusterAccess: o.getTaskClusterAccess("hooks"),
			PostgresAccess:    o.getPostgresAccess("hooks"),
			PulseAccess:       o.getPulseAccess("hooks"),
			CryptoConfig:      o.getCrypto("hooks"),
		},
		Index: IndexConfig{
			TaskClusterAccess: o.getTaskClusterAccess("index"),
			PostgresAccess:    o.getPostgresAccess("index"),
			PulseAccess:       o.getPulseAccess("index"),
		},
		Notify: NotifyConfig{
			TaskClusterAccess:  o.getTaskClusterAccess("notify"),
			PostgresAccess:     o.getPostgresAccess("notify"),
			PulseAccess:        o.getPulseAccess("notify"),
			EmailSourceAddress: spec.EmailSourceAddress,
		},
		Object: ObjectConfig{
			TaskClusterAccess: o.getTaskClusterAccess("object"),
			PostgresAccess:    o.getPostgresAccess("object"),
			CryptoConfig:      o.getCrypto("object"),
		},
		PurgeCache: PurgeCacheConfig{
			TaskClusterAccess: o.getTaskClusterAccess("purge_cache"),
			PostgresAccess:    o.getPostgresAccess("purge_cache"),
		},
		Queue: QueueConfig{
			TaskClusterAccess:      o.getTaskClusterAccess("queue"),
			PostgresAccess:         o.getPostgresAccess("queue"),
			PulseAccess:            o.getPulseAccess("queue"),
			PublicArtifactBucket:   spec.PublicArtifactBucket,
			PrivateArtifactBucket:  spec.PrivateArtifactBucket,
			SignPublicArtifactURLs: spec.SignPublicArtifactURLs,
			ArtifactRegion:         spec.ArtifactRegion,
		},
		Secrets: SecretsConfig{
			TaskClusterAccess: o.getTaskClusterAccess("secrets"),
			PostgresAccess:    o.getPostgresAccess("secrets"),
			CryptoConfig:      o.getCrypto("secrets"),
		},
		WebServer: WebServerConfig{
			TaskClusterAccess:           o.getTaskClusterAccess("web_server"),
			PostgresAccess:              o.getPostgresAccess("web_server"),
			PulseAccess:                 o.getPulseAccess("web_server"),
			CryptoConfig:                o.getCrypto("web_server"),
			PublicURL:                   spec.RootURL,
			AdditionalAllowedCORSOrigin: spec.AdditionalAllowedCORSOrigin,
			SessionSecret:               o.state.SessionSecret,
			RegisteredClients:           []string{},
		},
		WorkerManager: WorkerManagerConfig{
			TaskClusterAccess: o.getTaskClusterAccess("worker_manager"),
			PostgresAccess:    o.getPostgresAccess("worker_manager"),
			PulseAccess:       o.getPulseAccess("worker_manager"),
			CryptoConfig:      o.getCrypto("worker_manager"),
			Providers:         map[string]json.RawMessage{},
		},
		UI: UIConfig{
			GraphQLSubscriptionEndpoint: fmt.Sprintf("%s/subscription", rootURL),
			GraphQLEndpoint:             fmt.Sprintf("%s/graphql", rootURL),
			BannerMessage:               spec.BannerMessage,
			UILoginStrategyNames:        strings.Join(spec.LoginStrategies, " "),
		},
		RootURL:             rootURL,
		ApplicationName:     spec.ApplicationName,
		IngressStaticIPName: spec.Ingress.StaticIPName,
		IngressExternalDNS:  spec.Ingress.ExternalDNSName,
		PulseHostname:       spec.Pulse.Host,
		PulseVHost:          spec.Pulse.Vhost,
		DockerImage:         o.dockerImage(),
		AzureAccountID:      spec.AzureAccountID,
	}

	values.Auth.StaticAccounts = append(values.Auth.StaticAccounts, o.accessTokenObjects...)

	// Fetch WST secret
	if ref := spec.WebSockTunnelSecretRef; ref != nil {
		var secret corev1.Secret
		name := types.NamespacedName{
			Namespace: o.Namespace,
			Name:      ref.Name,
		}
		if err := o.Client.Get(ctx, name, &secret); err != nil {
			return nil, err
		}

		values.Auth.WebSockTunnelSecret = (string)(secret.Data["secret"])
	}

	// Fetch Azure credentials.
	azureAccounts := map[string]string{}
	if ref := spec.AzureSecretRef; ref != nil {
		var secret corev1.Secret
		name := types.NamespacedName{
			Namespace: o.Namespace,
			Name:      ref.Name,
		}
		if err := o.Client.Get(ctx, name, &secret); err != nil {
			return nil, err
		}

		for k, v := range secret.Data {
			azureAccounts[k] = (string)(v)
		}

		values.Auth.AzureAccounts = azureAccounts
	}

	// Fetch AWS credentials.
	if ref := spec.AWSSecretRef; ref != nil {
		var secret corev1.Secret
		name := types.NamespacedName{
			Namespace: o.Namespace,
			Name:      ref.Name,
		}
		if err := o.Client.Get(ctx, name, &secret); err != nil {
			return nil, err
		}

		awsAccess := AWSAccess{
			AccessKeyID:     (string)(secret.Data["access-key-id"]),
			SecretAccessKey: (string)(secret.Data["secret-access-key"]),
		}
		values.Notify.AWSAccess = awsAccess
		values.Queue.AWSAccess = awsAccess
	}

	// Fetch GitHub configuration.
	if ref := spec.GitHub.SecretRef; ref != nil {
		var secret corev1.Secret
		name := types.NamespacedName{
			Namespace: o.Namespace,
			Name:      ref.Name,
		}
		if err := o.Client.Get(ctx, name, &secret); err != nil {
			return nil, err
		}

		webhookSecret := ""
		for k, v := range secret.Data {
			if !strings.HasPrefix("webhook-secret", k) {
				continue
			}

			s := (string)(v)
			webhookSecret += fmt.Sprintf("'%s'", s)
		}

		values.WebServer.UILoginStrategies.GitHub = &GitHubLoginStrategy{
			ClientID:     (string)(secret.Data["client-id"]),
			ClientSecret: (string)(secret.Data["client-secret"]),
		}
		values.GitHub.GitHubAppID = (string)(secret.Data["app-id"])
		values.GitHub.GitHubPrivatePEM = (string)(secret.Data["private-pem"])
		values.GitHub.WebhookSecret = webhookSecret
	}

	// Add Matrix settings
	if ref := spec.MatrixSecretRef; ref != nil {
		var secret corev1.Secret
		name := types.NamespacedName{
			Namespace: o.Namespace,
			Name:      ref.Name,
		}
		if err := o.Client.Get(ctx, name, &secret); err != nil {
			return nil, err
		}

		values.Notify.MatrixConfig = MatrixConfig{
			AccessToken: string(secret.Data["access-token"]),
			BaseURL:     string(secret.Data["base-url"]),
			UserID:      string(secret.Data["user-id"]),
		}
	}

	// Add Slack settings
	if ref := spec.SlackSecretRef; ref != nil {
		var secret corev1.Secret
		name := types.NamespacedName{
			Namespace: o.Namespace,
			Name:      ref.Name,
		}
		if err := o.Client.Get(ctx, name, &secret); err != nil {
			return nil, err
		}

		values.Notify.SlackConfig = SlackConfig{
			APIURL:      string(secret.Data["api-url"]),
			AccessToken: string(secret.Data["access-token"]),
		}
	}

	// Include provided static access tokens.
	if ref := spec.AccessTokensSecretRef; ref != nil {
		var secret corev1.Secret
		name := types.NamespacedName{
			Namespace: o.Namespace,
			Name:      ref.Name,
		}
		if err := o.Client.Get(ctx, name, &secret); err != nil {
			return nil, err
		}

		for _, v := range secret.Data {
			var accessToken taskclusterv1beta1.StaticAccessToken
			err := json.Unmarshal(v, &accessToken)
			if err != nil {
				return nil, err
			}

			values.Auth.StaticAccounts = append(values.Auth.StaticAccounts, accessToken)
		}
	}

	// Load providers.
	if ref := spec.WorkerManagerProvidersSecretRef; ref != nil {
		var secret corev1.Secret
		name := types.NamespacedName{
			Namespace: o.Namespace,
			Name:      ref.Name,
		}
		if err := o.Client.Get(ctx, name, &secret); err != nil {
			return nil, err
		}

		for k, v := range secret.Data {
			values.WorkerManager.Providers[k] = v
		}
	}

	return values, nil
}

func (o *TaskClusterOperations) ensureServiceAccount(name string) *ServiceAccount {
	if sa, ok := o.state.ServiceAccounts[name]; ok {
		return sa
	}

	sa := &ServiceAccount{}
	o.state.ServiceAccounts[name] = sa
	return sa
}

func (o *TaskClusterOperations) ensurePostgresAccess(ctx context.Context, name string) error {
	sa := o.ensureServiceAccount(name)

	username := o.source.Spec.PostgresUserPrefix
	if name != "" {
		username = fmt.Sprintf("%s_%s", username, name)
	}

	if sa.PostgresPassword == "" {
		sa.PostgresPassword = pwgen.AlphaNumeric(20)
	}

	db, err := o.connectToPostgres(ctx)
	if err != nil {
		return err
	}

	rows, err := db.Query(ctx, "SELECT 1 from pg_roles WHERE rolname=$1", pgx.QuerySimpleProtocol(true), username)
	if err != nil {
		return fmt.Errorf("error checking postgres user: %w", err)
	}

	usernameSafe := pgx.Identifier{username}.Sanitize()
	sql := ""
	if rows.Next() {
		sql = fmt.Sprintf("ALTER USER %s WITH PASSWORD $1", usernameSafe)
	} else {
		sql = fmt.Sprintf("CREATE USER %s WITH PASSWORD $1", usernameSafe)
	}
	rows.Close()

	if _, err := db.Exec(ctx, sql, pgx.QuerySimpleProtocol(true), sa.PostgresPassword); err != nil {
		return err
	}

	return nil
}

func (o *TaskClusterOperations) getPostgresAccess(name string) PostgresAccess {
	sa := o.ensureServiceAccount(name)
	db := o.dbInfo

	username := o.source.Spec.PostgresUserPrefix
	if name != "" {
		username = fmt.Sprintf("%s_%s", username, name)
	}

	db.Username = username
	db.Password = sa.PostgresPassword
	url := db.ConnectionString(true, true)
	return PostgresAccess{
		ReadDBURL:  url,
		WriteDBURL: url,
	}
}

func (o *TaskClusterOperations) ensurePulseAccess(ctx context.Context, name string) error {
	sa := o.ensureServiceAccount(name)

	dashName := strings.Replace(name, "_", "-", -1)
	username := fmt.Sprintf("%s-taskcluster-%s", o.source.Spec.Pulse.Vhost, dashName)

	if sa.PulsePassword == "" {
		sa.PulsePassword = pwgen.AlphaNumeric(20)
	}

	pulse, err := o.connectToPulse(ctx)
	if err != nil {
		return err
	}

	_, err = pulse.PutUser(username, rabbithole.UserSettings{
		Name:     username,
		Tags:     "",
		Password: sa.PulsePassword,
	})
	if err != nil {
		return err
	}

	_, err = pulse.UpdatePermissionsIn(o.source.Spec.Pulse.Vhost, username, rabbithole.Permissions{
		Configure: ".*",
		Write:     ".*",
		Read:      ".*",
	})

	return err
}

func (o *TaskClusterOperations) getPulseAccess(name string) PulseAccess {
	sa := o.ensureServiceAccount(name)

	dashName := strings.Replace(name, "_", "-", -1)
	username := fmt.Sprintf("%s-taskcluster-%s", o.source.Spec.Pulse.Vhost, dashName)

	return PulseAccess{
		PulseUsername: username,
		PulsePassword: sa.PulsePassword,
	}
}

func (o *TaskClusterOperations) ensureAccessToken(name string) {
	sa := o.ensureServiceAccount(name)

	if sa.AccessToken == "" {
		sa.AccessToken = "TC" + pwgen.AlphaNumeric(22)
	}
}

func (o *TaskClusterOperations) getTaskClusterAccess(name string) TaskClusterAccess {
	sa := o.ensureServiceAccount(name)

	return TaskClusterAccess{
		AccessToken: sa.AccessToken,
	}
}

func (o *TaskClusterOperations) getStaticAccessToken(name string) taskclusterv1beta1.StaticAccessToken {
	sa := o.ensureServiceAccount(name)
	clientID := fmt.Sprintf("static/taskcluster/%s", strings.Replace(name, "_", "-", -1))

	return taskclusterv1beta1.StaticAccessToken{
		ClientID:    clientID,
		AccessToken: sa.AccessToken,
	}
}

func (o *TaskClusterOperations) ensureCrypto(name string) {
	sa := o.ensureServiceAccount(name)

	if sa.AzureCryptoKey != "" {
		id := strconv.Itoa(int(time.Now().Unix()))
		sa.DBCryptoKeys = append(sa.DBCryptoKeys, DBCryptoKey{
			ID:        id,
			Algorithm: "aes-256",
			Key:       sa.AzureCryptoKey,
		})
		sa.AzureCryptoKey = ""
	}

	if len(sa.DBCryptoKeys) < 1 {
		id := strconv.Itoa(int(time.Now().Unix()))
		rawKey := pwgen.AlphaNumeric(32)
		key := base64.StdEncoding.EncodeToString([]byte(rawKey))

		sa.DBCryptoKeys = []DBCryptoKey{
			{
				ID:        id,
				Algorithm: "aes-256",
				Key:       key,
			},
		}
	}
}

func (o *TaskClusterOperations) getCrypto(name string) CryptoConfig {
	return o.ensureServiceAccount(name).CryptoConfig
}

func (o *TaskClusterOperations) FinishDeployment(ctx context.Context) (reconcile.Result, error) {
	upgradeKey := types.NamespacedName{
		Namespace: o.dbUpgradeJob.Namespace,
		Name:      o.dbUpgradeJob.Name,
	}

	var job batchv1.Job

	err := o.Client.Get(ctx, upgradeKey, &job)
	if err != nil && !apierrors.IsNotFound(err) {
		return reconcile.Result{}, err
	} else if err != nil {
		// Job doesn't exist, we're good.
		return reconcile.Result{}, nil
	}

	// Check if it has finished.
	desiredCompletions := int32(1)
	if job.Spec.Completions != nil {
		desiredCompletions = *job.Spec.Completions
	}

	completions := job.Status.Failed + job.Status.Succeeded
	if completions < desiredCompletions {
		o.Logger.Info("waiting for migration job to complete")
		return reconcile.Result{
			RequeueAfter: time.Minute,
		}, nil
	}

	// Delete the old Job if it doesn't match.
	if job.Annotations == nil || job.Annotations[hashAnnotation] != o.dbUpgradeHash {
		o.Logger.Info("deleting old migration job")
		if err := o.Client.Delete(ctx, &job); err != nil {
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}
