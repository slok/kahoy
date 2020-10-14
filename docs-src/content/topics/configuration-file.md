---
title: "Configuration file"
weight: 365
---

Most of kahoy options are based on flags (and env vars), however it also has some options that are based in a configuration file. The file options.

By default it will look for `./kahoy.yml`, yet you can configure using `--config-file` flag.

Lets see what are the available options based on an example:

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
