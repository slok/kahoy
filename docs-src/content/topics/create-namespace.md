---
title: "Create namespace"
weight: 370
---

When we try dry-run style operations against the apiserver (like apply or diff) on namespaced resources, Kubernetes expets the namespaces of these resource to exist. However, on some of these executions the namespace has not been created yet, making the execution fail with a _namespace not found_ error.

Using `--create-namespace` flag in default or diff [modes]({{< ref "topics/modes.md" >}}), Kahoy will ensure the namespace exists before trying to apply these resources.

{{< hint warning >}}
When used in [diff mode]({{< ref "topics/modes.md#diff" >}}), a namespace will be created, so be aware of this side effect.
{{< /hint >}}
