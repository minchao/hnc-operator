package controllers

import (
	"context"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	hncapi "sigs.k8s.io/hierarchical-namespaces/api/v1alpha2"
)

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
