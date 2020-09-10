package composite

import (
	"context"
	"encoding/json"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	tinyerrors "github.com/wellplayedgames/tiny-operator/pkg/errors"
)

const (
	compositeStateKey = "hive.wellplayed.games/composite-state"
	parentLabel       = "hive.wellplayed.games/composite-parent"
)

type permanentError struct {
	error
}

// IsPermanentError returns true if the error should not result in a retry.
func IsPermanentError(err error) bool {
	if _, ok := err.(*permanentError); ok {
		return true
	}

	return false
}

func kindIndex(kinds []schema.GroupVersionKind, kind schema.GroupKind) int {
	for idx, k := range kinds {
		if k.Group == kind.Group && k.Kind == kind.Kind {
			return idx
		}
	}

	return -1
}

func hasUID(uids []types.UID, uid types.UID) bool {
	for _, i := range uids {
		if i == uid {
			return true
		}
	}

	return false
}

// State describes the state of the
type State struct {
	DeployedKinds []schema.GroupVersionKind `json:"deployedKinds,omitempty"`
}

// EnsureKinds makes sure the given kinds are included and returns true if
// any changes were made.
func (s *State) EnsureKinds(kinds []schema.GroupVersionKind) bool {
	madeChanges := false

	for _, k := range kinds {
		idx := kindIndex(s.DeployedKinds, k.GroupKind())
		if idx < 0 {
			madeChanges = true
			s.DeployedKinds = append(s.DeployedKinds, k)
			continue
		}

		gvk := s.DeployedKinds[idx]
		if gvk.Version != k.Version {
			madeChanges = true
			s.DeployedKinds[idx] = k
			continue
		}
	}

	return madeChanges
}

// StateAccessor is a type which can access the composite state of an object.
type StateAccessor interface {
	GetCompositeState() (*State, error)
	SetCompositeState(newState *State) error
}

type stateAccessor struct {
	metav1.Object
}

// AccessState provides an accessor for
func AccessState(obj metav1.Object) StateAccessor {
	return &stateAccessor{obj}
}

func (a *stateAccessor) GetCompositeState() (*State, error) {
	var state State

	anno := a.GetAnnotations()
	if anno == nil {
		return &state, nil
	}

	text, ok := anno[compositeStateKey]
	if !ok {
		return &state, nil
	}

	err := json.Unmarshal([]byte(text), &state)
	return &state, err
}

func (a *stateAccessor) SetCompositeState(state *State) error {
	by, err := json.Marshal(state)
	if err != nil {
		return err
	}

	anno := a.GetAnnotations()
	if anno == nil {
		anno = map[string]string{}
	}

	anno[compositeStateKey] = string(by)
	a.SetAnnotations(anno)
	return nil
}

// Reconciler reconciles composite resources.
type Reconciler struct {
	Log    logr.Logger
	Client client.Client
	Scheme *runtime.Scheme
}

// Reconcile child resources of a composite resource.
func (r *Reconciler) Reconcile(ctx context.Context, owner string, parent runtime.Object, children []runtime.Object, dryRun bool) error {
	parentMeta, err := meta.Accessor(parent)
	if err != nil {
		return &permanentError{err}
	}

	log := r.Log
	acc := AccessState(parentMeta)

	var updateOptions []client.UpdateOption
	var patchOptions []client.PatchOption
	var deleteOptions []client.DeleteOption

	if dryRun {
		updateOptions = []client.UpdateOption{client.DryRunAll}
		patchOptions = []client.PatchOption{client.DryRunAll}
		deleteOptions = []client.DeleteOption{client.DryRunAll}
	}

	applyOptions := append(patchOptions, client.ForceOwnership, client.FieldOwner(owner))

	parentKey := string(parentMeta.GetUID())
	selector := labels.SelectorFromSet(labels.Set{
		parentLabel: parentKey,
	})

	state, err := acc.GetCompositeState()
	if err != nil {
		return &permanentError{err}
	}

	// Mark all new kinds, to make sure they can't get forgotten.
	desiredKinds := make([]schema.GroupVersionKind, 0, len(children))

	for _, child := range children {
		// Add GVK of resource to the list of GVKs we are processing.
		gvk, err := apiutil.GVKForObject(child, r.Scheme)
		if err != nil {
			return &permanentError{err}
		}

		idx := kindIndex(desiredKinds, gvk.GroupKind())

		if idx >= 0 {
			desiredKinds[idx] = gvk
		} else {
			desiredKinds = append(desiredKinds, gvk)
		}

		meta, err := meta.Accessor(child)
		if err != nil {
			return &permanentError{err}
		}

		// Associate with parent.
		labels := meta.GetLabels()
		if labels == nil {
			labels = map[string]string{}
		}
		labels[parentLabel] = parentKey
		meta.SetLabels(labels)

		// Set resource owner to parent.
		if meta.GetNamespace() != "" {
			err = controllerutil.SetControllerReference(parentMeta, meta, r.Scheme)
			if err != nil {
				return &permanentError{err}
			}
		}

		// Ensure child GVK is set. (For structs this isn't true by default, but needed for apply.)
		child.GetObjectKind().SetGroupVersionKind(gvk)
	}

	if state.EnsureKinds(desiredKinds) && !dryRun {
		acc.SetCompositeState(state)
		if err := r.Client.Update(ctx, parent, updateOptions...); err != nil {
			return err
		}
	}

	// Update or create all child objects.
	var passError error
	desiredUIDs := make([]types.UID, 0, len(children))
	for _, child := range children {
		objToPatch := child
		if dryRun {
			objToPatch = child.DeepCopyObject()
		}

		err := r.Client.Patch(ctx, objToPatch, client.Apply, applyOptions...)
		if err != nil {
			passError = tinyerrors.Append(passError, err)
		}

		acc, err := meta.Accessor(objToPatch)
		if err != nil {
			log.Error(err, "failed to access child metadata")
			return &permanentError{err}
		}

		desiredUIDs = append(desiredUIDs, acc.GetUID())
	}

	// If updating any resources failed, fail now.
	if passError != nil {
		return passError
	}

	// Destroy all old objects.
	for _, gvk := range state.DeployedKinds {
		var list unstructured.UnstructuredList
		list.SetGroupVersionKind(gvk)

		match := client.MatchingLabelsSelector{Selector: selector}
		err := r.Client.List(ctx, &list, match)
		if err != nil {
			return err
		}

		err = list.EachListItem(func(obj runtime.Object) error {
			kind := obj.GetObjectKind()
			kind.SetGroupVersionKind(gvk)

			acc, err := meta.Accessor(obj)
			if err != nil {
				log.Error(err, "failed to access child metadata")
				return &permanentError{err}
			}

			if hasUID(desiredUIDs, acc.GetUID()) {
				return nil
			}

			err = r.Client.Delete(ctx, obj, deleteOptions...)
			if err != nil {
				passError = tinyerrors.Append(passError, err)
			}

			return nil
		})
		if err != nil {
			return err
		}
	}

	// If deleting any resources failed, fail now.
	if passError != nil {
		return passError
	}

	// Remove old types from state.
	if (len(state.DeployedKinds) != len(desiredKinds)) && !dryRun {
		state.DeployedKinds = desiredKinds
		acc.SetCompositeState(state)
		if err := r.Client.Update(ctx, parent, updateOptions...); err != nil {
			return err
		}
	}

	return nil
}
