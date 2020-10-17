---
title: "Scope and alternatives"
weight: 020
---

It's very important to identify whether Kahoy is the right tool for you. This documentation entry attempts to clarify what is the current scope of the tool and the comparison with other deployment alternative tools out there.

## Scope

Kahoy was born because the available tools for deployment required a good amount of knowledge and time just to get started with and felt too complex for our usecases. More on this under [History of the project]({{< ref "community/history.md" >}}) section. We really wanted that plug-and-play tool that felt familiar to what we already knew.

- This tool does not perform any form of templating. The generation, and mutation of the YAMLs are out of the scope. We believe the are powerful tools that can be used together with Kahoy for that matter e.g kustomize+kahoy or helm+kahoy.
- Manage the lifecycle of Kubernetes resources using raw YAML files and GitOps.
- Run on CI (dry run, diff, apply) for instant user feedback.
- Simple and flexible.
- Just a bit smarter than Kubectl.
- Ability to plan what should change declaring current and previous states (read about this in the [Concepts section]({{< ref "topics/concepts.md" >}}))

If you really want to get your applications up and running in a fairly simple and reliable way Kahoy is probably the right tool for the job. However if your desired flow consists in a more ellaborated process, and you need to bring in concepts like releases or bundles, we recommend you to take a look at the alternatives section.

## Alternatives

- [Helm]: You are forced to embrace its own ecosystem, in exchange it has templating (v2 tiller), the concept of releases, can be used to deploy single apps... It's important to note that if you like the templating capabilities of Helm but wish it was easier to deploy them don't forget you can combine both helm and kahoy to deploy generated manifests.
- [Kustomize]: Similar scope than helm but with a different approach. Also note that like with Helm, you can use kustomize for templating and kahoy for deployment.
- [Kapp]: Like Kahoy, tries to ease up the same complexity issues that come with Helm and Kustomize. In a sense its very similar to Kahoy but it introduces the concept of an Application and works on a higher level. In exchange it comes with more refined options and flows. If you need something more complex than Kahoy, is likely that Kapp is where you should look at.
- [Flux]: Controller-based flow. Asynchronous solution, pull based by definition. It's very powerful but lacks the desired feedback you'd like when releasing your applications.
- [Kubectl]: Official tool to interact with the cluster. Is what kahoy uses under the hood, very powerful tool with many different options. Unfortunately it falls behind when handling multiple manifests and dependencies between them. It also requires a fairly good amount of scripting to get things working. Kahoy really aims to remove that effort from you. _We could say that Kahoy is a small layer on top of Kubectl_.

[helm]: https://helm.sh/
[kustomize]: https://github.com/kubernetes-sigs/kustomize
[kapp]: https://github.com/k14s/kapp
[flux]: https://github.com/fluxcd/flux
[kubectl]: https://kubernetes.io/docs/reference/kubectl/overview/
