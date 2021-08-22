package controllers

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"

	opv1 "github.com/minchao/hnc-operator/api/v1"
)

var _ = Describe("NamespaceBinding controller", func() {
	var (
		ctx           = context.Background()
		interval      = int64(10)
		defaultConfig = opv1.NamespaceBinding{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "hnc.operator/v1",
				Kind:       "NamespaceBinding",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: opv1.Config,
			},
		}
	)

	BeforeEach(func() {
		removeConfig(ctx)
	})

	AfterEach(func() {
		removeConfig(ctx)
	})

	It("Should skip execution when invalid label selector is provided", func() {
		parent := createNamespace(ctx, "parent")

		config := defaultConfig
		config.Spec = opv1.NamespaceBindingSpec{
			Selector: "$",
			Parent:   parent.Name,
			Interval: &interval,
		}

		Expect(k8sClient.Create(ctx, &config)).Should(Succeed())

		var nb *opv1.NamespaceBinding

		Eventually(func() bool {
			var err error
			if nb, err = getConfig(ctx); err != nil {
				return false
			}
			return true
		}).Should(BeTrue())

		Expect(nb.Status.LastExecutionTime).Should(BeNil())
	})

	It("Should skip execution when parent namespace not found", func() {
		config := defaultConfig
		config.Spec = opv1.NamespaceBindingSpec{
			Selector: "omega",
			Parent:   "notfound",
			Interval: &interval,
		}

		Expect(k8sClient.Create(ctx, &config)).Should(Succeed())

		var nb *opv1.NamespaceBinding

		Eventually(func() bool {
			var err error
			if nb, err = getConfig(ctx); err != nil {
				return false
			}
			return true
		}).Should(BeTrue())

		Expect(nb.Status.LastExecutionTime).Should(BeNil())
	})

	It("Should execute namespace binding", func() {
		parent := createNamespace(ctx, "parent")

		config := defaultConfig
		config.Spec = opv1.NamespaceBindingSpec{
			Selector: "alpha",
			Parent:   parent.Name,
			Interval: &interval,
		}

		Expect(k8sClient.Create(ctx, &config)).Should(Succeed())

		Eventually(func() bool {
			nb, err := getConfig(ctx)
			if err != nil {
				return false
			}
			if nb.Status.LastExecutionTime == nil {
				return false
			}
			return true
		}).Should(BeTrue())
	})

	It("Should set parent for the selected namespaces", func() {
		parent := createNamespace(ctx, "parent")

		config := defaultConfig
		config.Spec = opv1.NamespaceBindingSpec{
			Selector: "beta",
			Parent:   parent.Name,
			Interval: &interval,
		}

		createNamespaceWithLabels(ctx, "beta-1", map[string]string{"beta": ""})
		createNamespaceWithLabels(ctx, "beta-2", map[string]string{"beta": ""})
		createNamespaceWithLabels(ctx, "gamma", map[string]string{"gamma": ""})

		Expect(k8sClient.Create(ctx, &config)).Should(Succeed())

		Eventually(func() error {
			if _, err := getConfig(ctx); err != nil {
				return err
			}
			return nil
		}).Should(Succeed())

		Eventually(func() bool {
			selector, _ := labels.Parse(fmt.Sprintf("%s.tree.hnc.x-k8s.io/depth", parent.Name))
			list := &corev1.NamespaceList{}
			if err := k8sClient.List(ctx, list, &client.ListOptions{LabelSelector: selector}); err != nil {
				return false
			}
			if len(list.Items) != 2 {
				return false
			}
			return true
		}).Should(BeTrue())
	})

	It("Should exclude specific namespaces by prefix", func() {
		parent := createNamespace(ctx, "parent")

		config := defaultConfig
		config.Spec = opv1.NamespaceBindingSpec{
			Selector:   "delta",
			Parent:     parent.Name,
			Interval:   &interval,
			Exclusions: []opv1.Exclusion{{Value: "delta-2"}},
		}

		createNamespaceWithLabels(ctx, "delta-1", map[string]string{"delta": ""})
		createNamespaceWithLabels(ctx, "delta-2", map[string]string{"delta": ""})
		createNamespaceWithLabels(ctx, "epsilon", map[string]string{"epsilon": ""})

		Expect(k8sClient.Create(ctx, &config)).Should(Succeed())

		Eventually(func() error {
			if _, err := getConfig(ctx); err != nil {
				return err
			}
			return nil
		}).Should(Succeed())

		Eventually(func() bool {
			selector, _ := labels.Parse(fmt.Sprintf("%s.tree.hnc.x-k8s.io/depth", parent.Name))
			list := &corev1.NamespaceList{}
			if err := k8sClient.List(ctx, list, &client.ListOptions{LabelSelector: selector}); err != nil {
				return false
			}
			if len(list.Items) != 1 {
				return false
			}
			return true
		}).Should(BeTrue())
	})
})
