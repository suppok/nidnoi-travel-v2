FROM golang:1.26-bookworm AS build
WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
# TODO: also build ./cmd/migrate once it exists, and copy /out/migrate below.
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api

FROM gcr.io/distroless/static-debian12
WORKDIR /app
COPY --from=build /out/api ./
COPY config/config_api.yml config/config_api.yml
COPY config/permissions config/permissions
COPY migrations migrations
EXPOSE 8080
ENTRYPOINT ["/app/api"]
