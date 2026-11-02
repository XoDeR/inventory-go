# ============================
# One Dockerfile for all services. The build context is the repository root.
#
#   docker build -f deploy/docker/Dockerfile \
#     --build-arg SERVICE=inventory --build-arg CMD=grpc_server -t inventory .
# ============================

# ============================
# Stage 1: Build stage
# ============================
FROM golang:1.27.2-alpine AS builder

# Which service to build and which of its cmd/ entrypoints to use
ARG SERVICE
ARG CMD

RUN apk add --no-cache git

WORKDIR /app

# go.work lists every module of the workspace, so all of them must be present
COPY go.work go.work.sum ./

# Copy only go.mod/go.sum first, so the dependency download is cached between builds
COPY platform/go.mod platform/go.sum ./platform/
COPY shared/go.mod shared/go.sum ./shared/
COPY inventory/go.mod inventory/go.sum ./inventory/
COPY order/go.mod order/go.sum ./order/
COPY payment/go.mod payment/go.sum ./payment/

RUN go mod download

COPY platform ./platform
COPY shared ./shared
COPY inventory ./inventory
COPY order ./order
COPY payment ./payment

# Build a static Linux binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/service ./${SERVICE}/cmd/${CMD}


# ============================
# Stage 2: Final image
# ============================
FROM alpine:3.21.3

# Run as a non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

USER appuser

COPY --from=builder /app/service .

# The configuration comes from environment variables
ENTRYPOINT ["./service"]
