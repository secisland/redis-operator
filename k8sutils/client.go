package k8sutils

import (
	"fmt"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// GenerateK8sClient create client for kubernetes
func GenerateK8sClient() *kubernetes.Clientset {
	config, err := rest.InClusterConfig()
	//config, err := clientcmd.BuildConfigFromFlags("", "/Users/shifu/.kube/config")
	if err != nil {
		fmt.Println("Error building kubeconfig:", err.Error())
	}
	clientset, _ := kubernetes.NewForConfig(config)
	return clientset
}
