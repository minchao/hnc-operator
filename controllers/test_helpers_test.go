package controllers

import (
	"context"
	"crypto/rand"
	"fmt"

	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	hncapi "sigs.k8s.io/hierarchical-namespaces/api/v1alpha2"

	opv1 "github.com/minchao/hnc-operator/api/v1"
)

type fakeHNCClient struct{}

func (f fakeHNCClient) getHierarchy(namespace string) (*hncapi.HierarchyConfiguration, error) {
	hier := &hncapi.HierarchyConfiguration{}
	hier.Name = "hierarchy"
	hier.Namespace = namespace
	return hier, nil
}

func (f fakeHNCClient) setParent(hier *hncapi.HierarchyConfiguration, parent string) error {
	namespace := &corev1.Namespace{}
	if err := k8sClient.Get(context.Background(), types.NamespacedName{Name: hier.Namespace}, namespace); err != nil {
		return err
	}
	namespace.Labels[fmt.Sprintf("%s.tree.hnc.x-k8s.io/depth", parent)] = "1"
	return k8sClient.Update(context.Background(), namespace)
}

func getConfig(ctx context.Context) (*opv1.NamespaceBinding, error) {
	nb := &opv1.NamespaceBinding{}
	if err := k8sClient.Get(ctx, types.NamespacedName{Name: opv1.Config}, nb); err != nil {
		return nil, err
	}
	return nb, nil
}

func removeConfig(ctx context.Context) {
	Eventually(func() error {
		config, err := getConfig(ctx)
		if err != nil {
			if apierrors.IsNotFound(err) {
				return nil
			}
			return err
		}
		return k8sClient.Delete(ctx, config)
	}).Should(Succeed())
}

func createNamespaceName(prefix string) string {
	suffix := make([]byte, 10)
	_, _ = rand.Read(suffix)
	return fmt.Sprintf("%s-%x", prefix, suffix)
}

func createNamespace(ctx context.Context, prefix string) *corev1.Namespace {
	return createNamespaceWithLabels(ctx, prefix, nil)
}

func createNamespaceWithLabels(ctx context.Context, prefix string, labels map[string]string) *corev1.Namespace {
	namespace := &corev1.Namespace{}
	namespace.Name = createNamespaceName(prefix)
	namespace.Labels = labels
	Expect(k8sClient.Create(ctx, namespace)).Should(Succeed())
	Eventually(func() string {
		return getNamespace(ctx, namespace.Name).Name
	}).Should(Equal(namespace.Name))
	return namespace
}

func getNamespace(ctx context.Context, name string) *corev1.Namespace {
	namespace := &corev1.Namespace{}
	Expect(k8sClient.Get(ctx, types.NamespacedName{Name: name}, namespace)).Should(Succeed())
	return namespace
}
