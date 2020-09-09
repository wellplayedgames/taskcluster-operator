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

package controllers

import (
	"context"
	"github.com/wellplayedgames/taskcluster-operator/pkg/pwgen"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"time"

	"github.com/go-logr/logr"
	"github.com/wellplayedgames/tiny-operator/pkg/composite"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	taskclusterv1beta1 "github.com/wellplayedgames/taskcluster-operator/api/v1beta1"
)

const (
	resourceOwner = "taskcluster.wellplayed.games"

	keyLastRotated = "last-rotated"
	keySecret = "secret"
	keySecretLast = "secret-last"
)

// WebSockTunnelReconciler reconciles a WebSockTunnel object
type WebSockTunnelReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=taskcluster.wellplayed.games,resources=websocktunnels,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=taskcluster.wellplayed.games,resources=websocktunnels/status,verbs=get;update;patch

func (r *WebSockTunnelReconciler) reconcileSecret(ctx context.Context, name types.NamespacedName) error {
	create := false
	var secret corev1.Secret
	secret.Name = name.Name
	secret.Namespace = name.Namespace
	secret.Data = map[string][]byte{}

	err := r.Client.Get(ctx, name, &secret)
	if errors.IsNotFound(err) {
		create = true
	} else if err != nil {
		return err
	}

	// Check if we have both keys at all.
	if len(secret.Data[keySecret]) == 0 || len(secret.Data[keySecretLast]) == 0 {
		now := time.Now().UTC()
		secret.Data[keyLastRotated] = []byte(now.Format(time.RFC3339))

		if len(secret.Data[keySecret]) == 0 {
			secret.Data[keySecret] = []byte(pwgen.AlphaNumeric(20))
		}

		if len(secret.Data[keySecretLast]) == 0 {
			secret.Data[keySecretLast] = []byte(pwgen.AlphaNumeric(20))
		}
	}

	if create {
		return r.Client.Create(ctx, &secret)
	}

	return r.Client.Update(ctx, &secret)
}

func (r *WebSockTunnelReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	logger := r.Log.WithValues("websocktunnel", req.NamespacedName)

	var wst taskclusterv1beta1.WebSockTunnel
	if err := r.Client.Get(ctx, req.NamespacedName, &wst); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Update secret
	if err := r.reconcileSecret(ctx, req.NamespacedName); err != nil {
		return ctrl.Result{}, err
	}

	builder := &WebSockTunnelBuilder{
		Logger: logger,
		Source: &wst,
	}
	objects, err := builder.Build()
	if err != nil {
		return ctrl.Result{}, err
	}

	compReconciler := &composite.Reconciler{
		Log:    logger,
		Client: r.Client,
		Scheme: r.Scheme,
	}
	if err := compReconciler.Reconcile(ctx, resourceOwner, &wst, objects, false); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *WebSockTunnelReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&taskclusterv1beta1.WebSockTunnel{}).
		Complete(r)
}
