---
title: "Installing Kahoy"
weight: 030
---

## Releases

Kahoy is compiled for all common platforms and uploaded to Github [releases].

To get the latest release, grab it [here][latest-release].

## Docker

In case you want to use Docker, you have a ready to be used image on [docker hub][docker-hub]

```bash
docker pull slok/kahoy
```

Note: This image has Kahoy and other dependencies like `kubectl` and `git`.

## Build from source

You can build binaries from source easily.

```bash
git clone git@github.com:slok/kahoy.git

cd ./kahoy

make build
```

[releases]: https://github.com/slok/kahoy/releases
[latest-release]: https://github.com/slok/kahoy/releases/latest
[docker-hub]: https://hub.docker.com/r/slok/kahoy
