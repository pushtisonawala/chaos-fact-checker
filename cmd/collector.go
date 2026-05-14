package cmd

import (
	"context"
	"fmt"
	"sort"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

type PodChaosDetails struct {
	ChaosType    string
	TargetedPods []string
	AffectedPods []string
}

// PodRecord represents one affected pod from the CR status
type PodRecord struct {
	Name      string
	Namespace string
	Action    string
	Message   string
}

func collectPodChaosDetails(
	client dynamic.Interface,
	name, namespace string,
) (PodChaosDetails, error) {

	gvr := schema.GroupVersionResource{
		Group:    "chaos-mesh.org",
		Version:  "v1alpha1",
		Resource: "podchaos",
	}

	cr, err := client.Resource(gvr).
		Namespace(namespace).
		Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return PodChaosDetails{}, fmt.Errorf("failed to get PodChaos CR: %w", err)
	}

	details := PodChaosDetails{
		ChaosType: "PodChaos",
	}

	if kind, ok := cr.Object["kind"].(string); ok && kind != "" {
		details.ChaosType = kind
	}

	if spec, ok := cr.Object["spec"].(map[string]interface{}); ok {
		details.TargetedPods = extractTargetedPods(spec, namespace)
	}

	if status, ok := cr.Object["status"].(map[string]interface{}); ok {
		details.AffectedPods = extractAffectedPods(status)
	}

	sort.Strings(details.TargetedPods)
	sort.Strings(details.AffectedPods)
	return details, nil
}

func extractTargetedPods(spec map[string]interface{}, defaultNamespace string) []string {
	selector, ok := spec["selector"].(map[string]interface{})
	if !ok {
		return nil
	}

	podsByNamespace, ok := selector["pods"].(map[string]interface{})
	if !ok {
		return nil
	}

	seen := make(map[string]struct{})
	var pods []string
	for ns, rawPods := range podsByNamespace {
		if ns == "" {
			ns = defaultNamespace
		}
		podList, ok := rawPods.([]interface{})
		if !ok {
			continue
		}
		for _, rawPod := range podList {
			podName := fmt.Sprintf("%v", rawPod)
			if podName == "" || podName == "<nil>" {
				continue
			}
			key := podName
			if ns != "" {
				key = ns + "/" + podName
			}
			if _, exists := seen[key]; exists {
				continue
			}
			seen[key] = struct{}{}
			pods = append(pods, key)
		}
	}

	return pods
}

func extractAffectedPods(status map[string]interface{}) []string {
	experiment, ok := status["experiment"].(map[string]interface{})
	if !ok {
		return nil
	}

	podRecords, ok := experiment["podRecords"].([]interface{})
	if ok {
		return collectPodsFromPodRecords(podRecords)
	}

	containerRecords, ok := experiment["containerRecords"].([]interface{})
	if ok {
		return collectPodsFromContainerRecords(containerRecords)
	}

	return nil
}

func collectPodsFromPodRecords(podRecords []interface{}) []string {
	seen := make(map[string]struct{})
	var pods []string
	for _, r := range podRecords {
		rec, ok := r.(map[string]interface{})
		if !ok {
			continue
		}
		name := fmt.Sprintf("%v", rec["name"])
		ns := fmt.Sprintf("%v", rec["namespace"])
		if name == "" || name == "<nil>" {
			continue
		}
		key := name
		if ns != "" {
			key = ns + "/" + name
		}
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		pods = append(pods, key)
	}

	return pods
}

func collectPodsFromContainerRecords(containerRecords []interface{}) []string {
	seen := make(map[string]struct{})
	var pods []string
	for _, r := range containerRecords {
		rec, ok := r.(map[string]interface{})
		if !ok {
			continue
		}
		id := fmt.Sprintf("%v", rec["id"])
		if id == "" || id == "<nil>" {
			continue
		}
		parts := strings.Split(id, "/")
		key := id
		if len(parts) >= 2 {
			key = parts[0] + "/" + parts[1]
		}
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		pods = append(pods, key)
	}

	return pods
}
