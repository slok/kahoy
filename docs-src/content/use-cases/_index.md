---
title: "Use cases"
weight: 600
---

{{< toc >}}

## Dry run on pull requests

When someone pushes a branch different from master, would be nice to execute Dry run on the changes (partial sync) of that branch.

Having our manifest in a git repository in `./manifests` and being our default branch `master`, we do this from a different branch.

```bash
kahoy apply \
    --dry-run \
    --include-changes \
    --fs-new-manifests-path "./manifests"
```

## Diff for CI Pull requests

Same as above but with `--diff` instead of `--dry-run`

## Deploying on master branch (when PR merged)

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
    --include-changes \
    --git-before-commit-sha "${GIT_BEFORE_COMMIT_SHA}" \
    --fs-new-manifests-path "./manifests"
```

Note: For Gitlab CI, this uses [env vars][gitlab-ci-env] (in `CI_COMMIT_BEFORE_SHA`).

## Schedule a full sync

Check this [Github actions example][github-actions-example] for more info.

## Exclude some manifest files (e.g encrypted secrets)

If you have some files that you don't want to be managed by kahoy, you can ignore them at file system level using `--fs-exclude`. Can be repeated.

E.g: exclude any file with the name secret on it.

```bash
kahoy apply \
    --fs-new-manifests-path "./manifests" \
    --fs-exclude "secret"
```

## Delete all

Instead of using git, use the fs by using the `paths` provider. Use the new state as `dev/null`.

```bash
kahoy apply \
    --provider="paths" \
    --fs-old-manifests-path "./manifests" \
    --fs-new-manifests-path "/dev/null"
```

## Deploy all

Instead of using git, use the fs by using the `paths` provider. Use the old state as `dev/null`.

```bash
kahoy apply \
    --provider="paths" \
    --fs-old-manifests-path "/dev/null" \
    --fs-new-manifests-path "./manifests"
```

## Deploy only some manifests

You can use the file filtering option `--fs-include`, works with any provider (`git`, `paths`...)

```bash
kahoy apply \
    --fs-new-manifests-path "./manifests" \
    --fs-include "./manifests/prometheus" \
    --fs-include "./manifests/grafana"
```

## Multiple envs

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

## batch by priorities

Kahoy knows how to manage priorities between groups. By default it will batch all the manifests with a default priority (`1000`), but maybe you want to deploy some groups first (e.g CRDs or the NS).

Given this `kahoy.yml` on your repo root (or any other path and use `--config-file`), with the group options:

```yaml
version: v1

groups:
  - id: crd
    priority: 200

  - id: ns
    priority: 100

  - id: system/roles
    priority: 300
```

it will make Kahoy apply first the `ns` group, then `crd` group, then `system/roles` group, and finally the rest.

## Kustomize and Kahoy

Check this [Kustomize example][kustomize-example].

[gh-context]: https://docs.github.com/en/actions/reference/context-and-expression-syntax-for-github-actions#github-context
[gitlab-ci-env]: https://docs.gitlab.com/ee/ci/variables/predefined_variables.html
[github-actions-example]: https://github.com/slok/kahoy-github-actions-example
[kustomize-example]: https://github.com/slok/kahoy-kustomize-example
