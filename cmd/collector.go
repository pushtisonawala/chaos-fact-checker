package cmd

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// PodRecord represents one affected pod from the CR status
type PodRecord struct {
	Name      string
	Namespace string
	Action    string
	Message   string
}

func collectPodChaosRecords(
	client dynamic.Interface,
	name, namespace string,
) ([]PodRecord, error) {

	gvr := schema.GroupVersionResource{
		Group:    "chaos-mesh.org",
		Version:  "v1alpha1",
		Resource: "podchaos",
	}

	cr, err := client.Resource(gvr).
		Namespace(namespace).
		Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get PodChaos CR: %w", err)
	}

	var records []PodRecord
	status, ok := cr.Object["status"].(map[string]interface{})
	if !ok {
		return records, nil
	}

	experiment, ok := status["experiment"].(map[string]interface{})
	if !ok {
		return records, nil
	}

	podRecords, ok := experiment["podRecords"].([]interface{})
	if !ok {
		return records, nil
	}

	for _, r := range podRecords {
		rec, ok := r.(map[string]interface{})
		if !ok {
			continue
		}
		records = append(records, PodRecord{
			Name:      fmt.Sprintf("%v", rec["name"]),
			Namespace: fmt.Sprintf("%v", rec["namespace"]),
			Action:    fmt.Sprintf("%v", rec["action"]),
			Message:   fmt.Sprintf("%v", rec["message"]),
		})
	}

	return records, nil
}
