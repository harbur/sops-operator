package main

import (
	"time"

	"github.com/harbur/sops-operator/api/types/v1alpha1"
	client_v1alpha1 "github.com/harbur/sops-operator/clientset/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
)

func WatchResources(clientSet client_v1alpha1.ExampleV1Alpha1Interface) cache.Store {
	sealedSecretStore, sealedSecretController := cache.NewInformer(
		&cache.ListWatch{
			ListFunc: func(lo metav1.ListOptions) (result runtime.Object, err error) {
				lo.IncludeUninitialized = true
				return clientSet.SealedSecrets("").List(lo)
			},
			WatchFunc: func(lo metav1.ListOptions) (watch.Interface, error) {
				lo.IncludeUninitialized = true
				return clientSet.SealedSecrets("").Watch(lo)
			},
		},
		&v1alpha1.SealedSecret{},
		1*time.Minute,
		cache.ResourceEventHandlerFuncs{},
	)

	go sealedSecretController.Run(wait.NeverStop)
	return sealedSecretStore
}
