# chaos-fact-checker

A CLI tool that verifies whether a Chaos Mesh experiment actually affected its targets at runtime.

## Problem

Chaos Mesh reports experiment lifecycle, but it does not confirm whether chaos was actually observed on the target. A `PodChaos` experiment can show `Running` while the selected pods remain unaffected.

## How It Works

1. Takes a chaos experiment name and namespace as input
2. Queries Kubernetes for runtime evidence
3. Applies deterministic rules to compare observed vs expected behavior
4. Prints either a human-readable verdict or structured JSON

## Architecture

The tool follows a small, linear flow instead of a multi-layered system:

```text
chaos-checker check --name <experiment> --namespace <ns> --output text
  -> load kubeconfig
  -> create Kubernetes clients
  -> collect evidence
  -> apply verdict rules
  -> print result
```

| Layer | Responsibility |
| --- | --- |
| CLI entrypoint | `main.go` starts the Cobra command tree. |
| Command handler | `cmd/check.go` parses flags, loads kubeconfig, and coordinates the check flow. |
| Evidence collector | Reads the PodChaos CR selector and `status.experiment.podRecords`, plus Kubernetes events. |
| Verdict logic | Produces `matched`, `partial`, or `mismatch` using deterministic rules. |

### Evidence Sources

- `PodChaos.spec.selector.pods`: source of truth for intended target pods when explicitly listed
- `PodChaos.status.experiment.podRecords`: strongest signal that Chaos Mesh recorded affected pods
- Kubernetes events: supporting evidence that injection activity was attempted

### Output Formats

- `--output text`: human-readable terminal summary
- `--output json`: pretty-printed JSON suitable for CI or dashboards

Example:

```bash
chaos-checker check --name pod-kill-test --namespace default --output json
```

Example JSON:

```json
{
  "experiment": "pod-kill-test",
  "chaos_type": "PodChaos",
  "verdict": "matched",
  "evidence": {
    "targeted_pods": ["nginx-abc123"],
    "affected_pods": ["nginx-abc123"],
    "events_found": 3
  },
  "explanation": "All targeted pods show disruption evidence"
}
```

### Verdict Rules

- `matched`: every targeted pod has disruption evidence
- `partial`: some disruption evidence was found, but not for every targeted pod
- `mismatch`: no runtime disruption evidence was found for the targeted pods

### Current Code Layout

```text
.
|-- main.go
|-- cmd/
|   |-- root.go
|   |-- check.go
|   `-- collector.go
`-- Readme.md
```

## Design Decisions

- Verdicts are rule-based; no LLM is used
- Reads CR status directly instead of relying on Kubernetes events alone
- Keeps collector logic separate so future chaos types can be added cleanly
