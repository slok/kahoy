---
title: "Getting started"
weight: 010
---

## Prerequisites

- A Kubernetes cluster and access to it.
- Install Kahoy.
- Install Kubectl.
- Path to a group of Kubernetes YAML manifests.

## Install Kahoy

Get latests binary release from [Github][latest-release].

## Use Kahoy

We are going to use [Kubernetes provider]({{< ref "topics/provider/kubernetes.md" >}}), this means that it will store kahoy's execution state on Kubernetes.

Our Kahoy state will be identified by `ci` id and stored on `default` namespace. Our manifest YAML files are on `./manifests` path.

First lets check the resources that will be applied (without applying) by Kahoy using `dry run`.

```bash
kahoy apply --dry-run --kube-provider-id "ci" -n "./manifests"
```

Now we are going to check the `diff` against the cluster to see the real changes that will be applied in the cluster (without applying).

```bash
kahoy apply --diff --kube-provider-id "ci" -n "./manifests"
```

Lets make them real by applying them.

{{< hint info >}} You can use `--auto-approve` to omit the interactive mode. {{< /hint >}}

```bash
kahoy apply --kube-provider-id "ci" -n "./manifests"
```

Now lets see how kahoy handles changes and deletions. Change any of the resources/manifests in `./manifests` and delete some others.

Tell kahoy that we only want to apply the resources that changed since latest execution. Lets check with `dry run` and `diff`.

```bash
kahoy apply --dry-run --kube-provider-id "ci" -n "./manifests" --include-changes

kahoy apply --diff --kube-provider-id "ci" -n "./manifests" --include-changes
```

And finally apply them on the cluster.

```bash
kahoy apply --kube-provider-id "ci" -n "./manifests" --include-changes
```

That's it! You have seen the common and basic usage, however Kahoy has lots of more options to fill your needs.

Lets summarize what have we seen in this starting guide:

- Maintain in sync a group of Kubernetes resources in YAML files in the FS.
- Apply only resource that changed since the last execution.
- Check how Kahoy handles any kind of resource and structure.
- Handle garbage collection of resource that have been removed.
- Multiple execution modes.

{{< hint warning >}}

### Clean

In case you want to clean kubernetes state (identified by Kahoy `ci` state id) and all the resources used/applied in this guide, you can use `/dev/null` as the wanted manifests state.

```bash
kahoy apply --kube-provider-id "ci" -n "/dev/null"
```

**This will delete everything applied in this guide.**

{{< /hint >}}

[latest-release]: https://github.com/slok/kahoy/releases/latest
