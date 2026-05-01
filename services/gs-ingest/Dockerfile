# syntax=docker/dockerfile:1.7
FROM golang:1.26-bookworm AS build
WORKDIR /src
COPY services/go.work services/go.work.sum ./services/
COPY services/packages ./services/packages
COPY services/gs-ingest ./services/gs-ingest
WORKDIR /src/services/gs-ingest
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags='-s -w' -o /out/gs-ingest ./cmd

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/gs-ingest /usr/local/bin/gs-ingest
COPY --from=build /src/services/gs-ingest/config /etc/gs-ingest/config
WORKDIR /etc/gs-ingest
USER nonroot:nonroot
EXPOSE 8080 9090
ENTRYPOINT ["/usr/local/bin/gs-ingest"]
