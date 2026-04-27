#!/bin/bash
# План 37.5d.1: снять снимок test-юзера (userid=1) из боевой legacy БД
# для использования как dev-fixture в game-origin (UI-сравнение).
#
# Запускать ДО старта game-origin:
#   bash projects/game-origin/tools/snapshot-legacy-user.sh
#
# Результат:
#   projects/game-origin/migrations/fixtures/test-user-snapshot.sql

set -euo pipefail

LEGACY_CONTAINER="oxsar2-mysql-1"
LEGACY_USER="root"
LEGACY_PASS="root"
LEGACY_DB="oxsar_db"
USERID="${1:-1}"

OUT="$(dirname "$0")/../migrations/fixtures/test-user-snapshot.sql"
mkdir -p "$(dirname "$OUT")"

echo "Snapshotting userid=${USERID} from ${LEGACY_CONTAINER}..." >&2

# Получить список planetid юзера (для taблиц связанных через planet)
PLANETIDS=$(docker exec "$LEGACY_CONTAINER" mysql -u"$LEGACY_USER" -p"$LEGACY_PASS" "$LEGACY_DB" \
  -N -B -e "SELECT GROUP_CONCAT(planetid) FROM na_planet WHERE userid=$USERID" 2>/dev/null)

if [ -z "$PLANETIDS" ]; then
  echo "ERROR: no planets for userid=$USERID" >&2
  exit 1
fi
echo "  planets: $PLANETIDS" >&2

# Header в выходной файл
cat > "$OUT" <<EOF
-- Snapshot test-юзера (userid=${USERID}) из legacy oxsar2 для UI-сравнения.
-- Сгенерировано $(date -u +%Y-%m-%dT%H:%M:%SZ) скриптом snapshot-legacy-user.sh.
-- Применять через apply-test-user-fixture.sh — он сохраняет global_user_id.
SET FOREIGN_KEY_CHECKS=0;
SET unique_checks=0;
SET autocommit=0;

EOF

# Группа 1: таблицы где у юзера есть колонка userid (фильтр WHERE userid=N).
# Исключено: na_event_src/na_event_dest — это VIEW, dump для них бессмыслен
# (содержимое генерируется из na_events).
USER_TABLES=(
  na_user na_password na_user2group na_user2ally
  na_planet na_research2user na_artefact2user na_officer
  na_referral na_user_experience na_credit_bonus_item
  na_res_log na_res_transfer
)

for tbl in "${USER_TABLES[@]}"; do
  echo "  dump $tbl WHERE userid=$USERID" >&2
  echo "-- $tbl" >> "$OUT"
  docker exec "$LEGACY_CONTAINER" mysqldump -u"$LEGACY_USER" -p"$LEGACY_PASS" "$LEGACY_DB" \
    --no-create-info --skip-extended-insert --skip-comments --hex-blob --complete-insert \
    --where="userid=$USERID" "$tbl" 2>/dev/null >> "$OUT" || {
      echo "  WARN: dump failed for $tbl, skipping" >&2
    }
  echo "" >> "$OUT"
done

# Группа 2: таблицы где фильтр по planetid IN (...). Только planetid —
# moonid есть только у na_galaxy (см. отдельную обработку ниже).
PLANET_TABLES=(
  na_building2planet na_unit2shipyard
  na_temp_fleet na_stargate_jump na_exchange_lots
)

for tbl in "${PLANET_TABLES[@]}"; do
  echo "  dump $tbl WHERE planetid IN (...)" >&2
  echo "-- $tbl" >> "$OUT"
  docker exec "$LEGACY_CONTAINER" mysqldump -u"$LEGACY_USER" -p"$LEGACY_PASS" "$LEGACY_DB" \
    --no-create-info --skip-extended-insert --skip-comments --hex-blob --complete-insert \
    --where="planetid IN ($PLANETIDS)" "$tbl" 2>/dev/null >> "$OUT" || true
  echo "" >> "$OUT"
done

# na_galaxy — отдельно с фильтром по planetid OR moonid
echo "  dump na_galaxy WHERE planetid OR moonid IN (...)" >&2
echo "-- na_galaxy" >> "$OUT"
docker exec "$LEGACY_CONTAINER" mysqldump -u"$LEGACY_USER" -p"$LEGACY_PASS" "$LEGACY_DB" \
  --no-create-info --skip-extended-insert --skip-comments --hex-blob --complete-insert \
  --where="planetid IN ($PLANETIDS) OR moonid IN ($PLANETIDS)" na_galaxy 2>/dev/null >> "$OUT" || true
echo "" >> "$OUT"

# Группа 3: связанные таблицы (alliance — нужны другие участники для рендера)
ALLY_AID=$(docker exec "$LEGACY_CONTAINER" mysql -u"$LEGACY_USER" -p"$LEGACY_PASS" "$LEGACY_DB" \
  -N -B -e "SELECT aid FROM na_user2ally WHERE userid=$USERID LIMIT 1" 2>/dev/null || echo "")

