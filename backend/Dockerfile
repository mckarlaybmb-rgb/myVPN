FROM golang:1.24-alpine AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/vpnos-api ./cmd/api
FROM alpine:3.22
WORKDIR /app
COPY --from=build /out/vpnos-api /app/vpnos-api
COPY migrations /app/migrations
EXPOSE 8080
ENTRYPOINT ["/app/vpnos-api"]