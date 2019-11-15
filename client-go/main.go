package main

import (
	"flag"
	"fmt"
	"k8s.io/client-go/rest"
	"log"
	"runtime"

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

	pod, err := clientSet.CoreV1().Pods("default").Get("example", metav1.GetOptions{})

	if err != nil {
		log.Fatalf("Cannot get pod %v", err)
	}

	log.Println(pod)

}
