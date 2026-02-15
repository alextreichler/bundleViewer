# --- Build Stage ---
FROM golang:1.25-bookworm AS builder

# Install Zig (Required for the parsing accelerator)
RUN apt-get update && apt-get install -y xz-utils curl 
    && curl -L https://ziglang.org/download/0.15.2/zig-linux-x86_64-0.15.2.tar.xz | tar -xJ -C /usr/local 
    && ln -s /usr/local/zig-linux-x86_64-0.15.2/zig /usr/local/bin/zig

# Set working directory
WORKDIR /app

# Copy dependency files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the native Zig library and the Go webapp
# We use CGO_ENABLED=1 and Zig as the C compiler
RUN CGO_ENABLED=1 CC="zig cc -target x86_64-linux" CXX="zig c++ -target x86_64-linux" \
    zig build-lib -lc -O ReleaseFast -target x86_64-linux internal/parser/native/parser.zig -femit-bin=internal/parser/native/libparser_linux_amd64.a \
    && CGO_ENABLED=1 CC="zig cc -target x86_64-linux" CXX="zig c++ -target x86_64-linux" \
    GOEXPERIMENT=greenteagc,jsonv2 \
    go build -tags "sqlite_fts5" -o bundleViewer cmd/webapp/main.go

# --- Run Stage ---
FROM debian:bookworm-slim

WORKDIR /app

# Install CA certificates for HTTPS bundle fetching
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

# Copy binary from builder
COPY --from=builder /app/bundleViewer .

# Default configuration for container environment
ENV HOST=0.0.0.0
ENV PORT=7575
ENV INACTIVITY_TIMEOUT=30m
ENV SHUTDOWN_TIMEOUT=4h
ENV ONE_SHOT=false
ENV MEMORY_LIMIT=0

EXPOSE 7575

# Run the viewer
ENTRYPOINT ["./bundleViewer"]
