# syntax=docker/dockerfile:1.7
FROM golang:1.24-bookworm AS build
WORKDIR /src
COPY go.work go.work.sum ./
COPY pkg ./pkg
COPY api ./api
COPY services/eo-pipeline ./services/eo-pipeline
WORKDIR /src/services/eo-pipeline
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags='-s -w' -o /out/eo-pipeline ./cmd/eo-pipeline

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/eo-pipeline /usr/local/bin/eo-pipeline
USER nonroot:nonroot
EXPOSE 8080 9090
ENTRYPOINT ["/usr/local/bin/eo-pipeline"]
