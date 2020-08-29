# Kahoy [![Build Status][ci-image]][ci-url] [![Go Report Card][goreport-image]][goreport-url]

_When [Kubectl] is to simple for your needs and available deployment solutions out there too complex_

Maintain Kubernetes resources in sync easily.

![kahoy run example](docs/img/kahoy.gif)

---

## :link: Table of contents

- [:tada: Introduction](#tada-introduction)
- [:checkered_flag: Features](#checkered_flag-features)
- [:shipit: Install](#shipit-install)
- [:mag: Scope](#mag-scope)
- [:pencil2: Concepts](#pencil2-concepts)
- [:wrench: How does it work](#wrench-how-does-it-work)
- [:computer: Execution options](#computer-execution-options)
- [:page_facing_up: Manifest source modes](#page_facing_up-manifest-source-modes)
- [:bulb: Use cases](#bulb-use-cases)

## :tada: Introduction

Kahoy is a minimal and flexible tool to sync/deploy your Kubernetes resource **raw** manifests and a cluster.

**Focuses on Gitops and Kubernetes resources, not apps/releases/services/whatever.**

Unlike other tools, Kahoy will adapt to your needs and not the other way around, its been designed and developed to be generic and flexible enought without adding unneed complexity.

## :checkered_flag: Features

- Simple, flexible and lightweight.
- Plans `apply`s and `delete`s based on manifests (fs, git...) state.
- Plans at Kubernetes resource level, not file/manifest level (no more resource deletions because files changed name).
- Gitops ready (understands git history to plan, plan based on git diffs...).
- Easy to integrate with an existing git repository of raw Kubernetes manifests.
- Easy to integrate with a file system path of raw Kubernetes manifests.
- Diff, Dry run, apply... (Can be executed separetly, helpful for CI state/jobs).
- Easy to set up on CI (Github actions, Gitlab CI...).
- Deploy priorities.
- Don't depend on concepts as application, releases, service... Just Kubernetes resource (CRDs included).
- Lots of filter options (file paths, resource namespace, types...).

## :shipit: Install

- Docker: `docker pull slok/kahoy` (has all the tools required like `kubectl` and `git`).
- Releases: Go to [releases](https://github.com/slok/kahoy/releases).
- Build from source: Clone the repo and `make build`.

## :mag: Scope

- No templating, the generation and mutation of the YAMLs is out of the scope (use other tools and then Kahoy, e.g kustomize+kahoy).
- Manage lifecycle of the Kubernetes resources (Deployment and deletion of resources) using raw YAML files.
- Focus on Gitops and CI step/jobs (dry run, diff, apply).
- Simplicity and flexibility.
- Just a bit smarter than Kubectl.
- Plan what will be applied and what deleted based in an old and a new manifest state (fs, git...).

If you need complex flows for your Kubernetes resources is likely that Kahoy is not for you.

## :pencil2: Concepts

Kahoy doesn't depend on app/service, labels/selectors or any other kind of app grouping concept:

### State

Kahoy plans what to apply or delete based on an `old` and a `new` state of manifests. These state can come from different sources:

- `paths`: Given 2 filesystem paths, it will use one for the old state and the other for the new state.

- `git`: Given 2 revisions (depends on the `git` mode), it will use the old git revision to get the manifests state in that moment, and the new git revision to get the manifests state on that moment.

### Resource

Is a Kubernetes resource, Kahoy will identify resources by type, ns and name, so, if the manifests file arrangement changes (grouping in files, splitting, rename...) will not affect at the plan.

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

These would be the groups IDs:

- `alertgram`
- `bilrost`
- `root` (this is the root id).
- `grafana`
- `grafana/grafana-dashboards`

**Note: Because resources are identified by its `type`, `ns` and `name`, you can move around in files without affecting on how Kahoy will identify them**

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

Will plan what resources need to exist on the cluster and what need to be removed (client-side, no cluster required).

![dry run](docs/img/dry-run.png)

### Diff

Will get a diff against the server of the planned resources (server-side, cluster required).

![diff](docs/img/diff.png)

### Default (Apply)

Will apply the resources that need to exist and delete the ones that don't.

## :page_facing_up: Manifest source modes

Kahoy needs two manifest states (old and new) to plan what resources need to exist/gone in the cluster. How these manifests are provided is using the `mode`

### `paths` (File system)

given 2 manifest file system paths, plans that needs to be applied against a cluster and what needs to be deleted.

This one is the most generic one and can be used when you want to manage almost everyhting, e.g previous Kahoy execution, prepare using bash scripts, kustomize, secrets...

### `git`

This is the best one for **gitops**.

This is the default mode, this mode understands git and can read states from a git repository, these 2 states are based on 2 git revisions.

Using `before-commit` will make a plan based on the manifests of `HEAD` (new state) and the commit provided (old state). Normally used when executed from `master/main` branch.

Instead of providing the `before-commit`, by default will get the base parent of the current branch `HEAD` (new state) against the default branch (old state), normally `master/main`). This mode is used when you are executing kahoy from a branch in a pull request.

Apart from knowing how to get an old and a new state from a git repository. **Git mode understands diff/patches**, this would make Kahoy only be applied what has been changed between these two revisions/commits. This is interesting in many cases:

- When you have lost of resources:
  - Have a clear view of what is changing.
  - Bazing fast deployments.
- Operators sometimes change deployed manifests, this mode avids overwriting every manifest on each deployment.
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
In this example we use `--git-before-commit-sha` flag. Normally this variable can be obtained in the executing CI:

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

TODO

### Exclude some manifest files (e.g encrypted secrets)

If you have some files that you don't want to me managed by kahoy, you can ignore them at file system level using `--fs-exclude`. Can be repeated.

E.g: exclude any file with the name secret on it.

```bash
kahoy apply \
    --fs-new-manifests-path "./manifests" \
    --fs-exclude "secret"
```

### Delete all

Instead using git, use the fs by using the `paths` mode. Use the new state as `dev/null`.

```bash
kahoy apply \
    --mode="paths" \
    --fs-old-manifests-path "./manifests" \
    --fs-new-manifests-path "/dev/null"
```

### Deploy all

Instead using git, use the fs by using the `paths` mode. Use the old state as `dev/null`.

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

Kahoy knows how to manage priorities. By defualt it will batch all the manifests with default priority (`1000`), but maybe you want to deploy some groups first (e.g CRDs or the NS.

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

### Alternatives

Kahoy born because available alternatives are too complex, Kubernetes is a complex system by itself, adding more complexity in the cases where is not needed, is not a good solution.

- [Helm]: Tries solving other kind of problems, has templating (v2 tiller), concept of releases, used to deploy single apps... However, you can use helm for templating and kahoy to deploy the generated manifests.
- [Kustomize]: Similar scope as helm but with a different approach, like Helm, you can use kustomize for the templating and kahoy for deploying raw manifests.
- [Kapp]: As Kahoy, tries solving the same problems of complexity that come with Helm, Kustomize... Very similar to Kahoy but with more complex options/flows, Kapp focuses on application level, Kahoy on Kubernetes resources, if you need something more complex than Kahoy, is likely that Kapp is your app.
- [Flux]: Controller based flow, very powerfull but complex. If you want a more `pull` than `push` approach, maybe you want this.
- [Kubectl]: Official tool. Is what kahoy uses under the hood, very powerfull tool, lots of options, although to make it work correctly with a group of manifests/repository... you will need scripting or something like Kahoy. _We could say that Kahoy is a small layer on top of Kubectl_.

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
