# syntax=docker/dockerfile:1.7
FROM golang:1.25-bookworm AS build
WORKDIR /src
COPY go.work go.work.sum ./
COPY pkg ./pkg
COPY api ./api
COPY services/sat-fsw ./services/sat-fsw
WORKDIR /src/services/sat-fsw
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags='-s -w' -o /out/sat-fsw ./cmd/sat-fsw

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/sat-fsw /usr/local/bin/sat-fsw
USER nonroot:nonroot
EXPOSE 8080 9090
ENTRYPOINT ["/usr/local/bin/sat-fsw"]
