FROM golang:1.26-bookworm AS build
WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api \
 && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/migrate ./cmd/migrate

FROM gcr.io/distroless/static-debian12
WORKDIR /app
COPY --from=build /out/api /out/migrate ./
COPY config/config_api.yml config/config_api.yml
COPY config/permissions config/permissions
COPY migrations migrations
EXPOSE 8080
ENTRYPOINT ["/app/api"]
