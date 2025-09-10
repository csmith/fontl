FROM golang:1.25.1 AS build
WORKDIR /go/src/app
COPY . .

RUN --mount=type=cache,target=/go/pkg/mod \
    set -eux; \
    CGO_ENABLED=0 GO111MODULE=on go install ./cmd/serve; \
    go run github.com/google/go-licenses@latest save ./... --save_path=/notices; \
    mkdir -p /mounts/fonts;

FROM ghcr.io/greboid/dockerbase/nonroot:1.20250803.0
COPY --from=build /go/bin/serve /serve
COPY --from=build /notices /notices
COPY --from=build --chown=65532:65532 /mounts /
VOLUME /fonts
ENV PORT=8080
EXPOSE 8080
ENTRYPOINT ["/serve", "-port", "8080", "-dir", "/fonts"]
