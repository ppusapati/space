# syntax=docker/dockerfile:1.7
FROM golang:1.26-bookworm AS build
WORKDIR /src
COPY services/go.work services/go.work.sum ./services/
COPY services/packages ./services/packages
COPY services/sat-fsw ./services/sat-fsw
WORKDIR /src/services/sat-fsw
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags='-s -w' -o /out/sat-fsw ./cmd

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/sat-fsw /usr/local/bin/sat-fsw
COPY --from=build /src/services/sat-fsw/config /etc/sat-fsw/config
WORKDIR /etc/sat-fsw
USER nonroot:nonroot
EXPOSE 8080 9090
ENTRYPOINT ["/usr/local/bin/sat-fsw"]
