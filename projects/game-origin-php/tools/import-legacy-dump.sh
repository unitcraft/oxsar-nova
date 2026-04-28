#!/bin/bash
# Импорт полного дампа БД из легаси oxsar2 в game-origin.
# Запускать когда нужно полностью пересоздать dev-окружение со всеми пользователями
# и их данными (планеты, флот, постройки, сообщения и т.д.).
#
# Требования:
# - Запущен legacy: docker ps | grep oxsar2-mysql-1
# - Запущен game-origin: docker ps | grep docker-mysql-1
#
# Использование:
#   bash projects/game-origin-php/tools/import-legacy-dump.sh
#
# Дамп (~1.5GB) кладётся в legacy_dump.sql (в .gitignore, не коммитится).

set -euo pipefail

DUMP="$(cd "$(dirname "$0")/.." && pwd)/legacy_dump.sql"

echo "=> Дамп legacy oxsar2 → $DUMP"
docker exec oxsar2-mysql-1 mysqldump -uroot -proot oxsar_db \
  --skip-comments --hex-blob --skip-add-locks --routines --triggers --events \
  2>/dev/null > "$DUMP"

echo "=> Размер дампа: $(du -h "$DUMP" | cut -f1)"

echo "=> Замена DEFINER на CURRENT_USER (избегаем 'oxsar@localhost не существует')"
sed -i \
  -e 's/DEFINER=`oxsar`@`localhost`/DEFINER=CURRENT_USER/g' \
  -e 's/DEFINER=`root`@`localhost`/DEFINER=CURRENT_USER/g' \
  "$DUMP"

echo "=> Очистка БД game-origin"
docker exec docker-mysql-1 mysql -uoxsar_user -poxsar_pass \
  -e "DROP DATABASE IF EXISTS oxsar_db; CREATE DATABASE oxsar_db CHARACTER SET utf8 COLLATE utf8_general_ci"

echo "=> Импорт дампа в game-origin (5-10 минут)"
docker exec -i docker-mysql-1 mysql -uoxsar_user -poxsar_pass oxsar_db < "$DUMP"

echo "=> Добавление global_user_id для JWT auth (plan-36)"
docker exec docker-mysql-1 mysql -uoxsar_user -poxsar_pass oxsar_db -e "
  ALTER TABLE na_user
    ADD COLUMN global_user_id VARCHAR(36) NULL DEFAULT NULL AFTER userid,
    ADD UNIQUE INDEX uq_global_user_id (global_user_id);
  UPDATE na_user SET global_user_id='dev-user-001' WHERE userid=1;
"

echo "=> Проверка"
docker exec docker-mysql-1 mysql -uoxsar_user -poxsar_pass oxsar_db -e "
  SELECT 'users' AS t, COUNT(*) AS cnt FROM na_user
  UNION SELECT 'planets', COUNT(*) FROM na_planet
  UNION SELECT 'phrases', COUNT(*) FROM na_phrases;
  SELECT userid, username, global_user_id FROM na_user WHERE global_user_id IS NOT NULL;
"

echo ""
echo "Готово. Логин: http://localhost:8092/dev-login.php"
