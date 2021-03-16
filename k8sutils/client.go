package k8sutils

import (
	"fmt"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
)

// GenerateK8sClient create client for kubernetes
func GenerateK8sClient() *kubernetes.Clientset {
	//config, err := rest.InClusterConfig()
	//config, err := clientcmd.BuildConfigFromFlags("", "/Users/shifu/.kube/config")
	config := ctrl.GetConfigOrDie()

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Println("Error building k8s clientset:", err.Error())
	}
	return clientset
}
