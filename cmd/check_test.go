package cmd

import (
	"testing"
)

func TestDeriveVerdictMatched(t *testing.T) {
	verdict, explanation := deriveVerdict(ReportEvidence{
		TargetedPods: []string{"default/nginx-abc123"},
		AffectedPods: []string{"default/nginx-abc123"},
		EventsFound:  3,
	})

	if verdict != "matched" {
		t.Fatalf("expected matched verdict, got %q", verdict)
	}
	if explanation != "All targeted pods show disruption evidence" {
		t.Fatalf("unexpected explanation: %q", explanation)
	}
}

func TestDeriveVerdictPartial(t *testing.T) {
	verdict, explanation := deriveVerdict(ReportEvidence{
		TargetedPods: []string{"default/nginx-abc123", "default/nginx-def456"},
		AffectedPods: []string{"default/nginx-abc123"},
		EventsFound:  1,
	})

	if verdict != "partial" {
		t.Fatalf("expected partial verdict, got %q", verdict)
	}
	if explanation == "" {
		t.Fatal("expected explanation for partial verdict")
	}
}

func TestExtractTargetedAndAffectedPods(t *testing.T) {
	spec := map[string]interface{}{
		"selector": map[string]interface{}{
			"pods": map[string]interface{}{
				"default": []interface{}{"nginx-abc123", "nginx-def456"},
			},
		},
	}

	status := map[string]interface{}{
		"experiment": map[string]interface{}{
			"podRecords": []interface{}{
				map[string]interface{}{
					"name":      "nginx-abc123",
					"namespace": "default",
				},
			},
		},
	}

	targeted := extractTargetedPods(spec, "default")
	affected := extractAffectedPods(status)

	if len(targeted) != 2 {
		t.Fatalf("expected 2 targeted pods, got %d", len(targeted))
	}
	if targeted[0] != "default/nginx-abc123" || targeted[1] != "default/nginx-def456" {
		t.Fatalf("unexpected targeted pods: %#v", targeted)
	}
	if len(affected) != 1 || affected[0] != "default/nginx-abc123" {
		t.Fatalf("unexpected affected pods: %#v", affected)
	}
}

func TestExtractAffectedPodsFromContainerRecords(t *testing.T) {
	status := map[string]interface{}{
		"experiment": map[string]interface{}{
			"containerRecords": []interface{}{
				map[string]interface{}{
					"id": "demo/nginx-56c45fd5ff-r5sl2",
				},
			},
		},
	}

	affected := extractAffectedPods(status)

	if len(affected) != 1 || affected[0] != "demo/nginx-56c45fd5ff-r5sl2" {
		t.Fatalf("unexpected affected pods from container records: %#v", affected)
	}
}
