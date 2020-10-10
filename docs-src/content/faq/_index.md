---
title: "F.A.Q"
weight: 999
---

{{< toc >}}

## Can I deploy anything?

Yes, Kahoy is focused on resource level, you will not need any app scope, labels to group them, or anything similar. You can deploy from 500 apps to 1 namespace.

## What does focused on resource level mean?

When we talk about resource level, it means that Kahoy identifies what to deploy/delete based on the Kubernetes resource ID (type + ns + name).

Other solutions add concepts like release ([Helm]) or app ([Kapp]), these use/add special fields like labels to identify them.

Not depending on these fields gives Kahoy, flexibility to deploy anything, and not depending on anything external to what the user defines in its manifests.

However, if you want to group them by app/release, you can always generate these manifests using helm or kustomize templating and let them add those grouping labels, Kahoy will handle correctly the manifests/resources as they are.

## What about CRDs?

CRDs are also Kubernetes resources, Kahoy knows how to handle them.

## Why old and new states?

In order to be able to track resource changes (e.g track deletion of resource). We need a way to detect what resources have been created/changed/gone.

As we explained in another question, Kahoy doesn't depend on special concepts like release/app/special labels... so to be able to track this we need to use something else, we use a previous (`old`) state and a current (`new`) state of the manifests, we compare them and then we can track what has changed.

Normally the manifests are already in Git, Git has history, so getting this information from git is enough and integrates perfectly with code review flows and Kahoy.

## Why file changes don't affect resources?

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

## Can I have multiple manifests envs on the same repository?

Yes, Kahoy takes a root manifest-path, as long as that root is the one for the environment, it should be ok.

You can invoke Kahoy `N` times, one per environment.

## Partial and full syncs?

Partial syncs filter the resources that will apply based on the changes from one state to another (checks diffs between kubernetes resources in both states). Use `--include-changes` for partial syncs.

Full syncs apply all the resources.

Check this [Github actions example][github-actions-example] for more info.

## How is Garbage collection handled?

Kahoy takes manifests in 2 states, an `old` state, and a `new` state. It compares both and checks what's missing in the `new` one comparing the `old` one. Those are the resources that will be deleted (garbage collected).

The deletion is made in a resource manner using `Kubectl delete`, this is safe because Kahoy selects what wants to delete so, it already knows what is going to be delete on the server.

Other methods like `prune` are not safe, and that's why Kahoy doesn't use them.

## Why git?

Git maintains history of the manifests, it tracks the changes, can be reverted, is known by almost everyone... this makes the manifests lifecycle to be reliable.

This gives us the opportunity to track changes on our resources, applying a reliable flow based on code reviews (Pull requests).

That's why Kahoy understands git, knows how to get two revisions, and compares the manifests that changed in those revisions, plan them and apply.

## When to use paths provider?

Kahoy understands git and most of the time you will not need it if you are using a repository. However, if you want to make everything yourself, using `paths` provider gives you full control. e.g:

- Prepare two manifest paths.
  - `new` manifests is the main repository
  - `old` manifests is a copy of `new` (`cp -r`) and checkout to a previous revision.
- Use `--provider=paths` to pass those manifest paths (`--fs-old-manifests-path`, `--fs-new-manifests-path`) to the two repo paths in different states.
- If you want to only apply on changes, use `--include-changes`.

Check an example [script][bash-git-example] that prepares two manifests paths with the different revisions.

## When to use Kubernetes provider?

When you want Kahoy manage the latest state for you instead of you managing the latest state (e.g: Using Git history).

This will make an easy and reliable way of managing the state.

## Env vars as options

You can use environment vars as options using `KAHOY_XXXX_XXXX`, cli args have priority. e.g:

- `--debug`: `KAHOY_DEBUG`
- `--kube-context`: `KAHOY_KUBE_CONTEXT`
- `--provider`: `KAHOY_PROVIDER`
- `--fs-include`: `KAHOY_FS_INCLUDE`
- ...

## Kustomize or helm manifests

You can maintain the generated manifests in git as a previous step to make the PRs, this would make that the final autogenerated manifests are committed and ready in the git history, ready to be used by Kahoy at any time (including CI) and cleaner on the PRs when multiple manifests change.

Check this [Kustomize example][kustomize-example].

## Encrypted secrets?

Encrypted secrets can't be understood by Kahoy, there are different solutions:

- Ignore encrypted files and apply them separately.
  - Invoke Kahoy ignoring them using `--fs-exclude`.
  - Decrypt the secrets.
  - Apply them using Kahoy with `--provider=paths` and `--fs-include` option.
- Move to a different solution where git repository doesn't have encrypted secrets (webhooks, controllers...).

## Non resource YAMLs

Kahoy will try loading all yamls as resources, if it fails, Kahoy will fail, this can be a problem when you have yamls that are not Kubernetes resources.

Use `--fs-exclude`, it works with `paths` and `git` providers.

## Ignore a resource

You can ignore resources at different levels and using multiple filters.

At file level you have `--fs-include` and `--fs-exclude`, these exclude or include based on filesystem path regexes.

At Kubernetes resource level you have others:

