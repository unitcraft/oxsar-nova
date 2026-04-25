# Runbook: Бэкапы и мониторинг

## Бэкапы PostgreSQL

### Что бэкапим

`pg_dump --format=custom` — весь oxsar-инстанс (схема + данные).
Кастомный формат: сжатый (~10× от plain SQL), восстанавливается
через `pg_restore --jobs N` параллельно.

### Как запустить вручную

```bash
export DB_URL="postgres://oxsar:PASSWORD@localhost:5432/oxsar"
export BACKUP_DIR="/var/backups/oxsar"
export BACKUP_RETAIN=14
# опционально:
export S3_BUCKET="s3://my-bucket/oxsar-backups"
export AWS_ENDPOINT="https://s3.selectel.ru"  # для Selectel / Yandex

bash deploy/backup.sh
```

### Установка cron на VPS

```bash
# Скопировать скрипт
cp deploy/backup.sh /opt/oxsar/backup.sh
chmod +x /opt/oxsar/backup.sh

# Создать .env файл с секретами (не в репо!)
cat > /opt/oxsar/backup.env << 'EOF'
DB_URL=postgres://oxsar:STRONG_PASSWORD@localhost:5432/oxsar
BACKUP_DIR=/var/backups/oxsar
BACKUP_RETAIN=14
S3_BUCKET=s3://my-bucket/oxsar
EOF
chmod 600 /opt/oxsar/backup.env

# Добавить в cron (каждые 6 часов)
crontab -e
# Добавить строку:
0 */6 * * * source /opt/oxsar/backup.env && /opt/oxsar/backup.sh >> /var/log/oxsar-backup.log 2>&1
```

### Восстановление из бэкапа

```bash
# Остановить сервисы
docker compose -f deploy/docker-compose.yml stop backend worker

# Создать чистую БД (или dropdb + createdb)
createdb -U oxsar oxsar_restore

# Восстановить
pg_restore \
  --host=localhost --port=5432 \
  --username=oxsar \
  --dbname=oxsar_restore \
  --jobs=4 \
  --no-owner \
  /var/backups/oxsar/oxsar_20260425_120000.dump

# Переключить DB_URL на oxsar_restore или переименовать
# После проверки — переименовать:
psql -U oxsar -c "ALTER DATABASE oxsar RENAME TO oxsar_old;"
psql -U oxsar -c "ALTER DATABASE oxsar_restore RENAME TO oxsar;"

# Запустить сервисы
docker compose -f deploy/docker-compose.yml up -d backend worker
```

### Проверка бэкапа

```bash
# Листинг дампов
ls -lh /var/backups/oxsar/

# Быстрая верификация (только список объектов, без восстановления)
pg_restore --list /var/backups/oxsar/oxsar_LATEST.dump | head -20
```

---

## Мониторинг (Prometheus + Grafana)

### Что мониторится

| Источник | Метрики |
|---|---|
| `backend:8080/metrics` | HTTP-запросы, latency, goroutines |
| `worker:9090/metrics` | Events processed, handler duration, queue lag |
| `postgres-exporter:9187` | Connections, locks, transaction rate, replication lag |

Ключевые метрики воркера:
- `oxsar_events_processed_total{state="error"}` — ошибки обработки
- `oxsar_events_handler_duration_seconds` — histogram latency
- `oxsar_events_lag_seconds` — отставание очереди от realtime

### Запуск мониторинга

```bash
docker compose \
  -f deploy/docker-compose.yml \
  -f deploy/docker-compose.prod.yml \
  -f deploy/docker-compose.monitoring.yml \
  up -d
```

Grafana: http://VPS_IP:3000 (admin / $GRAFANA_PASSWORD)

### Базовые алерты (настроить вручную в Grafana)

| Условие | Серьёзность | Действие |
|---|---|---|
| `oxsar_events_lag_seconds > 60` | Warning | Воркер отстаёт — проверить CPU/БД |
| `oxsar_events_lag_seconds > 300` | Critical | Воркер завис — перезапустить |
| `oxsar_events_processed_total{state="error"} rate > 0.1` | Warning | Ошибки в событиях — смотреть логи |
| `pg_stat_activity_count{state="idle in transaction"} > 10` | Warning | Утечки транзакций |
| Disk usage `/var/backups > 80%` | Warning | Почистить старые бэкапы |

### Закрытие /metrics от интернета (nginx)

```nginx
# В nginx.conf — только внутренняя сеть:
location /metrics {
    allow 10.0.0.0/8;
    allow 172.16.0.0/12;
    deny all;
    proxy_pass http://backend:8080;
}
```

---

## Checklist перед запуском в прод

- [ ] Cron бэкапа настроен и проверен (`bash deploy/backup.sh`)
- [ ] Первый бэкап создан и восстановление проверено
- [ ] S3-bucket настроен (или как минимум local retain=14)
- [ ] Grafana доступна, Prometheus scrapes all targets (зелёные)
- [ ] Алерты настроены (Slack / Telegram webhook)
- [ ] `/metrics` закрыт firewall / nginx от внешнего доступа
- [ ] `GRAFANA_PASSWORD` задан не дефолтным
