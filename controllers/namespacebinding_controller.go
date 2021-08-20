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

package controllers

import (
	"context"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	hncapi "sigs.k8s.io/hierarchical-namespaces/api/v1alpha2"

	opv1 "github.com/minchao/hnc-operator/api/v1"
)

type Client interface {
	getHierarchy(namespace string) (*hncapi.HierarchyConfiguration, error)
	setParent(hier *hncapi.HierarchyConfiguration, parent string) error
}

func init() {
	_ = hncapi.AddToScheme(scheme.Scheme)
}

// NamespaceBindingReconciler reconciles a NamespaceBinding object
type NamespaceBindingReconciler struct {
	client.Client
	hncClient Client
	Scheme    *runtime.Scheme
}

//+kubebuilder:rbac:groups=hnc.operator,resources=namespacebindings,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=hnc.operator,resources=namespacebindings/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=hnc.operator,resources=namespacebindings/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=namespaces,verbs=get;list;watch
//+kubebuilder:rbac:groups=hnc.x-k8s.io,resources=hierarchyconfigurations,verbs=get;create;update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the NamespaceBinding object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *NamespaceBindingReconciler) Reconcile(ctx context.Context, _ ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	inst, err := r.getInstance(ctx)
	if err != nil {
		log.Error(err, "unable to get NamespaceBinding")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	labelSelector, err := labels.Parse(inst.Spec.Selector)
	if err != nil {
		log.Error(err, "invalid selector")
		return ctrl.Result{}, err
	}

	if _, err := r.getNamespace(ctx, inst.Spec.Parent); err != nil {
		log.Error(err, "unable to get parent namespace")
		return ctrl.Result{}, err
	}

	namespaces := r.getNamespacesByLabelSelector(ctx, labelSelector)
	namespaces = r.filterNamespaces(namespaces, inst.Spec.Exclusions)
	r.setParentForSelectedNamespaces(ctx, namespaces, inst.Spec.Parent)

	return ctrl.Result{
		RequeueAfter: time.Duration(*inst.Spec.Interval) * time.Second,
	}, nil
}

func (r *NamespaceBindingReconciler) getInstance(ctx context.Context) (*opv1.NamespaceBinding, error) {
	namespaceBinding := &opv1.NamespaceBinding{}
	if err := r.Get(ctx, types.NamespacedName{Name: opv1.Config}, namespaceBinding); err != nil {
		return nil, err
	}
	return namespaceBinding, nil
}

func (r *NamespaceBindingReconciler) getNamespace(ctx context.Context, name string) (*v1.Namespace, error) {
	namespace := &v1.Namespace{}
	if err := r.Get(ctx, types.NamespacedName{Name: name}, namespace); err != nil {
		return nil, err
	}
	return namespace, nil
}

func (r *NamespaceBindingReconciler) getNamespacesByLabelSelector(ctx context.Context, selector labels.Selector) []v1.Namespace {
	list := &v1.NamespaceList{}
	if err := r.List(ctx, list, &client.ListOptions{LabelSelector: selector}); err != nil {
		logf.FromContext(ctx).Error(err, "unable to get namespaces")
	}
	return list.Items
}

// TODO
func (r *NamespaceBindingReconciler) filterNamespaces(namespaces []v1.Namespace, exclusions []opv1.Exclusion) (result []v1.Namespace) {
	for _, exclusion := range exclusions {
		for _, namespace := range namespaces {
			if !isExcludedNamespaceByPrefix(namespace.Name, exclusion.Value) {
				result = append(result, namespace)
			}
		}
	}
	return
}

func (r *NamespaceBindingReconciler) setParentForSelectedNamespaces(ctx context.Context, namespaces []v1.Namespace, parent string) {
	log := logf.FromContext(ctx)

	for _, namespace := range namespaces {
		hier, err := r.hncClient.getHierarchy(namespace.Name)
		if err != nil {
			log.Error(err, "unable to get hierarchy")
			continue
		}
		if err := r.hncClient.setParent(hier, parent); err != nil {
			log.Error(err, "unable to set parent")
		}

		log.Info("set parent namespace", "ns", namespace.Name, "parent", parent)
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *NamespaceBindingReconciler) SetupWithManager(mgr ctrl.Manager) error {
	var err error
	r.hncClient, err = newHncClient(*mgr.GetConfig())
	if err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&opv1.NamespaceBinding{}).
		Complete(r)
}

func isExcludedNamespaceByPrefix(namespace, prefix string) bool {
	return strings.HasPrefix(namespace, prefix)
}

type hncClient struct {
	client *rest.RESTClient
}

func newHncClient(config rest.Config) (client *hncClient, err error) {
	config.ContentConfig.GroupVersion = &hncapi.GroupVersion
	config.APIPath = "/apis"
	config.NegotiatedSerializer = serializer.WithoutConversionCodecFactory{CodecFactory: scheme.Codecs}
	config.UserAgent = rest.DefaultKubernetesUserAgent()

	var restClient *rest.RESTClient
	restClient, err = rest.UnversionedRESTClientFor(&config)
	if err != nil {
		return nil, err
	}
	return &hncClient{client: restClient}, err
}

func (c *hncClient) getHierarchy(namespace string) (*hncapi.HierarchyConfiguration, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	hier := &hncapi.HierarchyConfiguration{}
	hier.Name = hncapi.Singleton
	hier.Namespace = namespace
	err := c.client.Get().Resource(hncapi.HierarchyConfigurations).Namespace(namespace).Name(hncapi.Singleton).Do(ctx).Into(hier)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}
	return hier, nil
}

func (c *hncClient) setParent(hier *hncapi.HierarchyConfiguration, parent string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	hier.Spec.Parent = parent
	var err error
	if hier.CreationTimestamp.IsZero() {
		err = c.client.Post().Resource(hncapi.HierarchyConfigurations).Namespace(hier.Namespace).Name(hncapi.Singleton).Body(hier).Do(ctx).Error()
	} else {
		err = c.client.Put().Resource(hncapi.HierarchyConfigurations).Namespace(hier.Namespace).Name(hncapi.Singleton).Body(hier).Do(ctx).Error()
	}
	return err
}
