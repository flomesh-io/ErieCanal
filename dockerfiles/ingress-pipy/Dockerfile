# syntax = docker/dockerfile:1.4
ARG TARGETOS
ARG TARGETARCH
ARG DISTROLESS_TAG

# Build the ingress-pipy binary
FROM --platform=$BUILDPLATFORM golang:1.19-alpine3.15 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN --mount=type=cache,target=/go/pkg go mod download
# Copy the go source
COPY . .

# Build
RUN apk add make bash git
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH make build/ingress-pipy

# Builder image, pipy
FROM flomesh/pipy:0.70.0-46-nonroot as pipy

# Build the final image
FROM gcr.io/distroless/base-debian11:$DISTROLESS_TAG
WORKDIR /
COPY --from=builder /workspace/bin/ingress-pipy .
COPY --from=pipy /usr/local/bin/pipy /usr/local/bin/pipy

USER 65532:65532

ENTRYPOINT ["/ingress-pipy"]
