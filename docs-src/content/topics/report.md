---
title: "Report"
weight: 355
---

Kahoy can give a report at the end of the execution with the information of the resources that have been deleted and applied.

This is very flexible and powerful because it gives the ability to plug new apps after Kahoy execution e.g:

- Push notifications
- Wait for resources be available: [Example][wait-example].
- Push metrics.
- Execute sanity checks
- ...

This approach follows unix philosophy of having multiple tools, each one doing one thing well and combining them to solve an specific problem (e.g `Kahoy | jq | waiter`).

By default it doesn't give the report, use `--report-path` (`-r`) flag, using `-` for stdout (`-r -`), or a path to an output file (e.g `-r /tmp/kahoy-report.json`.

The format is in JSON because this way it can be combined with tools like [jq], example:

```json
{
  "version": "v1",
  "id": "01EHXWW5XNQF3V8WF14Z3GCAZT",
  "started_at": "2020-09-11T06:15:38Z",
  "ended_at": "2020-09-11T06:15:54Z",
  "applied_resources": [
    {
      "id": "apps/v1/Deployment/test-kahoy/grafana",
      "group": "monitoring/grafana",
      "gvk": "apps/v1/Deployment",
      "api_version": "apps/v1",
      "kind": "Deployment",
      "namespace": "test-kahoy",
      "name": "grafana"
    },
    {
      "id": "core/v1/Namespace/default/test-kahoy",
      "group": "ns",
      "gvk": "/v1/Namespace",
      "api_version": "v1",
      "kind": "Namespace",
      "namespace": "",
      "name": "test-kahoy"
    }
  ],
  "deleted_resources": [
    {
      "id": "rbac.authorization.k8s.io/v1/Role/test-kahoy/prometheus",
      "group": "monitoring/prometheus",
      "gvk": "rbac.authorization.k8s.io/v1/Role",
      "api_version": "rbac.authorization.k8s.io/v1",
      "kind": "Role",
      "namespace": "test-kahoy",
      "name": "prometheus"
    }
  ]
}
```

[wait-example]: https://github.com/slok/kahoy-app-deploy-example
