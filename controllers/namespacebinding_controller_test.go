package controllers

import (
	"context"
	"crypto/rand"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	opv1 "github.com/minchao/hnc-operator/api/v1"
)

var _ = Describe("NamespaceBinding controller", func() {
	var (
		parent *corev1.Namespace

		ctx      = context.Background()
		interval = int64(10)
	)

	BeforeEach(func() {
		parent = createNamespace(ctx, "parent")
	})

	It("Should execute namespace binding", func() {
		Eventually(func() string {
			return getNamespace(ctx, parent.Name).Name
		}).Should(Equal(parent.Name))

		config := &opv1.NamespaceBinding{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "hnc.operator/v1",
				Kind:       "NamespaceBinding",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: opv1.Config,
			},
			Spec: opv1.NamespaceBindingSpec{
				Selector: "test",
				Parent:   parent.Name,
				Interval: &interval,
			},
		}

		Expect(k8sClient.Create(ctx, config)).Should(Succeed())

		Eventually(func() bool {
			nb := &opv1.NamespaceBinding{}
			if err := k8sClient.Get(ctx, types.NamespacedName{Name: opv1.Config}, nb); err != nil {
				return false
			}
			if nb.Status.LastExecutionTime == nil {
				return false
			}
			return true
		}).Should(BeTrue())
	})
})

func createNamespaceName(prefix string) string {
	suffix := make([]byte, 10)
	_, _ = rand.Read(suffix)
	return fmt.Sprintf("%s-%x", prefix, suffix)
}

func createNamespace(ctx context.Context, prefix string) *corev1.Namespace {
	ns := &corev1.Namespace{}
	ns.Name = createNamespaceName(prefix)
	Expect(k8sClient.Create(ctx, ns)).Should(Succeed())
	return ns
}

func getNamespace(ctx context.Context, name string) *corev1.Namespace {
	ns := &corev1.Namespace{}
	Eventually(func() error {
		return k8sClient.Get(ctx, types.NamespacedName{Name: name}, ns)
	}).Should(Succeed())
	return ns
}
