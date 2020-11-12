# Changelog

## Unreleased

### Added

- Add `make` to docker image.

## [v2.1.0] - 2020-11-12

### Notes

After the cluster scoped resource IDs change (ignores namespaces), Kubernetes storage could duplicate some of the resources state in case your cluster scoped resources had a namespace set by error. Check the warnings of Kubernetes storage `cluster scoped resource has namespace set` message. After fixing your resources, identify Kahoy Kubernetes state and remove them manually using `kubectl delete secret ...`).

### Added

- Documentation page in https://docs.kahoy.dev/.
- `version` command.
- Override Kubectl path with `--kubectl-path` flag.
- Default 5 minute timeout for any apply operation.
- `--execution-timeout` flag for the apply method to override the default timeout
- `--logger` flag to set the logger type, available optios are: default, json and simple.
- `--apply-first` flag that inverts the actions order on resources, `apply` before `delete`.
- Kubernetes storage ID validates with the same requirements as a Kubernetes label value.
- Kahoy checks with the apiserver (using discovery API) if the loaded resource type is known by the API and fail if not.
- Use `-` in `--fs-new-manifests-path` to load data from `stdin`.
- Support same kubectl external diff option behaviour on diff of deleted resources, using `KUBECTL_EXTERNAL_DIFF` env var.

### Changed

- Cluster scoped IDs ignore the namespace field for the kahoy resource ID.
- Kubernetes storage resources, manifest path data shows to the Kubernetes resource instead of the original fs manifest path.
- By default Kahoy will first delete and then apply resources. We find this way of execution safer than applying and then deleting. Alternatively an option has been added to invert this behaviour.

## [v2.0.0] - 2020-10-05

### Breaking

- `--kube-exclude-type` short flag changed from `-a` to `-t`, `--kube-include-annotation` is `-a` short flag.
- `--mode` flag renamed to `--provider`.
- Provider default is `kubernetes`.

### Added

- Optional filter apply/delete plan based on K8s resources that had changes from old to new state using `--include-changes` flag.
- Optional label based filter for resources using Kubernetes standard label selectors using `--kube-include-label` flag.
- Optional annotation filter for resources using Kubernetes standard label selectors using `--kube-include-annotation` flag.
- Load `metav1.List` YAML resources as individual resources.
- Allow groups waiting specific time after apply.
- `fs-include` and `fs-exclude` arg options to kahoy app global configuration file as an alternative.
- JSON report with the resources applied and deleted after the execution.
- Confirmation prompt when running `kahoy apply` without diff or dry-run modes enabled.
- Optional `--auto-approve` to disable the confirmation prompt.
- Optional `--create-namespace` on regular and diff modes that will create missing namespaces of applied resources.
- Kubernetes provider.
- Optional `--include-namespace` to only apply resources of given namespaces.

### Changed

- On Diff, deleted resources now show the real fields and resource the server will delete (before we didn't check the server state).
- Fix YAML failing on load when YAML file was multiresource and had files only with comments.
- Fix using current directory as the manifests path, loads all resources as root group.
- Capture correctly OS sigansl and stop safely command execution.
- Batch executions stop in the different batch executions if context is cancelled.
- Group wait now stops if the context is cancelled.
- On dry-run, groups are printed in order.

### Removed

- Git filtering in favor of generic filtering based on Kubernetes resource diff.
- `--git-diff-filter` flag in favor of `--include-changes`.

## [v1.0.0] - 2020-08-31

### Added

- Apply/delete resource Plan.
- File based filtering (include exclude).
- Kubernetes type filtering (exclude).
- Add States repositories (old and new).
- Paths provider (load from fs).
- Git provider (load form Git repository).
- Git filtering based on `git diff`.
- Git states based on previous commit or `git merge-base`.
- Grouping of resources.
- Group priority options.
- YAML configuration for Kahoy.
- Same resource ID validation.
- Dry run mode.
- Diff mode.
- No color mode.
- Debug mode.

[unreleased]: https://github.com/slok/kahoy/compare/v2.1.0...HEAD
[v2.1.0]: https://github.com/slok/kahoy/compare/v2.0.0...v2.1.0
[v2.0.0]: https://github.com/slok/kahoy/compare/v1.0.0...v2.0.0
[v1.0.0]: https://github.com/slok/kahoy/releases/tag/v1.0.0
