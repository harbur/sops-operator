package v1alpha1

import (
	"github.com/harbur/sops-operator/api/types/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

type SealedSecretsInterface interface {
	List(opts metav1.ListOptions) (*v1alpha1.SealedSecretList, error)
	Get(name string, options metav1.GetOptions) (*v1alpha1.SealedSecret, error)
	Create(*v1alpha1.SealedSecret) (*v1alpha1.SealedSecret, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	Patch(name string, types types.PatchType, patchBytes []byte) (*v1alpha1.SealedSecret, error)
	Update(sealedSecret *v1alpha1.SealedSecret) (*v1alpha1.SealedSecret, error)
	// ...
}

type sealedSecretClient struct {
	restClient rest.Interface
	ns         string
}

func (c *sealedSecretClient) List(opts metav1.ListOptions) (*v1alpha1.SealedSecretList, error) {
	result := v1alpha1.SealedSecretList{}
	err := c.restClient.
		Get().
		Namespace(c.ns).
		Resource("sealedsecrets").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(&result)

	return &result, err
}

func (c *sealedSecretClient) Get(name string, opts metav1.GetOptions) (*v1alpha1.SealedSecret, error) {
	result := v1alpha1.SealedSecret{}
	err := c.restClient.
		Get().
		Namespace(c.ns).
		Resource("sealedsecrets").
		Name(name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(&result)

	return &result, err
}

func (c *sealedSecretClient) Update(sealedSecret *v1alpha1.SealedSecret) (*v1alpha1.SealedSecret, error) {
	result := v1alpha1.SealedSecret{}
	err := c.restClient.
		Put().
		Namespace(c.ns).
		Resource("sealedsecrets").
		Name(sealedSecret.Name).
		Body(sealedSecret).
		Do().
		Into(&result)

	return &result, err
}

func (c *sealedSecretClient) Create(sealedSecret *v1alpha1.SealedSecret) (*v1alpha1.SealedSecret, error) {
	result := v1alpha1.SealedSecret{}
	err := c.restClient.
		Post().
		Namespace(c.ns).
		Resource("sealedsecrets").
		Body(sealedSecret).
		Do().
		Into(&result)

	return &result, err
}

func (c *sealedSecretClient) Patch(name string, types types.PatchType, patchBytes []byte) (*v1alpha1.SealedSecret, error) {
	result := v1alpha1.SealedSecret{}
	err := c.restClient.
		Patch(types).
		Namespace(c.ns).
		Resource("sealedsecrets").
		Body(patchBytes).
		Name(name).
		Do().
		Into(&result)

	return &result, err
}

func (c *sealedSecretClient) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.restClient.
		Get().
		Namespace(c.ns).
		Resource("sealedsecrets").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}
