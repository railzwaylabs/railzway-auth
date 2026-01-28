FROM node:20-alpine AS frontend-builder

# Enable pnpm
RUN corepack enable && corepack prepare pnpm@latest --activate

WORKDIR /src/ui

# Copy frontend config
COPY ui/package.json ui/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

# Copy frontend source and build
COPY ui/ ./
RUN pnpm run build

FROM golang:1.25-alpine AS backend-builder

RUN apk add --no-cache git ca-certificates
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /workspace/auth ./cmd/auth

FROM gcr.io/distroless/static-debian12 AS runtime

WORKDIR /app

# Copy backend binary
COPY --from=backend-builder /workspace/auth /app/auth

# Copy frontend build artifacts
# The app expects ./ui/dist relative to WORKDIR
COPY --from=frontend-builder /src/ui/dist /app/ui/dist

EXPOSE 8080
USER 65532:65532

ENTRYPOINT ["/app/auth"]
