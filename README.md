<p align="center">
    <img src="docs/img/logo.png" width="25%" align="center" alt="kahoy">
</p>

# Kahoy

[![Build Status][ci-image]][ci-url] [![Go Report Card][goreport-image]][goreport-url]

> When [Kubectl] is too simple for your needs and available deployment solutions too complex.

Maintain Kubernetes resources in sync easily.

![kahoy run example](docs/img/kahoy.gif)

---

## :link: Table of contents

- [:tada: Introduction](#tada-introduction)
- [:checkered_flag: Features](#checkered_flag-features)
- [:shipit: Install](#shipit-install)
- [:key: Getting started](#key-getting-started)
- [:mag: Scope](#mag-scope)
- [:pencil2: Concepts](#pencil2-concepts)
- [:wrench: How does it work](#wrench-how-does-it-work)
- [:computer: Execution options](#computer-execution-options)
- [:page_facing_up: Manifest source modes](#page_facing_up-manifest-source-modes)
- [:bulb: Use cases](#bulb-use-cases)
- [:question: F.A.Q](#question-faq)
- [:tophat: Alternatives](#tophat-alternatives)

## :tada: Introduction

Kahoy is a minimal and flexible tool to sync/deploy your Kubernetes resource **raw** manifests and a cluster.

Focuses on Gitops, and Kubernetes resources (not apps/releases/services/whatever) and understands git repositories.

Unlike other tools, Kahoy will adapt to your needs and not the other way around, its been designed and developed to be generic and flexible enough for raw manifests without adding unneeded complexity.

## :checkered_flag: Features

- Simple, flexible, and lightweight.
- Deploys a deletes Kubernetes resources.
- Deploy anything, from a `Namespace`, `Deployment`... to a `CRD`.
- Plans what to delete or deploy based on two manifest states (old and new).
- Garbage collection resources.
- Load states from different sources (fs, git...).
- Plans at Kubernetes resource level (not file/manifest level, not app/release level)
- Different execution modes: Diff, Dry run...
- Gitops ready (split commands, understands git repositories).
- Use full syncs or partial syncs based on git diffs.
- Deploy priorities.
- Multiple filtering options (file paths, resource namespace, types...).
- Uses Kubernetes >=v1.18 and server-side apply.
- Push mode (triggered from CI), not pull (controller).

## :shipit: Install

- Docker: Ready image with all the required tools (e.g Kubectl) `docker pull slok/kahoy`.
- Releases: Go to [releases](https://github.com/slok/kahoy/releases).
- Build from source: `git clone git@github.com:slok/kahoy.git && cd ./kahoy && make build`.

## :key: Getting started

A simple example that deploys/deletes what changed between the states `HEAD` (new) and `2cd4a1c1a7921ec593432cfdb9307dc8d6584862` (old) git revisions.

```bash
kahoy apply -n "./manifests" -c "2cd4a1c1a7921ec593432cfdb9307dc8d6584862"
```

For more advanced ways of using this, check:

- `kahoy apply --help`.
- [Use cases](<(#bulb-use-cases)>) section.
- [Github actions example][github-actions-example].

## :mag: Scope

- No templating, the generation, and mutation of the YAMLs are out of the scope (use other tools and then Kahoy, e.g kustomize+kahoy).
- Manage the lifecycle of Kubernetes resources (Deployment and deletion of resources) using raw YAML files.
- Focus on Gitops and CI step/jobs (dry run, diff, apply).
- Simplicity and flexibility.
- Just a bit smarter than Kubectl.
- Plan what will be applied and what deleted based in an old and a new manifest state (fs, git...).

If you need complex flows for your Kubernetes resources is likely that Kahoy is not for you.

## :pencil2: Concepts

Kahoy doesn't depend on app/service, labels/selectors, or any other kind of app grouping concept, it uses these 3 concepts:

### State

Kahoy plans what to apply or delete based on an `old` and a `new` state of manifests. These states can come from different sources:

- `paths`: Given 2 filesystem paths, it will use one for the old state and the other for the new state.
- `git`: Given 2 revisions (depends on the `git` mode), it will use the old git revision to get the manifest's state at that moment, and the new git revision to get the manifest's state on that moment.

### Resource

Is a Kubernetes resource, Kahoy will identify resources by type, ns, and name, so, if the manifests file arrangement changes (grouping in files, splitting, rename...) will not affect at the plan. E.g:

Having these 2 manifests:

`grafana.yaml`:

```yaml
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: grafana
  namespace: monitoring
#...
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: grafana
  namespace: monitoring
#...
---
apiVersion: v1
kind: Service
metadata:
  name: grafana
  namespace: monitoring
#...
```

`ingress.yaml`:

```yaml
apiVersion: networking.k8s.io/v1beta1
kind: Ingress
metadata:
  name: grafana
  namespace: monitoring
#...
```

Kahoy would load 4 resources with these IDs:

- `apps/v1/Deployment/monitoring/grafana`
- `core/v1/Service/monitoring/grafana`
- `core/v1/ServiceAccount/monitoring/grafana`
- `networking.k8s.io/v1beta1/Ingress/monitoring/grafana`

> Note: Because resources are identified by its `type`, `ns`, and `name`, you can move around in files without affecting how Kahoy will identify them.

### Groups

A group is a way of adding options (e.g deployment priority) to the resources in the group.

Kahoy will identify the groups by the path where the manifests are based on the root of the manifests. E.g:

Having this tree and our manifests root in `./manifests`

```bash
./manifests/
├── alertgram
│   ├── alertgram-secret.yaml
│   └── alertgram.yaml
├── bilrost
│   └── bilrost.yaml
├── root-stuff.yaml
└── grafana
    ├── config.yaml
    ├── grafana-dashboards
    │   ├── grafana-dashboards-kubernetes.yaml
    │   └── grafana-dashboards-provision.yaml
    ├── grafana.yaml
    └── ingress.yaml
```

These would be the group IDs:

- `alertgram`
- `bilrost`
- `root` (this is the root id).
- `grafana`
- `grafana/grafana-dashboards`

## :wrench: How does it work

- Load manifests into K8s resources.
  - Filter manifest at file level if required.
  - Load old state `Resource`s and `Group`s.
  - Load new state `Resource`s and `Group`s.
- Plan by comparing old and new states.
  - Get Exist resources (`Apply` plan).
  - Get Missing resources (`Delete` plan).
- Process K8s resources.
  - Filter resources at Kubernetes resource level if required (ns, type, label...).
- Manage resources.
  - Batch resources (e.g by priority).
  - Apply.
  - Delete.

## :computer: Execution options

### Dry-run

Will plan what resources need to exist on the cluster and what needs to be removed (client-side, no cluster required).

![dry run](docs/img/dry-run.png)

### Diff

Will get a diff against the server of the planned resources (server-side, cluster required).

![diff](docs/img/diff.png)

### Default (Apply)

Will apply the resources that need to exist and delete the ones that don't. Apply uses Kubectl and [server-side][serverside-apply] apply.

## :page_facing_up: Manifest source modes

Kahoy needs two manifest states (old and new) to plan what resources need to exist/gone in the cluster. How these manifests are provided is using the `mode`

### `paths` (File system)

Given 2 manifest file system paths, plans what needs to be applied against a cluster, and what needs to be deleted.

This one is the most generic one and can be used when you want to manage almost everything, e.g previous Kahoy execution, prepare using bash scripts, kustomize, secrets...

### `git`

This is the best one for **gitops**.

This is the default mode, this mode understands git and can read states from a git repository, these 2 states are based on 2 git revisions.

Using `before-commit` will make a plan based on the manifests of `HEAD` (new state) and the commit provided (old state). Normally used when executed from `master/main` branch.

Instead of providing the `before-commit`, by default will get the base parent of the current branch `HEAD` (new state) against the default branch (old state), normally `master/main`). This mode is used when you are executing kahoy from a branch in a pull request.

Apart from knowing how to get an old and a new state from a git repository. **Git mode understands diff/patches**, this would make Kahoy only be applied what has been changed between these two revisions/commits. This is interesting in many cases:

- When you have lost of resources:
  - Have a clear view of what is changing.
  - Blazing fast deployments.
- Operators sometimes change deployed manifests, this mode avoids overwriting every manifest on each deployment.
- Split full syncs with partial syncs (you can continue making a full sync every hour or whatever).

## :bulb: Use cases

### Dry run on pull requests

When someone pushes a branch different from master, would be nice to execute Dry run on the changes (partial sync) of that branch.

Having our manifest in a git repository in `./manifests` and being our default branch `master`, we do this from a different branch.

```bash
kahoy apply \
    --dry-run \
    --git-diff-filter \
    --fs-new-manifests-path "./manifests"
```

### Diff for CI Pull requests

Same as above but with `--diff` instead of `--dry-run`

### Deploying on master branch (when PR merged)

Kahoy needs to compare our `HEAD` against the previous applied state, that's why we need the `before-commit`.
In this example, we use `--git-before-commit-sha` flag. Normally this variable can be obtained in the executing CI:

Github actions uses [Github context][gh-context], example:

```yaml
env:
  GIT_BEFORE_COMMIT_SHA: ${{ github.event.before }}
```

So... in Kahoy:

```bash
kahoy apply \
    --git-diff-filter \
    --git-before-commit-sha "${GIT_BEFORE_COMMIT_SHA}" \
    --fs-new-manifests-path "./manifests"
```

Note: For Gitlab CI, this uses [env vars][gitlab-ci-env] (in `CI_COMMIT_BEFORE_SHA`).

### Schedule a full sync

Check this [Github actions example][github-actions-example] for more info.

### Exclude some manifest files (e.g encrypted secrets)

If you have some files that you don't want to be managed by kahoy, you can ignore them at file system level using `--fs-exclude`. Can be repeated.

E.g: exclude any file with the name secret on it.

```bash
kahoy apply \
    --fs-new-manifests-path "./manifests" \
    --fs-exclude "secret"
```

### Delete all

Instead of using git, use the fs by using the `paths` mode. Use the new state as `dev/null`.

```bash
kahoy apply \
    --mode="paths" \
    --fs-old-manifests-path "./manifests" \
    --fs-new-manifests-path "/dev/null"
```

### Deploy all

Instead of using git, use the fs by using the `paths` mode. Use the old state as `dev/null`.

```bash
kahoy apply \
    --mode="paths" \
    --fs-old-manifests-path "/dev/null" \
    --fs-new-manifests-path "./manifests"
```

### Deploy only some manifests

You can use the file filtering option `--fs-include`, works with any mode (`git`, `paths`...)

```bash
kahoy apply \
    --fs-new-manifests-path "./manifests" \
    --fs-include "./manifests/prometheus" \
    --fs-include "./manifests/grafana"
```

### Multiple envs

If you have multiple envs on the same repository, you can have them in different manifests root.

```bash
kahoy apply \
    --kube-context "env1" \
    --fs-new-manifests-path "./manifests/env1"
```

```bash
kahoy apply \
    --kube-context "env2" \
    --fs-new-manifests-path "./manifests/env2"
```

### batch by priorities

Kahoy knows how to manage priorities. By default it will batch all the manifests with a default priority (`1000`), but maybe you want to deploy some groups first (e.g CRDs or the NS.

you would have `kahoy.yml` on your repo root (or any other path and use `--config-file`), with the group options:

```yaml
version: v1

# Groups configuration.
groups:
  - id: crd
    priority: 200
  - id: ns
    priority: 100
  - id: system/roles
    priority: 300
```

This will make Kahoy apply first the `ns` group, then `crd` group, then `system/roles` group, and finally the rest.

### Kustomize and Kahoy

TODO

## :question: F.A.Q

- [Can I deploy anything?](#can-i-deploy-anything)
- [What does focused on resource level mean?](#what-does-focused-on-resource-level-mean)
- [What about CRDs?](#what-about-crds)
- [Why old and new states?](#why-old-and-new-states)
- [Why file changes don't affect resources?](#why-file-changes-dont-affect-resources)
- [Can I have multiple manifests envs on the same repository?](#can-i-have-multiple-manifests-envs-on-the-same-repository)
- [Partial and full syncs?](#partial-and-full-syncs)
- [How is Garbage collection handled?](#how-is-garbage-collection-handled)
- [Why git?](#why-git)
- [When to use paths mode?](#when-to-use-paths-mode)
- [Env vars as options](#env-vars-as-options)
- [Kustomize or helm manifests](#kustomize-or-helm-manifests)
- [Encrypted secrets?](#encrypted-secrets)
- [Non resource YAMLs](#non-resource-yamls)
- [Ignore a resource](#ignore-a-resource)
- [Github actions integration](#github-actions-integration)
- [Configuration file](#configuration-file)

### Can I deploy anything?

Yes, Kahoy is focused on resource level, you will not need any app scope, labels to group them, or anything similar. You can deploy from 500 apps to 1 namespace.

### What does focused on resource level mean?

When we talk about resource level, it means that Kahoy identifies what to deploy/delete based on the Kubernetes resource ID (type + ns + name).

Other solutions add concepts like release ([Helm]) or app ([Kapp]), these use/add special fields like labels to identify them.

Not depending on these fields gives Kahoy, flexibility to deploy anything, and not depending on anything external to what the user defines in its manifests.

However, if you want to group them by app/release, you can always generate these manifests using helm or kustomize templating and let them add those grouping labels, Kahoy will handle correctly the manifests/resources as they are.

### What about CRDs?

CRDs are also Kubernetes resources, Kahoy knows how to handle them.

### Why old and new states?

In order to be able to track resource changes (e.g track deletion of resource). We need a way to detect what resources have been created/changed/gone.

As we explained in another question, Kahoy doesn't depend on special concepts like release/app/special labels... so to be able to track this we need to use something else, we use a previous (`old`) state and a current (`new`) state of the manifests, we compare them and then we can track what has changed.

Normally the manifests are already in Git, Git has history, so getting this information from git is enough and integrates perfectly with code review flows and Kahoy.

### Why file changes don't affect resources?

Because Kahoy loads resources and then plans what has changed, e.g:

You have a file called `app.yaml` and has these resources with the IDs:

- A service called `app1` on the ns `apps`: `v1/Service/apps/app1`
- A Deployment called `app1` on the ns `apps`: `apps/v1/Deployment/apps/app1`
- An ingress called `app1` on the ns `apps`: `networking.k8s.io/v1beta1/Ingress/apps/app1`

Now you split the file in

- `deployment.yaml`: `apps/v1/Deployment/apps/app1`
- `svc.yaml`: `v1/Service/apps/app1`
- `ingress.yaml`: `networking.k8s.io/v1beta1/Ingress/apps/app1`

For Kahoy internally, are the same.

### Can I have multiple manifests envs on the same repository?

Yes, Kahoy takes a root manifest-path, as long as that root is the one for the environment, it should be ok.

You can invoke Kahoy `N` times, one per environment.

### Partial and full syncs?

Yes, partial syncs will apply only the changes from one git revision to another (`git diff`). full syncs apply all the repository.

Check this [Github actions example][github-actions-example] for more info.

### How is Garbage collection handled?

Kahoy takes manifests in 2 states, an `old` state, and a `new` state. It compares both and checks what's missing in the `new` one comparing the `old` one. Those are the resources that will be deleted (garbage collected).

### Why git?

Git maintains history of the manifests, it tracks the changes, can be reverted, is known by almost everyone... this makes the manifests lifecycle to be reliable.

This gives us the opportunity to track changes on our resources, applying a reliable flow based on code reviews (Pull requests).

That's why Kahoy understands git, knows how to get two revisions, and compares the manifests that changed in those revisions, plan them and apply.

Also having the ability to apply only the changes of a `git diff`, gives us a way of scaling better, more visibility and easy to integrate

### When to use paths mode?

Kahoy understands git and most of the time you will not need it if you are using a repository. However, if you want to make everything yourself, using `paths` mode gives you full control. e.g:

- Prepare two manifest paths.
  - `new` manifests is the main repository
  - `old` manifests is a copy of `new` (`cp -r`) and checkout to a previous revision.
- Use `--mode=paths` to pass those manifest paths (`--fs-old-manifests-path`, `--fs-new-manifests-path`) to the two repo paths in different states.
- If you want to only apply on changes, use git diff to make N arguments with the option `--fs-include`.

Check an example [script][bash-git-example] that prepares a git diff file and the two manifests paths with the different revisions.

### Env vars as options

You can use environment vars as options using `KAHOY_XXXX_XXXX`, cli args have priority. e.g:

- `--debug`: `KAHOY_DEBUG`
- `--kube-context`: `KAHOY_KUBE_CONTEXT`
- `--mode`: `KAHOY_MODE`
- `--fs-include`: `KAHOY_FS_INCLUDE`
- ...

### Kustomize or helm manifests

You can maintain the generated manifests in git as a previous step to make the PRs, this would make that the final autogenerated manifests are committed and ready in the git history, ready to be used by Kahoy at any time (including CI) and cleaner on the PRs when multiple manifests change.

### Encrypted secrets?

Encrypted secrets can't be understood by Kahoy, there are different solutions:

- Ignore encrypted files and apply them separately.
  - Invoke Kahoy ignoring them using `--fs-exclude`.
  - Decrypt the secrets.
  - Apply them using Kahoy with `--mode=paths` and `--fs-include` option.
- Move to a different solution where git repository doesn't have encrypted secrets (webhooks, controllers...).

### Non resource YAMLs

Kahoy will try loading all yamls as resources, if it fails, Kahoy will fail, this can be a problem when you have yamls that are not Kubernetes resources.

Use `--fs-exclude`, it works with `paths` and `git` modes.

### Ignore a resource

You can ignore resources at different levels and using multiple filters.

At file level you have `--fs-include` and `--fs-exclude`, these exclude or include based on filesystem path regexes.

At Kubernetes resource level you have others:

- `--kube-exclude-type`: Exclude based on Kubernetes type regex (e.g: `apps/*/Deployment, v1/Pod`)
- TODO more of them

### Github actions integration

Check this [Github actions example][github-actions-example] for more info.

### Configuration file

Kahoy accepts a configuration file (by default `./kahoy.yml`) to set options, at this moment these are the options:

```yaml
version: v1

# List of groups configuration.
groups:
  - id: crd # Identifies by group ID.
    priority: 200 # Adds apply priority to all the resources in a group (by default `1000`).
```

## :tophat: Alternatives

Kahoy born because available alternatives are too complex, Kubernetes is a complex system by itself, adding more complexity in the cases where is not needed, is not a good solution.

- [Helm]: Tries solving other kinds of problems, has templating (v2 tiller), the concept of releases, used to deploy single apps... However, you can use helm for templating and kahoy to deploy the generated manifests.
- [Kustomize]: Similar scope as helm but with a different approach, like Helm, you can use kustomize for the templating and kahoy for deploying raw manifests.
- [Kapp]: As Kahoy, tries solving the same problems of complexity that come with Helm, Kustomize... Very similar to Kahoy but with more complex options/flows, Kapp focuses on application level, Kahoy on Kubernetes resources, if you need something more complex than Kahoy, is likely that Kapp is your app.
- [Flux]: Controller-based flow, very powerful but complex. If you want a more `pull` than `push` approach, maybe you want this.
- [Kubectl]: Official tool. Is what kahoy uses under the hood, very powerful tool, lots of options, although to make it work correctly with a group of manifests/repository... you will need scripting or something like Kahoy. _We could say that Kahoy is a small layer on top of Kubectl_.

[ci-image]: https://github.com/slok/kahoy/workflows/CI/badge.svg
[ci-url]: https://github.com/slok/kahoy/actions
[goreport-image]: https://goreportcard.com/badge/github.com/slok/kahoy
[goreport-url]: https://goreportcard.com/report/github.com/slok/kahoy
[gh-context]: https://docs.github.com/en/actions/reference/context-and-expression-syntax-for-github-actions#github-context
[gitlab-ci-env]: https://docs.gitlab.com/ee/ci/variables/predefined_variables.html
[helm]: https://helm.sh/
[kustomize]: https://github.com/kubernetes-sigs/kustomize
[kapp]: https://github.com/k14s/kapp
[flux]: https://github.com/fluxcd/flux
[kubectl]: https://kubernetes.io/docs/reference/kubectl/overview/
[serverside-apply]: https://kubernetes.io/blog/2020/04/01/kubernetes-1.18-feature-server-side-apply-beta-2/#what-is-server-side-apply
[github-actions-example]: https://github.com/slok/kahoy-github-actions-example
[bash-git-example]: https://gist.github.com/slok/3f37c2a0dd823d5b66db869a468109ce
