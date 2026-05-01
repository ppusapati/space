# syntax=docker/dockerfile:1.7
FROM golang:1.26-bookworm AS build
WORKDIR /src
COPY services/go.work services/go.work.sum ./services/
COPY services/packages ./services/packages
COPY services/gi-reports ./services/gi-reports
WORKDIR /src/services/gi-reports
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags='-s -w' -o /out/gi-reports ./cmd

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/gi-reports /usr/local/bin/gi-reports
COPY --from=build /src/services/gi-reports/config /etc/gi-reports/config
WORKDIR /etc/gi-reports
USER nonroot:nonroot
EXPOSE 8080 9090
ENTRYPOINT ["/usr/local/bin/gi-reports"]
