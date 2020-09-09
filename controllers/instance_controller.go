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
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

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

func (r *InstanceReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	logger := r.Log.WithValues("instance", req.NamespacedName)

	var instance taskclusterv1beta1.Instance
	if err := r.Client.Get(ctx, req.NamespacedName, &instance); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	ops := &TaskClusterOperations{
		Logger:         r.Log,
		Client:         r.Client,
		Scheme:         r.Scheme,
		NamespacedName: req.NamespacedName,
		UsePublicIPs:   r.UsePublicIPs,
		ChartPath:      r.ChartPath,
	}

	if err := ops.Prepare(ctx); err != nil {
		return ctrl.Result{}, err
	}

	r.Log.Info("Migrating state")

	if err := ops.MigrateState(ctx); err != nil {
		return ctrl.Result{}, err
	}

	r.Log.Info("Rendering values")

	objects, err := ops.Build(ctx)
	if err != nil {
		return ctrl.Result{}, err
	}

	r.Log.Info("Applying resources")

	compReconciler := &composite.Reconciler{
		Log:    logger,
		Client: r.Client,
		Scheme: r.Scheme,
	}
	if err := compReconciler.Reconcile(ctx, resourceOwner, &instance, objects, false); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *InstanceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&taskclusterv1beta1.Instance{}).
		Complete(r)
}
