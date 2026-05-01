# syntax=docker/dockerfile:1.7
FROM golang:1.26-bookworm AS build
WORKDIR /src
COPY services/go.work services/go.work.sum ./services/
COPY services/packages ./services/packages
COPY services/gi-analytics ./services/gi-analytics
WORKDIR /src/services/gi-analytics
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags='-s -w' -o /out/gi-analytics ./cmd

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/gi-analytics /usr/local/bin/gi-analytics
COPY --from=build /src/services/gi-analytics/config /etc/gi-analytics/config
WORKDIR /etc/gi-analytics
USER nonroot:nonroot
EXPOSE 8080 9090
ENTRYPOINT ["/usr/local/bin/gi-analytics"]
