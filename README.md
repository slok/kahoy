<p align="center">
    <img src="docs-src/static/img/logo.png" width="25%" align="center" alt="kahoy">
</p>

<p align="center">
  <a href="https://github.com/slok/kahoy/actions">
    <img src="https://github.com/slok/kahoy/workflows/CI/badge.svg" alt="ci">
  </a>
  
  <a href="https://goreportcard.com/report/github.com/slok/kahoy">
    <img src="https://goreportcard.com/badge/github.com/slok/kahoy" alt="Go Report">
  </a>
  
  <a href="https://hub.docker.com/r/slok/kahoy">
    <img src="https://img.shields.io/docker/pulls/slok/kahoy.svg" alt="Docker Pulls">
  </a>

  <a href="https://github.com/slok/kahoy/releases/latest">
    <img src="https://img.shields.io/github/v/release/slok/kahoy" alt="Release">
  </a>
</p>

# Kahoy - easy and reliable Kubernetes manifest deployments

> When [Kubectl] is too simple and available deployment solutions too complex.

- [Docs]
- [Releases]
- [Docker images][docker-images]

## Main features

- Simple, flexible, and lightweight. A **single CLI**.
- Deploy **any kind** of Kubernetes resource (core resources, CRDs...).
- Reliable **Garbage collection**.
- **Adapts** to any type of Kubernetes **manifest structure** (a single YAML, few manifests, big manifest repository...).
- Use to deploy as individual releases/services (**Helm style**) or group of manifest repository (**Flux style**).
- **Gitops** ready (apply only changed resources, diff, dry-run...).
- Multiple resource **filtering options** (file paths, resource namespace, types...).
- **Reports** of what applies and deletes (useful to combine with other apps, e.g: wait, checks, notifications…).

## Getting started

Check [concepts docs][concepts-docs] and start deploying any kind of Kubernetes resources:

```bash
# Dry run.
$ kahoy apply --dry-run --kube-provider-id "ci" -n "./manifests"

# Diff changes.
$ kahoy apply --diff --kube-provider-id "ci" -n "./manifests"

# Deploy.
$ kahoy apply --kube-provider-id "ci" -n "./manifests"
```

Get more information on the [docs] website.

## Contributing

Check [CONTRIBUTING.md](CONTRIBUTING.md) file.

[kubectl]: https://kubernetes.io/docs/reference/kubectl/overview/
[concepts-docs]: https://docs.kahoy.dev/topics/concepts/
[docs]: http://docs.kahoy.dev
[releases]: https://github.com/slok/kahoy/releases
[docker-images]: https://hub.docker.com/r/slok/kahoy
