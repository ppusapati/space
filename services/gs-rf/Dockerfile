# syntax=docker/dockerfile:1.7
FROM golang:1.26-bookworm AS build
WORKDIR /src
COPY services/go.work services/go.work.sum ./services/
COPY services/packages ./services/packages
COPY services/gs-rf ./services/gs-rf
WORKDIR /src/services/gs-rf
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags='-s -w' -o /out/gs-rf ./cmd

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/gs-rf /usr/local/bin/gs-rf
COPY --from=build /src/services/gs-rf/config /etc/gs-rf/config
WORKDIR /etc/gs-rf
USER nonroot:nonroot
EXPOSE 8080 9090
ENTRYPOINT ["/usr/local/bin/gs-rf"]
