---
title: "Git provider"
weight: 336
---

This provider understands git and can read states from a git repository, these 2 states are based on 2 git revisions.

Using `before-commit` will make a plan based on the manifests of `HEAD` (new state) and the commit provided (old state). Normally used when executed from `master/main` branch.

Instead of providing the `before-commit`, by default will get the base parent of the current branch `HEAD` (new state) against the default branch (old state), normally `master/main`). This provider is used when you are executing kahoy from a branch in a pull request.

Example of usage:

```bash
kahoy apply \
  --provider "git" \
  --git-before-commit-sha "b060762ef93bbe2d03e108d1788eb3505df519a3" \
  --fs-new-manifests-path "./manifests"
```
