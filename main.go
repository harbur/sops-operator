package main

import (
	"flag"
	"log"
	"time"

	"github.com/harbur/project-initializer/api/types/v1alpha1"
	clientV1alpha1 "github.com/harbur/project-initializer/clientset/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	defaultAnnotation      = "initializer.kubernetes.io/projects"
	defaultInitializerName = "project.initializer.kubernetes.io"
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

	log.Println("Starting the Kubernetes project initializer...")
	log.Printf("Initializer name set to: %s", initializerName)

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	_, projectController := cache.NewInformer(
		&cache.ListWatch{
			ListFunc: func(lo metav1.ListOptions) (result runtime.Object, err error) {
				lo.IncludeUninitialized = true
				return clientSet.Projects("").List(lo)
			},
			WatchFunc: func(lo metav1.ListOptions) (watch.Interface, error) {
				lo.IncludeUninitialized = true
				return clientSet.Projects("").Watch(lo)
			},
		},
		&v1alpha1.Project{},
		1*time.Minute,
		cache.ResourceEventHandlerFuncs{},
	)

	go projectController.Run(wait.NeverStop)

	if clientset != nil {
	}
	if err != nil {
		log.Fatal(err.Error())
	}

	//////////////////

	for {
		projectsFromStore := store.List()
		for _, project := range projectsFromStore {
			var project = project.(*v1alpha1.Project)

			initializeProject(project, clientset, clientSet)
			if project.Initializers != nil {

				if len(project.Initializers.Pending) > 0 {
				}
			}
		}

		time.Sleep(2 * time.Second)
	}

}

func initializeProject(project *v1alpha1.Project, clientset *kubernetes.Clientset, v1Alpha1Clientset *clientV1alpha1.ExampleV1Alpha1Client) error {
	if project.ObjectMeta.GetInitializers() != nil {
		pendingInitializers := project.ObjectMeta.GetInitializers().Pending

		if initializerName == pendingInitializers[0].Name {
			log.Printf("Initializing Project %s (%s)", project.Name, project.Namespace)

			initializedProject := project

			// Remove self from the list of pending Initializers while preserving ordering.
			if len(pendingInitializers) == 1 {
				initializedProject.ObjectMeta.Initializers = nil
			} else {
				initializedProject.ObjectMeta.Initializers.Pending = append(pendingInitializers[:0], pendingInitializers[1:]...)
			}

			// Create Namespaces
			createNamespace(project, "lab", clientset)
			createNamespace(project, "staging", clientset)
			createNamespace(project, "pre", clientset)
			createNamespace(project, "pro", clientset)

			_, err := v1Alpha1Clientset.Projects(project.Namespace).Update(initializedProject)
			if err != nil {
				log.Printf("Error %s", err)
				return err
			}
		}
	}

	return nil
}

func createRoleBinding(project *v1alpha1.Project, namespace string, clientset *kubernetes.Clientset) error {
	// Create RoleBinding
	roleName := project.Spec.Owner + "-cluster-admin"
	roleSpec := &rbacv1.RoleBinding{ObjectMeta: metav1.ObjectMeta{Name: roleName}}
	roleSpec.RoleRef.APIGroup = "rbac.authorization.k8s.io"
	roleSpec.RoleRef.Kind = "ClusterRole"
	roleSpec.RoleRef.Name = "silk:users:cluster-admin"
	roleSpec.Subjects = make([]rbacv1.Subject, 1)
	roleSpec.Subjects[0].APIGroup = "rbac.authorization.k8s.io"
	roleSpec.Subjects[0].Kind = "User"
	roleSpec.Subjects[0].Name = project.Spec.Owner

	_, err := clientset.RbacV1().RoleBindings(namespace).Create(roleSpec)
	if err != nil {
		log.Printf("- Error creating %s", err)
		return err
	} else {
		log.Printf("- Created RoleBinding %s (%s)", roleSpec.Name, namespace)
	}
	return err
}

func createNamespace(project *v1alpha1.Project, namespaceSuffix string, clientset *kubernetes.Clientset) error {
	namespace = project.Namespace + "-" + project.Name + "-" + namespaceSuffix
	nsSpec := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: namespace}}
	result, err := clientset.Core().Namespaces().Create(nsSpec)
	if err != nil {
		log.Printf("- Error creating %s", err)
		return err
	} else {
		log.Printf("- Created Namespace %s", result.Name)
	}
	createRoleBinding(project, namespace, clientset)
	return err
}
