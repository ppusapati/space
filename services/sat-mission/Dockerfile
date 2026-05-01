# syntax=docker/dockerfile:1.7
FROM golang:1.26-bookworm AS build
WORKDIR /src
COPY services/go.work services/go.work.sum ./services/
COPY services/packages ./services/packages
COPY services/sat-mission ./services/sat-mission
WORKDIR /src/services/sat-mission
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags='-s -w' -o /out/sat-mission ./cmd

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/sat-mission /usr/local/bin/sat-mission
COPY --from=build /src/services/sat-mission/config /etc/sat-mission/config
WORKDIR /etc/sat-mission
USER nonroot:nonroot
EXPOSE 8080 9090
ENTRYPOINT ["/usr/local/bin/sat-mission"]
