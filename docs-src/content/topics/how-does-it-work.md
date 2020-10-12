---
title: "How does it work"
weight: 310
---

![high level architecture](/img/kahoy-high-level.png)

- Load manifests into K8s resources.
  - Filter manifest at file level if required.
  - Load old state `Resource`s and `Group`s.
  - Load new state `Resource`s and `Group`s.
- Plan by comparing old and new states.
  - Get Exist resources (`Apply` plan).
  - Get Missing resources (`Delete` plan).
- Process K8s resources.
  - Filter resources at Kubernetes resource level if required (ns, type, label...).
- Manage resources.
  - Batch resources (e.g by priority).
  - Apply.
  - Delete.
- Post operations.
  - Store state.
  - ouput status.
