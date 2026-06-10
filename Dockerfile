# Stage 1: Build the Vue frontend
FROM node:20-alpine AS frontend-builder
WORKDIR /app/frontend
RUN npm install -g pnpm
COPY frontend/package.json frontend/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile
COPY frontend/ ./
RUN pnpm run build

# Stage 2: Build the Go backend (build entire package, not main.go alone)
FROM golang:1.26-alpine AS backend-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend-builder /app/frontend/dist ./frontend/dist

ARG TARGETOS=linux
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -ldflags="-s -w" -o docklog .

# Stage 3: Final runtime image
FROM alpine:3.21
WORKDIR /app
RUN apk add --no-cache ca-certificates libc6-compat
COPY --from=backend-builder /app/docklog /usr/local/bin/docklog
COPY --from=backend-builder /app/frontend/dist ./frontend/dist

EXPOSE 8000
CMD ["docklog", "server"]
