# syntax=docker/dockerfile:1.7
FROM golang:1.26-bookworm AS build
WORKDIR /src
COPY services/go.work services/go.work.sum ./services/
COPY services/packages ./services/packages
COPY services/sat-simulation ./services/sat-simulation
WORKDIR /src/services/sat-simulation
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags='-s -w' -o /out/sat-simulation ./cmd

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/sat-simulation /usr/local/bin/sat-simulation
COPY --from=build /src/services/sat-simulation/config /etc/sat-simulation/config
WORKDIR /etc/sat-simulation
USER nonroot:nonroot
EXPOSE 8080 9090
ENTRYPOINT ["/usr/local/bin/sat-simulation"]
