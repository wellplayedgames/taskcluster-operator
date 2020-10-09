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
	"github.com/go-logr/logr"
	"github.com/wellplayedgames/tiny-operator/pkg/composite"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"time"

	taskclusterv1beta1 "github.com/wellplayedgames/taskcluster-operator/api/v1beta1"
)

// InstanceReconciler reconciles a Instance object
type InstanceReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme

	ChartPath    string
	UsePublicIPs bool
}

// +kubebuilder:rbac:groups=taskcluster.wellplayed.games,resources=instances,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=taskcluster.wellplayed.games,resources=instances/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=configmaps;secrets;services;serviceaccounts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch,resources=jobs;cronjobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles;rolebindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cert-manager.io,resources=certificates,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=extensions,resources=ingresses,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=sql.cnrm.cloud.google.com,resources=sqlinstances;sqldatabases,verbs=get;list;watch

func (r *InstanceReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	logger := r.Log.WithValues("instance", req.NamespacedName)

	var instance taskclusterv1beta1.Instance
	if err := r.Client.Get(ctx, req.NamespacedName, &instance); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	now := time.Now()
	mnow := metav1.Time{Time: now}

	// Configure the Progressing status condition.
	progressing := taskclusterv1beta1.InstanceCondition{
		Type:               taskclusterv1beta1.InstanceProgressing,
		LastTransitionTime: mnow,
		Status:             corev1.ConditionFalse,
		Reason:             "Unknown",
	}
	defer func() {
		hasSet := false
		for idx := range instance.Status.Conditions {
			c := &instance.Status.Conditions[idx]
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
			instance.Status.Conditions = append(instance.Status.Conditions, progressing)
		}

		err := r.Client.Status().Update(ctx, &instance)
		if err != nil {
			logger.Error(err, "failed to update status")
		}
	}()

	ops := &TaskClusterOperations{
		Logger:         r.Log,
		Client:         r.Client,
		Scheme:         r.Scheme,
		NamespacedName: req.NamespacedName,
		UsePublicIPs:   r.UsePublicIPs,
		ChartPath:      r.ChartPath,
	}

	if err := ops.Prepare(ctx); err != nil {
		progressing.Reason = "PrepareFailed"
		progressing.Message = err.Error()
		return ctrl.Result{}, err
	}

	r.Log.Info("Migrating state")

	if err := ops.MigrateState(ctx); err != nil {
		progressing.Reason = "MigrateFailed"
		progressing.Message = err.Error()
		return ctrl.Result{}, err
	}

	r.Log.Info("Rendering values")

	objects, err := ops.Build(ctx)
	if err != nil {
		progressing.Reason = "BuildFailed"
		progressing.Message = err.Error()
		return ctrl.Result{}, err
	}

	r.Log.Info("Applying resources")

	compReconciler := &composite.Reconciler{
		Log:    logger,
		Client: r.Client,
		Scheme: r.Scheme,
	}
	if err := compReconciler.Reconcile(ctx, resourceOwner, &instance, objects, false); err != nil {
		progressing.Reason = "CompositeReconcileFailed"
		progressing.Message = err.Error()
		return ctrl.Result{}, err
	}

	progressing.Status = corev1.ConditionTrue
	progressing.Reason = "Reconciled"
	return ctrl.Result{}, nil
}

func (r *InstanceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&taskclusterv1beta1.Instance{}).
		Watches(&source.Kind{Type: &taskclusterv1beta1.AccessToken{}}, &enqueueRequestForInstance{}).
		Complete(r)
}
