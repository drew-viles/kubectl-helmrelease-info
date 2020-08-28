package main

import (
	//"context"
	"flag"
	"fmt"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
	"path/filepath"
)

const (
	ResourceHealthy   = "\xE2\x9C\x85"
	ResourceUnhealthy = "\xE2\x9D\x8C"
)

var (
	client                dynamic.Interface
	clientset             *kubernetes.Clientset
	clientConfig          *rest.Config
	kubeconfig, namespace *string
	err                   error
)

func main() {
	parseFlags()
	prepKubernetesConnection()

	serverVersion, _ := clientset.ServerVersion()

	if serverVersion == nil {
		log.Fatalf("Couldn't reach Kubernetes. Do you need to authenticate first? Maybe you use ADFS. Ensure `kubectl get nodes` works to ensure you can communicate with the cluster.")
	}

	fmt.Printf("Server Version %s\n", serverVersion)
	HelmReleases()
}

// homeDir fetches the home directory based on OS.
func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

// parseFlags does what it says, it parses any flags passed to the program and sets any defaults.
func parseFlags() {
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	namespace = flag.String("n", "", "specify the namespace to get the helm release data from")
	flag.Parse()
}

// prepKubernetesConnection
func prepKubernetesConnection() {
	clientConfig, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}

	client, err = dynamic.NewForConfig(clientConfig)
	if err != nil {
		panic(err)
	}

	clientset, err = kubernetes.NewForConfig(clientConfig)
	if err != nil {
		panic(err)
	}
}
