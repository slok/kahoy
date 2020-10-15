---
title: "Batch and priorities"
weight: 345
---

As we have seen in the [concepts]({{< ref "topics/concepts.md" >}}) section, Kahoy has the concept of groups. Kahoy can use these to give special options to all the resources in group.

## Priority

If `priority` is configured for a group, depending on the priority it will be batched and applied differently to the default one.

The priority is given with an integer that represents the order of the batches, by default is `1000`. This means that a group with a priority of `200`, will be batched and applied after a group with priority of `50` and before the default batch `1000`.

Normally should be fine to use the default one (no usage of priorities), however, you may want to deploy some groups first (e.g CRDs, NS...).

The priorities can be configured in Kahoy's [config file]({{< ref "topics/configuration-file.md" >}}), lets see an example:

```yaml
version: v1

groups:
  - id: crd
    priority: 200

  - id: ns
    priority: 100

  - id: system/roles
    priority: 300

  - id: apps/app1/dependencies
    priority: 200
```

With this configuration, Kahoy will create 4 batches to apply them in this order:

- 1st batch (`100`): `ns/` resources.
- 2nd batch (`200`): `crd/` and `apps/app1/dependencies` resources.
- 3rd batch (`300`): `system/roles/` resources.
- 4th batch (`1000`): Rest of the resources (default priority).

{{< hint info >}}Priorities are not used on resource deletion{{< /hint >}}

## Wait

Apart from priorities that specify the execution order, you can wait between batches.

Normally you would wait for some kind of ready condition, however there are lots of different resource types, each resource having lots of different conditions that can be seen as _ready_. This is complex.

Kahoy took the simplest approach. Instead of implementing tons of ways to wait for a ready condition, its based on eventual consistency concept and a waiting time.

Lets see an example:

```yaml
groups:
  - id: crd
    priority: 200
    wait:
      duration: 5s
```

This will make kahoy wait `5s` after applying the `crd` batch (`200`) before continuing applying the next one.

{{< hint info >}}If two groups with the same priority have different wait durations, Kahoy will take the highest wait value{{< /hint >}}
