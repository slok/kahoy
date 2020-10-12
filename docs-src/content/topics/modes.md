---
title: "Execution modes"
weight: 320
---

## Dry-run

Will plan and list the resources that need to exist and be deleted from the cluster.

Enable this mode using `--dry-run` flag.

{{< hint info >}}
This operation doesn't require to use the Kubernetes cluster unless [Kubernetes provider]({{< ref "topics/provider/kubernetes.md" >}}) is used.
{{< /hint >}}

![dry run](/img/dry-run.png)

## Diff

Will get the diff against the current cluster manifests. Requires to connect to the Kubernetes cluster.

Enable this mode using `--diff` flag.

![diff](/img/diff.png)

## Default (no options)

Will apply the resources that need to exist, and remove the ones that need to be deleted.

If we don't provide mode flags, this is the mode that will be used.

{{< hint info >}}
Apply uses Kubectl with [server-side](https://kubernetes.io/blog/2020/04/01/kubernetes-1.18-feature-server-side-apply-beta-2/#what-is-server-side-apply) apply.
{{< /hint >}}
