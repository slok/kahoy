---
title: "Kubernetes"
weight: 333
---

Given an storage ID and a namespace, at the end of the execution it will store the executed state (applied and deleted resources).

The ID is important because you can have different states for each Kahoy execution flows on the same cluster.

{{< hint info >}}
The ID has the same requirements as a Kubernetes label value. More info [here](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#syntax-and-character-set).
{{< /hint >}}

{{< hint warning >}}
The state is stored with a `Secret` per existing resource. Be aware of [object count quota](https://kubernetes.io/docs/concepts/policy/resource-quotas/#object-count-quota).
{{< /hint >}}

With this state storage, it will load the `old` manifest state from Kubernetes and `new` manifest state from an fs path.

{{< hint info >}}
You can use `stdin` as the new manifests path with `-` (`-n-`/`--fs-new-manifests-path -`).
{{< /hint >}}

This provider gives reliable and easy management, but is slower (Needs to get the state from the cluster) and requires space on the cluster to store the state (however the stored resources are compressed).

Example of usage:

```bash
kahoy apply \
  --provider "kubernetes" \
  --kube-provider-id "ci" \
  --fs-new-manifests-path "./manifests"
```

## Check kahoy state

If you want to check all the resource states, you can do (Check `Secret` annotations for more information):

```bash
kubectl -n {STORAGE_NAMESPACE} get secrets -l 'kahoy.slok.dev/storage-id={STORAGE_ID}'
```

## Move kahoy state

In case you want to move kahoy state to another namespaces you can get the resources and apply them in another namespace.

```bash
kubectl -n {STORAGE_OLD_NS}  get secrets -l 'kahoy.slok.dev/storage-id={STORAGE_ID}' -o json | \
  jq '.items[].metadata.namespace = "{STORAGE_NEW_NS}"' | \
  kubectl -n {STORAGE_NEW_NS} apply -f-
```

## Delete kahoy state

In the strange case that you want to reset Kahoy state, you can do it by removing these secrets and apply again all the manifests to create the latest state again:

```bash
kubectl -n {STORAGE_NAMESPACE} delete secrets -l 'kahoy.slok.dev/storage-id={STORAGE_ID}'
```

## Get the state of a resource

Identify the resource:

```bash
kubectl -n {STORAGE_NAMESPACE} get secrets -l 'kahoy.slok.dev/storage-id={STORAGE_ID}' -o jsonpath='{range .items[*]}{.metadata.name} {.metadata.annotations}{"\n"}{end}' | grep dex-config
```

Get the ID and fetch from Kubernetes.

```bash
kubectl -n {STORAGE_NAMESPACE} get secrets {GOT_ID} -o jsonpath='{.data.raw}' | base64 -d | gzip -d
```
