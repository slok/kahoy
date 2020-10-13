---
title: "Concepts"
weight: 305
---

{{< toc >}}

## Resource

When we refer to a Resource we mean any Kubernetes resource.

Kahoy internally identifies resources by type, namespace, and name. Thanks to this the user can arrange any number of kubernetes resources in a single yaml or split them in multiple yamls. Both scenarios will be considered equal and it will not affect any of the operations run by Kahoy. Take a look at the following example:

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

> Note: Because resources are identified by their `type`, `ns`, and `name`, you can safely move them around between files and it will not affect how Kahoy identifies them.

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

## State Provider

Kahoy plans what to apply or delete based on an `old` and a `new` state of manifests. These states can come from different sources and thats entirely up to
the user needs.

Check [providers section]({{< relref "provider/" >}}) for more information.

## Kahoy Plan

Kahoy plan is a command that creates an execution plan for the user to inspect.
It's calculated based on the old and new states and prompts the user with all
the meaningful information about the incoming changes.
