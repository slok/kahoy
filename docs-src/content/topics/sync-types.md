---
title: "Sync types"
weight: 340
---

Kahoy can sync Kubernetes resources in two different ways, we will call then `partial` and `full`. Using both will leverage the full power of kahoy and make the sync more reliable.

## Partial

A partial sync will sync **only the resources that have changes**.

Kahoy will compare old and new [state]({{< ref "topics/concepts.md#state-provider" >}}) resources, get the resources that changed (modifications and deletions) and filter the rest.

This makes the perfect way of execution on pull requests because:

- Maintains the scope of the changes on the branch/PR.
- Reduces the collisions of changes between changes of different people.

You can enable this filtering option with `--include-changes` in any [mode]({{< ref "topics/modes.md" >}}).

Example:

```bash
kahoy apply \
    --dry-run \
    --include-changes \
    --kube-provider-id ci \
    --fs-new-manifests-path "./manifests"
```

## Full

A `full` sync is the opposite of `partial` sync, it will apply all the resources, having changes or not.

It will not filter anything, will apply all existing resources and delete the missing ones.

Normally used to sanitize possible external changes. Making the perfect sync to be used at regular intervals (e.g. using in a scheduled pipeline every hour).

Example:

```bash
kahoy apply \
    --dry-run \
    --kube-provider-id ci \
    --fs-new-manifests-path "./manifests"
```
