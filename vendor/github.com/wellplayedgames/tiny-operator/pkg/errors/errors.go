package errors

import (
	"strings"

	"k8s.io/apimachinery/pkg/api/errors"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

// A CompositeError is an error formed from multiple other errors.
type CompositeError interface {
	Errors() []error
}

type compositeError struct {
	InnerErrors []error
}

var _ error = (*compositeError)(nil)
var _ CompositeError = (*compositeError)(nil)

func (e *compositeError) Error() string {
	if len(e.InnerErrors) == 1 {
		return e.InnerErrors[0].Error()
	}

	var result strings.Builder
	result.WriteString("multiple errors occurred: ")

	for idx, err := range e.InnerErrors {
		result.WriteString(err.Error())

		if idx < len(e.InnerErrors)-1 {
			result.WriteString(", ")
		}
	}

	return result.String()
}

func (e *compositeError) Errors() []error {
	return e.InnerErrors
}

// Append some errors to an existing error, creating a composite if
// necessary.
func Append(target error, toAdd ...error) error {
	var errors []error

	if target == nil {
		if len(toAdd) == 1 {
			return toAdd[0]
		}

		errors = toAdd
	} else if comp, ok := target.(*compositeError); ok {
		errors = append(comp.InnerErrors, toAdd...)
	} else {
		errors = make([]error, len(toAdd)+1)
		errors[0] = target
		copy(errors[1:], toAdd)
	}

	return &compositeError{
		InnerErrors: errors,
	}
}

// APIStatuses returns the list of API status contained in a given error,
// and a boolean which is true if the error consists only of API statuses.
func APIStatuses(err error) ([]metav1.Status, bool) {
	if err == nil {
		return nil, true
	} else if apiErr, ok := err.(k8serrors.APIStatus); ok {
		return []metav1.Status{apiErr.Status()}, true
	} else if compErr, ok := err.(CompositeError); ok {
		errors := compErr.Errors()
		statuses := make([]metav1.Status, 0, len(errors))
		hasOnlyStatuses := true

		for _, err := range errors {
			if apiErr, ok := err.(k8serrors.APIStatus); ok {
				statuses = append(statuses, apiErr.Status())
			} else {
				hasOnlyStatuses = false
			}
		}

		return statuses, hasOnlyStatuses
	} else {
		return nil, false
	}
}

// AllErrors returns a boolean which is true if the predicate for
// all errors in the given error return true.
func AllErrors(err error, pred func(error) bool) bool {
	var errors []error
	if comp, ok := err.(CompositeError); ok {
		errors = comp.Errors()
	} else if err != nil {
		errors = []error{err}
	}

	for _, err := range errors {
		if !pred(err) {
			return false
		}
	}

	return true
}

func findResource(scheme *runtime.Scheme, resources []runtime.Object, target *metav1.StatusDetails) runtime.Object {
	for _, resource := range resources {
		meta, err := meta.Accessor(resource)
		if err != nil {
			continue
		}

		gvk, err := apiutil.GVKForObject(resource, scheme)
		if err != nil {
			continue
		}

		if target.UID != meta.GetUID() {
			continue
		}

		if target.Group != gvk.Group {
			continue
		}

		if target.Kind != gvk.Kind {
			continue
		}

		return resource
	}

	return nil
}

// ResolveErrorStatuses tries to resolve all statuses in the given error with an
// action function. If any status cannot be resolved, an error is returned.
func ResolveErrorStatuses(
	scheme *runtime.Scheme,
	resources []runtime.Object,
	action func(obj runtime.Object, err *errors.StatusError) error,
	err error,
) error {
	statuses, onlyStatuses := APIStatuses(err)
	if !onlyStatuses {
		return err
	}

	matchedResources := make([]runtime.Object, len(statuses))

	for idx, status := range statuses {
		matched := findResource(scheme, resources, status.Details)
		if matched == nil {
			return err
		}

		matchedResources[idx] = matched
	}

	for idx, status := range statuses {
		resource := matchedResources[idx]
		statusErr := &errors.StatusError{
			ErrStatus: status,
		}
		err := action(resource, statusErr)
		if err != nil {
			return err
		}
	}

	return nil
}
