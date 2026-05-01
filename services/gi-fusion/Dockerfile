# syntax=docker/dockerfile:1.7
FROM golang:1.26-bookworm AS build
WORKDIR /src
COPY services/go.work services/go.work.sum ./services/
COPY services/packages ./services/packages
COPY services/gi-fusion ./services/gi-fusion
WORKDIR /src/services/gi-fusion
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags='-s -w' -o /out/gi-fusion ./cmd

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/gi-fusion /usr/local/bin/gi-fusion
COPY --from=build /src/services/gi-fusion/config /etc/gi-fusion/config
WORKDIR /etc/gi-fusion
USER nonroot:nonroot
EXPOSE 8080 9090
ENTRYPOINT ["/usr/local/bin/gi-fusion"]
