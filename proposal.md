
## Timeline

**Weeks 1–2**
Set up Chaos Mesh locally with kind, run real experiments, 
trace the code path from CR creation to fault injection, 
write a design doc and align with mentors before touching 
anything production.

**Weeks 3–4**
PodChaos first. Build the CR collector, Kubernetes event 
collector, pod status collector, and the rules engine. 
Write unit tests for all four verdict states.

**Weeks 5–6**
DNSChaos. Build the active DNS probe via kubectl exec 
and compare real responses against expected overrides.

**Weeks 7–8**
NetworkChaos. Active latency probe and tc rule 
verification. Integration tests for all three chaos 
types together.

**Weeks 9–10**
Verdict refinement, human-readable reports, JSON output, 
CLI flags with sensible defaults.

**Weeks 11–12**
Documentation, optional LLM summarization layer, 
code review cycles.
