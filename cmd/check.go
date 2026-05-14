package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	experimentName string
	namespace      string
	outputFormat   string
)

type ReportEvidence struct {
	TargetedPods []string `json:"targeted_pods"`
	AffectedPods []string `json:"affected_pods"`
	EventsFound  int      `json:"events_found"`
}

type ExperimentReport struct {
	Experiment  string         `json:"experiment"`
	ChaosType   string         `json:"chaos_type"`
	Verdict     string         `json:"verdict"`
	Evidence    ReportEvidence `json:"evidence"`
	Explanation string         `json:"explanation"`
}

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check if a chaos experiment actually affected targets",
	RunE:  runCheck,
}

func init() {
	checkCmd.Flags().StringVar(&experimentName, "name", "", "Chaos CR name")
	checkCmd.Flags().StringVar(&namespace, "namespace", "default", "Namespace")
	checkCmd.Flags().StringVar(&outputFormat, "output", "text", "Output format: text or json")
	checkCmd.MarkFlagRequired("name")
}

func runCheck(cmd *cobra.Command, args []string) error {
	report, err := buildReport()
	if err != nil {
		return err
	}

	return renderReport(report)
}

func buildReport() (ExperimentReport, error) {
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		kubeconfig = os.Getenv("HOME") + "/.kube/config"
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return ExperimentReport{}, fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return ExperimentReport{}, fmt.Errorf("failed to create k8s client: %w", err)
	}

	events, err := clientset.CoreV1().Events(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return ExperimentReport{}, fmt.Errorf("failed to list events: %w", err)
	}

	chaosEvents := countChaosEvents(events.Items)

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return ExperimentReport{}, fmt.Errorf("failed to create dynamic client: %w", err)
	}

	details, err := collectPodChaosDetails(dynamicClient, experimentName, namespace)
	if err != nil {
		return ExperimentReport{}, err
	}

	report := ExperimentReport{
		Experiment: experimentName,
		ChaosType:  details.ChaosType,
		Evidence: ReportEvidence{
			TargetedPods: details.TargetedPods,
			AffectedPods: details.AffectedPods,
			EventsFound:  chaosEvents,
		},
	}

	report.Verdict, report.Explanation = deriveVerdict(report.Evidence)
	return report, nil
}

func renderReport(report ExperimentReport) error {
	switch outputFormat {
	case "json":
		encoded, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to encode JSON report: %w", err)
		}
		fmt.Println(string(encoded))
		return nil
	case "text":
		fmt.Printf("Checking experiment: %s in namespace: %s\n\n", report.Experiment, namespace)
		fmt.Printf("Targeted pods: %d\n", len(report.Evidence.TargetedPods))
		for _, pod := range report.Evidence.TargetedPods {
			fmt.Printf("  -> %s\n", pod)
		}
		fmt.Printf("\nAffected pods: %d\n", len(report.Evidence.AffectedPods))
		for _, pod := range report.Evidence.AffectedPods {
			fmt.Printf("  -> %s\n", pod)
		}
		fmt.Printf("\nChaos-related events found: %d\n\n", report.Evidence.EventsFound)
		fmt.Printf("Verdict: %s\n", report.Verdict)
		fmt.Println(report.Explanation)
		return nil
	default:
		return fmt.Errorf("unsupported output format %q, use text or json", outputFormat)
	}
}

func countChaosEvents(events []corev1.Event) int {
	count := 0
	for _, event := range events {
		if event.Reason == "ChaosInjected" || event.Reason == "PodChaos" {
			count++
		}
	}
	return count
}

func deriveVerdict(evidence ReportEvidence) (string, string) {
	targetedCount := len(evidence.TargetedPods)
	affectedCount := len(intersection(evidence.TargetedPods, evidence.AffectedPods))

	switch {
	case targetedCount > 0 && affectedCount == targetedCount:
		return "matched", "All targeted pods show disruption evidence"
	case affectedCount > 0 || evidence.EventsFound > 0:
		return "partial", "Some disruption evidence was found, but not for every targeted pod"
	default:
		return "mismatch", "No runtime disruption evidence was found for the targeted pods"
	}
}

func intersection(targeted []string, affected []string) []string {
	if len(targeted) == 0 || len(affected) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(affected))
	for _, pod := range affected {
		seen[pod] = struct{}{}
	}

	var matched []string
	for _, pod := range targeted {
		if _, ok := seen[pod]; ok {
			matched = append(matched, pod)
		}
	}
	return matched
}
