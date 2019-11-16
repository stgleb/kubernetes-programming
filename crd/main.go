package main

import (
	"flag"
	"log"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	namespace = "crd-namespace"
)

func main() {
	var (
		config *rest.Config
		err    error
	)

	сrdName := flag.String("crdName", "pizzas.pizza.com", "crdName")
	groupName := flag.String("groupName", "pizza.com", "group name")
	pluralName := flag.String("pluralName", "pizzas", "crd plural name")
	singularName := flag.String("singularName", "pizza", "crd singular name")
	kind := flag.String("kind", "Pizza", "crd kind")
	version := flag.String("version", "v1alpha1", "version")

	if config, err = rest.InClusterConfig(); err != nil && config != nil {
	} else {
		kubeConfig := flag.String("kubeconfig", "kubeconfig.json", "kubeconfig file")
		flag.Parse()

		config, err = clientcmd.BuildConfigFromFlags("", *kubeConfig)

		if err != nil {
			log.Fatalf("kubeconfig %s can not be read", *kubeConfig)
		}
	}

	// Create general purpose client set for creating namespace
	clientSet, err := kubernetes.NewForConfig(config)

	if err != nil && !apierrors.IsAlreadyExists(err) {
		log.Fatalf("Cannot create client set %v", err)
	}

	_, err = clientSet.CoreV1().Namespaces().Create(&apiv1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	})

	if err != nil && !apierrors.IsAlreadyExists(err) {
		log.Fatalf("error create namespace %s %v", namespace, err)
	}

	// Create extension client for creating crd
	crdClientSet, err := apiextensionsclientset.NewForConfig(config)

	if err != nil {
		log.Fatalf("create client set %v", err)
	}

	// Read file with CRD
	crd := &apiextensionsv1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name:      *сrdName,
			Namespace: namespace,
		},
		Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
			Group: *groupName,
			Scope: apiextensionsv1beta1.NamespaceScoped,
			Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
				Plural:   *pluralName,
				Singular: *singularName,
				Kind:     *kind,
			},
			Versions: []apiextensionsv1beta1.CustomResourceDefinitionVersion{
				{
					Name:    *version,
					Served:  true,
					Storage: true,
				},
			},
		},
	}

	crd, err = crdClientSet.ApiextensionsV1beta1().CustomResourceDefinitions().Create(crd)

	if err != nil && !apierrors.IsAlreadyExists(err) {
		log.Fatalf("error create crd %v", err)
	}

	log.Printf("CRD %s has been created\n", crd.Name)
}
