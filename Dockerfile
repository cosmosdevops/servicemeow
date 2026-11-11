# Build on the native platform and cross-compile, so multi-arch images
# don't need emulation. BuildKit sets BUILDPLATFORM/TARGETOS/TARGETARCH; the
# default keeps the legacy builder working, building for the host platform.
ARG BUILDPLATFORM=linux/amd64
FROM --platform=$BUILDPLATFORM golang:1.27-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG TARGETOS
ARG TARGETARCH
ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath \
    -ldflags="-s -w -X github.com/CosmosDevops/servicemeow/cmd.version=${VERSION}" \
    -o /out/servicemeow .

FROM alpine:3.24 AS release
COPY --from=build /out/servicemeow /servicemeow

USER 65534:65534
ENTRYPOINT [ "/servicemeow" ]
CMD ["--help"]
