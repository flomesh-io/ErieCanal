# syntax = docker/dockerfile:1.4
ARG DISTROLESS_TAG

# Build the gateway binary
FROM --platform=$BUILDPLATFORM golang:1.19 as builder
ARG TARGETOS
ARG TARGETARCH

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
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH make build/gateway

# Build the final image
FROM flomesh/pipy:0.90.0-123-$DISTROLESS_TAG
WORKDIR /
COPY --from=builder /workspace/bin/gateway .

USER 65532:65532

ENTRYPOINT ["/gateway"]
