# syntax=docker/dockerfile:1.7
FROM golang:1.26-bookworm AS build
WORKDIR /src
COPY services/go.work services/go.work.sum ./services/
COPY services/packages ./services/packages
COPY services/eo-analytics ./services/eo-analytics
WORKDIR /src/services/eo-analytics
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags='-s -w' -o /out/eo-analytics ./cmd

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/eo-analytics /usr/local/bin/eo-analytics
USER nonroot:nonroot
EXPOSE 8080 9090
ENTRYPOINT ["/usr/local/bin/eo-analytics"]
