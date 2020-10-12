---
title: "Introduction"
weight: 000
---

Kahoy is a minimal and flexible tool to deploy your Kubernetes **raw** manifest resources to a cluster.

It's based on GitOps principles, and **out of the box** Kubernetes resources. It does not need apps/releases/services/or any other Custom Resource Definitions to manage deployments.

Kahoy will adapt to your needs and not the other way around, its been designed and developed to be generic and flexible enough for raw manifests without adding unneeded complexity.

## Features

- Simple, flexible, and lightweight.
- Deploys a deletes Kubernetes resources.
- Deploy anything, a `Namespace`, `Ingress`, `CRD`, domain apps (e.g `Deployment`+`Service`)...
- Garbage collection resources.
- Load states from different sources/providers (fs, git, kubernetes...).
- Plans at Kubernetes resource level (not file/manifest level, not app/release level)
- Gitops ready (split commands, understands git repositories, apply only changes, Diff, Dry run...).
- Use full syncs or partial syncs based on resource changes/diffs.
- Deploy priorities.
- Multiple filtering options (file paths, resource namespace, types...).
- Push mode (triggered from CI), not pull (controller).
- Use Kubectl under the hood (Kubernetes >=v1.18 and server-side apply).
- Safe deletion of resources (doesn't use `prune` method to delete K8s resources).
- Reports of what applies and deletes (useful to combine with other apps, e.g: wait, checks, notifications...).
