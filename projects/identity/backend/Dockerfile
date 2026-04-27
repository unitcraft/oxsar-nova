# Multi-stage build для identity-service. Бинарник — /app/server.
#
# Build context — корень репозитория (compose: context: ..),
# чтобы COPY мог достать конфиги и миграции по абсолютным путям внутри
# монорепы.

FROM golang:1.23-alpine AS build
WORKDIR /src

COPY projects/identity/backend/go.mod projects/identity/backend/go.sum ./projects/identity/backend/
RUN cd projects/identity/backend && go mod download

COPY projects/identity/backend ./projects/identity/backend
RUN cd projects/identity/backend && \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

FROM alpine:3.20 AS runtime
RUN apk add --no-cache ca-certificates tzdata wget
WORKDIR /app

COPY --from=build /out/server /app/server

ENV IDENTITY_ADDR=:9000

EXPOSE 9000

ENTRYPOINT ["/app/server"]
