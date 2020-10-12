---
title: "Installing Kahoy"
weight: 030
---

## Releases

You have ready to be used binaries in Github [releases].

To get the latest release, grab it [here][latest-release].

## Docker

In case you want to use Docker, you have a ready to be used image on [docker hub][docker-hub], this image has Kahoy and some dependencies like `kubectl` and `git`.

```bash
docker pull slok/kahoy
```

## Build from source

You can build binaries from source easily.

```bash
git clone git@github.com:slok/kahoy.git

cd ./kahoy

make build
```

[releases]: https://github.com/slok/kahoy/releases
[latest-release]: https://github.com/slok/kahoy/releases/latest
[docker-hub]: https://hub.docker.com/repository/docker/slok/kahoy
