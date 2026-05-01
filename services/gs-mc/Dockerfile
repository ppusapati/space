# syntax=docker/dockerfile:1.7
FROM golang:1.26-bookworm AS build
WORKDIR /src
COPY services/go.work services/go.work.sum ./services/
COPY services/packages ./services/packages
COPY services/gs-mc ./services/gs-mc
WORKDIR /src/services/gs-mc
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags='-s -w' -o /out/gs-mc ./cmd

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/gs-mc /usr/local/bin/gs-mc
COPY --from=build /src/services/gs-mc/config /etc/gs-mc/config
WORKDIR /etc/gs-mc
USER nonroot:nonroot
EXPOSE 8080 9090
ENTRYPOINT ["/usr/local/bin/gs-mc"]