- `--kube-exclude-type`: Exclude based on Kubernetes type regex (e.g: `apps/*/Deployment`, `v1/Pod`...).
- `--kube-include-label`: Kubernetes style selector that will select only the resources that match the label selector (e.g: `app=myapp,component!=database,env`)
- `--kube-include-annotation`: Kubernetes style selector that will select only the resources that match the annotation selector (e.g: `app=myapp,component!=database,!non-wanted-key`)

## Why so many filtering options?

There isn't a correct manifest structure, grouping, naming... These can be, spliting a repo per env, a monorepo for everything, a repo for each app...

Kahoy tries adapting to most use cases, so, having multiple ways of including/excluding resources/manifests is a good way of adapting to the different users.

All these filtering options give users a way of solving lots of use cases, for example splitting CI deployments in many ways, e.g:

- Deploy single env on a monorepo identified by paths (`--fs-include envs/prod`).
- Deploy single env on a monorepo on the same path, identified by labels, e.g Kustomize generated files (`--kube-include-label env=prod`).
- Split CI steps by env.
- Split CI steps by nature (Kahoy with everything except secrets -> decrypt secrets -> Kahoy all secrets).
- Updating a single service (`--fs-include apps/app1`).
- Exclude encrypted files (`--fs-exclude secret`)
- Exclude an specific app (`--fs-exclude apps/app1`)
- Ignore CRDs that have an annotation, becase controller change the information (`--kube-include-annotation ...`)
- Integrate Kahoy gradually including manifests (`--fs-include monitoring/grafana --fs-include monitoring/prometheus`)
- ...

## I have namespace not found error on regular apply or diff

When we apply a namespaced resource on a namespace that does not exists, the action will fail with an error like:

```text
Error from server (NotFound): namespaces "some-namespace" not found
```

This happens when you don't apply/create the `Namespaces` before the namespaced Kubernetes resources. Or when you do a diff `kahoy --diff`. Kahoy uses server-side diff, so it will try a fake/dry-run apply to get the diff and because there is no namespace, it will fail.

This is a tricky [known problem](https://github.com/kubernetes/kubernetes/issues/83562), to solve this at this moment is only one option, and is to create the namespace before the server-side apply.

By default Kahoy will not create the missing namespaces of applied resources, but with `--create-namespace`, it will. This will work with regular mode `kahoy apply` and also diff mode `kahoy apply --diff`.

**Be aware using it with `--diff` would create a namespace, this means that the diff would have a write operation on the cluster, the ns creation.**

Is a good practice that if you use `--create-namespace`, you add to your resources the `Namespace` manifest, this way in the case you delete anytime the resources along with the namespace, the created namespace will be garbage collected by Kahoy.

## Why don't use kubectl `prune` to delete resources?

TL;DR: Is unpredictable, then risky.

- [Official documentation][kubectl-delete-docs] discourages `--prune`.
- [Official documentation][kubectl-delete-docs] encourages `delete -f`.
- You never know what will be deleted exactly beforehand.
- Can delete resources that we didn't even know they exists (because the selector matches).
- Can have a big blast radius when an error is made in the `prune` execution.
- Some controllers/operators create resources and set the labels with the ones from the original resource, this would make prune delete the controller object on each `apply` with `prune`.

## Github actions integration

Check this [Github actions example][github-actions-example] for more info.

## Configuration file

Kahoy accepts a configuration file (by default `./kahoy.yml`) to set options, at this moment these are the options:

```yaml
# Version of the configuration format.
version: v1

# File system configuration.
fs:
  # Exclude regex for file paths (same as `--fs-exclude`, can be used both).
  exclude:
    - prometheus/secrets
    - secret*
  # Include regex for file paths (same as `--fs-include`, can be used both).
  include:
    - apps/

# List of groups configuration.
groups:
  # Represented by the group ID
  - id: crd
    # Priority of the group (by default is 1000). Applied in asc order.
    priority: 200
    # Wait options.
    wait:
      # The time will wait after being applied (Ts, Tm, Th format).
      duration: 5s

  - id: ns
    priority: 100
    wait:
      duration: 10s

  - id: system/roles
    priority: 300
```

## Report

Kahoy can give a report at the end of the execution with the information of the resources that have been deleted and applied.

This is very flexible and powerful because it gives the ability to plug new apps after Kahoy execution e.g:

- Push notifications
- Wait for resources be available: [Example][wait-example].
- Push metrics.
- Execute sanity checks
- ...

This approach follows unix philosophy of having N tools, each one doing one thing (e.g `Kahoy | jq | waiter`).

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

[bash-git-example]: https://gist.github.com/slok/3f37c2a0dd823d5b66db869a468109ce
[kustomize-example]: https://github.com/slok/kahoy-kustomize-example
[kubectl-delete-docs]: https://kubernetes.io/docs/tasks/manage-kubernetes-objects/declarative-config/#how-to-delete-objects
[github-actions-example]: https://github.com/slok/kahoy-github-actions-example
[wait-example]: https://github.com/slok/kahoy-app-deploy-example
[jq]: https://stedolan.github.io/jq/
[helm]: https://helm.sh/
[kapp]: https://github.com/k14s/kapp
