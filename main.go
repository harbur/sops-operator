package main

import (
	"flag"
	"log"
	"time"

	"github.com/harbur/sops-operator/api/types/v1alpha1"
	clientV1alpha1 "github.com/harbur/sops-operator/clientset/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	defaultAnnotation      = "initializer.kubernetes.io/sealedsecrets"
	defaultInitializerName = "sealedsecret.initializer.kubernetes.io"
	defaultConfigmap       = "envoy-initializer"
	defaultNamespace       = "default"
)

var (
	annotation        string
	configmap         string
	initializerName   string
	namespace         string
	requireAnnotation bool
)

type config struct {
	Containers []corev1.Container
	Volumes    []corev1.Volume
}

var kubeconfig string

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "path to Kubernetes config file")
	flag.Parse()
}

func main() {
	var config *rest.Config
	var err error

	if kubeconfig == "" {
		log.Printf("using in-cluster configuration")
		config, err = rest.InClusterConfig()
	} else {
		log.Printf("using configuration from '%s'", kubeconfig)
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	}

	if err != nil {
		panic(err)
	}

	v1alpha1.AddToScheme(scheme.Scheme)

	clientSet, err := clientV1alpha1.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	store := WatchResources(clientSet)

	/////////////////

	flag.StringVar(&annotation, "annotation", defaultAnnotation, "The annotation to trigger initialization")
	flag.StringVar(&configmap, "configmap", defaultConfigmap, "The envoy initializer configuration configmap")
	flag.StringVar(&initializerName, "initializer-name", defaultInitializerName, "The initializer name")
	flag.StringVar(&namespace, "namespace", "default", "The configuration namespace")
	flag.BoolVar(&requireAnnotation, "require-annotation", false, "Require annotation for initialization")
	flag.Parse()

	log.Println("Starting the Kubernetes sops operator...")
	log.Printf("Initializer name set to: %s", initializerName)

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	_, sealedSecretController := cache.NewInformer(
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

	if clientset != nil {
	}
	if err != nil {
		log.Fatal(err.Error())
	}

	//////////////////

	for {
		sealedSecretsFromStore := store.List()
		for _, sealedSecret := range sealedSecretsFromStore {
			var sealedSecret = sealedSecret.(*v1alpha1.SealedSecret)

			initializeSealedSecret(sealedSecret, clientset, clientSet)
			if sealedSecret.Initializers != nil {

				if len(sealedSecret.Initializers.Pending) > 0 {
				}
			}
		}

		time.Sleep(2 * time.Second)
	}

}

func initializeSealedSecret(sealedSecret *v1alpha1.SealedSecret, clientset *kubernetes.Clientset, v1Alpha1Clientset *clientV1alpha1.ExampleV1Alpha1Client) error {
	if sealedSecret.ObjectMeta.GetInitializers() != nil {
		pendingInitializers := sealedSecret.ObjectMeta.GetInitializers().Pending

		if initializerName == pendingInitializers[0].Name {
			log.Printf("Initializing SealedSecret %s (%s)", sealedSecret.Name, sealedSecret.Namespace)

			initializedSealedSecret := sealedSecret

			// Remove self from the list of pending Initializers while preserving ordering.
			if len(pendingInitializers) == 1 {
				initializedSealedSecret.ObjectMeta.Initializers = nil
			} else {
				initializedSealedSecret.ObjectMeta.Initializers.Pending = append(pendingInitializers[:0], pendingInitializers[1:]...)
			}

			// Create Namespaces
			createSecret(sealedSecret, clientset)

			_, err := v1Alpha1Clientset.SealedSecrets(sealedSecret.Namespace).Update(initializedSealedSecret)
			if err != nil {
				log.Printf("Error %s", err)
				return err
			}
		}
	}

	return nil
}

func createSecret(sealedSecret *v1alpha1.SealedSecret, clientset *kubernetes.Clientset) error {
	namespace = sealedSecret.Namespace
	nsSpec := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      sealedSecret.Name,
			Namespace: sealedSecret.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(sealedSecret, schema.GroupVersionKind{
					Group:   v1alpha1.SchemeGroupVersion.Group,
					Version: v1alpha1.SchemeGroupVersion.Version,
					Kind:    "SealedSecret",
				}),
			},
		},
		Data: sealedSecret.Data}
	result, err := clientset.Core().Secrets(sealedSecret.Namespace).Create(nsSpec)
	if err != nil {
		log.Printf("- Error unsealing %s", err)
		return err
	} else {
		log.Printf("- Unsealed %s", result.Name)
	}
	return err
}
