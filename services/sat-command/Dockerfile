# syntax=docker/dockerfile:1.7
FROM golang:1.26-bookworm AS build
WORKDIR /src
COPY services/go.work services/go.work.sum ./services/
COPY services/packages ./services/packages
COPY services/sat-command ./services/sat-command
WORKDIR /src/services/sat-command
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags='-s -w' -o /out/sat-command ./cmd

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/sat-command /usr/local/bin/sat-command
COPY --from=build /src/services/sat-command/config /etc/sat-command/config
WORKDIR /etc/sat-command
USER nonroot:nonroot
EXPOSE 8080 9090
ENTRYPOINT ["/usr/local/bin/sat-command"]
