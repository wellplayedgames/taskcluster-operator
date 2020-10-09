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
	"fmt"
	"github.com/wellplayedgames/taskcluster-operator/pkg/pwgen"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	fieldInstanceRef = ".spec.instanceRef"

	keyLastRotated = "last-rotated"
	keySecret      = "secret"
	keySecretLast  = "secret-last"
)

// WebSockTunnelReconciler reconciles a WebSockTunnel object
type WebSockTunnelReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=taskcluster.wellplayed.games,resources=websocktunnels,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=taskcluster.wellplayed.games,resources=websocktunnels/status,verbs=get;update;patch

func (r *WebSockTunnelReconciler) reconcileSecret(ctx context.Context, source *taskclusterv1beta1.WebSockTunnel) error {
	create := false
	var secret corev1.Secret
	secret.Name = source.Name
	secret.Namespace = source.Namespace
	secret.Data = map[string][]byte{}

	name := types.NamespacedName{
		Namespace: source.Namespace,
		Name:      source.Name,
	}
	err := r.Client.Get(ctx, name, &secret)
	if errors.IsNotFound(err) {
		create = true
	} else if err != nil {
		return err
	}

	// This has to be done afterwards as it would be wiped by the get.
	ctrl.SetControllerReference(source, &secret, r.Scheme)

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

	now := time.Now()
	mnow := metav1.Time{Time: now}

	// Configure the Progressing status condition.
	progressing := taskclusterv1beta1.WebSockTunnelCondition{
		Type:               taskclusterv1beta1.WebSockTunnelProgressing,
		LastTransitionTime: mnow,
		Status:             corev1.ConditionFalse,
		Reason:             "Unknown",
	}
	defer func() {
		hasSet := false
		for idx := range wst.Status.Conditions {
			c := &wst.Status.Conditions[idx]
			if c.Type == progressing.Type {
				hasSet = true

				if c.Status != progressing.Status {
					c.LastTransitionTime = progressing.LastTransitionTime
				}

				c.Status = progressing.Status
				c.Message = progressing.Message
				c.Reason = progressing.Reason
				break
			}
		}

		if !hasSet {
			wst.Status.Conditions = append(wst.Status.Conditions, progressing)
		}

		err := r.Client.Status().Update(ctx, &wst)
		if err != nil {
			logger.Error(err, "failed to update status")
		}
	}()

	// Update secret
	if err := r.reconcileSecret(ctx, &wst); err != nil {
		progressing.Reason = "ReconcileSecretFailed"
		progressing.Message = err.Error()
		return ctrl.Result{}, err
	}

	builder := &WebSockTunnelBuilder{
		Logger: logger,
		Source: &wst,
	}
	objects, err := builder.Build()
	if err != nil {
		progressing.Reason = "GenerateManifestFailed"
		progressing.Message = err.Error()
		return ctrl.Result{}, err
	}

	compReconciler := &composite.Reconciler{
		Log:    logger,
		Client: r.Client,
		Scheme: r.Scheme,
	}
	if err := compReconciler.Reconcile(ctx, resourceOwner, &wst, objects, false); err != nil {
		progressing.Reason = "CompositeReconcileFailed"
		progressing.Message = err.Error()
		return ctrl.Result{}, err
	}

	progressing.Status = corev1.ConditionTrue
	progressing.Reason = "Reconciled"
	return ctrl.Result{}, nil
}

func (r *WebSockTunnelReconciler) SetupWithManager(mgr ctrl.Manager) error {
	ctx := context.Background()

	if err := mgr.GetFieldIndexer().IndexField(ctx, &taskclusterv1beta1.AccessToken{}, fieldInstanceRef, func(obj runtime.Object) []string {
		token := obj.(*taskclusterv1beta1.AccessToken)
		instance := fmt.Sprintf("%s/%s", token.Spec.InstanceRef.Namespace, token.Spec.InstanceRef.Name)
		return []string{instance}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&taskclusterv1beta1.WebSockTunnel{}).
		Complete(r)
}
