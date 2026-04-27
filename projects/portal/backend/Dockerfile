# Multi-stage build для portal-backend. Бинарник — /app/server.

FROM golang:1.23-alpine AS build
WORKDIR /src

COPY projects/portal/backend/go.mod projects/portal/backend/go.sum ./projects/portal/backend/
RUN cd projects/portal/backend && go mod download

COPY projects/portal/backend ./projects/portal/backend
RUN cd projects/portal/backend && \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

FROM alpine:3.20 AS runtime
RUN apk add --no-cache ca-certificates tzdata wget
WORKDIR /app

COPY --from=build /out/server /app/server
COPY projects/game-nova/configs/universes.yaml /app/configs/universes.yaml

ENV UNIVERSES_CONFIG=/app/configs/universes.yaml
ENV PORTAL_ADDR=:8090

EXPOSE 8090

ENTRYPOINT ["/app/server"]
