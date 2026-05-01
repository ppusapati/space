# syntax=docker/dockerfile:1.7
FROM golang:1.26-bookworm AS build
WORKDIR /src
COPY services/go.work services/go.work.sum ./services/
COPY services/packages ./services/packages
COPY services/sat-telemetry ./services/sat-telemetry
WORKDIR /src/services/sat-telemetry
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags='-s -w' -o /out/sat-telemetry ./cmd

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/sat-telemetry /usr/local/bin/sat-telemetry
USER nonroot:nonroot
EXPOSE 8080 9090
ENTRYPOINT ["/usr/local/bin/sat-telemetry"]
