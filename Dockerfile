# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY *.go ./

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o tplink-ddm-exporter .

# Runtime stage
FROM scratch

COPY --from=builder /build/tplink-ddm-exporter /tplink-ddm-exporter

# Add CA certificates for HTTPS (if needed in future)
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

EXPOSE 9116

ENTRYPOINT ["/tplink-ddm-exporter"]

