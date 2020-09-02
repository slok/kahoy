# Changelog

## [unreleased]

Breaking: `--kube-exclude-type` short flag changed from `-a` to `-t`, `--kube-include-annotation` is `-a` short flag.

### Added

- Optional filter apply/delete plan based on K8s resources that had changes from old to new state using `--include-changes` flag.
- Optional label based filter for resources using Kubernetes standard label selectors using `--kube-include-label` flag.
- Optional annotation filter for resources using Kubernetes standard label selectors using `--kube-include-annotation` flag.

### Changed

- Deprecate `--git-diff-filter` flag in favor of `--include-changes`.
- `--kube-exclude-type` short flag changed from `-a` to `-t`, `--kube-include-annotation` is `-a` short flag.

### Removed

- Git filtering in favor of generic filtering based on Kubernetes resource diff.

## [v1.0.0] - 2020-08-31

### Added

- Apply/delete resource Plan.
- File based filtering (include exclude).
- Kubernetes type filtering (exclude).
- Add States repositories (old and new).
- Paths mode (load from fs).
- Git mode (load form Git repository).
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

[unreleased]: https://github.com/slok/kahoy/compare/v1.0.0...HEAD
[v1.0.0]: https://github.com/slok/kahoy/releases/tag/v1.0.0