if [ -n "$ALLY_AID" ]; then
  echo "  dump na_alliance WHERE aid=$ALLY_AID" >&2
  echo "-- na_alliance" >> "$OUT"
  docker exec "$LEGACY_CONTAINER" mysqldump -u"$LEGACY_USER" -p"$LEGACY_PASS" "$LEGACY_DB" \
    --no-create-info --skip-extended-insert --skip-comments --hex-blob --complete-insert \
    --where="aid=$ALLY_AID" na_alliance 2>/dev/null >> "$OUT" || true

  # Другие участники той же alliance — нужны их user-записи для отрисовки списка
  ALLY_MEMBERS=$(docker exec "$LEGACY_CONTAINER" mysql -u"$LEGACY_USER" -p"$LEGACY_PASS" "$LEGACY_DB" \
    -N -B -e "SELECT GROUP_CONCAT(userid) FROM na_user2ally WHERE aid=$ALLY_AID" 2>/dev/null)

  if [ -n "$ALLY_MEMBERS" ] && [ "$ALLY_MEMBERS" != "$USERID" ]; then
    echo "  dump na_user (alliance members: $ALLY_MEMBERS)" >&2
    echo "-- alliance members (na_user)" >> "$OUT"
    docker exec "$LEGACY_CONTAINER" mysqldump -u"$LEGACY_USER" -p"$LEGACY_PASS" "$LEGACY_DB" \
      --no-create-info --skip-extended-insert --skip-comments --hex-blob --complete-insert \
      --where="userid IN ($ALLY_MEMBERS) AND userid != $USERID" na_user 2>/dev/null >> "$OUT" || true

    echo "-- alliance members (na_user2ally)" >> "$OUT"
    docker exec "$LEGACY_CONTAINER" mysqldump -u"$LEGACY_USER" -p"$LEGACY_PASS" "$LEGACY_DB" \
      --no-create-info --skip-extended-insert --skip-comments --hex-blob --complete-insert \
      --where="aid=$ALLY_AID AND userid != $USERID" na_user2ally 2>/dev/null >> "$OUT" || true
  fi
  echo "" >> "$OUT"
fi

# Группа 4: чат — последние N сообщений + их авторы (для рендера Chat страницы)
echo "  dump last 100 chat messages + their authors" >&2
LAST_CHAT_AUTHORS=$(docker exec "$LEGACY_CONTAINER" mysql -u"$LEGACY_USER" -p"$LEGACY_PASS" "$LEGACY_DB" \
  -N -B -e "SELECT GROUP_CONCAT(DISTINCT userid) FROM (SELECT userid FROM na_chat ORDER BY messageid DESC LIMIT 100) t" 2>/dev/null)

if [ -n "$LAST_CHAT_AUTHORS" ]; then
  echo "-- na_chat (last 100 messages)" >> "$OUT"
  docker exec "$LEGACY_CONTAINER" mysqldump -u"$LEGACY_USER" -p"$LEGACY_PASS" "$LEGACY_DB" \
    --no-create-info --skip-extended-insert --skip-comments --hex-blob --complete-insert \
    --where="messageid IN (SELECT messageid FROM (SELECT messageid FROM na_chat ORDER BY messageid DESC LIMIT 100) t)" \
    na_chat 2>/dev/null >> "$OUT" || true

  echo "-- na_user (chat authors)" >> "$OUT"
  docker exec "$LEGACY_CONTAINER" mysqldump -u"$LEGACY_USER" -p"$LEGACY_PASS" "$LEGACY_DB" \
    --no-create-info --skip-extended-insert --skip-comments --hex-blob --complete-insert \
    --where="userid IN ($LAST_CHAT_AUTHORS) AND userid != $USERID" na_user 2>/dev/null >> "$OUT" || true
fi

# Группа 5: рейтинг — топ-50 юзеров по points для рендера Ranking
echo "  dump top-50 users for ranking" >&2
TOP_USERS=$(docker exec "$LEGACY_CONTAINER" mysql -u"$LEGACY_USER" -p"$LEGACY_PASS" "$LEGACY_DB" \
  -N -B -e "SELECT GROUP_CONCAT(userid) FROM (SELECT userid FROM na_user ORDER BY points DESC LIMIT 50) t" 2>/dev/null)
if [ -n "$TOP_USERS" ]; then
  echo "-- na_user (top-50 by points, кроме уже импортированных)" >> "$OUT"
  docker exec "$LEGACY_CONTAINER" mysqldump -u"$LEGACY_USER" -p"$LEGACY_PASS" "$LEGACY_DB" \
    --no-create-info --skip-extended-insert --skip-comments --hex-blob --complete-insert \
    --where="userid IN ($TOP_USERS) AND userid != $USERID" na_user 2>/dev/null >> "$OUT" || true
fi

cat >> "$OUT" <<EOF

COMMIT;
SET FOREIGN_KEY_CHECKS=1;
SET unique_checks=1;
EOF

SIZE=$(wc -c < "$OUT")
LINES=$(wc -l < "$OUT")
echo "" >&2
echo "Done: $OUT (${SIZE} bytes, ${LINES} lines)" >&2
