/*
Copyright 2021.

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

package v1

import (
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var namespacebindinglog = logf.Log.WithName("namespacebinding-resource")

func (r *NamespaceBinding) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-hnc-operator-v1-namespacebinding,mutating=true,failurePolicy=fail,sideEffects=None,groups=hnc.operator,resources=namespacebindings,verbs=create;update,versions=v1,name=mnamespacebinding.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Defaulter = &NamespaceBinding{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *NamespaceBinding) Default() {
	namespacebindinglog.Info("default", "name", r.Name)

	if r.Spec.Interval == nil {
		r.Spec.Interval = new(int64)
		*r.Spec.Interval = 30
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-hnc-operator-v1-namespacebinding,mutating=false,failurePolicy=fail,sideEffects=None,groups=hnc.operator,resources=namespacebindings,verbs=create;update,versions=v1,name=vnamespacebinding.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Validator = &NamespaceBinding{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *NamespaceBinding) ValidateCreate() error {
	namespacebindinglog.Info("validate create", "name", r.Name)

	return r.validateNamespaceBinding()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *NamespaceBinding) ValidateUpdate(old runtime.Object) error {
	namespacebindinglog.Info("validate update", "name", r.Name)

	return r.validateNamespaceBinding()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *NamespaceBinding) ValidateDelete() error {
	namespacebindinglog.Info("validate delete", "name", r.Name)

	return nil
}

func (r *NamespaceBinding) validateNamespaceBinding() error {
	var allErrs field.ErrorList
	if err := r.validateLabelSelector(); err != nil {
		allErrs = append(allErrs, err)
	}
	if len(allErrs) == 0 {
		return nil
	}
	return apierrors.NewInvalid(
		schema.GroupKind{Group: "labels.Parse", Kind: "NamespaceBinding"},
		r.Namespace, allErrs)
}

func (r *NamespaceBinding) validateLabelSelector() *field.Error {
	if _, err := labels.Parse(r.Spec.Selector); err != nil {
		return field.Invalid(field.NewPath("spec").Child("selector"), r.Spec.Selector, err.Error())
	}
	return nil
}
