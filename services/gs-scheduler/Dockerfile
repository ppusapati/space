# syntax=docker/dockerfile:1.7
FROM golang:1.26-bookworm AS build
WORKDIR /src
COPY services/go.work services/go.work.sum ./services/
COPY services/packages ./services/packages
COPY services/gs-scheduler ./services/gs-scheduler
WORKDIR /src/services/gs-scheduler
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags='-s -w' -o /out/gs-scheduler ./cmd

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/gs-scheduler /usr/local/bin/gs-scheduler
COPY --from=build /src/services/gs-scheduler/config /etc/gs-scheduler/config
WORKDIR /etc/gs-scheduler
USER nonroot:nonroot
EXPOSE 8080 9090
ENTRYPOINT ["/usr/local/bin/gs-scheduler"]
