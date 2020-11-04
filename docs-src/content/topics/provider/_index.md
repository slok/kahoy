---
title: "Provider"
weight: 330
---

Kahoy needs two manifest states (old and new) to plan what resources have changed and which delete/apply actions need to be handled.

How these manifest states are configured is determined by the `provider` parameters.

| Provider                                            | Plug and Play | Flexible | Fast | History |
| --------------------------------------------------- | ------------- | -------- | ---- | ------- |
| [Kubernetes]({{< ref "kubernetes.md" >}}) (default) | ✔             | ✔        | ✖    | ✖       |
| [Git]({{< ref "git.md" >}})                         | ✖             | ✖        | ✔    | ✔       |
| [Paths]({{< ref "paths.md" >}})                     | ✖             | ✔        | ✔    | ✖       |

- **Plug and Play**: Refers to how straightforward is to use Kahoy with this provider. For example using Kahoy with the `kubernetes` provider works out of the box. However using the `git` or `paths` providers would require a few configuration steps from the user.
- **Flexible**: Means that the resource can be mutated/processed before applying them, for example decrypting a Secret manifest file.
- **Fast**: In terms of how fast kahoy can calculate the changes between the states. For example using the `paths` provider means old and new manifests can be read from the filesystem so Kahoy will read through them as fast as the hardware can be. Using the `kubernetes` provider however means you need to fetch the state from the cluster which involves higher latency calls.
- **History**: Means that you can apply/delete an specific point in time.
