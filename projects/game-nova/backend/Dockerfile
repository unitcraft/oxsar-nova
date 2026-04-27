# Multi-stage build для game-nova: один образ, два бинаря (server, worker).
# Выбор entrypoint — через command в docker-compose.
#
# auth-service и portal живут в отдельных Go-модулях:
#   projects/auth/backend/Dockerfile
#   projects/portal/backend/Dockerfile

FROM golang:1.23-alpine AS build
WORKDIR /src

# Сначала копируем только манифест — кешируем слой go mod download,
# если go.mod/go.sum не менялись.
COPY projects/game-nova/backend/go.mod projects/game-nova/backend/go.sum ./projects/game-nova/backend/
RUN cd projects/game-nova/backend && go mod download

# Копируем исходники и собираем оба бинаря.
COPY projects/game-nova/backend ./projects/game-nova/backend
RUN cd projects/game-nova/backend && \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server && \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/worker ./cmd/worker

FROM alpine:3.20 AS runtime
# wget — для docker healthcheck. ca-certificates — для go-redis и HTTPS.
RUN apk add --no-cache ca-certificates tzdata wget
WORKDIR /app

COPY --from=build /out/server /app/server
COPY --from=build /out/worker /app/worker
COPY projects/game-nova/configs /app/configs
# План 19: вики читается на runtime — копируем сгенерированные .md.
COPY docs/wiki /app/wiki

ENV CATALOG_DIR=/app/configs
ENV I18N_DIR=/app/configs/i18n
ENV WIKI_ROOT=/app/wiki/ru
ENV SERVER_ADDR=:8080

EXPOSE 8080

# По умолчанию стартуем server. worker-сервис в compose переопределит
# entrypoint.
ENTRYPOINT ["/app/server"]
