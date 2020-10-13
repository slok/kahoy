---
title: "Paths"
weight: 339
---

Given 2 manifest file system paths, plans what needs to be applied against a cluster, and what needs to be deleted.

This one is the most generic one and can be used when you want to manage almost everything, e.g previous Kahoy execution, prepare using bash scripts, kustomize, secrets...

Example of usage:

```bash
kahoy apply \
  --provider "paths" \
  --fs-old-manifests-path "./old-manifests" \
  --fs-new-manifests-path "./manifests"
```
