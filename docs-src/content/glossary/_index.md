---
title: "Glossary"
weight: 2000
---

Here are the explanations of some words and phrases used along the documentation.

{{< toc >}}

## Resource

Sometimes also called Object.

A Kubernetes object or resource uniquely identified by Kahoy using the type, group, namespace and name. E.g:

- `apps/v1/Deployment/monitoring/grafana`
- `core/v1/Service/monitoring/grafana`
- `core/v1/ServiceAccount/monitoring/grafana`
- `networking.k8s.io/v1beta1/Ingress/monitoring/grafana`

## Group

A group of resources under the same path on the file system.

## Provider

The mechanism used to obtain the groups and resources. E.g:

- Git repository using the file system.
- Kubernetes secrets using the apiserver.
- A file system path.

## State

A Kahoy executed snapshot of the resource's data (yaml specs). A provider is used to obtain it. E.g:

- Git hash.
- Kubernetes secrets with some specific label.
- An file system path.

## Report

The output kahoy creates with the resources applied and deleted at the end of an execution.

## Kubernetes provider id (storage id)

The id used to identify Kahoy state on Kubernetes, giving the ability to execute multiple kahoy independent executions by using different ids on the same cluster and namespace.

## Mode

The way Kahoy is executed. E.g:

- Dry run.
- Diff.
- Default (regular).

## Multipurpose deploy tool

Means that Kahoy can deploy Kubernetes resources in multiple scenarios and use cases. E.g:

- Independent app/release (Helm style).
- As a full repository (Flux style).
- Multiple envs (single repository, env per path).
- Raw or simple (Kubectl style)
- ...

## Focused on Kubernetes resource

Means that Kahoy focuses on the Kubernetes resource to identify and work with each of them (check if resource changed from previous state, filtering...).

So, this means that it doesn't need extra information or mutate the resource itself to track it.

Normally this implies that you are free to organize the deployment the way you want:

- Multiple apps/releases at the same time.
- Single app/release.
- Bunch of random resources (e.g register the CRDs).

## Focused on app (or release, service...)

Means that the deployment (tool) identifies what deploys using the concept or the domain of a release/app/service.

Normally this implies creating a namespace for the resources, adding extra labels to track it..., and can't be used with multiple resources of different apps at the same time.

An example could be Helm or Kapp.

## Safe deletion

Kahoy uses per resource identification and tracks each resource independently. This implies that it deletes explicitly the required resources, this method avoids side effects unlike [prune][kubectl-delete-docs] that can have side effects and is dangerous to use it.

## Ready for gitops

Means that Kahoy has been designed with CI, Git and automation in mind, so, is ready to be used in systems like Github actions or Gitlab CI easily.

More information about Gitops concept: [here][gitops]

[kubectl-delete-docs]: https://kubernetes.io/docs/tasks/manage-kubernetes-objects/declarative-config/#how-to-delete-objects
[gitops]: https://www.weave.works/technologies/gitops/
