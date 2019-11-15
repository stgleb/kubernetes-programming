package main

import (
	"flag"
	"fmt"
	"k8s.io/client-go/rest"
	"log"
	"runtime"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
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

	if err != nil {
		log.Fatalf("Cannot create client set %v", err)
	}

	deploymentsClient := clientSet.AppsV1().Deployments(apiv1.NamespaceDefault)

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

	pods, err := clientSet.CoreV1().Pods(apiv1.NamespaceDefault).List(metav1.ListOptions{})

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
}

func int32Ptr(i int32) *int32 { return &i }
