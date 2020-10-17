---
title: "Garbage collection"
weight: 360
---

The prerequisites to be able to delete missing resources, is to track the existing resources.

Kahoy has a reliable garbage collection system. To implement this, uses the [state concept]({{< ref "topics/concepts.md#state-provider" >}}). If something that on the `old` state, doesn't exist on the `new` state, it will be deleted.

To know more of how resources are tracked and retrieved from the `old` and `new` states, you will need to check the different [providers]({{< ref "topics/provider" >}}).

{{< hint warning >}}Resources that haven't been handled by kahoy (manually, other deployment tools...) will not be tracked, so will not be garbage collected when missing.{{< /hint >}}
