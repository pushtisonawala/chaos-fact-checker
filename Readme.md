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
4. Outputs a verdict: ✅ Confirmed / ❌ Mismatch / ⚠️ Inconclusive

## Usage

```bash
./chaos-checker check --name my-podchaos --namespace default
```

## Built as part of LFX Term 2 2026 mentorship preparation
Chaos Mesh: Runtime Fact Checker for Experiments