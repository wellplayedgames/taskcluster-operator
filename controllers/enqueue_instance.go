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
	for _, req := range e.getInstanceReconcileRequest(evt.Meta) {
		q.Add(req)
	}
}

// Update implements EventHandler
func (e *enqueueRequestForInstance) Update(evt event.UpdateEvent, q workqueue.RateLimitingInterface) {
	for _, req := range e.getInstanceReconcileRequest(evt.MetaOld) {
		q.Add(req)
	}
	for _, req := range e.getInstanceReconcileRequest(evt.MetaNew) {
		q.Add(req)
	}
}

// Delete implements EventHandler
func (e *enqueueRequestForInstance) Delete(evt event.DeleteEvent, q workqueue.RateLimitingInterface) {
	for _, req := range e.getInstanceReconcileRequest(evt.Meta) {
		q.Add(req)
	}
}

// Generic implements EventHandler
func (e *enqueueRequestForInstance) Generic(evt event.GenericEvent, q workqueue.RateLimitingInterface) {
	for _, req := range e.getInstanceReconcileRequest(evt.Meta) {
		q.Add(req)
	}
}

// getInstanceReconcileRequest returns all valid requests for clusters based on the passed-in resource.
func (e *enqueueRequestForInstance) getInstanceReconcileRequest(object metav1.Object) (requests []reconcile.Request) {
	if token, ok := object.(*taskclusterv1beta1.AccessToken); ok {
		requests = append(requests, reconcile.Request{
			NamespacedName: types.NamespacedName{
				Namespace: token.Spec.InstanceRef.Namespace,
				Name:      token.Spec.InstanceRef.Name,
			},
		})
	}

	return requests
}
