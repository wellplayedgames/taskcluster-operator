package controllers

import (
	taskclusterv1beta1 "github.com/wellplayedgames/taskcluster-operator/api/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type enqueueRequestForInstance struct{}

// Create implements EventHandler
func (e *enqueueRequestForInstance) Create(evt event.CreateEvent, q workqueue.RateLimitingInterface) {
	e.addInstanceReconcileRequest(evt.Meta, q)
}

// Update implements EventHandler
func (e *enqueueRequestForInstance) Update(evt event.UpdateEvent, q workqueue.RateLimitingInterface) {
	e.addInstanceReconcileRequest(evt.MetaOld, q)
	e.addInstanceReconcileRequest(evt.MetaNew, q)
}

// Delete implements EventHandler
func (e *enqueueRequestForInstance) Delete(evt event.DeleteEvent, q workqueue.RateLimitingInterface) {
	e.addInstanceReconcileRequest(evt.Meta, q)
}

// Generic implements EventHandler
func (e *enqueueRequestForInstance) Generic(evt event.GenericEvent, q workqueue.RateLimitingInterface) {
	e.addInstanceReconcileRequest(evt.Meta, q)
}

// addInstanceReconcileRequest returns all valid requests for clusters based on the passed-in resource.
func (e *enqueueRequestForInstance) addInstanceReconcileRequest(object metav1.Object, q workqueue.RateLimitingInterface) {
	if token, ok := object.(*taskclusterv1beta1.AccessToken); ok {
		q.Add(reconcile.Request{
			NamespacedName: types.NamespacedName{
				Namespace: token.Spec.InstanceRef.Namespace,
				Name:      token.Spec.InstanceRef.Name,
			},
		})
	}
}
