# ============================================
# Stage 1: Build Frontend
# ============================================
FROM node:20-alpine AS frontend-builder

# Enable pnpm
RUN corepack enable && corepack prepare pnpm@latest --activate

WORKDIR /app/ui

# Copy frontend package files
COPY ui/package.json ui/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

# Copy frontend source and build
COPY ui/ ./
RUN pnpm run build

# ============================================
# Stage 2: Build Backend
# ============================================
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /workspace/auth ./cmd/auth

# ============================================
# Stage 3: Runtime
# ============================================
FROM gcr.io/distroless/static-debian12 AS runtime

COPY --from=builder /workspace/auth /usr/local/bin/auth
COPY --from=frontend-builder /app/ui/dist /ui/dist

EXPOSE 8080
USER 65532:65532
ENTRYPOINT ["/usr/local/bin/auth"]
