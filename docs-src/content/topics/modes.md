---
title: "Execution modes"
weight: 320
---

## Dry-run

Will prompt the user with an execution plan that lists the resources that would be applied and/or deleted from the cluster in the event of running kahoy in the default mode.

Enable this mode using `--dry-run` flag.

{{< hint info >}}
Dry run execution mode is a read-only operation.
{{< /hint >}}

![dry run](/img/dry-run.png)

## Diff

Will prompt the diff between current resource manifests and existing cluster resources. This mode requires connectivity to the Kubernetes cluster.

Enable this mode using `--diff` flag.

![diff](/img/diff.png)

{{< hint info >}}
Diff execution mode is a read-only operation.
{{< /hint >}}

## Default

Also known as the `apply` mode. This operation will **modify** the state of the cluster applying user desired changes.

Unless application flags provided, this is the default execution mode.

{{< hint info >}}
Under the hood the apply mode relies on Kubectl with [server-side](https://kubernetes.io/blog/2020/04/01/kubernetes-1.18-feature-server-side-apply-beta-2/#what-is-server-side-apply) apply.
{{< /hint >}}
