# Stage 1: Build the Vue frontend
FROM node:20-alpine AS frontend-builder
WORKDIR /app/frontend
RUN npm install -g pnpm
COPY frontend/package.json frontend/pnpm-lock.yaml ./
RUN pnpm install
COPY frontend/ ./
RUN pnpm run build

# Stage 2: Build the Go backend
FROM golang:alpine AS backend-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend-builder /app/frontend/dist ./frontend/dist
RUN go build -o docklog main.go

# Stage 3: Final runtime image
FROM alpine:latest
WORKDIR /app
RUN apk add --no-cache ca-certificates libc6-compat
COPY --from=backend-builder /app/docklog .
COPY --from=backend-builder /app/frontend/dist ./frontend/dist

# Expose the port
EXPOSE 8000

# Run the binary
CMD ["./docklog"]
