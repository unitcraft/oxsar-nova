# Multi-stage build для billing-service. Бинарник — /app/server.
# План 38.

FROM golang:1.23-alpine AS build
WORKDIR /src

COPY projects/billing/backend/go.mod projects/billing/backend/go.sum ./projects/billing/backend/
RUN cd projects/billing/backend && go mod download

COPY projects/billing/backend ./projects/billing/backend
RUN cd projects/billing/backend && \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

FROM alpine:3.20 AS runtime
RUN apk add --no-cache ca-certificates tzdata wget
WORKDIR /app

COPY --from=build /out/server /app/server

ENV BILLING_ADDR=:9100

EXPOSE 9100

ENTRYPOINT ["/app/server"]
