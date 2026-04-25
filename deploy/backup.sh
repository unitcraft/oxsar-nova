#!/usr/bin/env bash
# deploy/backup.sh — резервное копирование PostgreSQL.
#
# Переменные окружения (передавать через .env или systemd):
#   DB_URL          — postgres DSN, например postgres://user:pass@host/db
#   BACKUP_DIR      — локальная папка для хранения (default: /var/backups/oxsar)
#   BACKUP_RETAIN   — сколько архивов хранить (default: 14)
#   S3_BUCKET       — если задан, копирует в S3: s3://bucket/prefix
#   AWS_ENDPOINT    — для совместимых с S3 хранилищ (Selectel, Yandex Cloud)
#
# Cron (каждые 6 часов):
#   0 */6 * * * /opt/oxsar/deploy/backup.sh >> /var/log/oxsar-backup.log 2>&1

set -euo pipefail

BACKUP_DIR="${BACKUP_DIR:-/var/backups/oxsar}"
BACKUP_RETAIN="${BACKUP_RETAIN:-14}"
TIMESTAMP=$(date -u +"%Y%m%d_%H%M%S")
FILENAME="oxsar_${TIMESTAMP}.dump"
FILEPATH="${BACKUP_DIR}/${FILENAME}"

mkdir -p "${BACKUP_DIR}"

echo "[$(date -u)] Starting backup → ${FILEPATH}"

# pg_dump в custom-формате (сжатый, восстанавливается через pg_restore).
pg_dump \
  --format=custom \
  --compress=9 \
  --no-privileges \
  --no-owner \
  "${DB_URL}" \
  --file="${FILEPATH}"

SIZE=$(du -sh "${FILEPATH}" | cut -f1)
echo "[$(date -u)] Dump complete: ${SIZE}"

# Ротация: удаляем старые архивы.
find "${BACKUP_DIR}" -name "oxsar_*.dump" -type f \
  | sort \
  | head -n "-${BACKUP_RETAIN}" \
  | xargs --no-run-if-empty rm -v

# Копирование в S3-совместимое хранилище (если задан BUCKET).
if [[ -n "${S3_BUCKET:-}" ]]; then
  AWS_ARGS=""
  if [[ -n "${AWS_ENDPOINT:-}" ]]; then
    AWS_ARGS="--endpoint-url ${AWS_ENDPOINT}"
  fi
  # shellcheck disable=SC2086
  aws s3 cp ${AWS_ARGS} "${FILEPATH}" "${S3_BUCKET}/${FILENAME}"
  echo "[$(date -u)] Uploaded to ${S3_BUCKET}/${FILENAME}"
fi

echo "[$(date -u)] Backup done."
