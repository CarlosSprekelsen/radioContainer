# syntax=docker/dockerfile:1.4

ARG GO_VERSION=1.22
FROM golang:${GO_VERSION}-alpine AS build

WORKDIR /src

# Install build tooling
RUN apk add --no-cache ca-certificates git build-base

# Pre-cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy application source
COPY . .

# Build for target platform
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -o /out/rcc ./cmd/rcc

FROM gcr.io/distroless/base-debian12

ENV RCC_ADDR=0.0.0.0:8080
WORKDIR /opt/rcc

COPY --from=build /out/rcc /usr/local/bin/rcc
RUN mkdir -p /opt/rcc/logs

EXPOSE 8080
EXPOSE 8081

ENTRYPOINT ["/usr/local/bin/rcc"]
