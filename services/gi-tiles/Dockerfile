# syntax=docker/dockerfile:1.7
FROM golang:1.26-bookworm AS build
WORKDIR /src
COPY services/go.work services/go.work.sum ./services/
COPY services/packages ./services/packages
COPY services/gi-tiles ./services/gi-tiles
WORKDIR /src/services/gi-tiles
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags='-s -w' -o /out/gi-tiles ./cmd

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/gi-tiles /usr/local/bin/gi-tiles
COPY --from=build /src/services/gi-tiles/config /etc/gi-tiles/config
WORKDIR /etc/gi-tiles
USER nonroot:nonroot
EXPOSE 8080 9090
ENTRYPOINT ["/usr/local/bin/gi-tiles"]
