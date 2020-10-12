---
title: "Concepts"
weight: 305
---

{{< toc >}}

## Resource

Is a Kubernetes resource, Kahoy will identify resources by type, namespace, and name, so, if the manifests file arrangement changes (grouping in files, splitting, rename...) it will not affect at the plan. E.g:

Having these 2 manifests:

`grafana.yaml`:

```yaml
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: grafana
  namespace: monitoring
#...
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: grafana
  namespace: monitoring
#...
---
apiVersion: v1
kind: Service
metadata:
  name: grafana
  namespace: monitoring
#...
```

`ingress.yaml`:

```yaml
apiVersion: networking.k8s.io/v1beta1
kind: Ingress
metadata:
  name: grafana
  namespace: monitoring
#...
```

Kahoy would load 4 resources with these IDs:

- `apps/v1/Deployment/monitoring/grafana`
- `core/v1/Service/monitoring/grafana`
- `core/v1/ServiceAccount/monitoring/grafana`
- `networking.k8s.io/v1beta1/Ingress/monitoring/grafana`

> Note: Because resources are identified by its `type`, `ns`, and `name`, you can move around in files without affecting how Kahoy will identify them.

## Group

A group is a way of adding options (e.g deployment priority) to the resources in the group. You could have one or many based on what you need.

Kahoy will identify the groups from the directory structure that contains the manifests. See the following example:

Given this tree and our manifests root in `./manifests`

```bash
./manifests/
├── alertgram
│   ├── alertgram-secret.yaml
│   └── alertgram.yaml
├── bilrost
│   └── bilrost.yaml
├── root-stuff.yaml
└── grafana
    ├── config.yaml
    ├── grafana-dashboards
    │   ├── grafana-dashboards-kubernetes.yaml
    │   └── grafana-dashboards-provision.yaml
    ├── grafana.yaml
    └── ingress.yaml
```

These would be the group IDs:

- `alertgram`
- `bilrost`
- `root` (this is the root id).
- `grafana`
- `grafana/grafana-dashboards`

## State (provider)

Kahoy plans what to apply or delete based on an `old` and a `new` state of manifests. These states can come from different sources.

Check [providers section]({{< relref "provider/" >}}) for mor information.

## Plan

Is the result of resources to apply and/or delete. Its calculated based on the old and new states.
