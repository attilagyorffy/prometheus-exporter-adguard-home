FROM golang:1.25-alpine AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /go/bin/adguard-home-exporter .

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=builder /go/bin/adguard-home-exporter /usr/local/bin/adguard-home-exporter
ENTRYPOINT ["adguard-home-exporter"]
