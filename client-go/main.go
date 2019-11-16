package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	namespace = "example-namespace2"
)

func main() {
	var (
		config *rest.Config
		err    error
	)

	if config, err = rest.InClusterConfig(); err != nil && config != nil {
	} else {
		kubeConfig := flag.String("kubeconfig", "kubeconfig.json", "kubeconfig file")
		flag.Parse()

		config, err = clientcmd.BuildConfigFromFlags("", *kubeConfig)

		if err != nil {
			log.Fatalf("kubeconfig %s can not be read", *kubeConfig)
		}
	}

	config.AcceptContentTypes = "application/vnd.kubernetes.protobuf,application/json"
	config.UserAgent = fmt.Sprintf("book-example/v1.0 (%s/%s) kubernetes/v1.0",
		runtime.GOOS, runtime.GOARCH)
	clientSet, err := kubernetes.NewForConfig(config)

	if err != nil && !apierrors.IsAlreadyExists(err) {
		log.Fatalf("Cannot create client set %v", err)
	}

	_, err = clientSet.CoreV1().Namespaces().Create(&apiv1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	})

	if err != nil {
		log.Fatalf("error create namespace %s %v", namespace, err)
	}

	// Create informer
	informerFactory := informers.NewSharedInformerFactoryWithOptions(clientSet, time.Second*1,
		informers.WithNamespace(namespace))
	podInformer := informerFactory.Core().V1().Pods()
	podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			log.Printf("object %v has been added\n", obj)
		},
		DeleteFunc: func(obj interface{}) {
			log.Printf("object %v has been deleted\n", obj)
		},
	})
	informerFactory.Start(wait.NeverStop)
	informerFactory.WaitForCacheSync(wait.NeverStop)

	deploymentsClient := clientSet.AppsV1().Deployments(namespace)

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "demo-deployment",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(2),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "demo",
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "demo",
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  "web",
							Image: "nginx:1.12",
							Ports: []apiv1.ContainerPort{
								{
									Name:          "http",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: 80,
								},
							},
						},
					},
				},
			},
		},
	}

	// Create Deployment
	log.Println("Creating deployment...")
	result, err := deploymentsClient.Create(deployment)
	if err != nil {
		panic(err)
	}
	log.Printf("Created deployment %q.\n", result.GetObjectMeta().GetName())

	pods, err := clientSet.CoreV1().Pods(namespace).List(metav1.ListOptions{})

	if err != nil {
		log.Fatalf("error list pods %v", err)
	}

	for _, pod := range pods.Items {
		log.Printf("pod %s\n", pod.Name)
	}

	log.Println("Deleting deployment...")
	deletePolicy := metav1.DeletePropagationForeground
	if err := deploymentsClient.Delete("demo-deployment", &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}); err != nil {
		panic(err)
	}

	log.Println("Deleted deployment.")

	err = clientSet.CoreV1().Namespaces().Delete(namespace, &metav1.DeleteOptions{})

	if err != nil {
		log.Fatalf("error deleting namespace %s %v", namespace, err)
	}
	log.Printf("Deleted namespace %s\n", namespace)
}

func int32Ptr(i int32) *int32 { return &i }

func prompt() {
	fmt.Printf("-> Press Return key to continue.")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		break
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	fmt.Println()
}
