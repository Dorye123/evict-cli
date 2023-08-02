package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/api/policy/v1beta1"
)

var kubeconfig string

func init() {
	rootCmd.PersistentFlags().StringVar(&kubeconfig, "kubeconfig", "", "Path to the kubeconfig file")
}

var rootCmd = &cobra.Command{
	Use:   "get-pods ",
	Short: "gets pods",
	Run: func(cmd *cobra.Command, args []string) {
		// Create a Kubernetes client
		clientset, err := getClient()
		if err != nil {
			fmt.Println("Error creating Kubernetes client:", err)
			return
		}

		// Get pods
		pods, err := clientset.CoreV1().Pods("").List(cmd.Context(), metav1.ListOptions{})
		if err != nil {
			fmt.Println("Error getting pods:", err)
			return
		}

		// Print pod names
		fmt.Println("\nPods:")
		for _, pod := range pods.Items {
			fmt.Println(pod.Name)
		}
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func getClient() (*kubernetes.Clientset, error) {
	config, err := getClientConfig()
	if err != nil {
		return nil, err
	}

	// Create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}

func getClientConfig() (*rest.Config, error) {
	// Use the in-cluster configuration if kubeconfig is not provided
	if kubeconfig == "" {
		config, err := rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
		return config, nil
	}

	// Use the provided kubeconfig file
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}
	return config, nil
}


func evictPodCmd() *cobra.Command {
	var podName string

	cmd := &cobra.Command{
		Use:   "evict-pod [pod-name]",
		Short: "Evicts a pod from Kubernetes",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			podName = args[0]

			// Create a Kubernetes client
			clientset, err := getClient()
			if err != nil {
				fmt.Println("Error creating Kubernetes client:", err)
				return
			}

			// Get the namespace of the pod
			pod, err := clientset.CoreV1().Pods("").Get(cmd.Context(), podName, metav1.GetOptions{})
			if err != nil {
				fmt.Println("Error getting pod:", err)
				return
			}

			namespace := pod.GetNamespace()

			// Evict the pod
			eviction := &v1beta1.Eviction{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Eviction",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      podName,
					Namespace: namespace,
				},
			}

			err = clientset.CoreV1().Pods(namespace).Evict(cmd.Context(), eviction)
			if err != nil {
				fmt.Println("Error evicting pod:", err)
				return
			}

			fmt.Printf("Pod '%s' in namespace '%s' evicted successfully.\n", podName, namespace)
		},
	}

	return cmd
}