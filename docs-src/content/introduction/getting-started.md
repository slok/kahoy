---
title: "Getting started"
weight: 010
---

## Prerequisites

- A Kubernetes cluster and access to it.
- Kahoy and Kubectl installed.
- A folder with Kubernetes manifests.

## Install Kahoy

Get latest binary release from [Github][latest-release] or refer to the [Installing Kahoy]({{< ref "introduction/install.md" >}}) section.

## Using Kahoy

We're going to go through a brief example of how a normal interaction deploying manifests with Kahoy would look like.

We're given the following scenario:

- Kahoy is configured with the default [Kubernetes provider]({{< ref "topics/provider/kubernetes.md" >}}). Which means that it will store Kahoy's execution state on Kubernetes.
- We've decided to identify Kahoy's state with the id `ci`. It will be stored on the `default` namespace.
- There is a folder called `./manifests` containing kubernetes manifests.

First we want to check what would be applied on the first run so to be safe we're going to use the `dry run` mode.

```bash
kahoy apply --dry-run --kube-provider-id "ci" -n "./manifests"
```

We've checked the `dry run` output and it looks good. Optionally if you wanted more information about the difference between what you want to apply and what is on the cluster, the `diff` mode is available.

```bash
kahoy apply --diff --kube-provider-id "ci" -n "./manifests"
```

The output of the `diff` also looks good. It's time to really apply the manifest changes.

```bash
kahoy apply --kube-provider-id "ci" -n "./manifests"
```

{{< hint info >}} You will be prompted with a message to confirm the apply. You can use the argument `--auto-approve` to omit the interactive mode. {{< /hint >}}

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

That's it!

Lets summarize what have we seen in this starting guide:

- Sync a group of Kubernetes resources in YAML files we had on the filesystem.
- Apply a subset of resources that changed since the last execution.
- Check how Kahoy handles any kind of resource and structure.
- Handle garbage collection of resource that have been removed.
- Multiple execution modes.

This wraps up the common and basic usage of Kahoy. If you liked it there are many other options to fit your needs so keep reading!

{{< hint warning >}}

### Clean

In case you want to clean kubernetes state (identified by Kahoy `ci` state id) and all the resources used/applied in this guide, you can use `/dev/null` as the wanted manifests state.

```bash
kahoy apply --kube-provider-id "ci" -n "/dev/null"
```

**This will delete everything applied in this guide.**

{{< /hint >}}

[latest-release]: https://github.com/slok/kahoy/releases/latest
