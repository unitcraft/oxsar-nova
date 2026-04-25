# Multi-stage build для backend: один образ, два бинаря (server, worker).
# Выбор entrypoint — через command в docker-compose.
#
# Стадия build: выкачиваем зависимости, компилируем статические бинари
# (CGO=0), пишем их в /out.
#
# Стадия runtime: минимальный alpine + CA-сертификаты (для go-redis
# и будущих внешних HTTPS-вызовов). configs копируются внутрь, потому
# что LoadCatalog читает их при старте.

FROM golang:1.23-alpine AS build
WORKDIR /src

# Сначала копируем только манифест — кешируем слой go mod download,
# если go.mod/go.sum не менялись.
COPY backend/go.mod backend/go.sum ./backend/
RUN cd backend && go mod download

# Копируем исходники и собираем оба бинаря.
COPY backend ./backend
RUN cd backend && \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server && \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/worker ./cmd/worker

FROM alpine:3.20 AS runtime
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app

COPY --from=build /out/server /app/server
COPY --from=build /out/worker /app/worker
COPY configs /app/configs
# План 19: вики читается на runtime — копируем сгенерированные .md.
COPY docs/wiki /app/wiki

ENV CATALOG_DIR=/app/configs
ENV I18N_DIR=/app/configs/i18n
ENV WIKI_ROOT=/app/wiki/ru
ENV SERVER_ADDR=:8080

EXPOSE 8080

# По умолчанию стартуем server. worker-сервис в compose переопределит
# command.
ENTRYPOINT ["/app/server"]
