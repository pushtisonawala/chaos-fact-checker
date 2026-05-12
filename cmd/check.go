package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	experimentName string
	namespace      string
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check if a chaos experiment actually affected targets",
	RunE:  runCheck,
}

func init() {
	checkCmd.Flags().StringVar(&experimentName, "name", "", "Chaos CR name")
	checkCmd.Flags().StringVar(&namespace, "namespace", "default", "Namespace")
	checkCmd.MarkFlagRequired("name")
}

func runCheck(cmd *cobra.Command, args []string) error {
	// load kubeconfig
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		kubeconfig = os.Getenv("HOME") + "/.kube/config"
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create k8s client: %w", err)
	}

	fmt.Printf("Checking experiment: %s in namespace: %s\n\n",
		experimentName, namespace)

	// collect evidence — get pods in namespace
	pods, err := clientset.CoreV1().Pods(namespace).List(
		context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list pods: %w", err)
	}

	fmt.Printf("Found %d pods in namespace\n", len(pods.Items))

	// collect events related to experiment
	events, err := clientset.CoreV1().Events(namespace).List(
		context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list events: %w", err)
	}

	// simple rule — look for chaos-related events
	chaosEvents := 0
	for _, e := range events.Items {
		if e.Reason == "ChaosInjected" || e.Reason == "PodChaos" {
			chaosEvents++
			fmt.Printf("  → Event: %s on %s\n", e.Reason, e.InvolvedObject.Name)
		}
	}

	fmt.Println("\nChecking CR status.experiment.podRecords...")

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create dynamic client: %w", err)
	}

	podRecords, err := collectPodChaosRecords(dynamicClient, experimentName, namespace)
	if err != nil {
		fmt.Printf(" Could not read CR status: %v\n", err)
	} else if len(podRecords) == 0 {
		fmt.Println("No pod records found in CR status")
	} else {
		fmt.Printf("Found %d affected pod(s) in CR:\n", len(podRecords))
		for _, r := range podRecords {
			fmt.Printf("  → %s/%s | action: %s\n", r.Namespace, r.Name, r.Action)
		}
	}

	// final verdict
	fmt.Println()
	if len(podRecords) > 0 {
		fmt.Println("✅ Confirmed — CR reports chaos injected into targets")
	} else if chaosEvents > 0 {
		fmt.Println("⚠️  Partial — events found but no pod records in CR")
	} else {
		fmt.Println("❌ Mismatch — no evidence of chaos injection found")
	}

	return nil
}
