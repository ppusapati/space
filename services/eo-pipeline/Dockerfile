# syntax=docker/dockerfile:1.7
FROM golang:1.26-bookworm AS build
WORKDIR /src
COPY services/go.work services/go.work.sum ./services/
COPY services/packages ./services/packages
COPY services/eo-pipeline ./services/eo-pipeline
WORKDIR /src/services/eo-pipeline
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags='-s -w' -o /out/eo-pipeline ./cmd

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/eo-pipeline /usr/local/bin/eo-pipeline
USER nonroot:nonroot
EXPOSE 8080 9090
ENTRYPOINT ["/usr/local/bin/eo-pipeline"]
