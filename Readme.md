# chaos-fact-checker

A CLI tool that verifies whether a Chaos Mesh experiment 
actually affected its targets at runtime.

## Problem

Chaos Mesh reports experiment lifecycle but cannot verify 
if chaos actually happened. A PodChaos experiment might 
show "Running" while targeted pods are completely unaffected.

## How it works

1. Takes a chaos experiment name as input
2. Queries Kubernetes pods and events for evidence
3. Applies deterministic rules to compare observed vs expected
4. Outputs a verdict: Confirmed/Mismatch/Inconclusive


## Architecture
User input: CR name + namespace
↓
┌─────────────────────────────┐
│      Evidence Collectors    │
│  ① CR status.podRecords     │  ← reads PodChaos CR directly
│  ② Kubernetes Events        │  ← k8s event stream
│  ③ Pod status conditions    │  ← live pod state
└─────────────────────────────┘
↓
┌─────────────────────────────┐
│      Rules Engine           │  ← deterministic, no LLM
│  if podRecords > 0 → ✅     │
│  if events only   → ⚠️      │
│  if nothing       → ❌      │
└─────────────────────────────┘
↓
Verdict output

## Design decisions
- LLM is explicitly NOT used for verdict — deterministic rules only
- Reads CR status directly instead of relying on k8s events alone
- Pluggable collector design — DNSChaos and NetworkChaos next

