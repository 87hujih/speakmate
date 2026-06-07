# syntax=docker/dockerfile:1

FROM golang:1.26-bookworm AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/speakmate ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/migrate ./cmd/migrate

FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

COPY --from=build /out/speakmate /app/speakmate
COPY --from=build /out/migrate /app/migrate
COPY migrations /app/migrations

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/app/speakmate"]
