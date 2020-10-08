---
title: "Kubernetes provider"
---

Given an storage ID and a namespace, at the end of the execution it will store the executed state (applied and deleted resources).

The ID is important because you can have different states for each Kahoy execution flows on the same cluster.

> Note: The state is stored with a `Secret` per existing resource. Be aware of [object count quota](https://kubernetes.io/docs/concepts/policy/resource-quotas/#object-count-quota)

With this state storage, it will load the `old` manifest state from Kubernetes and `new` manifest state from an fs path. This means that unlike other modes, using dry-run with Kubernetes provider needs access to a cluster.

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
