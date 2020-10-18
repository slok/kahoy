---
title: "Kubernetes auth"
weight: 375
---

By default Kubernetes will use the default context and `$HOME/.kube/config` kubeconfig.

You can override these settings with:

- `--kube-config`.
- `--kube-context`.

{{< hint info >}}
Kahoy also will configure the kubeconfig path if `KUBECONFIG` standard env var is used.
{{< /hint >}}
