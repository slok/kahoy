---
title: "Kahoy documentation"
---

> When [Kubectl] is too simple for your needs and available deployment solutions too complex.

Maintain Kubernetes resources in sync easily.

![kahoy run example](/img/kahoy.gif)

You probably noticed a big gap between what can be done with kubectl and more advanced tools like Helm and FluxCD. But.. what about all the cases where we just need a bit more than what kubectl can offer?

Kahoy is a minimal and flexible tool to deploy your Kubernetes **raw** manifest resources to a cluster.

It's based on GitOps principles, and **out of the box** Kubernetes resources. It does not need apps/releases/services/or any other Custom Resource Definitions to manage deployments.

Kahoy will adapt to your needs and not the other way around, its been designed and developed to be generic and flexible enough for raw manifests without adding unneeded complexity.

[kubectl]: https://kubernetes.io/docs/reference/kubectl/overview/
