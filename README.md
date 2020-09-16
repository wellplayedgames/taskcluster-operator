# taskcluster-operator
`taskcluster-operator` is a Kubernetes operator designed to minimise operations required to run TaskCluster.

# Example Resources
## websocktunnel
```yaml
apiVersion: taskcluster.wellplayed.games/v1beta1
kind: WebSockTunnel
metadata:
  name: websocktunnel
spec:
  domainName: websocktunnel.my.org
  secretRef: { name: 'websocktunnel' }
  certificateIssuerRef:
    name: letsencrypt-prod
    kind: ClusterIssuer
```

## Taskcluster Instance
```yaml
apiVersion: taskcluster.wellplayed.games/v1beta1
kind: Instance
metadata:
  name: taskcluster
spec:
  webSockTunnelSecretRef: { name: 'websocktunnel' }
  awsSecretRef: { name: 'aws' }
  azureSecretRef: { name: 'azure' }
  workerManagerProvidersSecretRef: { name: 'taskcluster-providers' }
  databaseRef: { name: 'taskcluster' }
  authSecretRef: { name: 'taskcluster-auth' }
  accessTokensSecretRef: { name: 'taskcluster-access-tokens' }

  github:
    botUsername: Robot Gunslinger
    secretRef: { name: 'github' }
  pulse:
    host: pulse.my.org
    vhost: orgtc
    adminSecretRef: { name: 'pulse-rabbitmq-secret' }
  ingress:
    staticIpName: taskcluster
    externalDNSName: taskcluster.my.org
    tlsSecretRef: { name: 'taskcluster-tls' }
    issuerRef:
      kind: ClusterIssuer
      name: letsencrypt-prod

  rootUrl: https://taskcluster.my.org
  applicationName: Org Taskcluster
  bannerMessage: ''
  additionalAllowedCorsOrigin: ''
  emailSourceAddress: robot@my.org
  publicArtifactBucket: org-artifacts-public
  privateArtifactBucket: org-artifacts-private
  artifactRegion: eu-west-1
  azureAccountId: orgazure
  dockerImage: taskcluster/taskcluster:v30.0.2
  postgresUserPrefix: orgtc
  loginStrategies: ['github']
```

# License
This project is licensed under the [Apache 2.0 License](LICENSE).