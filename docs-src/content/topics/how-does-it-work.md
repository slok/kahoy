---
title: "How does it work"
weight: 310
---

A pipeline is probably the best way to illustrate all the things that are happening inside Kahoy. This pipeline consists of 5 major steps:

![high level architecture](/img/kahoy-high-level.png)

- Stage 1: Load kubernetes manifests.
  - Load old state [Resources]({{< relref "./concepts.md#resource" >}}) and [Groups]({{< relref "./concepts.md#group" >}}).
  - Load new state [Resources]({{< relref "./concepts.md#resource" >}}) and [Groups]({{< relref "./concepts.md#group" >}}).
  - Optional: There are filters available to exclude/include certain types of manifests.
- Stage 2: Plan by comparing old and new states.
  - Get existing resources and produce the section `Apply` in the [Plan]({{< relref "./concepts.md#kahoy-plan" >}}).
  - Get missing resources and produce the section `Delete` in the [Plan]({{< relref "./concepts.md#kahoy-plan" >}}).
- Stage 3: Process Kubernetes resources.
  - Optional: There are filters available to exclude/include certain resources by properties like namespace, type, labels, etcetera.
- Stage 4: Main command execution.
  - Apply resources.
  - Delete resources
  - Optiional: Perform the operations in batches with configurable priorities.
- Stage 5: Post operations.
  - Store state.
  - Output the resulting changes.

The above pipeline is also a good example to show what Kahoy doesn't do. Please visit the [scope of this project]({{< relref "/introduction/alternatives-and-scope.md#scope">}}) if you are missing any other functionality.
