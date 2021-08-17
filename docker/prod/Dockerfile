FROM golang:1.17.0-alpine as build-stage

RUN apk --no-cache add \
    g++ \
    git \
    make \
    curl \
    bash

ARG VERSION
ENV VERSION=${VERSION}

# Compile.
WORKDIR /src
COPY . .
RUN ./scripts/build/build.sh

# Get kubectl.
ARG KUBERNETES_VERSION="v1.21.2"
RUN curl -LO https://storage.googleapis.com/kubernetes-release/release/${KUBERNETES_VERSION}/bin/linux/amd64/kubectl && \
    chmod a+x kubectl && \
    mv kubectl /usr/bin/

# Final image with common utils that work along Kahoy to help the use of it,
# however this will give us a bigger image.
FROM alpine:latest

RUN apk --no-cache add \
    ca-certificates \
    bash \
    make \
    git \
    colordiff

COPY --from=build-stage /src/bin/kahoy /usr/local/bin/kahoy
COPY --from=build-stage /usr/bin/kubectl /usr/local/bin/kubectl

ENTRYPOINT ["/usr/local/bin/kahoy"]