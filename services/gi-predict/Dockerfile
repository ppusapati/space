# syntax=docker/dockerfile:1.7
FROM golang:1.26-bookworm AS build
WORKDIR /src
COPY services/go.work services/go.work.sum ./services/
COPY services/packages ./services/packages
COPY services/gi-predict ./services/gi-predict
WORKDIR /src/services/gi-predict
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags='-s -w' -o /out/gi-predict ./cmd

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/gi-predict /usr/local/bin/gi-predict
COPY --from=build /src/services/gi-predict/config /etc/gi-predict/config
WORKDIR /etc/gi-predict
USER nonroot:nonroot
EXPOSE 8080 9090
ENTRYPOINT ["/usr/local/bin/gi-predict"]
