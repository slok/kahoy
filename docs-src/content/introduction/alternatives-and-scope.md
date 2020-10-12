---
title: "Alternatives and scope"
weight: 020
---

## Alternatives

Kahoy was born because the available tools for deployment were too complex for our usecase, more on this under [History of the project]({{< ref "community/history.md" >}}) section. Kubernetes is a complex system by itself, adding more complexity in the cases where is not needed, is not a good solution.

- [Helm]: Tries solving other kinds of problems, has templating (v2 tiller), the concept of releases, used to deploy single apps... However, you can use helm for templating and kahoy to deploy the generated manifests.
- [Kustomize]: Similar scope as helm but with a different approach, like Helm, you can use kustomize for the templating and kahoy for deploying raw manifests.
- [Kapp]: As Kahoy, tries solving the same problems of complexity that come with Helm, Kustomize... Very similar to Kahoy but with more complex options/flows, Kapp focuses on application level, Kahoy on Kubernetes resources, if you need something more complex than Kahoy, is likely that Kapp is your app.
- [Flux]: Controller-based flow, very powerful but complex. If you want a more `pull` than `push` approach, maybe you want this. Direct benefits of Kahoy versus flux are the inmediate feedback you can get on CI as opposed to waiting until flux runs and applies your changes.
- [Kubectl]: Official tool to interact with the cluster. Is what kahoy uses under the hood, very powerful tool, lots of options, although to make it work correctly with a group of manifests... you will most likely need scripting. Kahoy really aims to remove that effort from you. _We could say that Kahoy is a small layer on top of Kubectl_.

## Scope

- This tool does not perform any form of templating, the generation, and mutation of the YAMLs are out of the scope (We believe the are powerful tools that can be used together with Kahoy for that matter e.g kustomize+kahoy or helm+kahoy).
- Manage the lifecycle of Kubernetes resources using raw YAML files and GitOps.
- Run on CI (dry run, diff, apply)
- Simplicity and flexibility.
- Just a bit smarter than Kubectl.
- Plan what should change declaring current and previous states (read about this in the Concepts section)

If you need complex flows for your Kubernetes resources is likely that Kahoy is not for you.

[helm]: https://helm.sh/
[kustomize]: https://github.com/kubernetes-sigs/kustomize
[kapp]: https://github.com/k14s/kapp
[flux]: https://github.com/fluxcd/flux
[kubectl]: https://kubernetes.io/docs/reference/kubectl/overview/
