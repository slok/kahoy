# How to Contribute

Kahoy is [Apache 2.0 licensed](LICENSE) and accepts contributions via GitHub
pull requests. This document outlines some of the conventions on development
workflow, commit message formatting, contact points and other resources to make
it easier to get your contribution accepted.

We gratefully welcome improvements to issues and documentation as well as to code.

## Getting Started

- Fork the repository on GitHub
- Read the [README](README.md#key-getting-started) for getting started.
- If you want to contribute as a developer, continue reading this document for further instructions
- Play with the project, submit bugs, submit pull requests!

## Contribution workflow

This is a rough outline of how to prepare a contribution:

- Fork the repository.
- Create a topic branch from where you want to base your work (usually branched from master).
- Make commits of logical units.
- Make sure your commit messages are clear and self-explanatory.
- Push your changes to a topic branch in your fork of the repository.
- If you changed code, add automated tests to cover your changes.
- Submit a pull request from your fork to the original repository.

## Automated checks

You can check your code satisfies some standards by using:

```bash
make check
```

You can run the unit tests by doing:

```bash
make test
```

## Manual checks

Unit tests are the way to go, however sometimes you want to check how you development is integrated with a real Kubernetes cluster (e.g Check Kubernetes versions, a cli option...). To make this testing easy, there are some files that will help you with it.

### Prepare cluster

Optional step if you already have one. We are going to use [Kind] to create a new cluster

```bash
kind create cluster --name kahoy
```

You will have a new kubectl context named `kind-kahoy`. Switch to it with

```bash
kubectl config set-context kind-kahoy
```

### Run tests

**Before starting you may want to export the kubectl context setter for kahoy: `export KAHOY_KUBE_CONTEXT=kind-kahoy`**

Kahoy comes with a handy set of use cases for manual testing. You can check them in [`tests/manual`][manual-tests]. These tests come with a `run.sh` file that will run the different use cases with some predefined flags for Kahoy.

How these use cases work is simple, they use `--mode=paths`, and we have different manifest directories in different states.

- [`manifests-all`][manifests-all]: It has all the resources.
- [`manifests-shuffle`][manifests-shuffle]: It has all resources but with different structure.
- [`manifests-some`][manifests-some]: It has some of the `manifest-all`, others have been deleted.
- [`manifests-some-changed`][manifests-some-changed]: Like `manifest-some` but has some resource fields changed.

By combining these paths with `new` and `old` states, we get common use cases.

> Tip: If you want to add more args, you can use Kahoy env vars, this is helpful so you don't need to edit the `run.sh` script args. but remember that you can always edit the run.sh to test things also. Lets see some examples:

- Apply all on the cluster: `./tests/manual/run.sh all`
- Apply all on the cluster with debug: `KAHOY_DEBUG=true ./tests/manual/run.sh all`
- Apply all on the cluster dry-run: `KAHOY_DRY_RUN=true ./tests/manual/run.sh all`
- Apply all on the cluster diff: `KAHOY_DIFF=true ./tests/manual/run.sh all`
- Apply all with only changes: `KAHOY_INCLUDE_CHANGES=true ./tests/manual/run.sh all`

From now on you get the idea of playing with the kahoy options and using the script. Now lets see what use cases come with the script:

#### Use cases

**Remember that you can modify the commands with dry-run, diff, debug, only changes...**

As we saw earlier, `all` applies all resources

```bash
./tests/manual/run.sh all
```

`shuffle` is like `all`, but shuffles the structure without changing it, we can check that our resources didn't change if the structure changes

```bash
# The obtained diff should be empty (well maybe some timestamp or internal managed field of Kubernetes).
KAHOY_DIFF=true ./tests/manual/run.sh shuffle
```

With `some` we can test that deletions work after applying `all`.

```bash
# This should show that some of the manifests have been deleted and needs to be deleted from the cluster.
KAHOY_DRY_RUN=true ./tests/manual/run.sh some
```

With `changed` we can test deletions and changes at the same time (handy with `include-changes` option). Check these two:

```bash
# Check that all resources will be handled (apply and delete).
KAHOY_DRY_RUN=true KAHOY_INCLUDE_CHANGES=false ./tests/manual/run.sh changed

# Check that only resources that change (apply) will be handled along with the ones to be deleted.
KAHOY_DRY_RUN=true KAHOY_INCLUDE_CHANGES=true ./tests/manual/run.sh changed
```

Finally to remove all, use `none`, it will compare `all` against `/dev/null`, so will delete everything. Check this sequence

```bash
# Diff the deletions.
KAHOY_DIFF=true ./tests/manual/run.sh none

# Apply deletions.
KAHOY_DIFF=false ./tests/manual/run.sh none

# Check deletions again, should return nothing.
KAHOY_DIFF=true ./tests/manual/run.sh none
```

Combining all of these, you can check losts of use cases working with a real Kubernetes cluster.

[kind]: https://github.com/kubernetes-sigs/kind
[manual-tests]: tests/manual
[manifests-all]: tests/manual/manifests-all
[manifests-shuffle]: tests/manual/manifests-shuffle
[manifests-some]: tests/manual/manifests-some
[manifests-some-changed]: tests/manual/manifests-some-changed
