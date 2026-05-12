# chaos-fact-checker

A CLI tool that verifies whether a Chaos Mesh experiment actually affected its targets at runtime.

## Problem

Chaos Mesh reports experiment lifecycle, but it does not confirm whether chaos was actually observed on the target. A `PodChaos` experiment can show `Running` while the selected pods remain unaffected.

## How It Works

1. Takes a chaos experiment name and namespace as input
2. Queries Kubernetes for runtime evidence
3. Applies deterministic rules to compare observed vs expected behavior
4. Prints a verdict: `Confirmed`, `Partial`, or `Mismatch`

## Architecture

The tool follows a small, linear flow instead of a multi-layered system:

```text
chaos-checker check --name <experiment> --namespace <ns>
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
| Evidence collector | Reads Kubernetes events and PodChaos CR `status.experiment.podRecords`. |
| Verdict logic | Produces `Confirmed`, `Partial`, or `Mismatch` using deterministic rules. |

### Evidence Sources

- `PodChaos.status.experiment.podRecords`: strongest signal that Chaos Mesh recorded affected pods
- Kubernetes events: supporting evidence that injection activity was attempted
- Pod listing: namespace-level context for the check

### Verdict Rules

- `Confirmed`: one or more pod records were found in the PodChaos CR
- `Partial`: chaos-related events were found, but no pod records were present
- `Mismatch`: no pod records and no chaos-related events were found

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
