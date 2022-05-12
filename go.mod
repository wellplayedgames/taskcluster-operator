module github.com/wellplayedgames/taskcluster-operator

go 1.13

require (
	github.com/go-logr/logr v1.2.0
	github.com/jackc/pgx/v4 v4.8.1
	github.com/jetstack/cert-manager v0.16.1
	github.com/michaelklishin/rabbit-hole v1.5.0
	github.com/onsi/ginkgo v1.16.5
	github.com/onsi/gomega v1.18.1
	github.com/streadway/amqp v1.0.0 // indirect
	github.com/wellplayedgames/tiny-operator v0.0.0-20200908164425-0b20788cc4c0
	helm.sh/helm/v3 v3.2.4
	k8s.io/api v0.24.0
	k8s.io/apimachinery v0.24.0
	k8s.io/client-go v0.24.0
	sigs.k8s.io/controller-runtime v0.12.0
)
