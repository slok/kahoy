---
title: "Filtering resources"
weight: 350
---

{{< toc >}}

Kahoy can filter resources at many levels, they can be used in combination to adapt to your needs, e.g:

- Ignore encrypted secrets
- Start introducing Kahoy step by step by adding specific paths
- Ignore specific resources with controller annotations.
- ...

Let's see the different levels and ways Kahoy can filter.

## File system level

This way of filtering is based on files and paths on the file system, normally used to filter what `--fs-new-manifests-path` loads, or if [`paths` provider]({{< ref "topics/provider/paths.md" >}}) is used, also `--fs-old-manifests-path`.

### Exclude paths or files

Enabled with `--fs-exclude`.

Is a regex to exclude the files or paths that match, it can be repeated multiple times.

Example (Don't apply CRDs path and any file with the name `secret`):

```bash
kahoy apply \
    --include-changes \
    --kube-provider-id ci \
    --fs-new-manifests-path "./manifests" \
    --fs-exclude "manifests/crd" \
    --fs-exclude "secret"
```

### Include paths or files

Enabled with `--fs-include`.

The opposite `exclude`, it will only include the ones that match.

Example (Only apply `grafana/` and `prometheus/` ):

```bash
kahoy apply \
    --include-changes \
    --kube-provider-id ci \
    --fs-new-manifests-path "./manifests" \
    --fs-include "manifests/monitoring/grafana" \
    --fs-include "manifests/monitoring/prometheus"
```

{{< hint info >}}If both are used at the same time and both match on any path or file, exclude will have preference.{{< /hint >}}

## Resource level

Filtering at resource level is very powerful because we have all the information of the resources.

{{< hint info >}}When we filter at resource level, Kahoy needs the resource loaded, this means loading all resources (unlike file system filters), normally not a problem nor a bottleneck.{{< /hint >}}

### Changes

Enabled with `--include-changes`.

It will check the same resource in an old and a new state and if there aren't changes it will exclude the resource.

Example:

```bash
kahoy apply \
    --include-changes \
    --kube-provider-id ci \
    --fs-new-manifests-path "./manifests"
```

For more information check [partial sync]({{< ref "topics/sync-types.md#partial" >}})

### Kubernetes type

Enabled with `--kube-exclude-type`.

Is a regular expression that can be repeated, if the resource Kubernetes type (apiVersion and kind) matches with the type, it will exclude.

Example:

```bash
kahoy apply \
    --kube-provider-id ci \
    --fs-new-manifests-path "./manifests" \
    --kube-exclude-type "v1/Pod" \
    --kube-exclude-type "rbac.authorization.k8s.io" \
    --kube-exclude-type "networking.k8s.io/v1beta1" \
    --kube-exclude-type "auth.bilrost.slok.dev/v1/IngressAuth"
```

### Kubernetes labels

Enabled with `--kube-include-label`.

Uses the same Kubernetes label system, and supports the same [selectors][label-selectors].

For example, `--kube-include-label "app=myapp,component!=database,env"` would include only the resources that:

- Have `app=myapp`
- Don't have `component=database`
- Have `env` label key and any value.

Example (only deploy production outside US):

```bash
kahoy apply \
    --kube-provider-id ci \
    --fs-new-manifests-path "./manifests" \
    --kube-include-label "env=production,region!=us"
```

### Kubernetes annotations

Enabled with `--kube-include-annotation`.

Follows the same rules as the label filter but for annotations.

Example (exclude resources marked to be managed by a controller):

```bash
kahoy apply \
    --kube-provider-id ci \
    --fs-new-manifests-path "./manifests" \
    --kube-include-annotation "!auth.bilrost.slok.dev/backend"
```

### Kubernetes namespaces

Enabled with `--include-namespace`.

A regex that can be repeated, if none of them match the resource namespace, it will exclude them.

Example (Exclude everything except `app-a` and `app-b` namespaces):

```bash
kahoy apply \
    --kube-provider-id ci \
    --fs-new-manifests-path "./manifests" \
    --include-namespace "app-a" \
    --include-namespace "app-b"
```

[label-selectors]: https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#label-selectors
