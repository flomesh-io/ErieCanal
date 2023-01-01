# syntax = docker/dockerfile:1.4
ARG TARGETOS
ARG TARGETARCH
ARG DISTROLESS_TAG

# Build the proxy-init binary
FROM --platform=$BUILDPLATFORM golang:1.19 as builder

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
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH make build/proxy-init

FROM gcr.io/distroless/static-debian11:$DISTROLESS_TAG
WORKDIR /
COPY --from=builder /workspace/bin/proxy-init .
USER 65532:65532

ENTRYPOINT ["/proxy-init"]